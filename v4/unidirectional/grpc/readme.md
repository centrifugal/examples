```
{
  "token_hmac_secret_key": "secret",
  "api_key": "secret",
  "uni_grpc": true,
  "grpc_api": true,
  "allowed_origins": [
    "*"
  ],
  "join_leave": true,
  "force_push_join_leave": true,
  "allow_user_limited_channels": true,
  "allow_subscribe_for_client": true,
  "namespaces": [
    {
      "name": "chat",
      "history_size": 10,
      "history_ttl": "300000s",
      "join_leave": true,
      "force_push_join_leave": true,
      "allow_user_limited_channels": true,
      "allow_subscribe_for_client": true,
      "allow_subscribe_for_anonymous": false,
      "allow_publish_for_client": true,
      "allow_publish_for_anonymous": true,
      "allow_history_for_subscriber": true,
      "allow_history_for_anonymous": true,
      "force_recovery": true
    },
    {
      "name": "raw",
      "join_leave": true,
      "force_push_join_leave": true,
      "allow_user_limited_channels": true,
      "allow_subscribe_for_client": true,
      "allow_subscribe_for_anonymous": false,
      "allow_publish_for_client": true,
      "allow_publish_for_anonymous": true
    },
    {
      "name": "traffic_light",
      "join_leave": true,
      "force_push_join_leave": true,
      "allow_user_limited_channels": true,
      "allow_subscribe_for_client": true,
      "allow_subscribe_for_anonymous": false,
      "allow_publish_for_client": true,
      "allow_publish_for_anonymous": true
    }
  ]
}
```