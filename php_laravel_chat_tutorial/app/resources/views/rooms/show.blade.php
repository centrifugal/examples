<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8" />
    <title>Chat Room</title>
    <script src="https://cdn.jsdelivr.net/gh/centrifugal/centrifuge-js@2.8.3/dist/centrifuge.min.js"></script>
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/5.0.0-alpha1/css/bootstrap.min.css">
    <style>
        @import 'https://fonts.googleapis.com/css?family=Noto+Sans';

        body {
            background-color: #f3f4f6;
        }

        ::-webkit-scrollbar {
            width: 10px;
        }

        ::-webkit-scrollbar-track {
            border-radius: 10px;
            background-color: rgba(25, 147, 147, 0.1);
        }

        ::-webkit-scrollbar-thumb {
            border-radius: 10px;
            background-color: rgba(25, 147, 147, 0.2);
        }

        .chat-thread {
            margin: 24px auto 0 auto;
            padding: 0 20px 0 0;
            list-style: none;
            overflow-y: scroll;
            overflow-x: hidden;
            position: absolute;
            top: 10px;
            bottom: 80px;
            left: 50%;
            transform: translate(-50%);
        }

        .chat-thread li {
            position: relative;
            clear: both;
            display: inline-block;
            padding: 16px 40px 16px 20px;
            margin: 0 0 20px 0;
            font: 16px/20px "Noto Sans", sans-serif;
            border-radius: 10px;
            background-color: rgba(25, 147, 147, 0.2);
        }

        /* Chat - Avatar */
        .chat-thread li:before {
            position: absolute;
            top: 0;
            width: 50px;
            height: 50px;
            border-radius: 50px;
            content: "";
        }

        /* Chat - Speech Bubble Arrow */
        .chat-thread li:after {
            position: absolute;
            top: 15px;
            content: "";
            width: 0;
            height: 0;
            border-top: 15px solid rgba(25, 147, 147, 0.2);
        }

        .chat-thread li {
            animation: show-chat-odd 0.15s 1 ease-in;
            -moz-animation: show-chat-odd 0.15s 1 ease-in;
            -webkit-animation: show-chat-odd 0.15s 1 ease-in;
            float: right;
            margin-right: 80px;
            color: #0AD5C1;
        }

        .chat-thread li:before {
            right: -80px;
            background-image: url(data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEAYABgAAD/4QBoRXhpZgAATU0AKgAAAAgABAEaAAUAAAABAAAAPgEbAAUAAAABAAAARgEoAAMAAAABAAIAAAExAAIAAAASAAAATgAAAAAAAABgAAAAAQAAAGAAAAABUGFpbnQuTkVUIHYzLjUuMTAA/9sAQwAHBQUGBQQHBgUGCAcHCAoRCwoJCQoVDxAMERgVGhkYFRgXGx4nIRsdJR0XGCIuIiUoKSssKxogLzMvKjInKisq/9sAQwEHCAgKCQoUCwsUKhwYHCoqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioq/8AAEQgAMgAyAwEiAAIRAQMRAf/EAB8AAAEFAQEBAQEBAAAAAAAAAAABAgMEBQYHCAkKC//EALUQAAIBAwMCBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYXGBkaJSYnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX29/j5+v/EAB8BAAMBAQEBAQEBAQEAAAAAAAABAgMEBQYHCAkKC//EALURAAIBAgQEAwQHBQQEAAECdwABAgMRBAUhMQYSQVEHYXETIjKBCBRCkaGxwQkjM1LwFWJy0QoWJDThJfEXGBkaJicoKSo1Njc4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoKDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uLj5OXm5+jp6vLz9PX29/j5+v/aAAwDAQACEQMRAD8A8wre0/w55qLLqM62ysMrEWAdh+PSl8M6fFLMbu5K7YziNT3b1/CqniRLq98UA2SlhHGobnA55FdbajHmZwxTnLlRtm78NabDGhs/tEwchmVfMGD0znvn0pqahoN3fCH+z0RcHcWiKY4745/KuUVtTtdSFtJCvmxHmN+317Gp5dTthGZXUSXPJJVsh1PUH1rP2rZt7GKOi1LwxHIv2jRg20jPks4b64b+h/OuZZWRyrqVZTggjBBrZ0PWftFwkA8xImTpu7gVJr1kpX7UhzJ0cZ5I9a0umtDJxcWYVFFFBJ1WlKkOmQASYyoYjZnk81aukuLHTP7YFtDeW6zJEFmyu4555H8PQemaz9NmR9PhO45C7T+HFa1rrk9iwSZ3vLJU+WyaVUCsDncMjJI64H41riYRVHmS7E4KbliOWT7nPePtPvbm8j1trWSyF3GpkgkYEbl4BBHTjHBrn49GkfRLnUpn2tBgtFj76k4GD/e749K63xb4gtdTtp4LeRismGVyOmDnpWBq2pXd74ZsdPjG2GCTdHDEnzSN/ebHLH+VeXBuyR7E4wu35EXhKKObVCRnMKMwY+hwAPz5rrJ7bzYXTfu3Ag5yKwfC+nNZvcPPJHvdQDGnJjOehPTPsOlb0zLFA8m/hVJ/SvXo0oundnhV6slU5UcjRRRXPY3Luk3giYwSHCscqfetfULXOlR3TXIhZ2ZYfLILHHDkjsO3PJPSuXqeC5aPzd5ZjIQdxOcEDFa+0fJyEKmvac5myyQWJlRZnkkTHytjoeuPepLaa9vi4ib7LbOMyFG5KjtnrVMabPd30gcrCjMSZGPQf1NbF0i29pHBAUlbABVT8pPqSOwrljDW53c+lrl/RWxvymyDAWPjsPSpdVnVF8iNsk8tz0HpVFLloowEYvJjBkIwB7AdhUGSxJJyT1JrpU2o8py1IwlJS6hRS0VAiKiiikUKKUUUUxC0ooooAKKKKQH/2Q==);
        }

        .chat-thread li:after {
            border-right: 15px solid transparent;
            right: -15px;
        }

        .chat-message {
            position: fixed;
            bottom: 18px;
        }

        .chat-message-input {
            width: 100%;
            height: 48px;
            font: 32px/48px "Noto Sans", sans-serif;
            background: none;
            color: #0AD5C1;
            border: 0;
            border-bottom: 1px solid rgba(25, 147, 147, 0.2);
            outline: none;
        }

        /* Small screens */
        @media all and (max-width: 767px) {
            .chat-thread {
                width: 90%;
            }

            .chat-message {
                left: 5%;
                width: 90%;
            }
        }

        /* Medium and large screens */
        @media all and (min-width: 768px) {
            .chat-thread {
                width: 50%;
            }

            .chat-message {
                left: 25%;
                width: 50%;
            }
        }

        @keyframes show-chat-even {
            0% {
                margin-left: -480px;
            }

            100% {
                margin-left: 0;
            }
        }

        @-moz-keyframes show-chat-even {
            0% {
                margin-left: -480px;
            }

            100% {
                margin-left: 0;
            }
        }

        @-webkit-keyframes show-chat-even {
            0% {
                margin-left: -480px;
            }

            100% {
                margin-left: 0;
            }
        }

        @keyframes show-chat-odd {
            0% {
                margin-right: -480px;
            }

            100% {
                margin-right: 0;
            }
        }

        @-moz-keyframes show-chat-odd {
            0% {
                margin-right: -480px;
            }

            100% {
                margin-right: 0;
            }
        }

        @-webkit-keyframes show-chat-odd {
            0% {
                margin-right: -480px;
            }

            100% {
                margin-right: 0;
            }
        }
    </style>
</head>

<body>
    <h3 id="room-name" style="color: darkseagreen">Room: {{ $currRoom->name }}</h3>
    <div class="col-md-5 col-lg-4 order-md-last">
        <ul class="list-group mb-3">
            @foreach($rooms as $room)
            @if ($room->id === $currRoom->id)
            @continue
            @endif
            <li id="room-{{ $room->id }}" class="list-group-item d-flex justify-content-between lh-sm">
                <div>
                    <h5 class="my-0">
                        <a href="{{ route('rooms.show', $room->id) }}" class="btn btn-sm btn-info mb-2">
                            {{ $room->name }}
                        </a>
                    </h5>
                    @if ($room->messages->count() > 0)
                    <div class="message-block">
                        <span class="message-text text-muted" style="display: block; font-size: 15px;">{{ $room->messages->last()->message }}</span>
                        <span class="message-date text-info" style="font-size: .700em;">{{ $room->messages->last()->created_at }}</span>
                    </div>
                    @else
                    <div class="message-block">
                        <span class="message-text text-muted" style="display: block; font-size: 15px;"></span>
                        <span class="message-date text-info" style="font-size: .700em;"></span>
                    </div>
                    @endif
                </div>
            </li>
            @endforeach
        </ul>
    </div>

    <ul id="chat-thread" class="chat-thread">
        @foreach($currRoom->messages as $message)
        <li>{{ $message->message }}</li>
        @endforeach
    </ul>
    <div class="chat-message">
        <input id="chat-message-input" class="chat-message-input" type="text" autocomplete="off" autofocus />
    </div>

    <script>
        const userId = {
            {
                $userId
            }
        }
        const roomId = {
            {
                $currRoom - > id
            }
        };
        const chatThread = document.querySelector('#chat-thread');
        const messageInput = document.querySelector('#chat-message-input');

        const centrifuge = new Centrifuge("ws://" + window.location.host + "/connection/websocket");

        centrifuge.on('connect', function(ctx) {
            console.log("connected", ctx);
        });

        centrifuge.on('disconnect', function(ctx) {
            console.log("disconnected", ctx);
        });

        centrifuge.on('publish', function(ctx) {
            const channel = ctx.channel;
            const payload = JSON.stringify(ctx.data);
            console.log('Publication from server-side channel', channel, payload);

            if (ctx.data.roomId === roomId) {
                addMessage(ctx.data.text)
            } else {
                const lastRoomMessageText = document.querySelector('#room-' + ctx.data.roomId + ' .message-block .message-text');
                const lastRoomMessageDate = document.querySelector('#room-' + ctx.data.roomId + ' .message-block .message-date');
                lastRoomMessageText.innerHTML = ctx.data.text;
                lastRoomMessageDate.innerHTML = ctx.data.createdAt;
            }
        });

        centrifuge.connect();

        messageInput.focus();
        var csrfToken = "{{ csrf_token() }}";
        messageInput.onkeyup = function(e) {
            if (e.keyCode === 13) { // enter, return
                e.preventDefault();
                const message = messageInput.value;
                if (!message) {
                    return;
                }

                var payload = JSON.stringify({
                    message: message
                })

                var xhttp = new XMLHttpRequest();
                xhttp.open("POST", "/rooms/" + roomId + "/publish")
                xhttp.setRequestHeader("X-CSRF-TOKEN", csrfToken)
                xhttp.send(payload);

                messageInput.value = '';
            }
        };

        function addMessage(text) {
            const chatNewThread = document.createElement('li');
            const chatNewMessage = document.createTextNode(text);
            chatNewThread.appendChild(chatNewMessage);
            chatThread.appendChild(chatNewThread);
            chatThread.scrollTop = chatThread.scrollHeight;
        }
    </script>
</body>

</html>