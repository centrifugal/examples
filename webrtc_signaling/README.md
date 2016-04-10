Centrifugo as signaling server for WebRTC.

This chat application uses WebRTC to create a peer-to-peer, server-less connection
between you and the recipient of your chat messages. This chat uses
<a href="https://github.com/centrifugal/centrifugo">Centrifugo</a> as signaling server
and <a href="https://github.com/muaz-khan/WebRTC-Experiment/tree/master/DataChannel">DataChannel</a>
javascript library for peer-to-peer communication. This demo inspired by
<a href="https://pusher.com/tutorials/webrtc_chat">WebRTC demo</a> of <a href="https://pusher.com">Pusher.com</a> real-time API service.

First run Centrifugo with config like this:

```javascript
{
  "secret": "secret",
  "publish": true,
  "anonymous": true
}
```

And run Centrifugo:

```
./centrifugo --config=config.json --insecure
```

Note that as usually in demos we use insecure options to simplify learning curve. In production you should
follow Centrifugo best practices.

Now start serving this chat application:

```
python -m SimpleHTTPServer 3000
```

Go to http://localhost:3000/ and follow instructions.
