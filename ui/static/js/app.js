// MQTT Agent Orchestration - Main Application
// ===========================================

// Global UI object for common functionality
const UI = {
    // Show toast notification
    showToast(message, type = 'info', duration = Config.ui.toastDuration) {
        const toastContainer = document.getElementById('toast-container');
        
        // Create toast element
        const toast = document.createElement('div');
        toast.className = `toast ${type} animate-slide-in`;
        
        const icon = {
            success: 'fa-check-circle',
            error: 'fa-exclamation-circle',
            warning: 'fa-exclamation-triangle',
            info: 'fa-info-circle'
        }[type] || 'fa-info-circle';
        
        toast.innerHTML = `
            <div class="toast-icon">
                <i class="fas ${icon}"></i>
            </div>
            <div class="toast-content">
                <p>${message}</p>
            </div>
            <button class="toast-close" onclick="this.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        `;
        
        // Add to container
        toastContainer.appendChild(toast);
        
        // Auto remove after duration
        setTimeout(() => {
            toast.classList.add('animate-fade-out');
            setTimeout(() => toast.remove(), 300);
        }, duration);
        
        // Limit number of toasts
        const toasts = toastContainer.querySelectorAll('.toast');
        if (toasts.length > Config.ui.maxToasts) {
            toasts[0].remove();
        }
    },
    
    // Show loading state
    showLoading(element, show = true) {
        if (show) {
            element.classList.add('loading');
            element.setAttribute('data-original-content', element.innerHTML);
            element.innerHTML = '<div class="spinner"></div>';
        } else {
            element.classList.remove('loading');
            const originalContent = element.getAttribute('data-original-content');
            if (originalContent) {
                element.innerHTML = originalContent;
                element.removeAttribute('data-original-content');
            }
        }
    },
    
    // Format date/time
    formatDateTime(date) {
        if (!date) return '';
        const d = new Date(date);
        return d.toLocaleString();
    },
    
    // Format relative time
    formatRelativeTime(date) {
        if (!date) return '';
        const d = new Date(date);
        const now = new Date();
        const diff = now - d;
        
        const minutes = Math.floor(diff / 60000);
        const hours = Math.floor(diff / 3600000);
        const days = Math.floor(diff / 86400000);
        
        if (minutes < 1) return 'just now';
        if (minutes < 60) return `${minutes}m ago`;
        if (hours < 24) return `${hours}h ago`;
        return `${days}d ago`;
    },
    
    // Format file size
    formatFileSize(bytes) {
        if (!bytes) return '0 B';
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`;
    },
    
    // Generate unique ID
    generateId() {
        return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    },
    
    // Debounce function
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    },
    
    // Throttle function
    throttle(func, limit) {
        let inThrottle;
        return function(...args) {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }
};

// Main Application Class
class MQTTAgentOrchestrationApp {
    constructor() {
        this.currentSection = 'dashboard';
        this.charts = {};
        this.refreshIntervals = {};
        this.initialized = false;
    }
    
    // Initialize the application
    async init() {
        console.log('Initializing MQTT Agent Orchestration UI...');
        
        try {
            // Initialize navigation
            this.initNavigation();
            
            // Initialize modals
            this.initModals();
            
            // Load saved configuration
            this.loadConfiguration();
            
            // Initialize MQTT connection
            MQTT.connect();
            
            // Set up MQTT event handlers
            this.setupMQTTHandlers();
            
            // Test API connection
            const connectionTest = await API.testConnection();
            if (connectionTest.connected) {
                UI.showToast('Connected to backend API', 'success');
                document.getElementById('system-status').classList.add('connected');
            } else {
                UI.showToast('Unable to connect to backend API', 'warning');
                document.getElementById('system-status').classList.add('disconnected');
            }
            
            // Initialize section-specific modules
            this.initDashboard();
            this.initWorkflows();
            this.initWorkers();
            this.initRAG();
            this.initModels();
            this.initConfiguration();
            this.initLogs();
            
            // Start periodic updates
            this.startPeriodicUpdates();
            
            // Initialize theme
            this.initTheme();
            
            this.initialized = true;
            console.log('Application initialized successfully');
            
        } catch (error) {
            console.error('Application initialization failed:', error);
            UI.showToast('Failed to initialize application', 'error');
        }
    }
    
    // Initialize navigation
    initNavigation() {
        const navLinks = document.querySelectorAll('.nav-link');
        
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = link.getAttribute('data-section');
                this.navigateToSection(section);
            });
        });
        
        // Handle browser back/forward
        window.addEventListener('popstate', (e) => {
            const section = e.state?.section || 'dashboard';
            this.navigateToSection(section, false);
        });
    }
    
    // Navigate to a section
    navigateToSection(section, updateHistory = true) {
        if (section === this.currentSection) return;
        
        // Update active nav link
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.toggle('active', link.getAttribute('data-section') === section);
        });
        
        // Update active section
        document.querySelectorAll('.content-section').forEach(sectionEl => {
            sectionEl.classList.toggle('active', sectionEl.id === section);
        });
        
        this.currentSection = section;
        
        // Update URL
        if (updateHistory) {
            history.pushState({ section }, '', `#${section}`);
        }
        
        // Trigger section-specific updates
        this.updateSection(section);
    }
    
    // Update section-specific content
    updateSection(section) {
        switch (section) {
            case 'dashboard':
                Dashboard.refresh();
                break;
            case 'workflows':
                Workflows.refresh();
                break;
            case 'workers':
                Workers.refresh();
                break;
            case 'rag':
                RAG.refresh();
                break;
            case 'models':
                Models.refresh();
                break;
            case 'logs':
                Logs.refresh();
                break;
        }
    }
    
    // Initialize modals
    initModals() {
        // Close modal on background click
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.closeModal(modal.id);
                }
            });
        });
        
        // Close modal buttons
        document.querySelectorAll('.modal-close, [data-modal]').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const modalId = btn.getAttribute('data-modal') || btn.closest('.modal').id;
                this.closeModal(modalId);
            });
        });
    }
    
    // Open modal
    openModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('active');
        }
    }
    
    // Close modal
    closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('active');
        }
    }
    
    // Load configuration from localStorage
    loadConfiguration() {
        const savedConfig = localStorage.getItem(Config.storage.config);
        if (savedConfig) {
            try {
                const config = JSON.parse(savedConfig);
                // Apply saved configuration
                Object.assign(Config, config);
            } catch (error) {
                console.error('Failed to load saved configuration:', error);
            }
        }
    }
    
    // Save configuration to localStorage
    saveConfiguration() {
        localStorage.setItem(Config.storage.config, JSON.stringify(Config));
        UI.showToast('Configuration saved', 'success');
    }
    
    // Set up MQTT event handlers
    setupMQTTHandlers() {
        // Connection events
        MQTT.on('connected', () => {
            document.getElementById('mqtt-status').classList.add('connected');
            document.getElementById('mqtt-status').classList.remove('disconnected');
        });
        
        MQTT.on('disconnected', () => {
            document.getElementById('mqtt-status').classList.remove('connected');
            document.getElementById('mqtt-status').classList.add('disconnected');
        });
        
        // Task events
        MQTT.on('task:status', (data) => {
            Dashboard.updateTaskStatus(data);
            Workflows.updateTaskStatus(data);
        });
        
        MQTT.on('task:result', (data) => {
            Dashboard.updateTaskResult(data);
            Workflows.updateTaskResult(data);
        });
        
        MQTT.on('task:progress', (data) => {
            Dashboard.updateTaskProgress(data);
            Workflows.updateTaskProgress(data);
        });
        
        // Worker events
        MQTT.on('worker:status', (data) => {
            Dashboard.updateWorkerStatus(data);
            Workers.updateWorkerStatus(data);
        });
        
        MQTT.on('worker:health', (data) => {
            Workers.updateWorkerHealth(data);
        });
        
        MQTT.on('worker:metrics', (data) => {
            Dashboard.updateWorkerMetrics(data);
            Workers.updateWorkerMetrics(data);
        });
        
        // System events
        MQTT.on('system:status', (data) => {
            Dashboard.updateSystemStatus(data);
        });
        
        MQTT.on('system:health', (data) => {
            Dashboard.updateSystemHealth(data);
        });
        
        MQTT.on('system:metrics', (data) => {
            Dashboard.updateSystemMetrics(data);
        });
        
        MQTT.on('system:logs', (data) => {
            Logs.addLogEntry(data);
        });
        
        // RAG events
        MQTT.on('rag:indexing', (data) => {
            RAG.updateIndexingStatus(data);
        });
        
        MQTT.on('rag:searchComplete', (data) => {
            RAG.handleSearchResults(data);
        });
    }
    
    // Initialize dashboard
    initDashboard() {
        if (typeof Dashboard !== 'undefined') {
            Dashboard.init();
        }
    }
    
    // Initialize workflows
    initWorkflows() {
        if (typeof Workflows !== 'undefined') {
            Workflows.init();
        }
    }
    
    // Initialize workers
    initWorkers() {
        if (typeof Workers !== 'undefined') {
            Workers.init();
        }
    }
    
    // Initialize RAG
    initRAG() {
        if (typeof RAG !== 'undefined') {
            RAG.init();
        }
    }
    
    // Initialize models
    initModels() {
        if (typeof Models !== 'undefined') {
            Models.init();
        }
    }
    
    // Initialize configuration
    initConfiguration() {
        // Configuration tab switching
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const tabId = btn.getAttribute('data-tab');
                
                // Update active tab button
                document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                
                // Update active tab content
                document.querySelectorAll('.config-tab').forEach(tab => {
                    tab.classList.toggle('active', tab.id === tabId);
                });
            });
        });
        
        // Save configuration button
        document.getElementById('save-config').addEventListener('click', () => {
            this.saveSystemConfiguration();
        });
    }
    
    // Initialize logs
    initLogs() {
        if (typeof Logs !== 'undefined') {
            Logs.init();
        }
    }
    
    // Start periodic updates
    startPeriodicUpdates() {
        // Dashboard metrics update
        this.refreshIntervals.dashboard = setInterval(() => {
            if (this.currentSection === 'dashboard') {
                Dashboard.updateMetrics();
            }
        }, Config.ui.refreshInterval);
        
        // System health check
        this.refreshIntervals.health = setInterval(() => {
            this.checkSystemHealth();
        }, 30000); // Every 30 seconds
    }
    
    // Check system health
    async checkSystemHealth() {
        try {
            const health = await API.getSystemHealth();
            const isHealthy = health.status === 'healthy';
            
            document.getElementById('system-status').classList.toggle('connected', isHealthy);
            document.getElementById('system-status').classList.toggle('disconnected', !isHealthy);
            
        } catch (error) {
            console.error('Health check failed:', error);
            document.getElementById('system-status').classList.add('disconnected');
        }
    }
    
    // Save system configuration
    async saveSystemConfiguration() {
        const config = {
            mqtt: {
                host: document.getElementById('mqtt-host').value,
                port: parseInt(document.getElementById('mqtt-port').value),
                username: document.getElementById('mqtt-username').value,
                password: document.getElementById('mqtt-password').value,
                clientPrefix: document.getElementById('mqtt-client-prefix').value
            },
            rag: {
                qdrantHost: document.getElementById('qdrant-host').value,
                qdrantPort: parseInt(document.getElementById('qdrant-port').value),
                embeddingModel: document.getElementById('embedding-model').value,
                collectionName: document.getElementById('collection-name').value
            },
            system: {
                logLevel: document.getElementById('log-level').value,
                maxConcurrent: parseInt(document.getElementById('max-concurrent').value),
                taskTimeout: parseInt(document.getElementById('task-timeout').value),
                healthInterval: parseInt(document.getElementById('health-interval').value)
            }
        };
        
        try {
            await API.saveConfig(config);
            this.saveConfiguration();
            UI.showToast('Configuration saved successfully', 'success');
            
            // Reconnect MQTT if settings changed
            if (config.mqtt.host !== Config.mqtt.broker.hostname || 
                config.mqtt.port !== Config.mqtt.broker.port) {
                MQTT.reconnect();
            }
        } catch (error) {
            console.error('Failed to save configuration:', error);
            UI.showToast('Failed to save configuration', 'error');
        }
    }
    
    // Initialize theme
    initTheme() {
        const savedTheme = localStorage.getItem(Config.storage.theme) || 'light';
        document.documentElement.setAttribute('data-theme', savedTheme);
        
        // Add theme toggle if needed
        // This can be extended to add a theme toggle button in the UI
    }
    
    // Clean up resources
    destroy() {
        // Clear intervals
        Object.values(this.refreshIntervals).forEach(interval => clearInterval(interval));
        
        // Disconnect MQTT
        MQTT.disconnect();
        
        // Disconnect WebSocket if connected
        API.disconnectWebSocket();
        
        console.log('Application destroyed');
    }
}

// Initialize application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.App = new MQTTAgentOrchestrationApp();
    window.App.init();
});

// Handle page unload
window.addEventListener('beforeunload', () => {
    if (window.App) {
        window.App.destroy();
    }
});

// Export UI utilities
window.UI = UI;