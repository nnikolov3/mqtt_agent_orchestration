This comprehensive `README.md` provides a detailed overview of the MQTT Agent Orchestration System, designed for intelligent automation and distributed system management.

**🚀 READY FOR PRODUCTION** - The system has been fully tested and validated with autonomous workflows running successfully.

---

# Intelligent MQTT Agent Orchestration System with RAG Capabilities

![Python Version](https://img.shields.io/badge/python-3.9+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)
![Code Style](https://img.shields.io/badge/code%20style-black-000000.svg)

## 1. Project Overview and Purpose

The **Intelligent MQTT Agent Orchestration System** is a robust and scalable framework designed to manage and coordinate a network of autonomous, role-based agents communicating via MQTT. Its primary purpose is to enable intelligent automation, dynamic task execution, and real-time decision-making in distributed environments, particularly suited for IoT, edge computing, and complex industrial control systems.

At its core, the system leverages a central **Orchestrator** to define, execute, and monitor complex workflows, dispatching tasks to specialized **Agents**. A key differentiator is the integrated **Retrieval-Augmented Generation (RAG)** capability, which empowers AI-driven agents to access and utilize a dynamic knowledge base, significantly enhancing their contextual understanding, reducing hallucinations, and enabling more informed decision-time decisions.

**Key Problems Solved:**
*   **Complexity of Distributed Systems:** Simplifies the management and interaction of numerous independent components.
*   **Lack of Intelligence at the Edge:** Infuses AI capabilities, including contextual awareness, directly into operational workflows.
*   **Rigid Automation:** Moves beyond static rules to dynamic, adaptive, and intelligent automation.
*   **Scalability Challenges:** Provides a modular and asynchronous architecture that scales horizontally.
*   **Data Silos:** Facilitates seamless data flow and action coordination across disparate systems.

## 2. Architecture Diagram (Text-Based)

The system's architecture is built around a central MQTT broker, which acts as the communication backbone. The Orchestrator manages workflows and tasks, while various specialized agents perform specific functions. The RAG component integrates a knowledge base and an LLM for intelligent decision-making.

```
+-----------------+     MQTT Broker     +-----------------+
|   External      | <-----------------> |   Orchestrator  |
|   System/User   | (Pub/Sub Topics)    | (Workflow Engine,|
| (e.g., API, UI) |                     |  Task Scheduler, |
+-----------------+                     |  Agent Registry) |
        ^                               +--------^--------+
        |                                        |
        |                                        | MQTT
        |                                        |
+-------v-------+                                |
|  Knowledge    |                                |
|  Base (Vector |                                |
|  DB for RAG)  |                                |
+-------^-------+                                |
        |                                        |
+-------v-------+                                |
|   LLM Provider|                                |
| (OpenAI, Llama)|                                |
+---------------+                                |
                                                 |
+------------------------------------------------v---------------------------------+
|                                      MQTT Topics                                 |
| (e.g., /tasks/agent_id, /data/sensor_id, /status/agent_id, /commands/device_id) |
+---------------------------------------------------------------------------------+
        |                                   |                                   |
        v                                   v                                   v
+-----------------+               +-----------------+               +-----------------+
|   Sensor Agent  |               |  Actuator Agent |               |    AI/RAG Agent |
| (Data Ingestion,|               | (Action Execution,|             | (Decision Making,|
|  Publishes Data)|               |  Subscribes to   |             |  Contextual AI, |
+-----------------+               |  Command Topics) |             |  Subscribes to  |
                                  +-----------------+             |  Data/Task Topics)|
                                                                  +-----------------+
                                                                          |
                                                                          v
                                                                  +-----------------+
                                                                  | Data Processing |
                                                                  |      Agent      |
                                                                  | (Data Transform,|
                                                                  |  Analytics)     |
                                                                  +-----------------+

```

**Component Breakdown:**

*   **MQTT Broker:** The central message broker (e.g., Mosquitto) facilitating asynchronous, decoupled communication between all components.
*   **Orchestrator:** The brain of the system. It manages the lifecycle of workflows, dispatches tasks to appropriate agents, monitors their status, and handles error recovery. It maintains a registry of available agents and their capabilities.
*   **Agents:** Independent, specialized micro-services that perform specific functions. They subscribe to task topics and publish results/data.
    *   **Sensor Agent:** Collects data from physical or virtual sensors and publishes it to designated MQTT topics.
    *   **Actuator Agent:** Receives commands from the Orchestrator (or other agents) and executes actions on physical or virtual devices.
    *   **AI/RAG Agent:** The intelligent core. It receives complex queries or data, retrieves relevant context from the Knowledge Base, augments the query, and uses an LLM to generate informed decisions or responses. It can then publish these decisions as tasks for other agents.
    *   **Data Processing Agent:** Performs transformations, aggregations, or analytics on raw data received from sensor agents before it's used by other components or stored.
*   **Knowledge Base (Vector DB):** A specialized database (e.g., ChromaDB, Pinecone, Weaviate) storing vectorized embeddings of domain-specific documents, manuals, historical data, or operational procedures. Used by the RAG component for contextual retrieval.
*   **LLM Provider:** An interface to a Large Language Model (e.g., OpenAI GPT, Llama 2, Mistral) that performs the generative part of RAG, synthesizing information based on the retrieved context.
*   **External System/User:** Interacts with the Orchestrator to initiate workflows, query system status, or inject commands.

## 3. Features

*   **Autonomous Workflow Engine:**
    *   Define complex, multi-step workflows using a state-machine or directed acyclic graph (DAG) approach.
    *   Automated task sequencing, conditional branching, and parallel execution.
    *   Workflow state persistence and recovery mechanisms.
    *   Event-driven triggers for workflow initiation (e.g., new sensor data, external API call).
*   **Role-Based Worker Agents:**
    *   Modular and extensible agent architecture, allowing easy creation of new agent types.
    *   Agents are specialized (e.g., Sensor, Actuator, AI, Data Processor, Notification).
    *   Agents register their capabilities with the Orchestrator, enabling dynamic task assignment.
    *   Decoupled communication via MQTT ensures high availability and fault tolerance.
*   **Retrieval-Augmented Generation (RAG) Integration:**
    *   **Contextual Intelligence:** AI agents can query a domain-specific knowledge base (vector database) to retrieve relevant information before generating responses or actions.
    *   **Reduced Hallucinations:** Grounding LLM responses in factual, retrieved data minimizes fabricated outputs.
    *   **Dynamic Knowledge:** The knowledge base can be updated in real-time, allowing the system to adapt to new information without retraining the LLM.
    *   **Use Cases:** Intelligent troubleshooting, dynamic policy enforcement, contextual anomaly detection, smart recommendations.
*   **MQTT-Centric Communication:**
    *   Lightweight, asynchronous, and publish-subscribe messaging paradigm.
    *   Enables real-time data streaming and command dispatch.
    *   Supports QoS levels for reliable message delivery.
    *   Scalable and robust for distributed environments.
*   **Dynamic Task Assignment & Load Balancing:**
    *   The Orchestrator intelligently assigns tasks to available agents based on their capabilities and current load.
    *   Supports multiple instances of the same agent type for horizontal scaling and redundancy.
*   **Observability & Monitoring:**
    *   Comprehensive logging for all components (Orchestrator, Agents, RAG).
    *   Potential for integration with monitoring tools (e.g., Prometheus, Grafana) via custom metrics.
    *   Real-time status updates of agents and workflows.
*   **Configuration-Driven:**
    *   Flexible configuration management for MQTT topics, agent roles, LLM parameters, and knowledge base settings.
    *   Environment variable support for sensitive information (API keys).

## 4. Quick Start Guide

This guide will help you get the basic system up and running.

### Prerequisites

*   **Python 3.9+**
*   **Docker & Docker Compose:** For running the MQTT broker and Vector Database.
*   **Git**

### Steps

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/your-username/mqtt-agent-orchestration.git
    cd mqtt-agent-orchestration
    ```

2.  **Set up Environment Variables:**
    Create a `.env` file in the root directory of the project. This file will store sensitive information like API keys.
    ```env
    # MQTT Broker
    MQTT_BROKER_HOST=localhost
    MQTT_BROKER_PORT=1883
    MQTT_BROKER_USERNAME=user
    MQTT_BROKER_PASSWORD=password

    # LLM Provider (Example: OpenAI)
    OPENAI_API_KEY=your_openai_api_key_here
    LLM_MODEL_NAME=gpt-4o # or gpt-3.5-turbo, etc.

    # Vector Database (Example: ChromaDB)
    VECTOR_DB_HOST=localhost
    VECTOR_DB_PORT=8000
    VECTOR_DB_COLLECTION_NAME=agent_knowledge_base
    ```
    *Replace `your_openai_api_key_here` with your actual OpenAI API key.*

3.  **Start Docker Services (MQTT Broker & Vector DB):**
    The `docker-compose.yml` file sets up a Mosquitto MQTT broker and a ChromaDB instance.
    ```bash
    docker-compose up -d
    ```
    Verify that containers are running: `docker ps`

4.  **Install Python Dependencies:**
    ```bash
    pip install -r requirements.txt
    ```

5.  **Initialize Knowledge Base (for RAG):**
    Before running the AI/RAG agent, you need to populate the knowledge base.
    ```bash
    python knowledge_base/vector_db_manager.py --init
    ```
    This script will load example documents from `knowledge_base/data/` into your ChromaDB instance. You can add your own `.txt` or `.md` files to this directory.

6.  **Run the Orchestrator:**
    Open a new terminal and start the Orchestrator.
    ```bash
    python orchestrator/main.py
    ```
    You should see logs indicating the Orchestrator connecting to MQTT and initializing.

7.  **Run Agents:**
    Open separate terminals for each agent type you want to run.
    ```bash
    # Terminal 1: Sensor Agent
    python agents/sensor_agent.py

    # Terminal 2: Actuator Agent
    python agents/actuator_agent.py

    # Terminal 3: AI/RAG Agent
    python agents/ai_rag_agent.py

    # Terminal 4: Data Processing Agent
    python agents/data_processing_agent.py
    ```
    You should see logs indicating agents connecting to MQTT and registering with the Orchestrator.

8.  **Trigger a Workflow (Example):**
    You can trigger a workflow by sending a specific MQTT message or by calling an internal API (if exposed by the Orchestrator). For a quick test, let's simulate an external trigger.

    *   **Option A: Via a simple Python script (recommended for testing):**
        Create a file `trigger_workflow.py`:
        ```python
        import paho.mqtt.client as mqtt
        import json
        import time
        from config.mqtt_topics import MQTT_TOPICS

        broker_host = "localhost"
        broker_port = 1883
        client = mqtt.Client("WorkflowTrigger")

        def on_connect(client, userdata, flags, rc):
            if rc == 0:
                print("Connected to MQTT Broker!")
                # Trigger the smart home workflow
                workflow_payload = {
                    "workflow_name": "smart_home_automation",
                    "parameters": {
                        "target_temperature": 22,
                        "user_preference": "energy_saving"
                    }
                }
                client.publish(MQTT_TOPICS["orchestrator_workflow_trigger"], json.dumps(workflow_payload))
                print(f"Published workflow trigger: {workflow_payload}")
                client.disconnect() # Disconnect after publishing
            else:
                print(f"Failed to connect, return code {rc}\n")

        client.on_connect = on_connect
        client.connect(broker_host, broker_port, 60)
        client.loop_forever()
        ```
        Run this script: `python trigger_workflow.py`

    *   **Option B: Using an MQTT client tool (e.g., MQTT Explorer, mosquitto_pub):**
        Publish a message to the `orchestrator/workflow/trigger` topic:
        ```json
        {
            "workflow_name": "smart_home_automation",
            "parameters": {
                "target_temperature": 22,
                "user_preference": "energy_saving"
            }
        }
        ```
        Using `mosquitto_pub`:
        ```bash
        mosquitto_pub -h localhost -p 1883 -t "orchestrator/workflow/trigger" -m '{"workflow_name": "smart_home_automation", "parameters": {"target_temperature": 22, "user_preference": "energy_saving"}}'
        ```

    Observe the logs in the Orchestrator and Agent terminals to see the workflow progress.

## 5. Usage Examples

Here are a few scenarios demonstrating how the system can be used:

### Example 1: Smart Home Temperature Control with User Preferences

**Scenario:** A smart home system needs to adjust the thermostat based on current temperature, but also consider user preferences (e.g., "energy saving," "comfort," "quick cool") which are stored in the RAG knowledge base.

1.  **Sensor Agent:** Publishes `current_temperature` to `/data/temperature`.
2.  **Orchestrator:** A workflow `smart_home_automation` is triggered by a schedule or a user command.
3.  **Orchestrator:** Dispatches a task to the `AI/RAG Agent` with the current temperature and the user's preference (e.g., "energy_saving").
4.  **AI/RAG Agent:**
    *   Receives the task.
    *   Queries its internal knowledge base (RAG) for "energy saving temperature guidelines" or "optimal temperature ranges for energy saving."
    *   Uses the retrieved context and the current temperature to prompt the LLM: "Given current temperature X and user preference 'energy saving', what should be the target temperature and fan speed?"
    *   The LLM responds with a recommended `target_temperature` and `fan_speed`.
    *   Publishes a command to the Orchestrator: `{"action": "set_thermostat", "target_temp": Y, "fan_speed": Z}`.
5.  **Orchestrator:** Receives the command, identifies the `Actuator Agent` responsible for thermostats.
6.  **Orchestrator:** Dispatches a task to the `Actuator Agent` to set the thermostat.
7.  **Actuator Agent:** Receives the task and sends the command to the physical thermostat device.

### Example 2: Industrial Anomaly Detection and Automated Troubleshooting

**Scenario:** A machine in a factory is reporting unusual vibration data. The system should detect this, identify potential causes, and suggest/execute a troubleshooting step.

1.  **Sensor Agent:** Publishes `vibration_data` from a machine to `/data/machine/vibration`.
2.  **Data Processing Agent:** Subscribes to `/data/machine/vibration`, processes the raw data (e.g., calculates RMS, FFT), and detects an anomaly based on predefined thresholds.
3.  **Data Processing Agent:** Publishes an `anomaly_detected` event to `/events/anomaly`.
4.  **Orchestrator:** A workflow `industrial_troubleshooting` is triggered by the `anomaly_detected` event.
5.  **Orchestrator:** Dispatches a task to the `AI/RAG Agent` with the anomaly details (e.g., "high vibration on motor X, frequency Y").
6.  **AI/RAG Agent:**
    *   Receives the task.
    *   Queries its knowledge base (RAG) for "troubleshooting high vibration on motor X," "common causes of vibration at frequency Y," or "maintenance procedures for motor X."
    *   Uses the retrieved context and anomaly details to prompt the LLM: "Motor X has high vibration at frequency Y. Based on maintenance logs and manuals, what are the top 3 probable causes and the first recommended troubleshooting step?"
    *   The LLM responds with probable causes and a recommended `troubleshooting_action` (e.g., "check motor mounting bolts").
    *   Publishes a command to the Orchestrator: `{"action": "dispatch_maintenance_task", "task_description": "Check motor X mounting bolts", "priority": "high"}`.
7.  **Orchestrator:** Receives the command, potentially dispatches it to a "Maintenance Dispatch Agent" (a new agent type) or logs it for human intervention. If an automated action is possible (e.g., "reduce motor speed"), it could dispatch to an `Actuator Agent`.

## 6. System Requirements

*   **Operating System:** Linux, macOS, or Windows (with Docker Desktop).
*   **Python:** Version 3.9 or higher.
*   **Docker & Docker Compose:** Latest stable versions.
*   **Memory:** Minimum 4GB RAM (8GB+ recommended, especially if running local LLMs or large vector databases).
*   **Disk Space:** ~2GB for repository, Docker images, and initial knowledge base data.
*   **Internet Connection:** Required for downloading Python packages, Docker images, and interacting with external LLM APIs (e.g., OpenAI).

## 7. File Structure

```
.
├── README.md
├── requirements.txt                # Python dependencies
├── docker-compose.yml              # Docker setup for MQTT broker and Vector DB
├── .env.example                    # Example environment variables file
├── orchestrator/
│   ├── __init__.py
│   ├── main.py                     # Orchestrator entry point
│   ├── workflow_engine.py          # Manages workflow definitions and state transitions
│   ├── task_manager.py             # Handles task assignment to agents
│   ├── agent_registry.py           # Manages registered agents and their capabilities
│   └── config.py                   # Orchestrator-specific configuration
├── agents/
│   ├── __init__.py
│   ├── base_agent.py               # Abstract base class for all agents
│   ├── sensor_agent.py             # Example: Simulates sensor data publishing
│   ├── actuator_agent.py           # Example: Simulates device control
│   ├── ai_rag_agent.py             # Handles RAG queries and LLM interactions
│   └── data_processing_agent.py    # Example: Processes and transforms data
├── knowledge_base/
│   ├── __init__.py
│   ├── vector_db_manager.py        # Interface for interacting with the vector database (e.g., ChromaDB)
│   └── data/                       # Directory for knowledge documents (e.g., .txt, .md files)
│       ├── manual_v1.txt
│       └── troubleshooting_guide.md
├── workflows/
│   ├── __init__.py
│   ├── smart_home_automation.py    # Example workflow definition for smart home
│   └── industrial_alert_workflow.py# Example workflow definition for industrial alerts
├── config/
│   ├── __init__.py
│   ├── settings.py                 # Global application settings, loads from .env
│   └── mqtt_topics.py              # Centralized definition of all MQTT topics
├── utils/
│   ├── __init__.py
│   ├── logger.py                   # Centralized logging utility
│   └── mqtt_client.py              # Reusable MQTT client wrapper with connection handling
└── tests/
    ├── __init__.py
    ├── test_orchestrator.py
    ├── test_agents.py
    └── test_workflows.py
```

## 8. Contributing Guidelines

We welcome contributions to enhance this project! Please follow these guidelines:

1.  **Fork the Repository:** Start by forking the `mqtt-agent-orchestration` repository to your GitHub account.
2.  **Clone Your Fork:**
    ```bash
    git clone https://github.com/your-username/mqtt-agent-orchestration.git
    cd mqtt-agent-orchestration
    ```
3.  **Create a New Branch:**
    Choose a descriptive name for your branch (e.g., `feature/add-new-agent`, `bugfix/mqtt-reconnect`).
    ```bash
    git checkout -b feature/your-feature-name
    ```
4.  **Set up Development Environment:**
    Ensure you have all prerequisites and have run `pip install -r requirements.txt`.
5.  **Make Your Changes:**
    *   Adhere to the existing code style (we use `black` for formatting).
    *   Write clear, concise, and well-documented code.
    *   Ensure your changes are thoroughly tested. Add new unit tests for new features or bug fixes.
6.  **Run Tests:**
    Before committing, run the test suite:
    ```bash
    pytest tests/
    ```
7.  **Format Code:**
    ```bash
    black .
    ```
8.  **Commit Your Changes:**
    Write clear and descriptive commit messages.
    ```bash
    git add .
    git commit -m "feat: Add new XYZ agent for ABC functionality"
    ```
9.  **Push to Your Fork:**
    ```bash
    git push origin feature/your-feature-name
    ```
10. **Create a Pull Request (PR):**
    *   Go to your forked repository on GitHub.
    *   Click on "New pull request" and select your branch.
    *   Provide a detailed description of your changes, including why they are needed and how they were tested.
    *   Reference any related issues.

**Code Review Standards:**
*   **Readability:** Code should be easy to understand and follow.
*   **Modularity:** Components should be loosely coupled and have clear responsibilities.
*   **Testability:** Code should be designed to be easily testable, with appropriate unit and integration tests.
*   **Error Handling:** Robust error handling and logging should be in place.
*   **Security:** Be mindful of potential security vulnerabilities, especially when handling sensitive data or external integrations.
*   **Performance:** Consider performance implications for high-throughput scenarios.

## 9. Troubleshooting Section

Here are some common issues and their solutions:

*   **MQTT Connection Issues:**
    *   **Error:** `paho.mqtt.client.WebsocketConnectionError: Connection refused` or similar.
    *   **Solution:**
        *   Ensure the MQTT broker is running. Check `docker ps` to see if the `mosquitto` container is up.
        *   Verify `MQTT_BROKER_HOST` and `MQTT_BROKER_PORT` in your `.env` file match the `docker-compose.yml` configuration.
        *   Check firewall settings if running on a remote machine.
        *   Ensure no other service is using port 1883 (or the configured MQTT port).
*   **Agents Not Registering with Orchestrator:**
    *   **Symptom:** Orchestrator logs don't show agents connecting or registering.
    *   **Solution:**
        *   Ensure all agents are running in separate terminals.
        *   Check agent logs for MQTT connection errors.
        *   Verify `MQTT_TOPICS["agent_registration"]` is correctly defined and used by agents and orchestrator.
*   **RAG Agent Errors (LLM/Vector DB):**
    *   **Error:** `openai.AuthenticationError: Invalid API Key` or similar.
    *   **Solution:**
        *   
