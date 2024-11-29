import { Log } from './types';
import { pickOne } from './utils';

export const setupWorkerTerminationChaos = ({
    logChaos,
    shouldCreateChaos,
    config
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    const originalWorker = window.Worker;
    const activeWorkers = new Set<Worker>();

    window.Worker = class ChaosWorker extends Worker {
        private readonly originalTerminate: () => void;
        private readonly scriptURL: string | URL;
        private cpuUsage = 0;
        private memoryUsage = 0;
        private performanceInterval?: number;

        constructor(scriptURL: string | URL, options?: WorkerOptions) {
            super(scriptURL, options);
            this.scriptURL = scriptURL;
            activeWorkers.add(this);
            this.originalTerminate = this.terminate;
            this.terminate = this.chaosTerminate;
            this.setupPerformanceChaos();
        }

        private readonly chaosTerminate = () => {
            if (shouldCreateChaos('worker')) {
                const chaos = pickOne([
                    'delay',
                    'zombie',
                    'resurrect',
                    'cascade'
                ]);

                switch (chaos) {
                    case 'delay': {
                        setTimeout(() => {
                            this.originalTerminate();
                        }, Math.random() * 3000);

                        logChaos({
                            type: 'worker.termination.delay',
                            description: 'Delayed worker termination',
                            duration: 3000,
                            impact: 'medium',
                            recoverable: true
                        });
                        break;
                    }
                    case 'zombie':
                        // Pretend to terminate but keep running
                        this.postMessage({ type: 'FAKE_TERMINATE' });
                        setTimeout(() => {
                            this.originalTerminate();
                        }, 10000);

                        logChaos({
                            type: 'worker.termination.zombie',
                            description: 'Created zombie worker',
                            duration: 10000,
                            impact: 'high',
                            recoverable: true
                        });
                        break;

                    case 'resurrect':
                        this.originalTerminate();
                        setTimeout(() => {
                            const newWorker = new originalWorker(this.scriptURL);
                            activeWorkers.add(newWorker);
                        }, Math.random() * 1000);

                        logChaos({
                            type: 'worker.termination.resurrect',
                            description: 'Resurrected terminated worker',
                            duration: 1000,
                            impact: 'medium',
                            recoverable: true
                        });
                        break;

                    case 'cascade':
                        // Terminate all workers in sequence
                        let delay = 0;
                        activeWorkers.forEach(worker => {
                            setTimeout(() => {
                                worker.terminate();
                            }, delay);
                            delay += 500;
                        });

                        logChaos({
                            type: 'worker.termination.cascade',
                            description: 'Triggered cascade termination',
                            duration: delay,
                            impact: 'high',
                            recoverable: false
                        });
                        break;
                }
            } else {
                this.originalTerminate();
            }
            activeWorkers.delete(this);
        };

        private setupPerformanceChaos() {
            this.performanceInterval = window.setInterval(() => {
                if (shouldCreateChaos('worker')) {
                    const chaos = pickOne([
                        'cpu_spike',
                        'memory_leak',
                        'deadlock',
                        'message_flood'
                    ]);

                    switch (chaos) {
                        case 'cpu_spike':
                            this.simulateCPUSpike();
                            break;

                        case 'memory_leak':
                            this.simulateMemoryLeak();
                            break;

                        case 'deadlock':
                            this.simulateDeadlock();
                            break;

                        case 'message_flood':
                            this.simulateMessageFlood();
                            break;
                    }
                }
            }, 5000);

            // Fix addEventListener type issues
            const originalAddEventListener = this.addEventListener;
            this.addEventListener = function (
                type: string,
                listener: EventListenerOrEventListenerObject | null,
                options?: boolean | AddEventListenerOptions
            ): void {
                if (type === 'error') {
                    const wrappedListener = (event: Event) => {
                        if (this.performanceInterval) {
                            clearInterval(this.performanceInterval);
                        }
                        if (listener && typeof listener === 'function') {
                            listener(event);
                        }
                    };
                    originalAddEventListener.call(this, type, wrappedListener as EventListener, options);
                    return;
                }
                if (listener) {
                    originalAddEventListener.call(this, type, listener, options);
                }
            };
        }

        private simulateCPUSpike() {
            this.postMessage({
                type: 'CHAOS_CPU_SPIKE',
                duration: Math.random() * 5000
            });

            this.cpuUsage = 100;
            setTimeout(() => {
                this.cpuUsage = 0;
            }, 5000);

            logChaos({
                type: 'worker.performance.cpu',
                description: 'Triggered CPU spike in worker',
                duration: 5000,
                impact: 'medium',
                recoverable: true
            });
        }

        private simulateMemoryLeak() {
            const leak = {
                size: Math.floor(Math.random() * 1000000),
                duration: Math.random() * 10000
            };

            this.postMessage({
                type: 'CHAOS_MEMORY_LEAK',
                ...leak
            });

            this.memoryUsage += leak.size;
            setTimeout(() => {
                this.memoryUsage = Math.max(0, this.memoryUsage - leak.size);
            }, leak.duration);

            logChaos({
                type: 'worker.performance.memory',
                description: `Created ${(leak.size / 1024 / 1024).toFixed(2)}MB memory leak`,
                duration: leak.duration,
                impact: 'high',
                recoverable: true
            });
        }

        private simulateDeadlock() {
            if (config.safeMode) return;

            this.postMessage({
                type: 'CHAOS_DEADLOCK',
                duration: 5000
            });

            logChaos({
                type: 'worker.performance.deadlock',
                description: 'Simulated worker deadlock',
                duration: 5000,
                impact: 'high',
                recoverable: false
            });
        }

        private simulateMessageFlood() {
            const messageCount = Math.floor(Math.random() * 100) + 50;
            let sent = 0;

            const floodInterval = setInterval(() => {
                if (sent >= messageCount) {
                    clearInterval(floodInterval);
                    return;
                }

                this.postMessage({
                    type: 'CHAOS_MESSAGE',
                    id: sent++,
                    timestamp: Date.now()
                });
            }, 10);

            logChaos({
                type: 'worker.performance.flood',
                description: `Flooding worker with ${messageCount} messages`,
                duration: messageCount * 10,
                impact: 'medium',
                recoverable: true
            });
        }

        get currentCPUUsage() {
            return this.cpuUsage;
        }

        get currentMemoryUsage() {
            return this.memoryUsage;
        }
    };

    return {
        getActiveWorkers: () => activeWorkers,
        cleanup: () => {
            window.Worker = originalWorker;
            activeWorkers.forEach(worker => worker.terminate());
        }
    };
};
