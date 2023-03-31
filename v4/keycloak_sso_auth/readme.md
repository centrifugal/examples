## Steps

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

Write React app (with Vite):

```
npm create vite@latest keycloak_sso_auth -- --template react
```

Then:

```
npm install --save @react-keycloak/web centrifuge keycloak-js
```

And:

```
npm run dev
```

Open http://localhost:5173/.

Try publishing message into user channel over Centrifugo Web UI at http://localhost:8000/#/actions.
