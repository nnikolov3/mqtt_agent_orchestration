// MQTT Agent Orchestration - MQTT Client
// ======================================

class MQTTClient {
    constructor() {
        this.client = null;
        this.connected = false;
        this.subscriptions = new Map();
        this.eventHandlers = new Map();
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 10;
    }

    // Initialize MQTT connection
    connect() {
        const { hostname, port, path, clientId, reconnectPeriod, keepalive } = Config.mqtt.broker;
        
        console.log(`Connecting to MQTT broker at ${hostname}:${port}${path}`);
        
        // Create Paho MQTT client
        this.client = new Paho.MQTT.Client(hostname, port, path, clientId);
        
        // Set callback handlers
        this.client.onConnectionLost = this.onConnectionLost.bind(this);
        this.client.onMessageArrived = this.onMessageArrived.bind(this);
        
        // Connection options
        const connectOptions = {
            onSuccess: this.onConnect.bind(this),
            onFailure: this.onConnectFailure.bind(this),
            keepAliveInterval: keepalive,
            reconnect: true,
            reconnectInterval: reconnectPeriod / 1000, // Convert to seconds
            useSSL: window.location.protocol === 'https:',
            timeout: 10
        };
        
        // Add credentials if configured
        const savedConfig = this.getSavedConfig();
        if (savedConfig.mqttUsername) {
            connectOptions.userName = savedConfig.mqttUsername;
        }
        if (savedConfig.mqttPassword) {
            connectOptions.password = savedConfig.mqttPassword;
        }
        
        // Connect to broker
        this.client.connect(connectOptions);
    }

    // Connection successful
    onConnect() {
        console.log('MQTT connected successfully');
        this.connected = true;
        this.reconnectAttempts = 0;
        
        // Update UI status
        this.updateConnectionStatus(true);
        
        // Subscribe to all topics
        this.subscribeToTopics();
        
        // Emit connection event
        this.emit('connected');
        
        // Show success notification
        UI.showToast('Connected to MQTT broker', 'success');
    }

    // Connection failed
    onConnectFailure(error) {
        console.error('MQTT connection failed:', error);
        this.connected = false;
        this.reconnectAttempts++;
        
        // Update UI status
        this.updateConnectionStatus(false);
        
        // Show error notification
        UI.showToast(`MQTT connection failed: ${error.errorMessage}`, 'error');
        
        // Emit disconnection event
        this.emit('disconnected', error);
        
        // Handle max reconnect attempts
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Max reconnection attempts reached');
            UI.showToast('Unable to connect to MQTT broker. Please check your settings.', 'error');
        }
    }

    // Connection lost
    onConnectionLost(response) {
        if (response.errorCode !== 0) {
            console.error('MQTT connection lost:', response.errorMessage);
            this.connected = false;
            
            // Update UI status
            this.updateConnectionStatus(false);
            
            // Show warning notification
            UI.showToast('MQTT connection lost. Attempting to reconnect...', 'warning');
            
            // Emit disconnection event
            this.emit('disconnected', response);
        }
    }

    // Message received
    onMessageArrived(message) {
        const topic = message.destinationName;
        const payload = message.payloadString;
        
        console.log(`MQTT message received on ${topic}:`, payload);
        
        try {
            const data = JSON.parse(payload);
            
            // Process message based on topic pattern
            this.processMessage(topic, data);
            
            // Call specific handlers
            const handlers = this.subscriptions.get(topic) || [];
            handlers.forEach(handler => {
                try {
                    handler(data, topic);
                } catch (error) {
                    console.error(`Error in message handler for ${topic}:`, error);
                }
            });
            
        } catch (error) {
            console.error('Error parsing MQTT message:', error);
        }
    }

    // Process messages by topic pattern
    processMessage(topic, data) {
        // Task status updates
        if (topic.startsWith('agent/task/status/')) {
            this.emit('task:status', data);
        }
        // Task results
        else if (topic.startsWith('agent/task/result/')) {
            this.emit('task:result', data);
        }
        // Task progress
        else if (topic.startsWith('agent/task/progress/')) {
            this.emit('task:progress', data);
        }
        // Worker status
        else if (topic.startsWith('agent/worker/status/')) {
            this.emit('worker:status', data);
        }
        // Worker health
        else if (topic.startsWith('agent/worker/health/')) {
            this.emit('worker:health', data);
        }
        // Worker metrics
        else if (topic.startsWith('agent/worker/metrics/')) {
            this.emit('worker:metrics', data);
        }
        // Workflow status
        else if (topic.startsWith('agent/workflow/status/')) {
            this.emit('workflow:status', data);
        }
        // Workflow progress
        else if (topic.startsWith('agent/workflow/progress/')) {
            this.emit('workflow:progress', data);
        }
        // System status
        else if (topic === 'agent/system/status') {
            this.emit('system:status', data);
        }
        // System health
        else if (topic === 'agent/system/health') {
            this.emit('system:health', data);
        }
        // System metrics
        else if (topic === 'agent/system/metrics') {
            this.emit('system:metrics', data);
        }
        // System logs
        else if (topic === 'agent/system/logs') {
            this.emit('system:logs', data);
        }
        // RAG indexing
        else if (topic.startsWith('agent/rag/indexing/')) {
            this.emit('rag:indexing', data);
        }
        // RAG search complete
        else if (topic.startsWith('agent/rag/search/complete/')) {
            this.emit('rag:searchComplete', data);
        }
    }

    // Subscribe to topics
    subscribeToTopics() {
        const topics = Object.values(Config.mqtt.topics);
        
        topics.forEach(topic => {
            this.subscribe(topic);
        });
    }

    // Subscribe to a topic
    subscribe(topic, handler) {
        if (!this.connected) {
            console.warn('Cannot subscribe - not connected to MQTT broker');
            return;
        }
        
        // Subscribe to topic if not already subscribed
        if (!this.subscriptions.has(topic)) {
            this.client.subscribe(topic, {
                onSuccess: () => {
                    console.log(`Subscribed to topic: ${topic}`);
                },
                onFailure: (error) => {
                    console.error(`Failed to subscribe to ${topic}:`, error);
                }
            });
            this.subscriptions.set(topic, []);
        }
        
        // Add handler if provided
        if (handler) {
            const handlers = this.subscriptions.get(topic);
            handlers.push(handler);
        }
    }

    // Unsubscribe from a topic
    unsubscribe(topic) {
        if (!this.connected) {
            return;
        }
        
        this.client.unsubscribe(topic, {
            onSuccess: () => {
                console.log(`Unsubscribed from topic: ${topic}`);
                this.subscriptions.delete(topic);
            },
            onFailure: (error) => {
                console.error(`Failed to unsubscribe from ${topic}:`, error);
            }
        });
    }

    // Publish a message
    publish(topic, payload, qos = 0, retained = false) {
        if (!this.connected) {
            console.warn('Cannot publish - not connected to MQTT broker');
            return Promise.reject(new Error('Not connected to MQTT broker'));
        }
        
        const message = new Paho.MQTT.Message(JSON.stringify(payload));
        message.destinationName = topic;
        message.qos = qos;
        message.retained = retained;
        
        try {
            this.client.send(message);
            console.log(`Published to ${topic}:`, payload);
            return Promise.resolve();
        } catch (error) {
            console.error(`Failed to publish to ${topic}:`, error);
            return Promise.reject(error);
        }
    }

    // Submit a task
    submitTask(task) {
        const taskId = `task-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        const payload = {
            id: taskId,
            ...task,
            timestamp: new Date().toISOString()
        };
        
        return this.publish(Config.mqtt.topics.taskSubmit, payload, 1)
            .then(() => taskId);
    }

    // Event emitter functionality
    on(event, handler) {
        if (!this.eventHandlers.has(event)) {
            this.eventHandlers.set(event, []);
        }
        this.eventHandlers.get(event).push(handler);
    }

    off(event, handler) {
        if (!this.eventHandlers.has(event)) {
            return;
        }
        
        const handlers = this.eventHandlers.get(event);
        const index = handlers.indexOf(handler);
        if (index > -1) {
            handlers.splice(index, 1);
        }
    }

    emit(event, data) {
        if (!this.eventHandlers.has(event)) {
            return;
        }
        
        const handlers = this.eventHandlers.get(event);
        handlers.forEach(handler => {
            try {
                handler(data);
            } catch (error) {
                console.error(`Error in event handler for ${event}:`, error);
            }
        });
    }

    // Update connection status in UI
    updateConnectionStatus(connected) {
        const statusElement = document.getElementById('mqtt-status');
        if (statusElement) {
            statusElement.classList.toggle('connected', connected);
            statusElement.classList.toggle('disconnected', !connected);
        }
    }

    // Get saved configuration
    getSavedConfig() {
        const saved = localStorage.getItem(Config.storage.config);
        return saved ? JSON.parse(saved) : {};
    }

    // Disconnect from broker
    disconnect() {
        if (this.client && this.connected) {
            this.client.disconnect();
            this.connected = false;
            this.updateConnectionStatus(false);
            console.log('Disconnected from MQTT broker');
        }
    }

    // Reconnect to broker
    reconnect() {
        this.disconnect();
        setTimeout(() => {
            this.connect();
        }, 1000);
    }

    // Check connection status
    isConnected() {
        return this.connected;
    }

    // Get connection info
    getConnectionInfo() {
        return {
            connected: this.connected,
            clientId: this.client ? this.client.clientId : null,
            reconnectAttempts: this.reconnectAttempts,
            subscriptions: Array.from(this.subscriptions.keys())
        };
    }
}

// Create singleton instance
const MQTT = new MQTTClient();

// Export for use in other modules
window.MQTT = MQTT;