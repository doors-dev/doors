self.addEventListener('install', () => {
  console.log('[SW] Installing…');
});

self.addEventListener('activate', () => {
  console.log('[SW] Activated!');
});

self.addEventListener('message', (e) => {
  console.log('[SW] Got message:', e.data);
});

