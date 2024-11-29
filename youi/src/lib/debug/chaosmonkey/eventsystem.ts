import { Log } from './types';
import { pickOne } from './utils';

export const setupEventChaos = ({
    logChaos,
    shouldCreateChaos,
    config
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    const eventHistory = new Map<string, number>();
    const originalAddEventListener = EventTarget.prototype.addEventListener;
    const originalDispatchEvent = EventTarget.prototype.dispatchEvent;
    const originalPostMessage = window.postMessage;
    const originalWorker = window.Worker;
    const originalWorkerPostMessage = Worker.prototype.postMessage;
    const workerInstances = new Set<Worker>();
    const interceptedEvents = new Set<string>();

    // Helper to identify important events that shouldn't be dropped
    const isImportantEvent = (event: Event): boolean => {
        const criticalTypes = new Set([
            'beforeunload',
            'unload',
            'error',
            'submit',
            'reset'
        ]);

        return criticalTypes.has(event.type);
    };

    const createEventChaos = (event: Event): Event | Promise<Event> => {
        if (!shouldCreateChaos('event')) return event;

        const chaosType = pickOne([
            'delay',
            'duplicate',
            'modify',
            'drop',
            'bubble'
        ]);

        switch (chaosType) {
            case 'delay':
                return createDelayedEvent(event);
            case 'duplicate':
                return createDuplicateEvent(event);
            case 'modify':
                return modifyEvent(event);
            case 'drop':
                return dropEvent(event);
            case 'bubble':
                return modifyBubbling(event);
            default:
                return event;
        }
    };

    const createDelayedEvent = (event: Event): Promise<Event> => {
        const delay = Math.random() * 1000 + 500;
        const delayedEvent = new Event(event.type, {
            bubbles: event.bubbles,
            cancelable: event.cancelable,
            composed: event.composed
        });

        logChaos({
            type: 'event.delay',
            description: `Delayed ${event.type} event by ${delay.toFixed(0)}ms`,
            duration: delay,
            impact: 'medium',
            recoverable: true
        });

        return new Promise(resolve => {
            setTimeout(() => resolve(delayedEvent), delay);
        });
    };

    const createDuplicateEvent = (event: Event): Event => {
        const count = Math.floor(Math.random() * 3) + 2;

        logChaos({
            type: 'event.duplicate',
            description: `Duplicating ${event.type} event ${count} times`,
            duration: 0,
            impact: 'medium',
            recoverable: true
        });

        for (let i = 0; i < count; i++) {
            setTimeout(() => {
                const clone = new Event(event.type, {
                    bubbles: event.bubbles,
                    cancelable: event.cancelable
                });
                event.target?.dispatchEvent(clone);
            }, i * 100);
        }

        return event;
    };

    const modifyEvent = (event: Event): Event => {
        if (event instanceof MouseEvent) {
            const modified = new MouseEvent(event.type, {
                ...event,
                clientX: event.clientX + (Math.random() * 10 - 5),
                clientY: event.clientY + (Math.random() * 10 - 5)
            });

            logChaos({
                type: 'event.modify',
                description: `Modified ${event.type} coordinates`,
                duration: 0,
                impact: 'low',
                recoverable: true
            });

            return modified;
        }

        if (event instanceof KeyboardEvent) {
            // Randomly capitalize or modify key
            const modified = new KeyboardEvent(event.type, {
                ...event,
                key: Math.random() > 0.5 ? event.key.toUpperCase() :
                    String.fromCharCode(event.key.charCodeAt(0) + 1)
            });

            logChaos({
                type: 'event.modify',
                description: `Modified ${event.type} key`,
                duration: 0,
                impact: 'medium',
                recoverable: true
            });

            return modified;
        }

        return event;
    };

    const dropEvent = (event: Event): Event => {
        if (config.safeMode && isImportantEvent(event)) {
            return event;
        }

        logChaos({
            type: 'event.drop',
            description: `Dropped ${event.type} event`,
            duration: 0,
            impact: 'high',
            recoverable: false
        });

        // Prevent event from reaching listeners
        event.stopImmediatePropagation();
        event.preventDefault();
        return event;
    };

    const modifyBubbling = (event: Event): Event => {
        const shouldBubble = !event.bubbles;

        logChaos({
            type: 'event.bubble',
            description: `${shouldBubble ? 'Enabled' : 'Disabled'} bubbling for ${event.type}`,
            duration: 0,
            impact: 'medium',
            recoverable: true
        });

        return new Event(event.type, {
            ...event,
            bubbles: shouldBubble
        });
    };

    // Intercept addEventListener
    EventTarget.prototype.addEventListener = function (
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | AddEventListenerOptions
    ) {
        interceptedEvents.add(type);
        const wrappedListener = async function (this: any, event: Event) {
            if (event instanceof Event) {
                const chaosEvent = await createEventChaos(event);
                if (chaosEvent) {
                    if (typeof listener === 'function') {
                        listener.call(this, chaosEvent);
                    } else {
                        listener.handleEvent.call(this, chaosEvent);
                    }
                }
            }
        };

        return originalAddEventListener.call(
            this,
            type,
            wrappedListener,
            options
        );
    };

    // Fix postMessage override with proper type handling
    const chaosPostMessage = function (
        message: any,
        targetOriginOrOptions: string | WindowPostMessageOptions
    ): void {
        if (shouldCreateChaos('event')) {
            const chaos = pickOne(['delay', 'modify', 'drop']);

            switch (chaos) {
                case 'delay':
                    setTimeout(() => {
                        if (typeof targetOriginOrOptions === 'string') {
                            originalPostMessage.call(window, message, { targetOrigin: targetOriginOrOptions });
                        } else {
                            originalPostMessage.call(window, message, targetOriginOrOptions);
                        }
                    }, Math.random() * 1000);
                    return;

                case 'modify':
                    if (typeof message === 'object') {
                        message = JSON.parse(JSON.stringify({
                            ...message,
                            chaosModified: true
                        }));
                    }
                    break;

                case 'drop':
                    if (!config.safeMode) {
                        return;
                    }
                    break;
            }
        }

        // Convert string targetOrigin to WindowPostMessageOptions
        if (typeof targetOriginOrOptions === 'string') {
            originalPostMessage.call(window, message, { targetOrigin: targetOriginOrOptions });
        } else {
            originalPostMessage.call(window, message, targetOriginOrOptions);
        }
    };

    window.postMessage = chaosPostMessage as typeof window.postMessage;

    // Fix Worker override with simplified message handling
    class ChaosWorker extends Worker {
        constructor(scriptURL: string | URL, options?: WorkerOptions) {
            super(scriptURL, options);
            workerInstances.add(this);

            const originalPostMessage = this.postMessage.bind(this);

            this.postMessage = function (
                message: any,
                transferOrOptions?: Transferable[] | StructuredSerializeOptions
            ): void {
                if (shouldCreateChaos('event')) {
                    const chaos = pickOne(['delay', 'modify', 'drop']);

                    switch (chaos) {
                        case 'delay':
                            setTimeout(() => {
                                if (Array.isArray(transferOrOptions)) {
                                    originalPostMessage(message, transferOrOptions);
                                } else {
                                    originalPostMessage(message, transferOrOptions as StructuredSerializeOptions);
                                }
                            }, Math.random() * 1000);
                            return;

                        case 'modify':
                            if (typeof message === 'object') {
                                message = JSON.parse(JSON.stringify({
                                    ...message,
                                    chaosModified: true
                                }));
                            }
                            break;

                        case 'drop':
                            if (!config.safeMode) {
                                return;
                            }
                            break;
                    }
                }

                if (Array.isArray(transferOrOptions)) {
                    originalPostMessage(message, transferOrOptions);
                } else {
                    originalPostMessage(message, transferOrOptions as StructuredSerializeOptions);
                }
            };
        }
    }

    window.Worker = ChaosWorker;

    return {
        getInterceptedEvents: () => Array.from(interceptedEvents),
        getEventFrequency: () => Object.fromEntries(eventHistory),
        cleanup: () => {
            EventTarget.prototype.addEventListener = originalAddEventListener;
            EventTarget.prototype.dispatchEvent = originalDispatchEvent;
            window.postMessage = originalPostMessage;
            window.Worker = originalWorker;
            workerInstances.forEach(worker => {
                worker.postMessage = originalWorkerPostMessage;
            });
        }
    };
};
