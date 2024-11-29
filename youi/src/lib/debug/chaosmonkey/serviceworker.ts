import { Log } from './types';
import { pickOne } from './utils';

export const setupWorkerChaos = ({
    logChaos,
    shouldCreateChaos,
    config
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    const originalServiceWorkerContainer = navigator.serviceWorker;
    const originalSharedWorker = window.SharedWorker;
    const registeredWorkers = new Set<ServiceWorkerRegistration>();
    const sharedWorkers = new Set<SharedWorker>();

    // ServiceWorker Chaos
    if ('serviceWorker' in navigator) {
        const createServiceWorkerChaos = () => {
            // Override register method
            navigator.serviceWorker.register = async (scriptURL: string | URL, options?: RegistrationOptions) => {
                const registration = await originalServiceWorkerContainer.register(scriptURL, options);
                registeredWorkers.add(registration);

                if (shouldCreateChaos('worker')) {
                    const chaos = pickOne([
                        'unregister',
                        'update',
                        'skipWaiting',
                        'claimClients',
                        'cacheChaos'
                    ]);

                    await handleServiceWorkerChaos(chaos, registration, scriptURL, options);
                }

                return registration;
            };

            // Override message handling
            navigator.serviceWorker.addEventListener('message', handleServiceWorkerMessage);
        };

        const handleServiceWorkerChaos = async (
            chaos: string,
            registration: ServiceWorkerRegistration,
            scriptURL: string | URL,
            options?: RegistrationOptions
        ) => {
            switch (chaos) {
                case 'unregister': {
                    setTimeout(async () => {
                        await registration.unregister();
                        setTimeout(() => {
                            originalServiceWorkerContainer.register(scriptURL, options);
                        }, Math.random() * 5000);

                        logChaos({
                            type: 'serviceWorker.unregister',
                            description: 'Temporarily unregistered ServiceWorker',
                            duration: 5000,
                            impact: 'high',
                            recoverable: true
                        });
                    }, Math.random() * 10000);
                    break;
                }

                case 'update': {
                    const interval = setInterval(async () => {
                        if (Math.random() < 0.2) {
                            await registration.update();
                            logChaos({
                                type: 'serviceWorker.update',
                                description: 'Forced ServiceWorker update',
                                duration: 0,
                                impact: 'medium',
                                recoverable: true
                            });
                        }
                    }, 30000);
                    break;
                }

                case 'skipWaiting': {
                    if (registration.waiting) {
                        registration.waiting.postMessage({ type: 'SKIP_WAITING' });
                        logChaos({
                            type: 'serviceWorker.skip',
                            description: 'Forced skipWaiting',
                            duration: 0,
                            impact: 'medium',
                            recoverable: true
                        });
                    }
                    break;
                }

                case 'claimClients': {
                    registration.active?.postMessage({ type: 'CLAIM_CLIENTS' });
                    logChaos({
                        type: 'serviceWorker.claim',
                        description: 'Forced clients.claim()',
                        duration: 0,
                        impact: 'medium',
                        recoverable: true
                    });
                    break;
                }

                case 'cacheChaos':
                    await createCacheChaos();
                    break;
            }
        };

        const handleServiceWorkerMessage = (event: MessageEvent) => {
            if (shouldCreateChaos('worker')) {
                const chaos = pickOne(['delay', 'modify', 'drop']);

                switch (chaos) {
                    case 'delay': {
                        event.stopImmediatePropagation();
                        setTimeout(() => {
                            navigator.serviceWorker.dispatchEvent(new MessageEvent('message', {
                                data: event.data,
                                source: event.source
                            }));
                        }, Math.random() * 2000);
                        break;
                    }

                    case 'modify': {
                        event.stopImmediatePropagation();
                        navigator.serviceWorker.dispatchEvent(new MessageEvent('message', {
                            data: typeof event.data === 'object' ?
                                { ...event.data, chaosModified: true } :
                                `${event.data} (chaos)`,
                            source: event.source
                        }));
                        break;
                    }

                    case 'drop':
                        if (!config.safeMode) {
                            event.stopImmediatePropagation();
                        }
                        break;
                }
            }
        };

        createServiceWorkerChaos();
    }

    // Cache API Chaos
    const createCacheChaos = async () => {
        const cacheNames = await caches.keys();

        for (const cacheName of cacheNames) {
            const cache = await caches.open(cacheName);
            const requests = await cache.keys();

            for (const request of requests) {
                if (Math.random() < 0.1) { // 10% chance per cached item
                    const chaos = pickOne(['expire', 'modify', 'duplicate']);

                    switch (chaos) {
                        case 'expire':
                            await cache.delete(request);
                            logChaos({
                                type: 'cache.expire',
                                description: `Expired cache entry for ${request.url}`,
                                duration: 0,
                                impact: 'medium',
                                recoverable: true
                            });
                            break;

                        case 'modify':
                            const response = await cache.match(request);
                            if (response) {
                                const modifiedResponse = new Response(
                                    `${await response.text()} <!-- chaos -->`,
                                    response
                                );
                                await cache.put(request, modifiedResponse);
                                logChaos({
                                    type: 'cache.modify',
                                    description: `Modified cached response for ${request.url}`,
                                    duration: 0,
                                    impact: 'medium',
                                    recoverable: true
                                });
                            }
                            break;

                        case 'duplicate':
                            const copies = Math.floor(Math.random() * 3) + 2;
                            for (let i = 0; i < copies; i++) {
                                const copyUrl = new URL(request.url);
                                copyUrl.searchParams.set('chaos', String(i));
                                await cache.add(copyUrl);
                            }
                            logChaos({
                                type: 'cache.duplicate',
                                description: `Created ${copies} cache variants for ${request.url}`,
                                duration: 0,
                                impact: 'low',
                                recoverable: true
                            });
                            break;
                    }
                }
            }
        }
    };

    // SharedWorker Chaos
    window.SharedWorker = class ChaosSharedWorker extends SharedWorker {
        constructor(scriptURL: string | URL, options?: string | WorkerOptions) {
            super(scriptURL, options);
            sharedWorkers.add(this);

            if (shouldCreateChaos('worker')) {
                const chaos = pickOne(['disconnect', 'delay', 'error']);

                switch (chaos) {
                    case 'disconnect':
                        setInterval(() => {
                            if (Math.random() < 0.1) {
                                this.port.close();
                                logChaos({
                                    type: 'sharedWorker.disconnect',
                                    description: 'Disconnected SharedWorker port',
                                    duration: 0,
                                    impact: 'high',
                                    recoverable: true
                                });
                            }
                        }, 30000);
                        break;

                    case 'delay':
                        const originalPostMessage = this.port.postMessage;
                        this.port.postMessage = function (message: any) {
                            setTimeout(() => {
                                originalPostMessage.call(this, message);
                            }, Math.random() * 1000);
                        };
                        break;

                    case 'error':
                        if (!config.safeMode) {
                            setTimeout(() => {
                                this.port.postMessage({ type: 'TRIGGER_ERROR' });
                            }, Math.random() * 5000);
                        }
                        break;
                }
            }

            // Track message frequency
            let messageCount = 0;
            const originalPostMessage = this.port.postMessage;
            this.port.postMessage = function (message: any) {
                messageCount++;
                if (messageCount > 100) { // More than 100 messages in 1 second
                    logChaos({
                        type: 'sharedWorker.flood',
                        description: 'High message frequency detected',
                        duration: 0,
                        impact: 'medium',
                        recoverable: true
                    });
                    messageCount = 0;
                }
                originalPostMessage.call(this, message);
            };

            setInterval(() => {
                messageCount = 0;
            }, 1000);
        }
    };

    return {
        getRegisteredWorkers: () => registeredWorkers,
        getSharedWorkers: () => sharedWorkers,
        cleanup: () => {
            window.SharedWorker = originalSharedWorker;
            sharedWorkers.forEach(worker => worker.port.close());
            registeredWorkers.forEach(reg => reg.unregister());
        }
    };
};
