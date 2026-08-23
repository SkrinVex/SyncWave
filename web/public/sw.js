const CACHE_NAME = 'syncwave-v1';

self.addEventListener('install', (e) => {
  self.skipWaiting();
});

self.addEventListener('activate', (e) => {
  e.waitUntil(self.clients.claim());
});

self.addEventListener('fetch', (e) => {
  // Let the browser handle everything normally, we just need a fetch handler for PWA installability
});
