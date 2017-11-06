var centrifuge;

var log;

function initConnection() {

    if (centrifuge && centrifuge.isConnected()) {
        centrifuge.disconnect();
    }

    var defaultEndpoint = "http://localhost:8000/connection";
    var defaultUserID = "42";
    var defaultTimestamp = parseInt(new Date().getTime()/1000).toString();
    var defaultInfo = "";

    var defaultSecret = "secret";

    var url = $("#connection-endpoint").val();
    if (!url) {
        url = defaultEndpoint;
        $("#connection-endpoint").val(defaultEndpoint);
    }

    var user = $("#connection-user-id").val();
    if (!user) {
        user = defaultUserID;
        $("#connection-user-id").val(defaultUserID)
    }

    var timestamp = $("#connection-timestamp").val();
    if (!timestamp) {
        timestamp = defaultTimestamp;
        $("#connection-timestamp").val(defaultTimestamp);
    }

    var info = $("#connection-info").val();
    if (!info) {
        info = defaultInfo;
        $("#connection-info").val(defaultInfo);
    }

    var secret = $("#secret").val();
    if (!secret) {
        secret = defaultSecret;
        $("#secret").val(defaultSecret);
    } 

    var hmacBody = user + timestamp + info;
    var shaObj = new jsSHA("SHA-256", "TEXT");
    shaObj.setHMACKey(secret, "TEXT");
    shaObj.update(hmacBody);
    var token = shaObj.getHMAC("HEX");

    $("#hmac-token").text(token);

    centrifuge = new Centrifuge({
        "url": url,
        "user": user,
        "timestamp": timestamp,
        "info": info,
        "token": token,
        "debug": true
    });

    centrifuge.on('connect', function(ctx) {
        addMessage("connected to Centrifugo", ctx);
        subscribe();
    });

    centrifuge.on('disconnect', function(ctx){
        addMessage('disconnected from Centrifugo', ctx);
    });

    centrifuge.connect();
}

$(window).load(function(){
    log = $('#log');

    initConnection();

    $("#credentials input").on("keyup", function(){
        initConnection();
    });
});


function getCurrentTime() {
    var pad = function (n) {return ("0" + n).slice(-2);};
    var d = new Date();
    return pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
}

function createMessage(text, data) {
    var time = getCurrentTime();
    var add_class = "";
    var message = $('<div class="message ' + add_class + '"></div>');
    var time_span = $('<span class="time"></span>');
    var text_span = $('<span class="text"></span>');
    var dataBlock = null
    if (data) {
        dataBlock = $('<pre class="event-data">'+ prettifyJson(data) +'</pre>')
    }
    time_span.text(time);
    text_span.text(text);
    message.append(time_span).append(text_span).append(dataBlock);
    return message;
}

function addMessage(text, from) {
    log.prepend(createMessage(text, from))
}

var subscription;

function subscribe() {
    var channel = 'public:developer_index';

    subscription = centrifuge.subscribe(channel, function(message) {
        if (message.data) {
            addMessage(message.data["input"], message.data["nick"]);
        }
    });

    subscription.on('subscribe', function(message) {
        addMessage("successfully subscribed on channel", message);
    });

    subscription.on('error', function(message) {
        addMessage("error subscribing on channel", message);
    });

    subscription.on('join', function(message) {
        addMessage('join event received', message);
    });

    subscription.on('leave', function(message) {
        addMessage('leave event received', message);
    });

    subscription.presence().then(function(message) {
        var count = 0;
        for (var key in message.data){
            count++;
        }
        addMessage('presence response received: ' + count + ' clients connected', message);
    }, function(err) {
        addMessage('presence error', err);
    });

    subscription.history().then(function(message) {
        addMessage('history response received', message);
    }, function(err) {
        addMessage('presence error', err);
    });

}

function prettifyJson(json) {
    return syntaxHighlight(JSON.stringify(json, undefined, 4));
}

function syntaxHighlight(json) {
    json = json.replace(/&/g, '&').replace(/</g, '<').replace(/>/g, '>');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        var cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}
