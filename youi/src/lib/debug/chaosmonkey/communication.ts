import { Log } from './types';
import { pickOne } from './utils';

export const setupCommunicationChaos = ({
    logChaos,
    shouldCreateChaos,
    config
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    const originalBroadcastChannel = window.BroadcastChannel;
    const originalEventSource = window.EventSource;
    const channels = new Map<string, BroadcastChannel>();
    const eventSources = new Set<EventSource>();

    // Intercept BroadcastChannel
    window.BroadcastChannel = class ChaosBroadcastChannel extends BroadcastChannel {
        constructor(name: string) {
            super(name);
            channels.set(name, this);

            // Intercept postMessage
            const originalPost = this.postMessage.bind(this);
            this.postMessage = (message: any) => {
                if (shouldCreateChaos('communication')) {
                    const chaos = pickOne([
                        'delay', 'duplicate', 'modify', 'drop', 'broadcast'
                    ]);

                    switch (chaos) {
                        case 'delay': {
                            setTimeout(() => {
                                originalPost(message);
                            }, Math.random() * 2000);

                            logChaos({
                                type: 'broadcast.delay',
                                description: `Delayed broadcast on ${name}`,
                                duration: 2000,
                                impact: 'low',
                                recoverable: true
                            });
                            return;
                        }
                        case 'duplicate': {
                            const count = Math.floor(Math.random() * 3) + 2;
                            for (let i = 0; i < count; i++) {
                                setTimeout(() => {
                                    originalPost(message);
                                }, i * 500);
                            }

                            logChaos({
                                type: 'broadcast.duplicate',
                                description: `Duplicated broadcast ${count}x on ${name}`,
                                duration: count * 500,
                                impact: 'medium',
                                recoverable: true
                            });
                            return;
                        }

                        case 'modify':
                            if (typeof message === 'object') {
                                message = {
                                    ...message,
                                    chaosModified: true,
                                    timestamp: Date.now()
                                };
                            }
                            break;

                        case 'drop':
                            if (!config.safeMode) {
                                logChaos({
                                    type: 'broadcast.drop',
                                    description: `Dropped broadcast on ${name}`,
                                    duration: 0,
                                    impact: 'high',
                                    recoverable: false
                                });
                                return;
                            }
                            break;

                        case 'broadcast':
                            // Broadcast to random other channels
                            channels.forEach((channel, channelName) => {
                                if (channelName !== name && Math.random() > 0.5) {
                                    channel.postMessage(message);
                                }
                            });
                            break;
                    }
                }
                originalPost(message);
            };
        }
    };

    // Intercept EventSource
    window.EventSource = class ChaosEventSource extends EventSource {
        private readonly originalAddEventListener: typeof EventSource.prototype.addEventListener;

        constructor(url: string | URL, eventSourceInitDict?: EventSourceInit) {
            super(url, eventSourceInitDict);
            eventSources.add(this);

            this.originalAddEventListener = this.addEventListener;
            this.addEventListener = this.chaosEventListener as typeof EventSource.prototype.addEventListener;

            // Randomly close and reopen connection
            if (shouldCreateChaos('communication')) {
                setInterval(() => {
                    if (Math.random() < 0.1) {
                        this.close();
                        setTimeout(() => {
                            eventSources.delete(this);
                            const newSource = new ChaosEventSource(url, eventSourceInitDict);
                            eventSources.add(newSource);
                        }, Math.random() * 1000);

                        logChaos({
                            type: 'sse.reconnect',
                            description: `Forced SSE reconnection to ${url}`,
                            duration: 1000,
                            impact: 'medium',
                            recoverable: true
                        });
                    }
                }, 10000);
            }
        }

        private readonly chaosEventListener = (
            type: string,
            listener: EventListenerOrEventListenerObject,
            options?: boolean | AddEventListenerOptions
        ) => {
            const handleListener = (event: Event) => {
                if (typeof listener === 'function') {
                    listener(event);
                } else {
                    listener.handleEvent(event);
                }
            };

            const handleChaosEvent = (event: Event) => {
                type ChaosType = 'delay' | 'modify' | 'drop' | 'duplicate';
                const chaos = pickOne(['delay', 'modify', 'drop', 'duplicate'] as const);
                const chaosHandlers: Record<ChaosType, () => void> = {
                    delay: () => {
                        setTimeout(() => handleListener(event), Math.random() * 1000);
                        logChaos({
                            type: 'sse.delay',
                            description: `Delayed SSE event ${type}`,
                            duration: 1000,
                            impact: 'low',
                            recoverable: true
                        });
                    },
                    modify: () => {
                        if (event instanceof MessageEvent) {
                            const modifiedEvent = new MessageEvent(event.type, {
                                data: typeof event.data === 'object' ?
                                    { ...event.data, chaosModified: true } :
                                    `${event.data} (chaos)`,
                                origin: event.origin,
                                lastEventId: event.lastEventId,
                                ports: [...event.ports]
                            });
                            handleListener(modifiedEvent);
                        }
                    },
                    drop: () => {
                        if (!config.safeMode) {
                            logChaos({
                                type: 'sse.drop',
                                description: `Dropped SSE event ${type}`,
                                duration: 0,
                                impact: 'high',
                                recoverable: false
                            });
                        }
                    },
                    duplicate: () => {
                        const count = Math.floor(Math.random() * 2) + 2;
                        for (let i = 0; i < count; i++) {
                            setTimeout(() => handleListener(event), i * 200);
                        }
                        logChaos({
                            type: 'sse.duplicate',
                            description: `Duplicated SSE event ${type} ${count}x`,
                            duration: count * 200,
                            impact: 'medium',
                            recoverable: true
                        });
                    }
                };

                chaosHandlers[chaos]();
            };

            const chaosListener = (event: Event) => {
                if (!shouldCreateChaos('communication')) {
                    handleListener(event);
                    return;
                }
                handleChaosEvent(event);
            };

            return this.originalAddEventListener(type, chaosListener, options);
        };
    };

    // Create random cross-talk between SSE sources
    const createSSECrossTalk = () => {
        setInterval(() => {
            if (!shouldCreateChaos('communication')) return;

            const sources = Array.from(eventSources);
            if (sources.length < 2) return;

            const source1 = sources[Math.floor(Math.random() * sources.length)];
            const source2 = sources[Math.floor(Math.random() * sources.length)];

            if (source1 !== source2) {
                const event = new MessageEvent('message', {
                    data: 'Cross-talk from another SSE connection'
                });

                source1.dispatchEvent(event);

                logChaos({
                    type: 'sse.crosstalk',
                    description: 'Created SSE cross-talk',
                    duration: 0,
                    impact: 'medium',
                    recoverable: true
                });
            }
        }, 5000);
    };

    createSSECrossTalk();

    return {
        getChannels: () => channels,
        getEventSources: () => eventSources,
        cleanup: () => {
            window.BroadcastChannel = originalBroadcastChannel;
            window.EventSource = originalEventSource;
            eventSources.forEach(source => source.close());
            channels.forEach(channel => channel.close());
        }
    };
};
