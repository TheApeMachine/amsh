import { pickOne } from '../utils';

interface Log {
    type: string;
    description: string;
    duration: number;
    impact: 'low' | 'medium' | 'high';
    recoverable: boolean;
}

export const setupResourceTimingChaos = ({ logChaos, shouldCreateChaos, config }: { logChaos: (log: Log) => void, shouldCreateChaos: (type: string) => boolean, config: { safeMode: boolean } }) => {
    const originalGetEntries = window.performance.getEntries;
    const originalGetEntriesByType = window.performance.getEntriesByType;
    const originalGetEntriesByName = window.performance.getEntriesByName;
    const originalClearResourceTimings = window.performance.clearResourceTimings;
    const originalSetResourceTimingBufferSize = window.performance.setResourceTimingBufferSize;

    // Store original entries for restoration
    const entriesCache = new Map<string, PerformanceEntry[]>();

    const modifyTimingEntry = (entry: PerformanceEntry): PerformanceEntry => {
        if (entry instanceof PerformanceResourceTiming) {
            const chaos = pickOne([
                'latency',
                'dns',
                'connection',
                'cache',
                'compression'
            ]);

            const proxy = new Proxy(entry, {
                get(target, prop) {
                    const value = Reflect.get(target, prop);

                    switch (chaos) {
                        case 'latency':
                            if (prop === 'duration' || prop === 'responseEnd') {
                                return value + Math.random() * 1000;
                            }
                            break;

                        case 'dns':
                            if (prop === 'domainLookupEnd' || prop === 'domainLookupStart') {
                                return value + Math.random() * 200;
                            }
                            break;

                        case 'connection':
                            if (prop === 'connectEnd' || prop === 'connectStart') {
                                return value + Math.random() * 300;
                            }
                            break;

                        case 'cache':
                            if (prop === 'transferSize' || prop === 'encodedBodySize') {
                                return Math.random() < 0.3 ? 0 : value * (1 + Math.random());
                            }
                            break;

                        case 'compression':
                            if (prop === 'decodedBodySize') {
                                return value * (1 + Math.random() * 0.5);
                            }
                            break;
                    }

                    return value;
                }
            });

            logChaos({
                type: 'resourceTiming.modify',
                description: `Modified timing metrics: ${chaos}`,
                duration: 0,
                impact: 'low',
                recoverable: true
            });

            return proxy;
        }
        return entry;
    };

    // Override performance methods
    window.performance.getEntries = function () {
        const entries = originalGetEntries.call(this);
        if (shouldCreateChaos('resourceTiming')) {
            return entries.map(modifyTimingEntry);
        }
        return entries;
    };

    window.performance.getEntriesByType = function (type: string) {
        const entries = originalGetEntriesByType.call(this, type);
        if (type === 'resource' && shouldCreateChaos('resourceTiming')) {
            const chaos = pickOne(['modify', 'filter', 'duplicate', 'order']);

            switch (chaos) {
                case 'modify':
                    return entries.map(modifyTimingEntry);

                case 'filter':
                    return entries.filter(() => Math.random() > 0.2);

                case 'duplicate':
                    const duplicates = entries.map(entry => {
                        const clone = { ...entry };
                        Object.setPrototypeOf(clone, Object.getPrototypeOf(entry));
                        return modifyTimingEntry(clone as PerformanceEntry);
                    });
                    return [...entries, ...duplicates];

                case 'order':
                    return entries.sort(() => Math.random() - 0.5);
            }
        }
        return entries;
    };

    window.performance.getEntriesByName = function (name: string, type?: string) {
        const entries = originalGetEntriesByName.call(this, name, type);
        if ((!type || type === 'resource') && shouldCreateChaos('resourceTiming')) {
            const chaos = pickOne(['cache', 'split', 'merge']);

            switch (chaos) {
                case 'cache':
                    // Simulate cached vs non-cached responses
                    return entries.map(entry => {
                        if (entry instanceof PerformanceResourceTiming) {
                            const isCached = Math.random() < 0.5;
                            const proxy = new Proxy(entry, {
                                get(target, prop) {
                                    if (isCached && (
                                        prop === 'transferSize' ||
                                        prop === 'encodedBodySize'
                                    )) {
                                        return 0;
                                    }
                                    return Reflect.get(target, prop);
                                }
                            });
                            return proxy;
                        }
                        return entry;
                    });

                case 'split':
                    // Split entry into multiple partial loads
                    if (entries.length > 0) {
                        const original = entries[0];
                        if (original instanceof PerformanceResourceTiming) {
                            const parts = Array(3).fill(null).map((_, i) => {
                                const part = {
                                    ...original,
                                    name: `${original.name}#part${i}`,
                                    startTime: original.startTime + (i * 100),
                                    responseEnd: original.responseEnd + (i * 100)
                                };

                                Object.setPrototypeOf(part, Object.getPrototypeOf(original));
                                return part as PerformanceResourceTiming;
                            });
                            return parts;
                        }
                    }
                    break;

                case 'merge':
                    // Merge similar entries
                    const resourceName = name.split('/').pop() ?? '';
                    const similar = originalGetEntriesByType.call(this, 'resource')
                        .filter(entry => entry.name.includes(resourceName));
                    if (similar.length > 0) {
                        return [...entries, ...similar];
                    }
                    break;
            }
        }
        return entries;
    };

    // Monitor resource loading
    const observer = new PerformanceObserver((list) => {
        if (shouldCreateChaos('resourceTiming')) {
            list.getEntries().forEach(entry => {
                const chaos = pickOne(['delay', 'size', 'error']);

                switch (chaos) {
                    case 'delay':
                        // Simulate network delay
                        setTimeout(() => {
                            const modifiedEntry = modifyTimingEntry(entry);
                            entriesCache.set(entry.name, [modifiedEntry]);
                        }, Math.random() * 1000);
                        break;

                    case 'size':
                        // Modify reported sizes
                        if (entry instanceof PerformanceResourceTiming) {
                            const sizeMultiplier = 1 + Math.random();
                            Object.defineProperties(entry, {
                                transferSize: { value: entry.transferSize * sizeMultiplier },
                                encodedBodySize: { value: entry.encodedBodySize * sizeMultiplier },
                                decodedBodySize: { value: entry.decodedBodySize * sizeMultiplier }
                            });
                        }
                        break;

                    case 'error':
                        if (!config.safeMode && Math.random() < 0.1) {
                            // Simulate failed resource load
                            const errorEvent = new Event('error');
                            entry.name && document.dispatchEvent(errorEvent);
                        }
                        break;
                }
            });
        }
    });

    observer.observe({ entryTypes: ['resource'] });

    // Add buffer size chaos
    window.performance.setResourceTimingBufferSize = function (maxSize: number) {
        if (shouldCreateChaos('resourceTiming')) {
            const chaos = pickOne(['reduce', 'expand', 'fluctuate']);

            switch (chaos) {
                case 'reduce':
                    maxSize = Math.floor(maxSize * 0.5);
                    break;

                case 'expand':
                    maxSize = Math.floor(maxSize * 1.5);
                    break;

                case 'fluctuate':
                    setInterval(() => {
                        originalSetResourceTimingBufferSize.call(
                            this,
                            Math.floor(maxSize * (0.5 + Math.random()))
                        );
                    }, 5000);
                    return;
            }
        }
        return originalSetResourceTimingBufferSize.call(this, maxSize);
    };

    // Extract chaos handlers to reduce cognitive complexity
    const handlePartialClear = function (this: Performance) {
        const entries = originalGetEntriesByType.call(this, 'resource');
        entries.forEach(entry => {
            if (Math.random() < 0.3) {
                entriesCache.set(entry.name, [entry]);
            }
        });
    };

    const handleSelectiveClear = function (this: Performance) {
        const types = ['script', 'css', 'img', 'fetch'];
        const typeToKeep = pickOne(types);
        const filteredEntries = originalGetEntriesByType.call(this, 'resource')
            .filter(entry => entry instanceof PerformanceResourceTiming && entry.initiatorType === typeToKeep);
        originalClearResourceTimings.call(this);
        filteredEntries.forEach(entry => {
            entriesCache.set(entry.name, [entry]);
        });
    };

    // Add clear timing chaos
    window.performance.clearResourceTimings = function () {
        if (shouldCreateChaos('resourceTiming')) {
            const chaosType = pickOne(['partial', 'delayed', 'selective']);

            // Move declarations outside case blocks
            let timeoutId: number;

            switch (chaosType) {
                case 'partial':
                    handlePartialClear.call(this);
                    break;

                case 'delayed':
                    timeoutId = window.setTimeout(() => {
                        originalClearResourceTimings.call(this);
                    }, Math.random() * 2000);
                    return;

                case 'selective':
                    handleSelectiveClear.call(this);
                    break;
            }
        }
        return originalClearResourceTimings.call(this);
    };

    return {
        getEntriesCache: () => entriesCache,
        cleanup: () => {
            window.performance.getEntries = originalGetEntries;
            window.performance.getEntriesByType = originalGetEntriesByType;
            window.performance.getEntriesByName = originalGetEntriesByName;
            window.performance.clearResourceTimings = originalClearResourceTimings;
            window.performance.setResourceTimingBufferSize = originalSetResourceTimingBufferSize;
            observer.disconnect();
            entriesCache.clear();
        }
    };
};
