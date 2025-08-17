// MQTT Agent Orchestration - Configuration
// ========================================

const Config = {
    // API Configuration
    api: {
        baseUrl: window.location.hostname === 'localhost' ? 'http://localhost:8080' : '',
        endpoints: {
            // Workflow Management
            workflows: '/api/workflows',
            workflowById: (id) => `/api/workflows/${id}`,
            workflowTasks: (id) => `/api/workflows/${id}/tasks`,
            createWorkflow: '/api/workflows',
            updateWorkflow: (id) => `/api/workflows/${id}`,
            deleteWorkflow: (id) => `/api/workflows/${id}`,
            
            // Worker Management
            workers: '/api/workers',
            workerById: (id) => `/api/workers/${id}`,
            workerHealth: '/api/workers/health',
            workerMetrics: '/api/workers/metrics',
            scaleWorkers: '/api/workers/scale',
            
            // Task Management
            tasks: '/api/tasks',
            taskById: (id) => `/api/tasks/${id}`,
            submitTask: '/api/tasks/submit',
            taskStatus: (id) => `/api/tasks/${id}/status`,
            taskResult: (id) => `/api/tasks/${id}/result`,
            
            // RAG Service
            ragSearch: '/api/rag/search',
            ragCollections: '/api/rag/collections',
            ragUpload: '/api/rag/upload',
            ragTrain: '/api/rag/train',
            ragDocuments: '/api/rag/documents',
            ragEmbeddings: '/api/rag/embeddings',
            
            // Model Management
            models: '/api/models',
            modelById: (id) => `/api/models/${id}`,
            downloadModel: '/api/models/download',
            modelPerformance: '/api/models/performance',
            
            // System Configuration
            config: '/api/config',
            configBySection: (section) => `/api/config/${section}`,
            saveConfig: '/api/config/save',
            
            // System Status
            systemStatus: '/api/system/status',
            systemHealth: '/api/system/health',
            systemMetrics: '/api/system/metrics',
            systemLogs: '/api/system/logs',
            
            // AI Services
            aiServices: '/api/ai/services',
            aiServiceStatus: (service) => `/api/ai/services/${service}/status`,
            aiServiceConfig: (service) => `/api/ai/services/${service}/config`
        }
    },
    
    // MQTT Configuration
    mqtt: {
        broker: {
            hostname: window.location.hostname || 'localhost',
            port: 9001, // WebSocket port
            path: '/mqtt',
            clientId: `mqtt-agent-ui-${Math.random().toString(16).substr(2, 8)}`,
            reconnectPeriod: 5000,
            keepalive: 60
        },
        topics: {
            // Task Topics
            taskSubmit: 'agent/task/submit',
            taskStatus: 'agent/task/status/+',
            taskResult: 'agent/task/result/+',
            taskProgress: 'agent/task/progress/+',
            
            // Worker Topics
            workerStatus: 'agent/worker/status/+',
            workerHealth: 'agent/worker/health/+',
            workerMetrics: 'agent/worker/metrics/+',
            workerRegister: 'agent/worker/register',
            workerUnregister: 'agent/worker/unregister',
            
            // Workflow Topics
            workflowStatus: 'agent/workflow/status/+',
            workflowProgress: 'agent/workflow/progress/+',
            workflowComplete: 'agent/workflow/complete/+',
            
            // System Topics
            systemStatus: 'agent/system/status',
            systemHealth: 'agent/system/health',
            systemMetrics: 'agent/system/metrics',
            systemLogs: 'agent/system/logs',
            
            // RAG Topics
            ragIndexing: 'agent/rag/indexing/+',
            ragSearchComplete: 'agent/rag/search/complete/+'
        }
    },
    
    // UI Configuration
    ui: {
        refreshInterval: 5000, // 5 seconds
        chartUpdateInterval: 1000, // 1 second
        logBufferSize: 1000,
        maxToasts: 5,
        toastDuration: 5000,
        animationDuration: 300,
        
        // Chart Configuration
        charts: {
            maxDataPoints: 50,
            colors: {
                primary: '#4f46e5',
                secondary: '#06b6d4',
                success: '#10b981',
                warning: '#f59e0b',
                danger: '#ef4444',
                info: '#3b82f6'
            }
        },
        
        // Worker Role Icons
        roleIcons: {
            developer: 'fa-code',
            reviewer: 'fa-search',
            approver: 'fa-check-circle',
            tester: 'fa-vial',
            default: 'fa-user-cog'
        },
        
        // Status Colors
        statusColors: {
            active: 'success',
            idle: 'warning',
            error: 'danger',
            pending: 'info',
            completed: 'success',
            failed: 'danger'
        }
    },
    
    // Local Storage Keys
    storage: {
        theme: 'mqttAgentTheme',
        config: 'mqttAgentConfig',
        recentTasks: 'mqttAgentRecentTasks',
        userPreferences: 'mqttAgentPreferences'
    },
    
    // Default Values
    defaults: {
        mqttHost: 'localhost',
        mqttPort: 1883,
        qdrantHost: 'localhost',
        qdrantPort: 6333,
        logLevel: 'info',
        maxConcurrentTasks: 10,
        taskTimeout: 300,
        healthCheckInterval: 30
    },
    
    // AI Service Providers
    aiProviders: [
        { id: 'cerebras', name: 'Cerebras', icon: 'fa-brain', color: '#ff6b6b' },
        { id: 'nvidia', name: 'NVIDIA', icon: 'fa-microchip', color: '#76b900' },
        { id: 'gemini', name: 'Gemini', icon: 'fa-gem', color: '#4285f4' },
        { id: 'grok', name: 'Grok', icon: 'fa-robot', color: '#1da1f2' },
        { id: 'groq', name: 'Groq', icon: 'fa-bolt', color: '#ff9800' },
        { id: 'local', name: 'Local Models', icon: 'fa-hdd', color: '#4f46e5' }
    ],
    
    // Worker Roles
    workerRoles: [
        { id: 'developer', name: 'Developer', description: 'Code generation and implementation' },
        { id: 'reviewer', name: 'Reviewer', description: 'Code review and quality assessment' },
        { id: 'approver', name: 'Approver', description: 'Final validation and approval' },
        { id: 'tester', name: 'Tester', description: 'Testing strategy and execution' }
    ],
    
    // Task Priorities
    taskPriorities: [
        { id: 'low', name: 'Low', color: 'info' },
        { id: 'medium', name: 'Medium', color: 'warning' },
        { id: 'high', name: 'High', color: 'danger' },
        { id: 'critical', name: 'Critical', color: 'danger' }
    ]
};

// Export for use in other modules
window.Config = Config;