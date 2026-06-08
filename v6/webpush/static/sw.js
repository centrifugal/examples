// Service worker: receives push events and shows a notification.
self.addEventListener('push', (event) => {
  console.log('[sw] push event received, raw:', event.data ? event.data.text() : '(no data)');
  let data = {};
  try {
    data = event.data ? event.data.json() : {};
  } catch (e) {
    data = { title: 'Notification', body: event.data ? event.data.text() : '' };
  }
  if (typeof data !== 'object' || data === null) {
    data = { title: 'Notification', body: String(data) };
  }
  const title = data.title || 'Notification';
  const options = {
    body: data.body || '',
    icon: data.icon,
    data: data,
  };
  event.waitUntil(self.registration.showNotification(title, options));
});

// Focus or open the app when the notification is clicked.
self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if ('focus' in client) return client.focus();
      }
      if (clients.openWindow) return clients.openWindow('/');
    })
  );
});
