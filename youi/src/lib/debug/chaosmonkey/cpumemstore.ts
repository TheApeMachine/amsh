import { Log } from './types';
import { pickOne } from './utils';

export const setupCPUChaos = ({
    logChaos,
    shouldCreateChaos
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    const createCPUSpike = (duration: number) => {
        logChaos({
            type: 'cpu.spike',
            description: `Creating CPU spike for ${duration}ms`,
            duration,
            impact: 'medium',
            recoverable: true
        });

        // Create CPU load using Web Workers
        const workerCode = `
            onmessage = function() {
                while (true) {
                    // Intensive calculation
                    Math.random() * Math.random();
                }
            };
        `;

        const workerCount = navigator.hardwareConcurrency || 4;
        const workers: Worker[] = [];

        for (let i = 0; i < workerCount; i++) {
            const blob = new Blob([workerCode], { type: 'application/javascript' });
            const worker = new Worker(URL.createObjectURL(blob));
            workers.push(worker);
            worker.postMessage('start');
        }

        setTimeout(() => {
            workers.forEach(worker => worker.terminate());
        }, duration);
    };

    setInterval(() => {
        if (shouldCreateChaos('cpu')) {
            createCPUSpike(Math.random() * 2000 + 1000); // 1-3 second spike
        }
    }, 10000); // Check every 10 seconds
};

export const setupMemoryChaos = ({
    logChaos,
    shouldCreateChaos,
    config
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    let memoryLeaks: any[] = [];

    const createMemoryLeak = () => {
        const size = Math.floor(Math.random() * 1000000) + 500000; // 0.5-1.5MB
        const leak = new Array(size).fill('🐒');

        logChaos({
            type: 'memory.leak',
            description: `Leaking ${(size / 1024 / 1024).toFixed(2)}MB of memory`,
            duration: 0,
            impact: 'high',
            recoverable: true
        });

        if (!config.safeMode) {
            memoryLeaks.push(leak); // Actual leak in unsafe mode
        } else {
            setTimeout(() => {
                const index = memoryLeaks.indexOf(leak);
                if (index > -1) {
                    memoryLeaks.splice(index, 1);
                }
            }, 5000); // Cleanup after 5 seconds in safe mode
        }
    };

    const createGarbageCollection = () => {
        logChaos({
            type: 'memory.gc',
            description: 'Forcing garbage collection pressure',
            duration: 2000,
            impact: 'low',
            recoverable: true
        });

        // Create and destroy lots of objects
        for (let i = 0; i < 1000000; i++) {
            const obj = {
                id: i,
                data: new Array(100).fill(Math.random())
            };
            obj.data = [];
        }
    };

    setInterval(() => {
        if (shouldCreateChaos('memory')) {
            pickOne([createMemoryLeak, createGarbageCollection])();
        }
    }, 15000); // Check every 15 seconds

    return {
        cleanup: () => {
            memoryLeaks = [];
        }
    };
};

export const setupStorageChaos = ({
    logChaos,
    shouldCreateChaos,
    config
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    const originalSetItem = localStorage.setItem;
    const originalGetItem = localStorage.getItem;
    const storageChaos = new Map<string, number>();

    // Helper to corrupt JSON data
    const corruptJson = (data: any): any => {
        if (typeof data !== 'object' || data === null) return data;

        const copy = Array.isArray(data) ? [...data] : { ...data };
        const keys = Object.keys(copy);

        if (keys.length === 0) return copy;

        // Corrupt a random key
        const keyToCorrupt = keys[Math.floor(Math.random() * keys.length)];
        const originalValue = copy[keyToCorrupt];

        // Corrupt based on type
        if (typeof originalValue === 'string') {
            copy[keyToCorrupt] = originalValue.split('').reverse().join('');
        } else if (typeof originalValue === 'number') {
            copy[keyToCorrupt] = originalValue + (Math.random() * 10 - 5);
        } else if (typeof originalValue === 'boolean') {
            copy[keyToCorrupt] = !originalValue;
        }

        return copy;
    };

    // Corrupt storage data
    localStorage.setItem = function (key: string, value: string) {
        if (shouldCreateChaos('storage')) {
            const chaosType = pickOne(['corrupt', 'delay', 'error']);

            switch (chaosType) {
                case 'corrupt':
                    value = corruptStorageValue(value);
                    logChaos({
                        type: 'storage.corruption',
                        description: `Corrupted localStorage value for ${key}`,
                        duration: 0,
                        impact: 'medium',
                        recoverable: true
                    });
                    break;

                case 'delay': {
                    const delay = Math.random() * 1000;
                    storageChaos.set(key, Date.now() + delay);
                    logChaos({
                        type: 'storage.delay',
                        description: `Added ${delay.toFixed(0)}ms delay to storage write`,
                        duration: delay,
                        impact: 'low',
                        recoverable: true
                    });
                    break;
                }

                case 'error':
                    if (!config.safeMode) {
                        logChaos({
                            type: 'storage.error',
                            description: `Blocked storage write to ${key}`,
                            duration: 0,
                            impact: 'high',
                            recoverable: false
                        });
                        throw new Error('Chaos Monkey: Storage write failed');
                    }
                    break;
            }
        }

        originalSetItem.call(localStorage, key, value);
    };

    // Add chaos to storage reads
    localStorage.getItem = function (key: string) {
        const chaosDelay = storageChaos.get(key);
        if (chaosDelay && Date.now() < chaosDelay) {
            return null; // Simulate delay by returning null
        }

        const value = originalGetItem.call(localStorage, key);

        if (shouldCreateChaos('storage') && value) {
            if (Math.random() < 0.3) { // 30% chance of read error
                logChaos({
                    type: 'storage.read',
                    description: `Simulated read error for ${key}`,
                    duration: 0,
                    impact: 'medium',
                    recoverable: true
                });
                return null;
            }
        }

        return value;
    };

    // Helper to corrupt storage values
    const corruptStorageValue = (value: string): string => {
        try {
            // If it's JSON, corrupt the data structure
            const data = JSON.parse(value);
            return JSON.stringify(corruptJson(data));
        } catch {
            // If it's not JSON, corrupt the string
            return value.split('').reverse().join('');
        }
    };

    // Periodically clear chaos delays
    setInterval(() => {
        for (const [key, delay] of storageChaos.entries()) {
            if (Date.now() > delay) {
                storageChaos.delete(key);
            }
        }
    }, 5000);

    return {
        cleanup: () => {
            localStorage.setItem = originalSetItem;
            localStorage.getItem = originalGetItem;
            storageChaos.clear();
        }
    };
};
