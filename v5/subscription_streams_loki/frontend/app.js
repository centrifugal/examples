const logs = document.getElementById('logs');
const lines = document.getElementById('lines');
const queryInput = document.getElementById('query');
const button = document.getElementById('submit');

function subscribeToLogs(e) {
    e.preventDefault();

    const query = queryInput.value;
    if (!query) {
        alert('Please enter a query.');
        return;
    }
    queryInput.disabled = true;
    button.disabled = true;

    const centrifuge = new Centrifuge('ws://localhost:9000/connection/websocket');

    const subscription = centrifuge.newSubscription('logs:stream', {
        data: { query: query }
    });

    subscription.on('publication', function(ctx) {
        const logLine = ctx.data.line;
        const logItem = document.createElement('li');
        logItem.textContent = logLine;
        lines.appendChild(logItem);
        logs.scrollTop = logs.scrollHeight;
    });

    subscription.subscribe();
    centrifuge.connect();
}
