/**
 * htmx-centrifugo - htmx extension for Centrifugo real-time messaging
 *
 * This extension enables htmx applications to receive real-time updates
 * from a Centrifugo server using WebSocket, SSE, or HTTP-streaming transports.
 *
 * @version 0.1.0
 * @license MIT
 */

(function() {
    // Use global Centrifuge from CDN
    const Centrifuge = window.Centrifuge;

    if (!Centrifuge) {
        console.error('htmx-centrifugo: Centrifuge library not found. Please load it before this extension.');
        return;
    }

    /** @type {Object} API reference from htmx.defineExtension */
    let api;

    /** @type {Map<Element, Object>} */
    const centrifugeInstances = new Map();

    /** @type {Map<Element, Set<Object>>} */
    const subscriptions = new Map();

    /**
     * Check if element is still in the document body
     * @param {Element} element
     * @returns {boolean}
     */
    function bodyContains(element) {
        return api ? api.bodyContains(element) : document.body.contains(element);
    }

    /**
     * Disconnect and remove Centrifuge instance for an element
     * @param {Element} element
     */
    function disconnectCentrifugeInstance(element) {
        if (centrifugeInstances.has(element)) {
            const centrifuge = centrifugeInstances.get(element);
            centrifuge.disconnect();
            centrifugeInstances.delete(element);
        }
        if (subscriptions.has(element)) {
            subscriptions.delete(element);
        }
    }

    /**
     * Get or create Centrifuge instance for an element
     * @param {Element} element
     * @param {boolean} forceRecreate - Force recreation of instance
     * @returns {Centrifuge|null}
     */
    function getCentrifugeInstance(element, forceRecreate = false) {
        // Check if instance already exists and we're not forcing recreation
        if (!forceRecreate && centrifugeInstances.has(element)) {
            return centrifugeInstances.get(element);
        }

        // Disconnect existing instance if recreating
        if (forceRecreate) {
            disconnectCentrifugeInstance(element);
        }

        // Build transports array based on configuration
        const transports = [];

        // Detect protocol
        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const httpProtocol = window.location.protocol;
        const host = window.location.host;

        // Check for explicit transport configuration
        const wsEndpoint = element.getAttribute('centrifugo-ws-endpoint');
        const httpStreamEndpoint = element.getAttribute('centrifugo-http-stream-endpoint');
        const sseEndpoint = element.getAttribute('centrifugo-sse-endpoint');

        if (wsEndpoint) {
            transports.push({
                transport: 'websocket',
                endpoint: wsEndpoint.startsWith('ws://') || wsEndpoint.startsWith('wss://')
                    ? wsEndpoint
                    : `${wsProtocol}//${host}${wsEndpoint}`
            });
        }

        if (httpStreamEndpoint) {
            transports.push({
                transport: 'http_stream',
                endpoint: httpStreamEndpoint.startsWith('http')
                    ? httpStreamEndpoint
                    : `${httpProtocol}//${host}${httpStreamEndpoint}`
            });
        }

        if (sseEndpoint) {
            transports.push({
                transport: 'sse',
                endpoint: sseEndpoint.startsWith('http')
                    ? sseEndpoint
                    : `${httpProtocol}//${host}${sseEndpoint}`
            });
        }

        // If no transports configured, use defaults
        if (transports.length === 0) {
            transports.push(
                {
                    transport: 'websocket',
                    endpoint: `${wsProtocol}//${host}/connection/websocket`
                },
                {
                    transport: 'http_stream',
                    endpoint: `${httpProtocol}//${host}/connection/http_stream`
                },
                {
                    transport: 'sse',
                    endpoint: `${httpProtocol}//${host}/connection/sse`
                }
            );
        }

        // Parse configuration from attributes
        const config = {
            debug: element.getAttribute('centrifugo-debug') === 'true'
        };

        // Token configuration - use getToken function for dynamic token support
        const tokenUrl = element.getAttribute('centrifugo-token-url');
        if (tokenUrl) {
            config.getToken = async function() {
                const response = await fetch(tokenUrl, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' }
                });
                const data = await response.json();
                return data.token;
            };
        } else {
            // Always use getToken function to support dynamic token updates
            config.getToken = function() {
                const token = element.getAttribute('centrifugo-token');
                if (token) {
                    return Promise.resolve(token);
                }
                return Promise.resolve('');
            };
        }

        // Optional: init endpoint workaround for HTTP/2
        const initEnabled = element.getAttribute('centrifugo-init') === 'true';
        if (initEnabled) {
            config.getData = function() {
                return fetch(`${httpProtocol}//${host}/connection/init`, {method: 'GET'}).then(function() {
                    return null;
                });
            };
        }

        // Create Centrifuge instance with transports array
        const centrifuge = new Centrifuge(transports, config);

        // Set up event handlers following htmx naming convention
        centrifuge.on('connecting', (ctx) => {
            element.dispatchEvent(new CustomEvent('htmx:centrifugo-connecting', {
                detail: ctx,
                bubbles: true
            }));
        });

        centrifuge.on('connected', (ctx) => {
            element.dispatchEvent(new CustomEvent('htmx:centrifugo-connected', {
                detail: ctx,
                bubbles: true
            }));
        });

        centrifuge.on('disconnected', (ctx) => {
            element.dispatchEvent(new CustomEvent('htmx:centrifugo-disconnected', {
                detail: ctx,
                bubbles: true
            }));
        });

        centrifuge.on('error', (ctx) => {
            element.dispatchEvent(new CustomEvent('htmx:centrifugo-error', {
                detail: ctx,
                bubbles: true
            }));
        });

        // Store instance and connect
        centrifugeInstances.set(element, centrifuge);

        // Only auto-connect if centrifugo-connect attribute is present
        // This allows delayed connection after token is set
        if (element.hasAttribute('centrifugo-connect')) {
            centrifuge.connect();
        }

        return centrifuge;
    }

    /**
     * Subscribe to a channel
     * @param {Element} element
     * @param {Centrifuge} centrifuge
     */
    function subscribeToChannel(element, centrifuge) {
        const channel = element.getAttribute('centrifugo-subscribe');
        if (!channel) {
            return;
        }

        // Check if already subscribed (prevent duplicate subscriptions)
        if (element._centrifugoSubscribed) {
            return;
        }
        element._centrifugoSubscribed = true;

        // Get swap strategy
        const swapStrategy = element.getAttribute('centrifugo-swap') || 'innerHTML';
        const target = element.getAttribute('centrifugo-target') || null;
        const targetElement = target ? document.querySelector(target) : element;

        if (!targetElement) {
            console.error('htmx-centrifugo: target element not found:', target);
            return;
        }

        // Subscription token configuration
        const subConfig = {};
        const subTokenUrl = element.getAttribute('centrifugo-sub-token-url');
        if (subTokenUrl) {
            subConfig.getToken = async function() {
                const response = await fetch(subTokenUrl, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ channel: channel })
                });
                const data = await response.json();
                return data.token;
            };
        }

        const subscription = centrifuge.newSubscription(channel, subConfig);

        // Handle publications
        subscription.on('publication', (ctx) => {
            // Fire before-message event (can be cancelled)
            const content = ctx.data.html || ctx.data;

            if (!api.triggerEvent(element, 'htmx:centrifugo-before-message', {
                message: content,
                channel: channel,
                data: ctx.data
            })) {
                return; // Event was cancelled
            }

            // Update DOM based on swap strategy
            if (typeof content === 'string') {
                switch (swapStrategy) {
                    case 'innerHTML':
                        targetElement.innerHTML = content;
                        break;
                    case 'outerHTML':
                        targetElement.outerHTML = content;
                        break;
                    case 'beforebegin':
                        targetElement.insertAdjacentHTML('beforebegin', content);
                        break;
                    case 'afterbegin':
                        targetElement.insertAdjacentHTML('afterbegin', content);
                        break;
                    case 'beforeend':
                        targetElement.insertAdjacentHTML('beforeend', content);
                        break;
                    case 'afterend':
                        targetElement.insertAdjacentHTML('afterend', content);
                        break;
                    case 'delete':
                        targetElement.remove();
                        break;
                    case 'none':
                        // Do nothing, just fire event
                        break;
                    default:
                        targetElement.innerHTML = content;
                }

                // Process htmx attributes in new content
                if (typeof htmx !== 'undefined') {
                    htmx.process(targetElement);
                }

                // Fire htmx after-settle event for hx-on handlers
                targetElement.dispatchEvent(new CustomEvent('htmx:after-settle', {
                    detail: { target: targetElement },
                    bubbles: true
                }));

                // Fire after-message event
                api.triggerEvent(element, 'htmx:centrifugo-after-message', {
                    message: content,
                    channel: channel,
                    data: ctx.data
                });
            }
        });

        subscription.on('subscribing', (ctx) => {
            api.triggerEvent(element, 'htmx:centrifugo-subscribing', {
                channel: channel,
                context: ctx
            });
        });

        subscription.on('subscribed', (ctx) => {
            api.triggerEvent(element, 'htmx:centrifugo-subscribed', {
                channel: channel,
                context: ctx
            });
        });

        subscription.on('unsubscribed', (ctx) => {
            api.triggerEvent(element, 'htmx:centrifugo-unsubscribed', {
                channel: channel,
                context: ctx
            });
        });

        subscription.on('error', (ctx) => {
            api.triggerEvent(element, 'htmx:centrifugo-subscription-error', {
                channel: channel,
                error: ctx
            });
        });

        subscription.subscribe();

        // Store subscription
        if (!subscriptions.has(element)) {
            subscriptions.set(element, new Set());
        }
        subscriptions.get(element).add(subscription);
    }

    /**
     * Handle sending messages through Centrifugo
     * @param {Element} element
     * @param {Centrifuge} centrifuge
     */
    function setupSend(element, centrifuge) {
        if (!element.hasAttribute('centrifugo-send')) {
            return;
        }

        // Check if already set up to avoid duplicate listeners
        if (element._centrifugoSendSetup) {
            return;
        }
        element._centrifugoSendSetup = true;

        const channel = element.getAttribute('centrifugo-channel');
        const method = element.getAttribute('centrifugo-method') || 'publish';

        element.addEventListener('submit', async (evt) => {
            evt.preventDefault();

            const formData = new FormData(element);
            const data = {};
            formData.forEach((value, key) => {
                data[key] = value;
            });

            try {
                // Fire before-send event (can be cancelled)
                if (!api.triggerEvent(element, 'htmx:centrifugo-before-send', {
                    data: data,
                    method: method
                })) {
                    return; // Event was cancelled
                }

                if (method === 'publish' && channel) {
                    // Publish to channel (requires server-side permission)
                    await centrifuge.publish(channel, data);
                } else if (method === 'rpc') {
                    // Send RPC call
                    const rpcMethod = element.getAttribute('centrifugo-rpc-method') || 'default';
                    const result = await centrifuge.rpc(rpcMethod, data);

                    api.triggerEvent(element, 'htmx:centrifugo-rpc-result', {
                        method: rpcMethod,
                        result: result
                    });
                }

                api.triggerEvent(element, 'htmx:centrifugo-after-send', {
                    data: data
                });

                // Reset form
                element.reset();
            } catch (error) {
                api.triggerEvent(element, 'htmx:centrifugo-send-error', {
                    error: error
                });
                console.error('htmx-centrifugo: send error:', error);
            }
        });
    }

    /**
     * Cleanup function
     * @param {Element} element
     */
    function cleanup(element) {
        // Clear subscription flag
        delete element._centrifugoSubscribed;
        delete element._centrifugoSendSetup;

        // Unsubscribe from all channels
        if (subscriptions.has(element)) {
            subscriptions.get(element).forEach(sub => {
                sub.unsubscribe();
            });
            subscriptions.delete(element);
        }

        // Disconnect and remove Centrifuge instance
        if (centrifugeInstances.has(element)) {
            const centrifuge = centrifugeInstances.get(element);
            centrifuge.disconnect();
            centrifugeInstances.delete(element);
        }
    }

    // Public API for external usage
    window.htmxCentrifugo = {
        /**
         * Connect a Centrifuge instance for an element
         * @param {Element} element - Element with centrifugo-connect attribute
         */
        connect: function(element) {
            const centrifuge = getCentrifugeInstance(element);
            if (centrifuge) {
                // Set up subscriptions for child elements
                const subscribers = element.querySelectorAll('[centrifugo-subscribe]');
                subscribers.forEach(sub => subscribeToChannel(sub, centrifuge));

                // Also check if the element itself subscribes
                if (element.hasAttribute('centrifugo-subscribe')) {
                    subscribeToChannel(element, centrifuge);
                }

                // Set up send handlers
                const senders = element.querySelectorAll('[centrifugo-send]');
                senders.forEach(sender => setupSend(sender, centrifuge));

                if (element.hasAttribute('centrifugo-send')) {
                    setupSend(element, centrifuge);
                }

                // Connect if not already connected
                if (centrifuge.state !== 'connected' && centrifuge.state !== 'connecting') {
                    centrifuge.connect();
                }
            }
        },

        /**
         * Disconnect a Centrifuge instance for an element
         * @param {Element} element - Element with centrifugo-connect attribute
         */
        disconnect: function(element) {
            disconnectCentrifugeInstance(element);
        },

        /**
         * Reconnect (disconnect and create new instance with updated token)
         * @param {Element} element - Element with centrifugo-connect attribute
         */
        reconnect: function(element) {
            const centrifuge = getCentrifugeInstance(element, true); // Force recreate
            if (centrifuge) {
                // Set up subscriptions and handlers
                const subscribers = element.querySelectorAll('[centrifugo-subscribe]');
                subscribers.forEach(sub => subscribeToChannel(sub, centrifuge));

                if (element.hasAttribute('centrifugo-subscribe')) {
                    subscribeToChannel(element, centrifuge);
                }

                const senders = element.querySelectorAll('[centrifugo-send]');
                senders.forEach(sender => setupSend(sender, centrifuge));

                if (element.hasAttribute('centrifugo-send')) {
                    setupSend(element, centrifuge);
                }

                centrifuge.connect();
            }
        }
    };

    // Define the htmx extension
    if (typeof htmx !== 'undefined') {
        htmx.defineExtension('centrifugo', {
            init: function(apiRef) {
                // Store API reference for later use
                api = apiRef;
            },
            onEvent: function(name, evt) {
                const element = evt.detail.elt;

                if (name === 'htmx:beforeCleanupElement') {
                    // Element being removed, cleanup
                    cleanup(element);
                    return;
                }

                if (name === 'htmx:afterProcessNode') {
                    // Element was added to DOM, set up Centrifugo if needed
                    if (!bodyContains(element)) {
                        return;
                    }

                    if (element.hasAttribute('centrifugo-connect')) {
                        const centrifuge = getCentrifugeInstance(element);
                        if (centrifuge) {
                            // Set up subscriptions for child elements
                            const subscribers = element.querySelectorAll('[centrifugo-subscribe]');
                            subscribers.forEach(sub => subscribeToChannel(sub, centrifuge));

                            // Also check if the element itself subscribes
                            if (element.hasAttribute('centrifugo-subscribe')) {
                                subscribeToChannel(element, centrifuge);
                            }

                            // Set up send handlers
                            const senders = element.querySelectorAll('[centrifugo-send]');
                            senders.forEach(sender => setupSend(sender, centrifuge));

                            if (element.hasAttribute('centrifugo-send')) {
                                setupSend(element, centrifuge);
                            }
                        }
                    } else if (element.hasAttribute('centrifugo-subscribe') || element.hasAttribute('centrifugo-send')) {
                        // Find nearest parent with centrifugo-connect
                        const parent = element.closest('[centrifugo-connect]');
                        if (parent) {
                            const centrifuge = getCentrifugeInstance(parent);
                            if (centrifuge) {
                                if (element.hasAttribute('centrifugo-subscribe')) {
                                    subscribeToChannel(element, centrifuge);
                                }
                                if (element.hasAttribute('centrifugo-send')) {
                                    setupSend(element, centrifuge);
                                }
                            }
                        }
                    }
                }
            }
        });
    }

    // Export for module systems
    if (typeof module !== 'undefined' && module.exports) {
        module.exports = {
            getCentrifugeInstance,
            cleanup
        };
    }
})();
