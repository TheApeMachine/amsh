import { Log } from './types';
import { pickOne } from './utils';

// Define WebSQL types
interface SQLTransaction {
    executeSql(sql: string, ...args: any[]): void;
}

interface WebSQLDatabase {
    transaction(
        callback: (transaction: SQLTransaction) => void,
        errorCallback?: (error: Error) => void,
        successCallback?: () => void
    ): void;
}

interface OpenDatabaseFunction {
    (name: string, version: string, displayName: string, size: number): WebSQLDatabase;
}

export const setupDatabaseChaos = ({
    logChaos,
    shouldCreateChaos,
    config
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    const activeDatabases = new Set<IDBDatabase>();
    const originalIndexedDB = window.indexedDB;
    const sqlTransactions = new Map<number, SQLTransaction>();
    let transactionId = 0;
    let savedOpenDatabase: OpenDatabaseFunction | undefined;

    // Helper functions to reduce nesting
    const handleVersionMismatch = (dbName: string, dbVersion?: number) => {
        return originalIndexedDB.open(dbName, (dbVersion || 1) + 1);
    };

    const handleQuotaExceeded = () => {
        return Promise.reject(new Error('QuotaExceededError: Database quota exceeded'));
    };

    const handleReadonlyUpgrade = (transaction: IDBTransaction) => {
        if (transaction.mode === 'readonly') {
            Object.defineProperty(transaction, 'mode', {
                get: () => 'readwrite'
            });
        }
    };

    // Move handleTimeout inside setupDatabaseChaos to access logChaos
    const handleTimeout = (transaction: IDBTransaction) => {
        const originalObjectStore = transaction.objectStore;
        transaction.objectStore = function (name: string) {
            const store = originalObjectStore.call(this, name);
            setTimeout(() => {
                logChaos({
                    type: 'indexeddb.timeout',
                    description: 'Delayed object store operation',
                    duration: 1000,
                    impact: 'medium',
                    recoverable: true
                });
            }, 1000);
            return store;
        };
    };

    // Fix type for storeNames
    const handleConcurrentChaos = (
        db: IDBDatabase,
        storeNames: string | string[] | Iterable<string>,
        mode?: IDBTransactionMode,
        options?: IDBTransactionOptions
    ) => {
        const safeStoreNames = typeof storeNames === 'string'
            ? [storeNames]
            : Array.isArray(storeNames)
                ? storeNames
                : Array.from(storeNames);
        for (let i = 0; i < 3; i++) {
            setTimeout(() => {
                const parallelTx = originalTransaction.call(db, safeStoreNames, mode, options);
                parallelTx.oncomplete = () => {
                    logChaos({
                        type: 'indexeddb.concurrent',
                        description: 'Created parallel transaction',
                        duration: 0,
                        impact: 'medium',
                        recoverable: true
                    });
                };
            }, i * 100);
        }
    };

    // IndexedDB Chaos
    window.indexedDB = new Proxy(originalIndexedDB, {
        get(target, prop) {
            if (prop === 'open') {
                return new Proxy(target.open, {
                    apply: (target, thisArg, args: [string, number?]) => {
                        if (shouldCreateChaos('database')) {
                            const [dbName, dbVersion] = args;
                            const chaos = pickOne([
                                'version_mismatch',
                                'schema_chaos',
                                'delayed_open',
                                'quota_exceed'
                            ]);

                            switch (chaos) {
                                case 'version_mismatch':
                                    return handleVersionMismatch(dbName, dbVersion);
                                case 'schema_chaos':
                                    return createSchemaChaos(target.apply(thisArg, [dbName, dbVersion]));
                                case 'delayed_open':
                                    return createDelayedOpen(target.apply(thisArg, [dbName, dbVersion]));
                                case 'quota_exceed':
                                    return handleQuotaExceeded();
                            }
                        }
                        return target.apply(thisArg, args);
                    }
                });
            }
            return Reflect.get(target, prop);
        }
    });

    // Schema Chaos: randomly modify object stores
    const createSchemaChaos = (request: IDBOpenDBRequest): Promise<IDBDatabase> => {
        return new Promise((resolve, reject) => {
            request.onerror = () => reject(request.error || new Error('Database operation failed'));
            request.onsuccess = () => resolve(request.result);

            request.onupgradeneeded = (event: IDBVersionChangeEvent) => {
                const db = request.result;
                const existingStores = Array.from(db.objectStoreNames);

                if (Math.random() < 0.3 && !config.safeMode) {
                    const storeToDelete = pickOne(existingStores);
                    if (storeToDelete) {
                        db.deleteObjectStore(storeToDelete);
                        logChaos({
                            type: 'indexeddb.schema',
                            description: `Deleted object store ${storeToDelete}`,
                            duration: 0,
                            impact: 'high',
                            recoverable: false
                        });
                    }
                }

                // Add chaotic indexes
                existingStores.forEach(storeName => {
                    if (event.target instanceof IDBOpenDBRequest) {
                        const transaction = event.target.transaction;
                        if (transaction) {
                            const store = transaction.objectStore(storeName);
                            if (store && Math.random() < 0.4) {
                                try {
                                    store.createIndex(
                                        `chaos_${Date.now()}`,
                                        Math.random() < 0.5 ? '_id' : 'timestamp',
                                        { unique: Math.random() < 0.3 }
                                    );
                                } catch { }
                            }
                        }
                    }
                });
            };
        });
    };

    // Delayed database operations
    const createDelayedOpen = (request: IDBOpenDBRequest) => {
        return new Promise((resolve, reject) => {
            const delay = Math.random() * 3000;

            setTimeout(() => {
                request.onerror = () => reject(request.error || new Error('Database operation failed'));
                request.onsuccess = () => {
                    const db = request.result;
                    activeDatabases.add(db);
                    resolve(db);
                };
            }, delay);

            logChaos({
                type: 'indexeddb.delay',
                description: `Delayed database open by ${delay.toFixed(0)}ms`,
                duration: delay,
                impact: 'medium',
                recoverable: true
            });
        });
    };

    // Transaction Chaos
    const originalTransaction = IDBDatabase.prototype.transaction;
    IDBDatabase.prototype.transaction = function (...args) {
        const transaction = originalTransaction.apply(this, args);

        if (shouldCreateChaos('database')) {
            const chaos = pickOne([
                'abort',
                'timeout',
                'readonly_upgrade',
                'concurrent_chaos'
            ]);

            switch (chaos) {
                case 'abort':
                    setTimeout(() => {
                        transaction.abort();
                        logChaos({
                            type: 'indexeddb.transaction',
                            description: 'Aborted transaction',
                            duration: 0,
                            impact: 'high',
                            recoverable: true
                        });
                    }, Math.random() * 1000);
                    break;

                case 'timeout':
                    handleTimeout(transaction);
                    break;

                case 'readonly_upgrade':
                    handleReadonlyUpgrade(transaction);
                    break;

                case 'concurrent_chaos':
                    handleConcurrentChaos(this, args[0], args[1], args[2]);
                    break;
            }
        }

        return transaction;
    };

    // WebSQL support
    if ('openDatabase' in window) {
        savedOpenDatabase = (window as any).openDatabase as OpenDatabaseFunction;
        (window as any).openDatabase = function (
            name: string,
            version: string,
            displayName: string,
            size: number
        ): WebSQLDatabase {
            if (!savedOpenDatabase) {
                throw new Error('WebSQL is not supported in this browser');
            }
            const db = savedOpenDatabase(name, version, displayName, size);

            const originalTransaction = db.transaction;
            db.transaction = function (callback, errorCallback, successCallback) {
                const wrappedCallback = (tx: SQLTransaction) => {
                    const txId = transactionId++;
                    sqlTransactions.set(txId, tx);

                    if (shouldCreateChaos('database')) {
                        const chaos = pickOne([
                            'syntax_error',
                            'constraint_violation',
                            'delayed_execution',
                            'partial_failure'
                        ]);

                        switch (chaos) {
                            case 'syntax_error': {
                                const originalExecuteSql = tx.executeSql;
                                tx.executeSql = function (sql: string, ...args: unknown[]) {
                                    if (Math.random() < 0.3) {
                                        sql = sql.replace(/SELECT|INSERT|UPDATE|DELETE/i, 'CHAOS');
                                    }
                                    return originalExecuteSql.call(this, sql, ...args);
                                };
                                break;
                            }
                            case 'constraint_violation': {
                                if (!config.safeMode) {
                                    tx.executeSql('DROP TABLE IF EXISTS chaos_temp');
                                    logChaos({
                                        type: 'websql.constraint',
                                        description: 'Violated SQL constraints',
                                        duration: 0,
                                        impact: 'high',
                                        recoverable: false
                                    });
                                }
                                break;
                            }
                            case 'delayed_execution': {
                                setTimeout(() => callback(tx), Math.random() * 2000);
                                return;
                            }
                            case 'partial_failure': {
                                const originalSuccess = successCallback;
                                if (originalSuccess) {
                                    successCallback = function (this: unknown) {
                                        if (Math.random() < 0.5) {
                                            errorCallback?.(new Error('Chaos partial failure'));
                                        } else {
                                            originalSuccess.apply(this);
                                        }
                                    };
                                }
                                break;
                            }
                        }
                    }

                    callback(tx);
                    sqlTransactions.delete(txId);
                };

                return originalTransaction.call(
                    this,
                    wrappedCallback,
                    errorCallback,
                    successCallback
                );
            };

            return db;
        };
    }

    return {
        getActiveDatabases: () => activeDatabases,
        getSQLTransactions: () => sqlTransactions,
        cleanup: () => {
            window.indexedDB = originalIndexedDB;
            activeDatabases.forEach(db => db.close());
            if ('openDatabase' in window && savedOpenDatabase) {
                (window as any).openDatabase = savedOpenDatabase;
            }
        }
    };
};