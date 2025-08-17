# MQTT Agent Orchestration - Web UI

A comprehensive web-based dashboard for managing and monitoring the MQTT Agent Orchestration system.

## Features

### 🎯 Complete System Dashboard
- **Real-time Metrics**: Active workflows, worker status, RAG documents, success rates
- **Activity Charts**: Visual representation of system activity using Chart.js
- **Recent Tasks**: Quick overview of recent task executions
- **System Health**: Real-time connection status for MQTT and backend services

### 🔄 Workflow Management
- **Create Workflows**: Submit new workflows with role assignment and priority
- **Monitor Progress**: Real-time workflow status updates via MQTT
- **Workflow Visualization**: Visual representation of workflow stages
- **Filter & Search**: Find workflows by status, role, or search terms

### 👥 Worker Management
- **Worker Overview**: Monitor active workers by role (Developer, Reviewer, Tester)
- **Health Monitoring**: Real-time worker health status
- **Performance Metrics**: Throughput and latency charts
- **Scale Workers**: Dynamically adjust worker count

### 🧠 RAG Service Interface
- **Semantic Search**: Search knowledge base with natural language queries
- **Document Upload**: Add new documents to the RAG system
- **Collection Management**: View and manage vector collections
- **Embedding Visualization**: Visual representation of document embeddings

### 🤖 Model Management
- **Local Models**: Manage GGUF models for local inference
- **External AI Services**: Configure Cerebras, NVIDIA, Gemini, Grok, and Groq
- **Performance Comparison**: Compare model performance metrics
- **Download Models**: Easy model download interface

### ⚙️ System Configuration
- **MQTT Settings**: Configure broker connection
- **AI Service Config**: Set up API keys and endpoints
- **RAG Settings**: Configure Qdrant and embedding models
- **Worker Settings**: Adjust worker parameters
- **System Settings**: Log levels, concurrency, timeouts

### 📊 Real-time Updates
- **MQTT Integration**: Live updates via MQTT WebSocket
- **Auto-refresh**: Periodic data updates
- **Toast Notifications**: User-friendly status notifications
- **Connection Status**: Visual indicators for system connectivity

## Technology Stack

- **Frontend**: Vanilla JavaScript, HTML5, CSS3
- **Charts**: Chart.js for data visualization
- **Real-time**: Paho MQTT.js for WebSocket MQTT
- **Styling**: Modern CSS with animations and responsive design
- **Icons**: Font Awesome 6

## Project Structure

```
ui/
├── index.html              # Main dashboard HTML
├── static/
│   ├── css/
│   │   ├── styles.css     # Main stylesheet
│   │   └── animations.css # Animation definitions
│   └── js/
│       ├── config.js      # Configuration and constants
│       ├── api.js         # REST API client
│       ├── mqtt-client.js # MQTT WebSocket client
│       ├── app.js         # Main application controller
│       ├── dashboard.js   # Dashboard module
│       ├── workflows.js   # Workflow management
│       ├── workers.js     # Worker monitoring
│       ├── rag.js         # RAG service interface
│       ├── models.js      # Model management
│       └── logs.js        # Log viewer
└── test-server.py         # Development server
```

## Quick Start

1. **Start the test server** (for development):
   ```bash
   cd ui
   python3 test-server.py
   ```

2. **Open in browser**:
   ```
   http://localhost:8000
   ```

3. **Configure MQTT** (optional):
   - Navigate to Configuration section
   - Update MQTT broker settings
   - Save configuration

## API Integration

The UI expects the backend API to be available at:
- Development: `http://localhost:8080`
- Production: Same origin as UI

### Required API Endpoints

- `/api/workflows` - Workflow management
- `/api/workers` - Worker management
- `/api/tasks` - Task operations
- `/api/rag/*` - RAG service endpoints
- `/api/models` - Model management
- `/api/config` - Configuration
- `/api/system/*` - System status and metrics

## MQTT Topics

The UI subscribes to the following MQTT topics:

- `agent/task/status/+` - Task status updates
- `agent/worker/status/+` - Worker status changes
- `agent/workflow/status/+` - Workflow progress
- `agent/system/status` - System health
- `agent/rag/indexing/+` - RAG indexing status

## Configuration

Configuration is stored in localStorage and includes:
- MQTT broker settings
- API endpoints
- UI preferences
- Theme settings

## Development

### Adding New Features

1. Create a new module in `static/js/`
2. Initialize in `app.js`
3. Add navigation item in `index.html`
4. Create corresponding section in HTML
5. Style with CSS variables for consistency

### Code Style

- Use ES6+ features
- Follow modular pattern
- Use Config object for constants
- Implement error handling
- Add loading states
- Include real-time update handlers

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Requires WebSocket support

## Security Considerations

- CORS headers required for API
- MQTT over WebSocket (ws:// or wss://)
- No hardcoded credentials
- Configuration in localStorage

## Future Enhancements

- [ ] Dark mode toggle
- [ ] Advanced workflow designer
- [ ] Metrics export functionality
- [ ] User authentication
- [ ] Role-based access control
- [ ] Mobile responsive improvements
- [ ] Offline capability
- [ ] Internationalization
