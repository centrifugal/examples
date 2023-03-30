import React from 'react'
import ReactDOM from 'react-dom/client'
import { ReactKeycloakProvider } from '@react-keycloak/web'
import App from './App'
import './index.css'

import Keycloak from "keycloak-js";

const keycloakClient = new Keycloak({
  url: "http://localhost:8080",
  realm: "myrealm",
  clientId: "myclient"
})

ReactDOM.createRoot(document.getElementById('root')).render(
  <ReactKeycloakProvider authClient={keycloakClient}>
    <React.StrictMode>
      <App />
    </React.StrictMode>
  </ReactKeycloakProvider>,
)
