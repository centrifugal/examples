$(window).load(function(){
    var url = "http://localhost:8000/connection";

    // note that you MUST NEVER reveal project secret key in production
    // this is just a demo where we generate connection token on client side
    var secret = "secret";

    var user = "42";
    var timestamp = parseInt(new Date().getTime()/1000).toString();

    var hmacBody = user + timestamp;

    var shaObj = new jsSHA("SHA-256", "TEXT");
    shaObj.setHMACKey(secret, "TEXT");
    shaObj.update(hmacBody);
    var token = shaObj.getHMAC("HEX");

    var log = $('#log');
    var nickname = $('#nickname');
    var input = $('#input');

    var channel = 'public:jsfiddle-chat';

    var get_current_time = function() {
        var pad = function (n) {return ("0" + n).slice(-2);};
        var d = new Date();
        return pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
    }

    var create_message = function(text, from) {
        var time = get_current_time();
        var add_class = "";
        if (typeof(from) == "undefined") {
            add_class += " system";
        }
        var from  = from || "system";
        var message = $('<div class="message ' + add_class + '"></div>');
        var time_span = $('<span class="time"></span>');
        var from_span = $('<span class="from"></span>');
        var text_span = $('<span class="text"></span>');
        time_span.text(time);
        from_span.text(from + ':');
        text_span.text(text);
        message.append(time_span).append(from_span).append(text_span);
        return message;
    }

    var add_message = function(text, from) {
        log.prepend(create_message(text, from))
    }

    var centrifuge = new Centrifuge({
        // please, read Centrifuge documentation to understand 
        // what does each option mean here
        "url": url,
        "user": user,
        "timestamp": timestamp,
        "token": token,
        "debug": true
    });

    var subscription;

    var subscribe = function() {
        subscription = centrifuge.subscribe(channel, function(message) {
            if (message.data) {
                add_message(message.data["input"], message.data["nick"]);
            }
        });

        subscription.on('subscribe', function() {
            add_message("subscribed on channel jsfiddle-chat");
        });
        
        subscription.presence().then(function(message) {
            var count = 0;
            for (var key in message.data){
                count++;
            }
            add_message('now connected ' + count + ' clients');
        }, function(err) {}); 
        
        subscription.on('join', function(message) {
            add_message('someone joined channel');
        });

        subscription.on('leave', function(message) {
            add_message('someone left channel');
        });
    }

    centrifuge.on('connect', function() {
        add_message("connected to Centrifugo");
        subscribe();
        setInterval(function() {
            // Heroku closes inactive websocket connection after 55 sec,
            // so let's send ping message periodically
            centrifuge.ping();
        }, 40000);
    });

    centrifuge.on('disconnect', function(){
        add_message('disconnected from Centrifugo');
    });

    input.on('keypress', function(e) {
        if (e.keyCode === 13 && centrifuge.isConnected() === true) {
            var text = input.val();
            if (text.length === 0) {
                return;
            }
            var nick = nickname.val();
            if (nick.length === 0) {nick = "anonymous";}
            data = {
                "nick": nick,
                "input": input.val()
            }
            subscription.publish(data);
            input.val('');
        }
    });

    centrifuge.connect();
});