function drawText(text) {
    const div = document.createElement('div');
    div.innerHTML = text;
    document.getElementById('log').appendChild(div);
}

const centrifuge = new Centrifuge('ws://localhost:9000/connection/websocket');

centrifuge.on('connect', function () {
    drawText('Connected to Centrifugo');
});

centrifuge.on('disconnect', function () {
    drawText('Disconnected from Centrifugo');
});

centrifuge.on('publish', function (ctx) {
    drawText('Publication, time = ' + ctx.data.time);
});

centrifuge.connect();
