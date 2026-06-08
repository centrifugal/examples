const logEl = document.getElementById('log');
function log(msg) {
  console.log(msg);
  logEl.textContent += msg + '\n';
}

// VAPID public keys are base64url; the browser's pushManager.subscribe needs a Uint8Array.
function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const raw = atob(base64);
  return Uint8Array.from([...raw].map((c) => c.charCodeAt(0)));
}

async function enable() {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    log('This browser does not support the Web Push API.');
    return;
  }

  const permission = await Notification.requestPermission();
  if (permission !== 'granted') {
    log('Notification permission was not granted: ' + permission);
    return;
  }

  const reg = await navigator.serviceWorker.register('sw.js');
  await navigator.serviceWorker.ready;
  log('Service worker registered.');

  const { publicKey } = await (await fetch('/api/vapid')).json();
  if (!publicKey || !/^[A-Za-z0-9_-]+$/.test(publicKey)) {
    log('Invalid VAPID public key from server: ' + JSON.stringify(publicKey) + '. Did you replace the placeholder in centrifugo.json and run `go run . -genvapid`?');
    return;
  }
  let subscription = await reg.pushManager.getSubscription();
  if (!subscription) {
    subscription = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(publicKey),
    });
  }
  log('Push subscription obtained.');

  const user = document.getElementById('user').value;
  const res = await fetch('/api/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ subscription, user }),
  });
  const data = await res.json();
  if (data.error) {
    log('device_register error: ' + JSON.stringify(data.error));
  } else {
    log('Registered device id: ' + (data.result && data.result.id));
  }
}

async function send() {
  const res = await fetch('/api/send', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      user: document.getElementById('user').value,
      title: document.getElementById('title').value,
      body: document.getElementById('body').value,
    }),
  });
  const data = await res.json();
  if (data.error) {
    log('send_push_notification error: ' + JSON.stringify(data.error));
  } else {
    log('Push queued, uid: ' + (data.result && data.result.uid));
  }
}

document.getElementById('enable').addEventListener('click', () => enable().catch((e) => log('error: ' + e)));
document.getElementById('send').addEventListener('click', () => send().catch((e) => log('error: ' + e)));
