// Initialise DataChannel.js
var datachannel = new DataChannel();

// Set the userid based on what has been defined by DataChannel
// https://github.com/muaz-khan/WebRTC-Experiment/tree/master/DataChannel#use-custom-user-ids
// datachannel.userid ="1";

// initialize Centrifuge object first
var centrifuge = new Centrifuge({
    url: 'ws://localhost:8000/connection/websocket',
    debug: true,
    insecure: true
});

// Storage of Centrifugo connection client ID
var socketId;

// bind listeners on centrifuge object instance events.
centrifuge.on('connect', function (ctx) {
    socketId = ctx.client;
});

// Set custom Pusher signalling channel
// https://github.com/muaz-khan/WebRTC-Experiment/blob/master/Signaling.md
datachannel.openSignalingChannel = function (config) {
    var channel = config.channel || this.channel || "default-channel";

    var publishErrorCount = 0;

    var sub = centrifuge.subscribe(channel, function (message) {
        if (message.client === socketId) {
            return;
        }
        config.onmessage(message.data);
    });

    centrifuge.connect();

    var socket = {
        send: function (message) {
            sub.publish(message).then(function () {
                publishErrorCount = 0;
            }, function (err) {
                console.log(err);
                publishErrorCount++;
                // Stop sending signaller messages if it's down
                if (publishErrorCount > 5) {
                    console.log("Disabling signaller due to connection failure");
                    datachannel.transmitRoomOnce = true;
                }
            });
        },
        channel: channel
    };

    sub.on("subscribe", function () {
        if (config.callback) config.callback(socket);
    });

    return socket;
};

// No Centrifugo specific code after this comment!!!

var onCreateChannel = function () {
    var channelName = cleanChannelName(channelInput.value);

    if (!channelName) {
        console.log("No channel name given");
        return;
    }

    disableConnectInput();

    datachannel.open(channelName);
};

var onJoinChannel = function () {
    var channelName = cleanChannelName(channelInput.value);

    if (!channelName) {
        console.log("No channel name given");
        return;
    }

    disableConnectInput();

    // Search for existing data channels
    datachannel.connect(channelName);
};

var cleanChannelName = function (channel) {
    return channel.replace(/(\W)+/g, "-").toLowerCase();
};

var onSendMessage = function () {
    var message = messageInput.value;

    if (!message) {
        console.log("No message given");
        return;
    }

    datachannel.send(message);
    addMessage(message, window.userid, true);

    messageInput.value = "";
};

var onMessageKeyDown = function (event) {
    if (event.keyCode == 13) {
        onSendMessage();
    }
};

var addMessage = function (message, userId, self) {
    var messages = messageList.getElementsByClassName("list-group-item");

    // Check for any messages that need to be removed
    var messageCount = messages.length;
    for (var i = 0; i < messageCount; i++) {
        var msg = messages[i];

        if (msg.dataset.remove === "true") {
            messageList.removeChild(msg);
        }
    }

    var newMessage = document.createElement("li");
    newMessage.classList.add("list-group-item");

    if (self) {
        newMessage.classList.add("self");
        newMessage.innerHTML = "<span class='badge'>You</span><p>" + message + "</p>";
    } else {
        newMessage.innerHTML = "<span class='badge'>" + userId + "</span><p>" + message + "</p>"
    }

    messageList.appendChild(newMessage);
};

var disableConnectInput = function () {
    channelInput.disabled = true;
    createChannelBtn.disabled = true;
    joinChannelBtn.disabled = true;
};

// Demo DOM elements
var channelInput = document.querySelector(".demo-chat-channel-input");
var createChannelBtn = document.querySelector(".demo-chat-create");
var joinChannelBtn = document.querySelector(".demo-chat-join");
var messageInput = document.querySelector(".demo-chat-message-input");
var sendBtn = document.querySelector(".demo-chat-send");
var messageList = document.querySelector(".demo-chat-messages");

// Set up DOM listeners
createChannelBtn.addEventListener("click", onCreateChannel);
joinChannelBtn.addEventListener("click", onJoinChannel);
sendBtn.addEventListener("click", onSendMessage);
messageInput.addEventListener("keydown", onMessageKeyDown);

// Set up DataChannel handlers
datachannel.onopen = function (userId) {
    document.querySelector(".demo-connect").classList.add("inactive");
    document.querySelector(".demo-chat").classList.remove("inactive");
    messageInput.focus();
};

datachannel.onmessage = function (message, userId) {
    addMessage(message, userId);
};