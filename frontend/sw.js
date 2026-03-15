const CACHE_NAME = 'markdown-server-v1';
const STATIC_ASSETS = [
    '/',
    '/index.html',
    '/files.html',
    '/viewer.html',
    '/styles.css',
    '/logo.svg',
    '/logo-dark.svg',
    '/favicon.svg'
];

const EXTERNAL_ASSETS = [
    'https://cdnjs.cloudflare.com/ajax/libs/marked/12.0.0/marked.min.js',
    'https://cdnjs.cloudflare.com/ajax/libs/mermaid/11.0.2/mermaid.min.js',
    'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.1/highlight.min.js',
    'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.1/styles/github.min.css',
    'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.1/styles/github-dark.min.css',
    'https://cdnjs.cloudflare.com/ajax/libs/fuse.js/7.0.0/fuse.min.js',
    'https://cdnjs.cloudflare.com/ajax/libs/vis-network/9.1.9/dist/vis-network.min.js',
    'https://cdnjs.cloudflare.com/ajax/libs/vis-network/9.1.9/dist/vis-network.min.css'
];

// Install event - cache static assets
self.addEventListener('install', (event) => {
    console.log('[Service Worker] Installing...');
    
    event.waitUntil(
        Promise.all([
            // Cache static assets
            caches.open(CACHE_NAME).then((cache) => {
                console.log('[Service Worker] Caching static assets');
                return cache.addAll(STATIC_ASSETS);
            }),
            // Pre-cache external CDN assets
            caches.open(CACHE_NAME + '-external').then((cache) => {
                console.log('[Service Worker] Caching external assets');
                return cache.addAll(EXTERNAL_ASSETS);
            })
        ]).then(() => {
            console.log('[Service Worker] Installation complete');
            return self.skipWaiting();
        })
    );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
    console.log('[Service Worker] Activating...');
    
    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames
                    .filter((cacheName) => {
                        return cacheName !== CACHE_NAME && 
                               cacheName !== CACHE_NAME + '-external';
                    })
                    .map((cacheName) => {
                        console.log('[Service Worker] Deleting old cache:', cacheName);
                        return caches.delete(cacheName);
                    })
            );
        }).then(() => {
            console.log('[Service Worker] Activation complete');
            return self.clients.claim();
        })
    );
});

// Fetch event - serve from cache, fallback to network
self.addEventListener('fetch', (event) => {
    const { request } = event;
    const url = new URL(request.url);

    // Skip non-GET requests
    if (request.method !== 'GET') {
        return;
    }

    // Skip API requests (they should go directly to server)
    if (url.pathname.startsWith('/api/')) {
        return;
    }

    // For navigation requests (page loads), try network first, then cache
    if (request.mode === 'navigate') {
        event.respondWith(
            fetch(request)
                .then((response) => {
                    // Cache the response for future use
                    const responseClone = response.clone();
                    caches.open(CACHE_NAME).then((cache) => {
                        cache.put(request, responseClone);
                    });
                    return response;
                })
                .catch(() => {
                    return caches.match(request);
                })
        );
        return;
    }

    // For static assets, try cache first, then network
    event.respondWith(
        caches.match(request).then((cachedResponse) => {
            if (cachedResponse) {
                // Return cached response and update cache in background
                event.waitUntil(
                    fetch(request).then((response) => {
                        if (response.ok) {
                            caches.open(CACHE_NAME).then((cache) => {
                                cache.put(request, response);
                            });
                        }
                    }).catch(() => {})
                );
                return cachedResponse;
            }

            // Not in cache, fetch from network
            return fetch(request)
                .then((response) => {
                    if (!response.ok) {
                        return response;
                    }

                    // Cache the response
                    const responseClone = response.clone();
                    caches.open(CACHE_NAME).then((cache) => {
                        cache.put(request, responseClone);
                    });

                    return response;
                })
                .catch(() => {
                    // Return offline fallback if available
                    if (request.destination === 'document') {
                        return caches.match('/index.html');
                    }
                    return new Response('Offline', { status: 503, statusText: 'Service Unavailable' });
                });
        })
    );
});

// Handle messages from the main thread
self.addEventListener('message', (event) => {
    if (event.data && event.data.type === 'SKIP_WAITING') {
        self.skipWaiting();
    }

    if (event.data && event.data.type === 'CLEAR_CACHE') {
        caches.keys().then((cacheNames) => {
            cacheNames.forEach((cacheName) => {
                caches.delete(cacheName);
            });
        });
    }
});