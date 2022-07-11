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

document.getElementById('btn1').addEventListener("click", async () => {
    resp = await centrifuge.namedRPC("s1:test", {});
    drawText('RPC response from ' + resp.data.service);
});

document.getElementById('btn2').addEventListener("click", async () => {
    resp = await centrifuge.namedRPC("s2:test", {});
    drawText('RPC response from ' + resp.data.service);
});
