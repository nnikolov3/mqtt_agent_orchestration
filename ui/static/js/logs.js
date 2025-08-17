// MQTT Agent Orchestration - Logs Module
// ======================================

const Logs = {
    logBuffer: [],
    maxLogs: 1000,
    
    init() {
        console.log('Logs module initialized');
        
        // Clear logs button
        document.getElementById('clear-logs')?.addEventListener('click', () => {
            this.clearLogs();
        });
        
        // Export logs button
        document.getElementById('export-logs')?.addEventListener('click', () => {
            this.exportLogs();
        });
    },
    
    refresh() {
        console.log('Refreshing logs...');
    },
    
    addLogEntry(data) {
        this.logBuffer.push(data);
        if (this.logBuffer.length > this.maxLogs) {
            this.logBuffer.shift();
        }
        
        // Add to viewer if logs section is active
        if (window.App.currentSection === 'logs') {
            this.appendLogToViewer(data);
        }
    },
    
    appendLogToViewer(log) {
        const viewer = document.getElementById('log-viewer');
        if (!viewer) return;
        
        const entry = document.createElement('div');
        entry.className = 'log-entry';
        entry.innerHTML = `
            <span class="log-timestamp">${new Date(log.timestamp).toLocaleTimeString()}</span>
            <span class="log-level ${log.level}">${log.level.toUpperCase()}</span>
            <span class="log-message">${log.message}</span>
        `;
        
        viewer.appendChild(entry);
        viewer.scrollTop = viewer.scrollHeight;
    },
    
    clearLogs() {
        this.logBuffer = [];
        const viewer = document.getElementById('log-viewer');
        if (viewer) viewer.innerHTML = '';
        UI.showToast('Logs cleared', 'success');
    },
    
    exportLogs() {
        const data = JSON.stringify(this.logBuffer, null, 2);
        const blob = new Blob([data], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `logs-${new Date().toISOString()}.json`;
        a.click();
        URL.revokeObjectURL(url);
        UI.showToast('Logs exported', 'success');
    }
};

window.Logs = Logs;