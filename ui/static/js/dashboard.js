// MQTT Agent Orchestration - Dashboard Module
// ===========================================

const Dashboard = {
    charts: {},
    metrics: {
        activeWorkflows: 0,
        activeWorkers: { total: 0, developer: 0, reviewer: 0, tester: 0 },
        ragDocuments: 0,
        successRate: 0,
        recentTasks: []
    },
    
    // Initialize dashboard
    init() {
        console.log('Initializing dashboard...');
        
        // Initialize charts
        this.initCharts();
        
        // Load initial data
        this.loadDashboardData();
        
        // Set up event handlers
        this.setupEventHandlers();
        
        // Initialize real-time updates
        this.initRealTimeUpdates();
    },
    
    // Initialize charts
    initCharts() {
        // Activity Chart
        const activityCtx = document.getElementById('activity-chart')?.getContext('2d');
        if (activityCtx) {
            this.charts.activity = new Chart(activityCtx, {
                type: 'line',
                data: {
                    labels: this.generateTimeLabels(20),
                    datasets: [
                        {
                            label: 'Tasks Processed',
                            data: new Array(20).fill(0),
                            borderColor: Config.ui.charts.colors.primary,
                            backgroundColor: 'rgba(79, 70, 229, 0.1)',
                            tension: 0.4
                        },
                        {
                            label: 'Active Workers',
                            data: new Array(20).fill(0),
                            borderColor: Config.ui.charts.colors.secondary,
                            backgroundColor: 'rgba(6, 182, 212, 0.1)',
                            tension: 0.4
                        }
                    ]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    interaction: {
                        mode: 'index',
                        intersect: false
                    },
                    plugins: {
                        legend: {
                            position: 'top'
                        },
                        tooltip: {
                            mode: 'index',
                            intersect: false
                        }
                    },
                    scales: {
                        x: {
                            display: true,
                            title: {
                                display: true,
                                text: 'Time'
                            }
                        },
                        y: {
                            display: true,
                            title: {
                                display: true,
                                text: 'Count'
                            },
                            beginAtZero: true
                        }
                    }
                }
            });
        }
        
        // Success Rate Mini Chart
        const successCtx = document.getElementById('success-chart')?.getContext('2d');
        if (successCtx) {
            this.charts.successRate = new Chart(successCtx, {
                type: 'doughnut',
                data: {
                    labels: ['Success', 'Failed'],
                    datasets: [{
                        data: [0, 0],
                        backgroundColor: [
                            Config.ui.charts.colors.success,
                            Config.ui.charts.colors.danger
                        ],
                        borderWidth: 0
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            display: false
                        },
                        tooltip: {
                            callbacks: {
                                label: (context) => {
                                    const label = context.label || '';
                                    const value = context.parsed || 0;
                                    return `${label}: ${value}%`;
                                }
                            }
                        }
                    }
                }
            });
        }
    },
    
    // Generate time labels for charts
    generateTimeLabels(count) {
        const labels = [];
        const now = new Date();
        
        for (let i = count - 1; i >= 0; i--) {
            const time = new Date(now - i * 60000); // 1-minute intervals
            labels.push(time.toLocaleTimeString('en-US', { 
                hour: '2-digit', 
                minute: '2-digit' 
            }));
        }
        
        return labels;
    },
    
    // Load dashboard data
    async loadDashboardData() {
        try {
            // Load workflows
            const workflows = await API.getWorkflows({ status: 'active' });
            this.metrics.activeWorkflows = workflows.length;
            document.getElementById('active-workflows').textContent = workflows.length;
            
            // Load workers
            const workers = await API.getWorkers();
            this.updateWorkerCounts(workers);
            
            // Load RAG documents count
            const collections = await API.getRAGCollections();
            let totalDocs = 0;
            collections.forEach(col => totalDocs += col.documentCount || 0);
            this.metrics.ragDocuments = totalDocs;
            document.getElementById('rag-documents').textContent = totalDocs;
            
            // Load recent tasks
            const tasks = await API.getTasks({ limit: 10 });
            this.updateRecentTasks(tasks);
            
            // Calculate success rate
            this.calculateSuccessRate(tasks);
            
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            UI.showToast('Failed to load dashboard data', 'error');
        }
    },
    
    // Update worker counts
    updateWorkerCounts(workers) {
        const counts = { total: 0, developer: 0, reviewer: 0, tester: 0 };
        
        workers.forEach(worker => {
            if (worker.status === 'active') {
                counts.total++;
                counts[worker.role] = (counts[worker.role] || 0) + 1;
            }
        });
        
        this.metrics.activeWorkers = counts;
        
        document.getElementById('active-workers').textContent = counts.total;
        document.getElementById('dev-workers').textContent = counts.developer;
        document.getElementById('review-workers').textContent = counts.reviewer;
        document.getElementById('test-workers').textContent = counts.tester;
    },
    
    // Update recent tasks
    updateRecentTasks(tasks) {
        const container = document.getElementById('recent-tasks');
        container.innerHTML = '';
        
        tasks.slice(0, 10).forEach(task => {
            const taskElement = this.createTaskElement(task);
            container.appendChild(taskElement);
        });
        
        this.metrics.recentTasks = tasks;
    },
    
    // Create task element
    createTaskElement(task) {
        const div = document.createElement('div');
        div.className = 'task-item';
        
        const statusClass = {
            active: 'active',
            completed: 'completed',
            failed: 'failed',
            pending: 'pending'
        }[task.status] || '';
        
        div.innerHTML = `
            <div class="task-info">
                <div class="task-status ${statusClass}"></div>
                <div class="task-details">
                    <h4>${task.name || task.id}</h4>
                    <div class="task-meta">
                        <span>Role: ${task.role}</span>
                        <span>•</span>
                        <span>${UI.formatRelativeTime(task.createdAt)}</span>
                    </div>
                </div>
            </div>
            <div class="task-actions">
                <button class="btn btn-sm" onclick="Dashboard.viewTaskDetails('${task.id}')">
                    <i class="fas fa-eye"></i>
                </button>
            </div>
        `;
        
        return div;
    },
    
    // Calculate success rate
    calculateSuccessRate(tasks) {
        const completed = tasks.filter(t => t.status === 'completed').length;
        const failed = tasks.filter(t => t.status === 'failed').length;
        const total = completed + failed;
        
        if (total > 0) {
            const rate = Math.round((completed / total) * 100);
            this.metrics.successRate = rate;
            document.getElementById('success-rate').textContent = `${rate}%`;
            
            // Update chart
            if (this.charts.successRate) {
                this.charts.successRate.data.datasets[0].data = [rate, 100 - rate];
                this.charts.successRate.update();
            }
        }
    },
    
    // Set up event handlers
    setupEventHandlers() {
        // Refresh button
        document.getElementById('refresh-dashboard')?.addEventListener('click', () => {
            this.refresh();
        });
    },
    
    // Initialize real-time updates
    initRealTimeUpdates() {
        // Update chart data periodically
        setInterval(() => {
            this.updateActivityChart();
        }, Config.ui.chartUpdateInterval);
    },
    
    // Update activity chart with new data
    updateActivityChart() {
        if (!this.charts.activity) return;
        
        const chart = this.charts.activity;
        const tasksData = chart.data.datasets[0].data;
        const workersData = chart.data.datasets[1].data;
        
        // Shift data and add new random values (replace with real data)
        tasksData.shift();
        tasksData.push(Math.floor(Math.random() * 10) + this.metrics.activeWorkflows);
        
        workersData.shift();
        workersData.push(this.metrics.activeWorkers.total);
        
        // Update labels
        chart.data.labels = this.generateTimeLabels(20);
        
        // Update chart
        chart.update('none'); // No animation for smooth updates
    },
    
    // Refresh dashboard
    async refresh() {
        console.log('Refreshing dashboard...');
        const refreshBtn = document.getElementById('refresh-dashboard');
        
        UI.showLoading(refreshBtn, true);
        await this.loadDashboardData();
        UI.showLoading(refreshBtn, false);
        
        UI.showToast('Dashboard refreshed', 'success');
    },
    
    // Update metrics
    async updateMetrics() {
        // This is called periodically to update metrics
        try {
            const metrics = await API.getSystemMetrics();
            
            // Update activity chart with real data
            if (this.charts.activity && metrics.taskMetrics) {
                const chart = this.charts.activity;
                const tasksData = chart.data.datasets[0].data;
                
                tasksData.shift();
                tasksData.push(metrics.taskMetrics.processed || 0);
                
                chart.update('none');
            }
            
        } catch (error) {
            console.error('Failed to update metrics:', error);
        }
    },
    
    // View task details
    viewTaskDetails(taskId) {
        // Navigate to workflows section with task details
        window.App.navigateToSection('workflows');
        // Trigger workflow module to show task details
        if (window.Workflows) {
            window.Workflows.showTaskDetails(taskId);
        }
    },
    
    // Real-time update handlers
    updateTaskStatus(data) {
        // Update recent tasks if this task is in the list
        const taskElement = document.querySelector(`[data-task-id="${data.taskId}"]`);
        if (taskElement) {
            const statusEl = taskElement.querySelector('.task-status');
            statusEl.className = `task-status ${data.status}`;
        }
        
        // Update metrics if needed
        if (data.status === 'active') {
            this.metrics.activeWorkflows++;
            document.getElementById('active-workflows').textContent = this.metrics.activeWorkflows;
        }
    },
    
    updateTaskResult(data) {
        // Refresh recent tasks to show updated results
        this.loadRecentTasks();
    },
    
    updateTaskProgress(data) {
        // Could add progress indicators to task items
        console.log('Task progress update:', data);
    },
    
    updateWorkerStatus(data) {
        // Update worker counts based on status changes
        if (data.previousStatus === 'active' && data.status !== 'active') {
            this.metrics.activeWorkers.total--;
            this.metrics.activeWorkers[data.role]--;
        } else if (data.previousStatus !== 'active' && data.status === 'active') {
            this.metrics.activeWorkers.total++;
            this.metrics.activeWorkers[data.role]++;
        }
        
        this.updateWorkerCounts([]);
    },
    
    updateWorkerMetrics(data) {
        // Update worker-related charts
        console.log('Worker metrics update:', data);
    },
    
    updateSystemStatus(data) {
        // Update system status indicators
        console.log('System status update:', data);
    },
    
    updateSystemHealth(data) {
        // Update health indicators
        const isHealthy = data.status === 'healthy';
        document.getElementById('system-status').classList.toggle('connected', isHealthy);
        document.getElementById('system-status').classList.toggle('disconnected', !isHealthy);
    },
    
    updateSystemMetrics(data) {
        // Update system-wide metrics
        this.updateMetrics();
    },
    
    // Load recent tasks
    async loadRecentTasks() {
        try {
            const tasks = await API.getTasks({ limit: 10 });
            this.updateRecentTasks(tasks);
        } catch (error) {
            console.error('Failed to load recent tasks:', error);
        }
    }
};

// Export for use in other modules
window.Dashboard = Dashboard;