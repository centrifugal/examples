import React, { useState, useEffect } from 'react';
import logo from './assets/centrifugo.svg'
import { Centrifuge } from "centrifuge";
import { useKeycloak } from '@react-keycloak/web'
import './App.css'

function App() {
  const { keycloak, initialized } = useKeycloak()
  const [connectionState, setConnectionState] = useState("disconnected");
  const [publishedData, setPublishedData] = useState("");
  const stateToEmoji = {
    "disconnected": "🔴",
    "connecting": "🟠",
    "connected": "🟢"
  }

  useEffect(() => {
    if (!initialized || !keycloak.authenticated) {
      return;
    }
    const centrifuge = new Centrifuge("ws://localhost:8000/connection/websocket", {
      token: keycloak.token,
      getToken: function () {
        return new Promise((resolve, reject) => {
          keycloak.updateToken(5).then(function () {
            resolve(keycloak.token);
          }).catch(function (err) {
            reject(err);
            keycloak.logout();
          });
        })
      }
    });
    centrifuge.on('state', function (ctx) {
      setConnectionState(ctx.newState);
    })

    const userChannel = "#" + keycloak.tokenParsed?.sub;
    const sub = centrifuge.newSubscription(userChannel);
    sub.on("publication", function (ctx) {
      setPublishedData(JSON.stringify(ctx.data));
    }).subscribe();

    centrifuge.connect();

    return () => {
      centrifuge.disconnect();
    };
  }, [keycloak, initialized]);

  if (!initialized) {
    return null;
  }

  return (
    <div>
      <header>
        <p>
          <img src={logo} width="100px" height="100px" />
        </p>
        <p>
          SSO with Keycloak and Centrifugo
          &nbsp;
          <span className={"connectionState " + connectionState}>
            {stateToEmoji[connectionState]}
          </span>
        </p>
        {keycloak.authenticated ? (
          <div>
            <p>Logged in as {keycloak.tokenParsed?.preferred_username + ", channel: #" + keycloak.tokenParsed?.sub}</p>
            {publishedData && (
              <pre>{publishedData}</pre>
            )}
            <button type="button" onClick={() => keycloak.logout()}>
              Logout
            </button>
          </div>
        ) : (
          <button type="button" onClick={() => keycloak.login()}>
            Login
          </button>
        )}
      </header>
    </div >
  );
}

export default App
