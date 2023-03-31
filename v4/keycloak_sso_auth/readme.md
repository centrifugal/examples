This is a source code for [Setting up Keycloak SSO authentication flow and connecting to Centrifugo WebSocket](https://centrifugal.dev/blog/2023/03/31/keycloak-sso-centrifugo) blog post in Centrifugal blog. 

## Steps to run an example

Run Keycloak:

```
docker run --rm -it -p 8080:8080 \
    -e KEYCLOAK_ADMIN=admin \
    -e KEYCLOAK_ADMIN_PASSWORD=admin \
    quay.io/keycloak/keycloak:21.0.1 start-dev
```

1. Create `myrealm`
1. Create `myclient`. Set valid redirect URIs `http://localhost:5173/*`, web origins as `http://localhost:5173`.
1. Create `myuser`, set password to it.

Run Centrifugo:

```
docker run --rm -it -p 8000:8000 \
    -e CENTRIFUGO_ALLOWED_ORIGINS="http://localhost:5173" \
    -e CENTRIFUGO_TOKEN_JWKS_PUBLIC_ENDPOINT="http://host.docker.internal:8080/realms/myrealm/protocol/openid-connect/certs" \
    -e CENTRIFUGO_ALLOW_USER_LIMITED_CHANNELS=true \
    -e CENTRIFUGO_ADMIN=true \
    -e CENTRIFUGO_ADMIN_SECRET=secret \
    -e CENTRIFUGO_ADMIN_PASSWORD=admin \
    centrifugo/centrifugo:v4.1.2 centrifugo
```

Run this app:

```
npm install
npm run dev
```
