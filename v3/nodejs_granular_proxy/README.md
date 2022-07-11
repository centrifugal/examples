Example of Centrifugo integration with NodeJS project.

Features covered:

* Using granular proxy mode
* Using connect proxy feature to authenticate over standard express.js session
* Distributing RPC requests sent over WebSocket to different app endpoints (could be differrent microservices in practice)  

Why integrate Centrifugo with NodeJS backend:

* Centrifugo scales well
* Centrifugo is pretty fast
* Centrifugo provides a variety of features out-of-the-box 
* Centrifugo works as a separate service – so can be a universal tool in your pocket

To run:

```
docker compose up
```

Then go to [http://localhost:9000](http://localhost:9000).
