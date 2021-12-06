<x-app-layout>
    <link href="https://maxcdn.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" rel="stylesheet" />
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/5.0.0-alpha1/css/bootstrap.min.css">
    <script src="https://cdn.jsdelivr.net/gh/centrifugal/centrifuge-js@2.8.3/dist/centrifuge.min.js"></script>

    <style>
        body {
            background-color: #f4f7f6;
            margin-top: 20px;
        }

        .card {
            background: #fff;
            transition: .5s;
            border: 0;
            margin-bottom: 30px;
            border-radius: .55rem;
            position: relative;
            width: 100%;
            box-shadow: 0 1px 2px 0 rgb(0 0 0 / 10%);
        }

        .chat-app .people-list {
            width: 280px;
            position: absolute;
            left: 0;
            top: 0;
            padding: 20px;
            z-index: 7
        }

        .chat-app .chat {
            margin-left: 280px;
            border-left: 1px solid #eaeaea
        }

        .people-list {
            -moz-transition: .5s;
            -o-transition: .5s;
            -webkit-transition: .5s;
            transition: .5s
        }

        .people-list .chat-list li {
            padding: 10px 15px;
            list-style: none;
            border-radius: 3px
        }

        .people-list .chat-list li:hover {
            background: #efefef;
            cursor: pointer
        }

        .people-list .chat-list li.active {
            background: #efefef
        }

        .people-list .chat-list li .name {
            font-size: 15px
        }

        .people-list .chat-list img {
            width: 45px;
            border-radius: 50%
        }

        .people-list img {
            float: left;
            border-radius: 50%
        }

        .people-list .about {
            float: left;
            padding-left: 8px
        }

        .people-list .status {
            color: #999;
            font-size: 13px
        }

        .chat .chat-header {
            padding: 15px 20px;
            border-bottom: 2px solid #f4f7f6
        }

        .chat .chat-header img {
            float: left;
            border-radius: 40px;
            width: 40px
        }

        .chat .chat-header .chat-about {
            float: left;
            padding-left: 10px
        }

        .chat .chat-history {
            padding: 20px;
            border-bottom: 2px solid #fff
        }

        .chat .chat-history ul {
            padding: 0
        }

        .chat .chat-history ul li {
            list-style: none;
            margin-bottom: 30px
        }

        .chat .chat-history ul li:last-child {
            margin-bottom: 0px
        }

        .chat .chat-history .message-data {
            margin-bottom: 15px
        }

        .chat .chat-history .message-data img {
            border-radius: 40px;
            width: 40px;
            display: inline;
        }

        .chat .chat-history .message-data-time {
            color: #434651;
            padding-left: 6px
        }

        .chat .chat-history .message {
            color: #444;
            padding: 18px 20px;
            line-height: 26px;
            font-size: 16px;
            border-radius: 7px;
            display: inline-block;
            position: relative
        }

        .chat .chat-history .message:after {
            bottom: 100%;
            left: 7%;
            border: solid transparent;
            content: " ";
            height: 0;
            width: 0;
            position: absolute;
            pointer-events: none;
            border-bottom-color: #fff;
            border-width: 10px;
            margin-left: -10px
        }

        .chat .chat-history .my-message {
            background: #efefef
        }

        .chat .chat-history .my-message:after {
            bottom: 100%;
            left: 30px;
            border: solid transparent;
            content: " ";
            height: 0;
            width: 0;
            position: absolute;
            pointer-events: none;
            border-bottom-color: #efefef;
            border-width: 10px;
            margin-left: -10px
        }

        .chat .chat-history .other-message {
            background: #e8f1f3;
            text-align: right
        }

        .chat .chat-history .other-message:after {
            border-bottom-color: #e8f1f3;
            left: 93%
        }

        .chat .chat-message {
            padding: 20px
        }

        .online,
        .offline,
        .me {
            margin-right: 2px;
            font-size: 8px;
            vertical-align: middle
        }

        .online {
            color: #86c541
        }

        .offline {
            color: #e47297
        }

        .me {
            color: #1d8ecd
        }

        .float-right {
            float: right
        }

        .clearfix:after {
            visibility: hidden;
            display: block;
            font-size: 0;
            content: " ";
            clear: both;
            height: 0
        }

        @media only screen and (max-width: 767px) {
            .chat-app .people-list {
                height: 465px;
                width: 100%;
                overflow-x: auto;
                background: #fff;
                left: -400px;
                display: none
            }

            .chat-app .people-list.open {
                left: 0
            }

            .chat-app .chat {
                margin: 0
            }

            .chat-app .chat .chat-header {
                border-radius: 0.55rem 0.55rem 0 0
            }

            .chat-app .chat-history {
                height: 300px;
                overflow-x: auto
            }
        }

        @media only screen and (min-width: 768px) and (max-width: 992px) {
            .chat-app .chat-list {
                height: 650px;
                overflow-x: auto
            }

            .chat-app .chat-history {
                height: 600px;
                overflow-x: auto
            }
        }

        @media only screen and (min-device-width: 768px) and (max-device-width: 1024px) and (orientation: landscape) and (-webkit-min-device-pixel-ratio: 1) {
            .chat-app .chat-list {
                height: 480px;
                overflow-x: auto
            }

            .chat-app .chat-history {
                height: calc(100vh - 350px);
                overflow-x: auto
            }
        }
    </style>

    <x-slot name="header">
        <h2 class="font-semibold text-xl text-gray-800 leading-tight">
            {{ __('Chat Rooms') }}
        </h2>
    </x-slot>

    <div class="py-12">
        <div class="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <form class="my-5" method="post" action="{{ route('rooms.store') }}">
                @csrf
                <div>
                    <x-input class="block mt-1 w-full" type="text" name="name" required autofocus />
                </div>
                <div class="flex items-center justify-end mt-4">
                    <x-button>
                        {{ __('Add room') }}
                    </x-button>
                </div>
            </form>

            <div class="container">
                <div class="row clearfix">
                    <div class="col-lg-12">
                        <div class="card chat-app">
                            <div id="plist" class="people-list">
                                <ul class="list-unstyled chat-list mt-2 mb-0">
                                    @foreach($rooms as $room)
                                        <li onclick="location.href='{{ route('rooms.show', $room->id) }}'" id="room-{{ $room->id }}" class="clearfix {{ !empty($currRoom) && $currRoom->id === $room->id ? 'active' : ''}}">
                                            <img src="http://127.0.0.1/chat-icon.png" alt="avatar">
                                            <div class="about">
                                                <div class="name">{{ $room->name }}</div>
                                                <div class="status">{{ ($room->messages->count() > 0) ? $room->messages->last()->message : '' }}</div>
                                                <div class="status date">{{ ($room->messages->count() > 0) ? $room->messages->last()->created_at->toDateTimeString() : '' }}</div>
                                            </div>
                                        </li>
                                    @endforeach
                                </ul>
                            </div>
                            <div class="chat">
                                <div class="chat-header clearfix">
                                    <div class="row">
                                        <div class="col-lg-6">
                                            <a href="javascript:void(0);" data-toggle="modal" data-target="#view_info">
                                                <img src="https://bootdey.com/img/Content/avatar/avatar2.png" alt="avatar">
                                            </a>
                                            @if (!empty($currRoom))
                                                <div class="chat-about">
                                                    <h6 class="m-b-0">{{ Auth::user()->name }}</h6>
                                                    <small>Num room participants: {{ $currRoom->users->count() }}</small>
                                                </div>
                                            @endif
                                        </div>
                                    </div>
                                </div>
                                @if (!empty($currRoom))
                                    <div class="chat-history">
                                        <ul class="m-b-0">
                                            @foreach($currRoom->messages as $message)
                                                <li class="clearfix">
                                                    @if ($message->sender_id === Auth::user()->id)
                                                        <div class="message-data">
                                                            <span class="message-data-time">
                                                                {{ $message->created_at->toFormattedDateString() }}, {{ $message->created_at->toTimeString() }}
                                                            </span>
                                                        </div>
                                                        <div class="message my-message">{{ $message->message }}</div>
                                                    @else
                                                        <div class="message-data text-right">
                                                            <span class="message-data-time">
                                                                {{ $message->created_at->toFormattedDateString() }}, {{ $message->created_at->toTimeString() }}
                                                            </span>
                                                            <img src="https://bootdey.com/img/Content/avatar/avatar7.png" alt="avatar">
                                                        </div>
                                                        <div class="message other-message float-right">{{ $message->message }}</div>
                                                    @endif
                                                </li>
                                            @endforeach
                                        </ul>
                                    </div>
                                @else
                                    <div style="position: relative; text-align: center; color: #4a5568;">
                                        <img src="http://127.0.0.1/background.jpeg" alt="background">
                                        <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);">Please choose a room</div>
                                    </div>
                                @endif
                                @if (!empty($currRoom))
                                    @if ($isJoin)
                                        <div class="chat-message clearfix">
                                            <div class="form-group">
                                                <div class="input-group mb-3">
                                                    <span class="input-group-text" id="basic-addon1"><i class="fa fa-send"></i></span>
                                                    <input type="text" id="chat-message-input" class="form-control" placeholder="Enter text here..." aria-label="Username" aria-describedby="basic-addon1">
                                                </div>
                                            </div>
                                        </div>
                                    @else
                                        <form class="inline-block px-4 py-2 bg-blue-700 rounded-md text-xs text-white hover:bg-blue-500" method="post" action="{{ route('rooms.join', $room->id) }}">
                                            @csrf
                                            <button type="submit">JOIN</button>
                                        </form>
                                    @endif
                                @endif
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        const userId = {{ Auth::user()->id }}
        const roomId = {{ !empty($currRoom) ? $currRoom-> id : 0 }};
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
                isSelf = ctx.data.senderId === userId
                addMessage(ctx.data.text, ctx.data.createdAtFormatted, isSelf)
            }

            const lastRoomMessageText = document.querySelector('#room-' + roomId + ' .status');
            const lastRoomMessageDate = document.querySelector('#room-' + roomId + ' .status.date');

            var text = ctx.data.text.substr(0,10) +  ctx.data.text.substr(ctx.data.text.length+1);
            if (ctx.data.text.length > 10) {
                text += "..."
            }

            lastRoomMessageText.innerHTML = text;
            lastRoomMessageDate.innerHTML = ctx.data.createdAt;
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

        function addMessage(text, date, isSelf) {
            const chatThreads = document.querySelector('.chat-history ul');

            var data = '<div class="message-data text-right">' +
                '<span class="message-data-time">' + date + '</span>' +
                '<img src="https://bootdey.com/img/Content/avatar/avatar7.png" alt="avatar">' +
                '</div>' +
                '<div class="message other-message float-right">' + text + '</div>'

            if (isSelf) {
                data = '<div class="message-data">' +
                    '<span class="message-data-time">' + date + '</span>' +
                    '</div>' +
                    '<div class="message my-message">' + text + '</div>'
            }

            const chatNewThread = document.createElement('li');
            chatNewThread.className = "clearfix";
            chatNewThread.innerHTML = data
            chatThreads.appendChild(chatNewThread)
        }
    </script>
</x-app-layout>
