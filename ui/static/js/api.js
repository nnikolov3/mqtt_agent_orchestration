// MQTT Agent Orchestration - API Client
// =====================================

class APIClient {
    constructor() {
        this.baseUrl = Config.api.baseUrl;
        this.headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
    }

    // Helper method for API calls
    async request(endpoint, options = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        const config = {
            ...options,
            headers: {
                ...this.headers,
                ...options.headers
            }
        };

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                const error = await response.json().catch(() => ({ message: response.statusText }));
                throw new Error(error.message || `HTTP error! status: ${response.status}`);
            }

            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }
            
            return await response.text();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    // GET request
    async get(endpoint, params = {}) {
        const queryString = new URLSearchParams(params).toString();
        const fullEndpoint = queryString ? `${endpoint}?${queryString}` : endpoint;
        return this.request(fullEndpoint, { method: 'GET' });
    }

    // POST request
    async post(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    // PUT request
    async put(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    // DELETE request
    async delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    }

    // Workflow Management
    async getWorkflows(filters = {}) {
        return this.get(Config.api.endpoints.workflows, filters);
    }

    async getWorkflow(id) {
        return this.get(Config.api.endpoints.workflowById(id));
    }

    async createWorkflow(workflow) {
        return this.post(Config.api.endpoints.createWorkflow, workflow);
    }

    async updateWorkflow(id, updates) {
        return this.put(Config.api.endpoints.updateWorkflow(id), updates);
    }

    async deleteWorkflow(id) {
        return this.delete(Config.api.endpoints.deleteWorkflow(id));
    }

    async getWorkflowTasks(id) {
        return this.get(Config.api.endpoints.workflowTasks(id));
    }

    // Worker Management
    async getWorkers() {
        return this.get(Config.api.endpoints.workers);
    }

    async getWorker(id) {
        return this.get(Config.api.endpoints.workerById(id));
    }

    async getWorkerHealth() {
        return this.get(Config.api.endpoints.workerHealth);
    }

    async getWorkerMetrics() {
        return this.get(Config.api.endpoints.workerMetrics);
    }

    async scaleWorkers(role, count) {
        return this.post(Config.api.endpoints.scaleWorkers, { role, count });
    }

    // Task Management
    async getTasks(filters = {}) {
        return this.get(Config.api.endpoints.tasks, filters);
    }

    async getTask(id) {
        return this.get(Config.api.endpoints.taskById(id));
    }

    async submitTask(task) {
        return this.post(Config.api.endpoints.submitTask, task);
    }

    async getTaskStatus(id) {
        return this.get(Config.api.endpoints.taskStatus(id));
    }

    async getTaskResult(id) {
        return this.get(Config.api.endpoints.taskResult(id));
    }

    // RAG Service
    async searchRAG(query, options = {}) {
        return this.post(Config.api.endpoints.ragSearch, { query, ...options });
    }

    async getRAGCollections() {
        return this.get(Config.api.endpoints.ragCollections);
    }

    async uploadDocument(document) {
        const formData = new FormData();
        if (document.file) {
            formData.append('file', document.file);
        } else {
            formData.append('content', document.content);
        }
        formData.append('title', document.title);
        formData.append('collection', document.collection);
        formData.append('type', document.type || 'text');

        return this.request(Config.api.endpoints.ragUpload, {
            method: 'POST',
            body: formData,
            headers: {} // Let browser set content-type for FormData
        });
    }

    async trainEmbeddings(collection) {
        return this.post(Config.api.endpoints.ragTrain, { collection });
    }

    async getRAGDocuments(collection) {
        return this.get(Config.api.endpoints.ragDocuments, { collection });
    }

    async getEmbeddings(collection) {
        return this.get(Config.api.endpoints.ragEmbeddings, { collection });
    }

    // Model Management
    async getModels() {
        return this.get(Config.api.endpoints.models);
    }

    async getModel(id) {
        return this.get(Config.api.endpoints.modelById(id));
    }

    async downloadModel(modelId) {
        return this.post(Config.api.endpoints.downloadModel, { modelId });
    }

    async getModelPerformance() {
        return this.get(Config.api.endpoints.modelPerformance);
    }

    // System Configuration
    async getConfig() {
        return this.get(Config.api.endpoints.config);
    }

    async getConfigSection(section) {
        return this.get(Config.api.endpoints.configBySection(section));
    }

    async saveConfig(config) {
        return this.post(Config.api.endpoints.saveConfig, config);
    }

    // System Status
    async getSystemStatus() {
        return this.get(Config.api.endpoints.systemStatus);
    }

    async getSystemHealth() {
        return this.get(Config.api.endpoints.systemHealth);
    }

    async getSystemMetrics() {
        return this.get(Config.api.endpoints.systemMetrics);
    }

    async getSystemLogs(filters = {}) {
        return this.get(Config.api.endpoints.systemLogs, filters);
    }

    // AI Services
    async getAIServices() {
        return this.get(Config.api.endpoints.aiServices);
    }

    async getAIServiceStatus(service) {
        return this.get(Config.api.endpoints.aiServiceStatus(service));
    }

    async getAIServiceConfig(service) {
        return this.get(Config.api.endpoints.aiServiceConfig(service));
    }

    async updateAIServiceConfig(service, config) {
        return this.put(Config.api.endpoints.aiServiceConfig(service), config);
    }

    // Utility Methods
    async testConnection() {
        try {
            const status = await this.getSystemStatus();
            return { connected: true, status };
        } catch (error) {
            return { connected: false, error: error.message };
        }
    }

    async exportData(type, format = 'json') {
        const endpoint = `/api/export/${type}`;
        const response = await this.get(endpoint, { format });
        
        if (format === 'json') {
            return response;
        }
        
        // For CSV or other formats, trigger download
        const blob = new Blob([response], { type: 'text/csv' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${type}-export-${new Date().toISOString()}.${format}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
    }

    // WebSocket for real-time updates (if needed beyond MQTT)
    connectWebSocket(onMessage) {
        const wsUrl = this.baseUrl.replace('http', 'ws') + '/ws';
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            console.log('WebSocket connected');
        };
        
        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                onMessage(data);
            } catch (error) {
                console.error('WebSocket message parse error:', error);
            }
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
        
        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            // Attempt to reconnect after 5 seconds
            setTimeout(() => this.connectWebSocket(onMessage), 5000);
        };
    }

    disconnectWebSocket() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}

// Create singleton instance
const API = new APIClient();

// Export for use in other modules
window.API = API;