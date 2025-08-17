const workflowForm = document.getElementById('workflow-form');
const workflowsContainer = document.getElementById('workflows');

workflowForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const content = document.getElementById('content').value;
    const complexity = document.getElementById('complexity').value;
    const apiKey = document.getElementById('api-key').value;

    const response = await fetch('/workflow/submit', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Key': apiKey,
        },
        body: JSON.stringify({ content, complexity }),
    });

    if (response.ok) {
        const message = await response.text();
        alert(message);
        workflowForm.reset();
        fetchWorkflows();
    } else {
        const error = await response.text();
        alert(`Error: ${error}`);
    }
});

async function fetchWorkflows() {
    // In a real application, you would fetch the list of workflows from the server.
    // For this example, we'll just display a message.
    workflowsContainer.innerHTML = '<p>Workflows are being processed...</p>';
}

fetchWorkflows();
