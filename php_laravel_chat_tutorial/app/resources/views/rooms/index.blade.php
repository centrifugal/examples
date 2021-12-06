<x-app-layout>
    <link href="https://maxcdn.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" rel="stylesheet" />
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/5.0.0-alpha1/css/bootstrap.min.css">
    <script src="https://cdn.jsdelivr.net/gh/centrifugal/centrifuge-js@2.8.4/dist/centrifuge.min.js"></script>

    <style>
        .chat-app {
            background: #fff;
            border: 0;
            margin-bottom: 30px;
            border-radius: .55rem;
            position: relative;
            width: 100%;
            box-shadow: 0 1px 2px 0 rgb(0 0 0 / 10%);
        }

        .chat-app .room-list {
            width: 280px;
            position: absolute;
            left: 0;
            top: 0;
            padding: 20px;
            height: 600px;
            overflow-y: scroll;
        }

        .chat-app .chat {
            margin-left: 280px;
            border-left: 1px solid #eaeaea;
            height: 600px;
        }

        .room-list .chat-list li {
            padding: 8px 10px;
            list-style: none;
            border-radius: 3px;
        }

        .room-list .chat-list li:hover {
            background: #efefef;
            cursor: pointer
        }

        .room-list .chat-list li.active {
            background: #efefef
        }

        .room-list .chat-list li .name {
            font-size: 16px;
            font-weight: bold;
        }

        .room-list .chat-list img {
            width: 47px;
            border-radius: 50%
        }

        .room-list img {
            float: left;
            border-radius: 50%
        }

        .room-list .about {
            float: left;
            padding-left: 8px;
            max-width: 170px;
        }

        .room-list .status {
            color: #999;
            font-size: 13px
        }

        .chat .chat-header {
            padding: 15px 20px;
            border-bottom: 2px solid #f4f7f6;
            height: 80px;
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
            border-bottom: 2px solid #fff;
            position: relative;
            height: 450px;
            overflow-y: scroll;
        }

        .chat .chat-history ul {
            padding: 0
        }

        .chat .chat-history ul li {
            list-style: none;
            margin-bottom: 30px
        }

        .chat .chat-history ul li:last-child {
            margin-bottom: 0px;
        }

        .chat .chat-history .message-data {
            margin-bottom: 5px;
        }

        .chat .chat-history .message-data img {
            border-radius: 40px;
            width: 40px;
            display: inline;
        }

        .chat .chat-history .message-data-time {
            color: #434651;
            padding-left: 6px;
            font-size: 12px;
        }

        .chat .chat-history .message {
            color: #444;
            padding: 10px 20px;
            line-height: 26px;
            font-size: 14px;
            border-radius: 7px;
            display: inline-block;
            position: relative;
            max-width: 450px;
        }

        .chat .chat-history .my-message {
            background: #efefef;
        }

        .chat .chat-history .other-message {
            background: #e8f1f3;
            text-align: left;
        }

        .chat .chat-message {
            padding: 20px;
            height: 50px;
        }

        .float-right {
            float: right
        }

        .float-left {
            float: left
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
            .chat-app .room-list {
                width: 100%;
                overflow-x: auto;
                background: #fff;
                left: -400px;
                display: none
            }

            .chat-app .room-list.open {
                left: 0
            }

            .chat-app .chat {
                margin: 0
            }

            .chat-app .chat .chat-header {
                border-radius: 0.55rem 0.55rem 0 0
            }

            .chat-app .chat-history {
                overflow-x: auto
            }
        }

        @media only screen and (min-width: 768px) and (max-width: 992px) {
            .chat-app .chat-list {
                overflow-x: auto
            }

            .chat-app .chat-history {
                overflow-x: auto
            }
        }

        @media only screen and (min-device-width: 768px) and (max-device-width: 1024px) and (orientation: landscape) and (-webkit-min-device-pixel-ratio: 1) {
            .chat-app .chat-list {
                overflow-x: auto
            }

            .chat-app .chat-history {
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
                <div class="input-group mb-3">
                    <input type="text" class="form-control" name="name" placeholder="Type a new group name" aria-label="Recipient's username" aria-describedby="button-addon2">
                    <button class="btn btn-outline-secondary" type="submit" id="button-addon2">{{ __('Add room') }}</button>
                </div>
            </form>

            <div class="container">
                <div class="row clearfix">
                    <div class="col-lg-12">
                        <div class="chat-app">
                            <div id="plist" class="room-list">
                                <ul class="list-unstyled chat-list mt-2 mb-0">
                                    @foreach($rooms as $room)
                                    <li onclick="location.href='{{ route('rooms.show', $room->id) }}'" id="room-{{ $room->id }}" class="clearfix {{ !empty($currRoom) && $currRoom->id === $room->id ? 'active' : ''}}">
                                        <img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAYAAAAeP4ixAAAABmJLR0QA/wD/AP+gvaeTAAAD3ElEQVRoge3a2WudRRjH8U+apaatTVp6o7hGLdEalyKKpVQF8Q9wr3pvsUoQoS6I6I1XXqh4I4J4oYiI+0rrhm3d21pjtVEi7i1KQEujbSU5Xjzzek7qOT3be05ykS+8DGeWZ37zzjszz8wc5pijJXTkaOs4XIjzcCpORh8WpfT9+BPfYQyf4kP8kqOGhjkF9+ELFCo8f6SnUvrnuBcDbdYOVuN1TJYI2o4HsRYr0V+mXH9KW4uHsKOk/CReFb3acgbwSknlY9iAE5uweRLuEJ9cZvfFFJ87HRjGRKpoN67CvBzrmIdrMZrq2I/1Odq3EM8k43/hdnTlWcFhdOMu/J3qfDppaIolYmbJeuGMZg3WwRC+SXVvVX7M1cQifJYMvYPFeairkz68lzR8ooGe6RQzSAFv4qg81dVJLzYlLS+rc1zemQruwNG5S6ufxdgpNG2otdBZOIR9YsGbLZwmZrKDWFFLgY2i5etaKKpRbhba3qiWcU3KOCLGyWyjC7uExtVHyvhsynRdFYObVfadmn02V6n7BsX1pSyLxdjYi54ZbMj7Veqej99wQNGznsaVydBjVQzNBh4XWi/PIkrn5FUp3NRORQ2yMYWZ5mkNGUrhzrbJaZxM41C5xK9Fd3W3TU7j9AitX2URpT3SJ9z0f8oUzGtwl5uRqtkuV+aQ8MT7yjVkSr77i3IU2lFmJBWaSQexVo74ae1N4fHtVNQgmQ/4UxZR2pBtKTy/bXIa55wUjmQRpQ35OIUXtU1O41yWwq3lEhcKN3nczLoo1Xyt+fhd7On/c1FKe2QCL2EprqhirJVUm6WuxjI8L158WVaKaXi32enGdysu3Kuq5PVCynhLi0U1wrDQ9lotmQdEl01geQtF1cug0HQAp9daaL3iTrGvSt520I8vhabb6inYgScUz7R6c5dWOwuShoL47Ot2o3rwVjKwRcxm7aZf7BgL4sRzQaOGehVP30dxbh7qauRsfKu4tjR90tmNR5LBg1p/TNSNu8WgLuBJTfREOa5Jhn/O02gJnbhesRf24cZWVHRsquCDnO0OiGuEH5L9KTyHE+oxUs89x5oUbquQvk6sO7vEnvp7cflZuuNcKoQP4gLhoGb77knhIt0vTt5bRnY6f+lh8V3iPrCSEzghGlQpfTvu0eRVW63X08fgR/wqrp2nUvxyPIqLxcZsOAk6M+Vbkp7O1KBx0VNj4q1/hD3NNKBeHhZv76b0exkeELNYQThyg+0U1AgrhOAJ3Crc52xqLOApFY4uZxtv+/93PYV3cckM6ppGLbPWHrF2jItve4v4s8BoC3XNMcds4V/d6V223MWyRgAAAABJRU5ErkJggg==" alt="avatar">
                                        <div class="about">
                                            <div class="name">{{ $room->name }}</div>
                                            <span class="user-name">{{ ($room->messages->count() > 0) ? $room->messages->last()->user->name : '' }}</span>
                                            <span class="status">{{ ($room->messages->count() > 0) ? Str::limit($room->messages->last()->message, 20) : '' }}</span>
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
                                                <img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAYAAAAeP4ixAAAABmJLR0QA/wD/AP+gvaeTAAAD3ElEQVRoge3a2WudRRjH8U+apaatTVp6o7hGLdEalyKKpVQF8Q9wr3pvsUoQoS6I6I1XXqh4I4J4oYiI+0rrhm3d21pjtVEi7i1KQEujbSU5Xjzzek7qOT3be05ykS+8DGeWZ37zzjszz8wc5pijJXTkaOs4XIjzcCpORh8WpfT9+BPfYQyf4kP8kqOGhjkF9+ELFCo8f6SnUvrnuBcDbdYOVuN1TJYI2o4HsRYr0V+mXH9KW4uHsKOk/CReFb3acgbwSknlY9iAE5uweRLuEJ9cZvfFFJ87HRjGRKpoN67CvBzrmIdrMZrq2I/1Odq3EM8k43/hdnTlWcFhdOMu/J3qfDppaIolYmbJeuGMZg3WwRC+SXVvVX7M1cQifJYMvYPFeairkz68lzR8ooGe6RQzSAFv4qg81dVJLzYlLS+rc1zemQruwNG5S6ufxdgpNG2otdBZOIR9YsGbLZwmZrKDWFFLgY2i5etaKKpRbhba3qiWcU3KOCLGyWyjC7uExtVHyvhsynRdFYObVfadmn02V6n7BsX1pSyLxdjYi54ZbMj7Veqej99wQNGznsaVydBjVQzNBh4XWi/PIkrn5FUp3NRORQ2yMYWZ5mkNGUrhzrbJaZxM41C5xK9Fd3W3TU7j9AitX2URpT3SJ9z0f8oUzGtwl5uRqtkuV+aQ8MT7yjVkSr77i3IU2lFmJBWaSQexVo74ae1N4fHtVNQgmQ/4UxZR2pBtKTy/bXIa55wUjmQRpQ35OIUXtU1O41yWwq3lEhcKN3nczLoo1Xyt+fhd7On/c1FKe2QCL2EprqhirJVUm6WuxjI8L158WVaKaXi32enGdysu3Kuq5PVCynhLi0U1wrDQ9lotmQdEl01geQtF1cug0HQAp9daaL3iTrGvSt520I8vhabb6inYgScUz7R6c5dWOwuShoL47Ot2o3rwVjKwRcxm7aZf7BgL4sRzQaOGehVP30dxbh7qauRsfKu4tjR90tmNR5LBg1p/TNSNu8WgLuBJTfREOa5Jhn/O02gJnbhesRf24cZWVHRsquCDnO0OiGuEH5L9KTyHE+oxUs89x5oUbquQvk6sO7vEnvp7cflZuuNcKoQP4gLhoGb77knhIt0vTt5bRnY6f+lh8V3iPrCSEzghGlQpfTvu0eRVW63X08fgR/wqrp2nUvxyPIqLxcZsOAk6M+Vbkp7O1KBx0VNj4q1/hD3NNKBeHhZv76b0exkeELNYQThyg+0U1AgrhOAJ3Crc52xqLOApFY4uZxtv+/93PYV3cckM6ppGLbPWHrF2jItve4v4s8BoC3XNMcds4V/d6V223MWyRgAAAABJRU5ErkJggg==" alt="avatar">
                                            </a>
                                            @if (!empty($currRoom))
                                            <div class="chat-about">
                                                <h6 class="m-b-0">Room: {{ $currRoom->name }}</h6>
                                                <small>Num participants: {{ $currRoom->users->count() }}</small>
                                            </div>
                                            @endif
                                        </div>
                                    </div>
                                </div>
                                <div class="chat-history" id="chat-history">
                                    @if (!empty($currRoom))
                                    <ul class="m-b-0">
                                        @foreach($currRoom->messages as $message)
                                        <li class="clearfix">
                                            @if ($message->sender_id === Auth::user()->id)
                                            <div class="message-data text-right">
                                                <span class="message-data-time">
                                                    {{ $message->created_at->toFormattedDateString() }}, {{ $message->created_at->toTimeString() }}
                                                </span>
                                            </div>
                                            <div class="message my-message float-right">{{ $message->message }}</div>
                                            @else
                                            <div class="message-data text-left">
                                                <img src="https://bootdey.com/img/Content/avatar/avatar7.png" alt="avatar">
                                                <span class="message-data-time">
                                                    {{ $message->created_at->toFormattedDateString() }}, {{ $message->created_at->toTimeString() }}
                                                </span>
                                            </div>
                                            <div class="message other-message float-left">
                                                <b>{{ $message->user->name }}</b><br>
                                                {{ $message->message }}
                                            </div>
                                            @endif
                                        </li>
                                        @endforeach
                                    </ul>
                                    @else
                                    <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);">Please choose a room</div>
                                    @endif
                                </div>

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
                                <div class="chat-message clearfix">
                                    <form class="inline-block" method="post" style="text-align: center;" action="{{ route('rooms.join', $currRoom->id) }}">
                                        @csrf
                                        <button type="submit" class="btn btn-primary">Join this room</button>
                                    </form>
                                </div>
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
        const userId = "{{ Auth::user() -> id }}";
        const roomId = "{{ !empty($currRoom) ? $currRoom -> id : 0 }}";

        const chatHistory = document.querySelector('#chat-history');
        const messageInput = document.querySelector('#chat-message-input');

        function scrollToLastMessage() {
            chatHistory.scrollTop = chatHistory.scrollHeight;
        }
        scrollToLastMessage();

        const centrifuge = new Centrifuge("ws://" + window.location.host + "/connection/websocket");

        centrifuge.on('connect', function(ctx) {
            console.log("connected", ctx);
        });

        centrifuge.on('disconnect', function(ctx) {
            console.log("disconnected", ctx);
        });

        centrifuge.on('publish', function(ctx) {
            if (ctx.data.roomId.toString() === roomId) {
                isSelf = ctx.data.senderId.toString() === userId;
                addMessage(ctx.data.text, ctx.data.createdAtFormatted, isSelf);
                scrollToLastMessage();
            }
            const lastRoomMessageText = document.querySelector('#room-' + ctx.data.roomId + ' .status');
            const lastRoomMessageUserName = document.querySelector('#room-' + ctx.data.roomId + ' .user-name');

            var text = ctx.data.text.substr(0, 15) + ctx.data.text.substr(ctx.data.text.length + 1);
            if (ctx.data.text.length > 15) {
                text += "..."
            }

            lastRoomMessageText.innerHTML = text;
            lastRoomMessageUserName.innerHTML = ctx.data.senderName;
        });

        centrifuge.connect();

        if (messageInput !== null) {
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
        }

        function addMessage(text, date, isSelf) {
            const chatThreads = document.querySelector('#chat-history ul');

            var data = '<div class="message-data text-left">' +
                '<span class="message-data-time">' + date + '</span>' +
                '<img src="https://bootdey.com/img/Content/avatar/avatar7.png" alt="avatar">' +
                '</div>' +
                '<div class="message other-message float-left">' + text + '</div>'

            if (isSelf) {
                data = '<div class="message-data text-right">' +
                    '<span class="message-data-time">' + date + '</span>' +
                    '</div>' +
                    '<div class="message my-message float-right">' + text + '</div>'
            }

            const chatNewThread = document.createElement('li');
            chatNewThread.className = "clearfix";
            chatNewThread.innerHTML = data
            chatThreads.appendChild(chatNewThread)
        }
    </script>
</x-app-layout>