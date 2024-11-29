import { Log } from './types';
import { pickOne } from './utils';

interface StorageOperation {
    key: string;
    value: string;
}

export const setupStorageChaos = ({
    logChaos,
    shouldCreateChaos
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    const originalSetItem = localStorage.setItem.bind(localStorage);
    const originalGetItem = localStorage.getItem.bind(localStorage);
    const storageAccessLog = new Map<string, number>();
    const modifiedKeys = new Set<string>();

    // Helper functions to reduce nesting and complexity
    const handleStorageCorruption = (value: string): string => {
        try {
            const parsed = JSON.parse(value);
            if (typeof parsed === 'object') {
                return JSON.stringify({
                    ...parsed,
                    chaosModified: true,
                    timestamp: Date.now()
                });
            }
        } catch {
            return value.split('').reverse().join('');
        }
        return value;
    };

    const handleStorageDelay = (operation: StorageOperation): void => {
        const delay = Math.random() * 1000;
        setTimeout(() => {
            originalSetItem(operation.key, operation.value);
        }, delay);

        logChaos({
            type: 'storage.delay',
            description: `Added ${delay.toFixed(0)}ms delay to storage write`,
            duration: delay,
            impact: 'low',
            recoverable: true
        });
    };

    const handleStorageExpiration = (key: string, value: string): void => {
        const expiry = Date.now() + Math.random() * 10000;
        const expiringValue = JSON.stringify({
            value,
            expiry
        });
        originalSetItem(key, expiringValue);

        setTimeout(() => {
            originalSetItem(key, '');
        }, expiry - Date.now());
    };

    const handleStorageCrossover = (key: string, value: string): void => {
        sessionStorage.setItem(key, value);
        logChaos({
            type: 'storage.crossover',
            description: 'Cross-contaminated localStorage to sessionStorage',
            duration: 0,
            impact: 'medium',
            recoverable: true
        });
    };

    // Create storage proxy
    const createStorageProxy = (): ProxyHandler<Storage> => ({
        get(target: Storage, prop: string) {
            if (prop === 'setItem') {
                return (key: string, value: string) => {
                    if (shouldCreateChaos('storage')) {
                        const chaos = pickOne([
                            'corruption',
                            'delay',
                            'expiration',
                            'crossover'
                        ]);

                        switch (chaos) {
                            case 'corruption': {
                                const corruptedValue = handleStorageCorruption(value);
                                originalSetItem(key, corruptedValue);
                                modifiedKeys.add(key);
                                return;
                            }
                            case 'delay': {
                                handleStorageDelay({ key, value });
                                return;
                            }
                            case 'expiration': {
                                handleStorageExpiration(key, value);
                                return;
                            }
                            case 'crossover': {
                                handleStorageCrossover(key, value);
                                return;
                            }
                        }
                    }
                    originalSetItem(key, value);
                };
            }

            if (prop === 'getItem') {
                return (key: string): string | null => {
                    const value = originalGetItem(key);

                    if (shouldCreateChaos('storage') && value) {
                        const chaos = pickOne([
                            'missing',
                            'incorrect',
                            'type'
                        ]);

                        switch (chaos) {
                            case 'missing': {
                                if (Math.random() < 0.2) {
                                    return null;
                                }
                                break;
                            }
                            case 'incorrect': {
                                if (modifiedKeys.has(key)) {
                                    return value.split('').reverse().join('');
                                }
                                break;
                            }
                            case 'type': {
                                try {
                                    const parsed = JSON.parse(value);
                                    if (typeof parsed === 'number') {
                                        return JSON.stringify(String(parsed));
                                    }
                                    if (typeof parsed === 'string') {
                                        return JSON.stringify(Number(parsed));
                                    }
                                } catch {
                                    return value;
                                }
                            }
                        }
                    }
                    return value;
                };
            }

            return Reflect.get(target, prop);
        }
    });

    // Apply proxies to storage objects
    Object.defineProperty(window, 'localStorage', {
        get: () => new Proxy(localStorage, createStorageProxy())
    });

    Object.defineProperty(window, 'sessionStorage', {
        get: () => new Proxy(sessionStorage, createStorageProxy())
    });

    return {
        getModifiedKeys: () => modifiedKeys,
        getAccessLog: () => storageAccessLog,
        cleanup: () => {
            Object.defineProperty(window, 'localStorage', {
                value: localStorage
            });
            Object.defineProperty(window, 'sessionStorage', {
                value: sessionStorage
            });
            modifiedKeys.clear();
            storageAccessLog.clear();
        }
    };
};
