// MQTT Agent Orchestration - Workflows Module
// ===========================================

const Workflows = {
    workflows: [],
    currentWorkflow: null,
    filters: {
        search: '',
        status: '',
        role: ''
    },
    
    // Initialize workflows
    init() {
        console.log('Initializing workflows...');
        
        // Set up event handlers
        this.setupEventHandlers();
        
        // Load initial data
        this.loadWorkflows();
    },
    
    // Set up event handlers
    setupEventHandlers() {
        // Create workflow button
        document.getElementById('create-workflow')?.addEventListener('click', () => {
            window.App.openModal('create-workflow-modal');
        });
        
        // Submit workflow form
        document.getElementById('submit-workflow')?.addEventListener('click', () => {
            this.submitNewWorkflow();
        });
        
        // Filters
        document.getElementById('workflow-search')?.addEventListener('input', (e) => {
            this.filters.search = e.target.value;
            this.applyFilters();
        });
        
        document.getElementById('workflow-status-filter')?.addEventListener('change', (e) => {
            this.filters.status = e.target.value;
            this.applyFilters();
        });
        
        document.getElementById('workflow-role-filter')?.addEventListener('change', (e) => {
            this.filters.role = e.target.value;
            this.applyFilters();
        });
    },
    
    // Load workflows
    async loadWorkflows() {
        try {
            const workflows = await API.getWorkflows();
            this.workflows = workflows;
            this.renderWorkflows();
        } catch (error) {
            console.error('Failed to load workflows:', error);
            UI.showToast('Failed to load workflows', 'error');
        }
    },
    
    // Render workflows
    renderWorkflows() {
        const container = document.getElementById('workflow-list');
        container.innerHTML = '';
        
        const filteredWorkflows = this.getFilteredWorkflows();
        
        if (filteredWorkflows.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-project-diagram"></i>
                    <p>No workflows found</p>
                    <button class="btn btn-primary" onclick="window.App.openModal('create-workflow-modal')">
                        Create First Workflow
                    </button>
                </div>
            `;
            return;
        }
        
        filteredWorkflows.forEach(workflow => {
            const element = this.createWorkflowElement(workflow);
            container.appendChild(element);
        });
    },
    
    // Create workflow element
    createWorkflowElement(workflow) {
        const div = document.createElement('div');
        div.className = 'workflow-card';
        
        const statusClass = Config.ui.statusColors[workflow.status] || 'info';
        const statusIcon = {
            active: 'fa-play-circle',
            completed: 'fa-check-circle',
            failed: 'fa-exclamation-circle',
            pending: 'fa-clock'
        }[workflow.status] || 'fa-question-circle';
        
        div.innerHTML = `
            <div class="workflow-info">
                <h4>${workflow.name}</h4>
                <p>${workflow.description || 'No description'}</p>
                <div class="workflow-meta">
                    <span class="badge badge-${statusClass}">
                        <i class="fas ${statusIcon}"></i> ${workflow.status}
                    </span>
                    <span>Role: ${workflow.currentRole || workflow.initialRole}</span>
                    <span>Created: ${UI.formatRelativeTime(workflow.createdAt)}</span>
                </div>
            </div>
            <div class="workflow-actions">
                <button class="btn btn-sm" onclick="Workflows.viewWorkflow('${workflow.id}')">
                    <i class="fas fa-eye"></i> View
                </button>
                ${workflow.status === 'active' ? `
                    <button class="btn btn-sm btn-danger" onclick="Workflows.stopWorkflow('${workflow.id}')">
                        <i class="fas fa-stop"></i> Stop
                    </button>
                ` : ''}
            </div>
        `;
        
        return div;
    },
    
    // Get filtered workflows
    getFilteredWorkflows() {
        return this.workflows.filter(workflow => {
            // Search filter
            if (this.filters.search) {
                const searchLower = this.filters.search.toLowerCase();
                if (!workflow.name.toLowerCase().includes(searchLower) &&
                    !workflow.description?.toLowerCase().includes(searchLower)) {
                    return false;
                }
            }
            
            // Status filter
            if (this.filters.status && workflow.status !== this.filters.status) {
                return false;
            }
            
            // Role filter
            if (this.filters.role && workflow.currentRole !== this.filters.role &&
                workflow.initialRole !== this.filters.role) {
                return false;
            }
            
            return true;
        });
    },
    
    // Apply filters
    applyFilters() {
        this.renderWorkflows();
    },
    
    // Submit new workflow
    async submitNewWorkflow() {
        const form = document.getElementById('create-workflow-form');
        
        const workflow = {
            name: document.getElementById('workflow-name').value,
            description: document.getElementById('workflow-description').value,
            initialRole: document.getElementById('workflow-initial-role').value,
            priority: document.getElementById('workflow-priority').value,
            content: document.getElementById('workflow-content').value
        };
        
        // Validate
        if (!workflow.name || !workflow.content) {
            UI.showToast('Please fill in all required fields', 'warning');
            return;
        }
        
        try {
            // Submit via MQTT
            const taskId = await MQTT.submitTask({
                workflowName: workflow.name,
                description: workflow.description,
                role: workflow.initialRole,
                priority: workflow.priority,
                content: workflow.content,
                type: 'workflow'
            });
            
            UI.showToast('Workflow created successfully', 'success');
            window.App.closeModal('create-workflow-modal');
            form.reset();
            
            // Reload workflows
            this.loadWorkflows();
            
        } catch (error) {
            console.error('Failed to create workflow:', error);
            UI.showToast('Failed to create workflow', 'error');
        }
    },
    
    // View workflow details
    async viewWorkflow(workflowId) {
        try {
            const workflow = await API.getWorkflow(workflowId);
            const tasks = await API.getWorkflowTasks(workflowId);
            
            this.currentWorkflow = workflow;
            this.showWorkflowDetails(workflow, tasks);
            
        } catch (error) {
            console.error('Failed to load workflow details:', error);
            UI.showToast('Failed to load workflow details', 'error');
        }
    },
    
    // Show workflow details
    showWorkflowDetails(workflow, tasks) {
        // This would show a detailed view of the workflow
        // For now, let's update the visualization
        this.renderWorkflowDiagram(workflow, tasks);
    },
    
    // Render workflow diagram
    renderWorkflowDiagram(workflow, tasks) {
        const container = document.getElementById('workflow-diagram');
        
        // Simple visualization - in production, use a proper diagram library
        container.innerHTML = `
            <div class="workflow-viz">
                <h3>${workflow.name}</h3>
                <div class="workflow-nodes">
                    ${tasks.map(task => `
                        <div class="workflow-node ${task.status}">
                            <i class="fas ${Config.ui.roleIcons[task.role]}"></i>
                            <span>${task.role}</span>
                            <small>${task.status}</small>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    },
    
    // Stop workflow
    async stopWorkflow(workflowId) {
        if (!confirm('Are you sure you want to stop this workflow?')) {
            return;
        }
        
        try {
            await API.updateWorkflow(workflowId, { status: 'stopped' });
            UI.showToast('Workflow stopped', 'success');
            this.loadWorkflows();
        } catch (error) {
            console.error('Failed to stop workflow:', error);
            UI.showToast('Failed to stop workflow', 'error');
        }
    },
    
    // Show task details
    showTaskDetails(taskId) {
        // Find task in workflows
        let task = null;
        for (const workflow of this.workflows) {
            if (workflow.tasks) {
                task = workflow.tasks.find(t => t.id === taskId);
                if (task) break;
            }
        }
        
        if (task) {
            // Show task details in a modal or panel
            console.log('Show task details:', task);
        }
    },
    
    // Refresh workflows
    refresh() {
        this.loadWorkflows();
    },
    
    // Real-time update handlers
    updateTaskStatus(data) {
        // Update workflow if it contains this task
        const workflow = this.workflows.find(w => w.id === data.workflowId);
        if (workflow) {
            workflow.status = data.workflowStatus || workflow.status;
            this.renderWorkflows();
        }
    },
    
    updateTaskResult(data) {
        // Update workflow visualization if viewing
        if (this.currentWorkflow && this.currentWorkflow.id === data.workflowId) {
            this.viewWorkflow(data.workflowId);
        }
    },
    
    updateTaskProgress(data) {
        // Update progress indicators
        console.log('Workflow task progress:', data);
    },
    
    updateWorkflowStatus(data) {
        const workflow = this.workflows.find(w => w.id === data.workflowId);
        if (workflow) {
            workflow.status = data.status;
            workflow.currentRole = data.currentRole;
            this.renderWorkflows();
        }
    },
    
    updateWorkflowProgress(data) {
        // Update workflow progress visualization
        console.log('Workflow progress:', data);
    }
};

// Export for use in other modules
window.Workflows = Workflows;