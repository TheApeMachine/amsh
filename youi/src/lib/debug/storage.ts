import { DebugEntry } from "./types";

interface StorageState {
    localStorage: Record<string, string>;
    sessionStorage: Record<string, string>;
    cookies: Record<string, string>;
}

export const setupStorageTracking = ({ addLog }: { addLog: (entry: DebugEntry) => void }) => {
    // Helper to parse cookies into an object
    const parseCookies = (): Record<string, string> => {
        return Object.fromEntries(
            document.cookie.split(';')
                .map(cookie => cookie.trim().split('='))
                .map(([key, value]) => [
                    key,
                    (() => {
                        try {
                            return decodeURIComponent(value);
                        } catch {
                            return value;
                        }
                    })()
                ])
        );
    };

    // Helper to get all storage values
    const getStorageSnapshot = (): StorageState => ({
        localStorage: Object.fromEntries(
            Object.keys(localStorage).map(key => [key, localStorage.getItem(key) ?? ''])
        ),
        sessionStorage: Object.fromEntries(
            Object.keys(sessionStorage).map(key => [key, sessionStorage.getItem(key) ?? ''])
        ),
        cookies: parseCookies()
    });

    // Log initial state
    const logStorageState = () => {
        const state = getStorageSnapshot();
        addLog({
            id: crypto.randomUUID(),
            timestamp: new Date().toISOString(),
            type: 'storage',
            category: 'storage.snapshot',
            summary: `Storage State: ${Object.keys(state.localStorage).length
                } local, ${Object.keys(state.sessionStorage).length
                } session, ${Object.keys(state.cookies).length
                } cookies`,
            details: state
        });
    };

    // Track localStorage changes
    const originalSetItem = localStorage.setItem;
    localStorage.setItem = function (key: string, value: string) {
        const oldValue = localStorage.getItem(key);
        originalSetItem.apply(localStorage, [key, value]);
        addLog({
            id: crypto.randomUUID(),
            timestamp: new Date().toISOString(),
            type: 'storage',
            category: 'storage.local',
            summary: `localStorage.${key} changed`,
            details: {
                key,
                oldValue,
                newValue: value,
                action: 'set'
            }
        });
    };

    const originalRemoveItem = localStorage.removeItem;
    localStorage.removeItem = function (key: string) {
        const oldValue = localStorage.getItem(key);
        originalRemoveItem.apply(localStorage, [key]);
        addLog({
            id: crypto.randomUUID(),
            timestamp: new Date().toISOString(),
            type: 'storage',
            category: 'storage.local',
            summary: `localStorage.${key} removed`,
            details: {
                key,
                oldValue,
                action: 'remove'
            }
        });
    };

    const originalClear = localStorage.clear;
    localStorage.clear = function () {
        const oldState = { ...localStorage };
        originalClear.apply(localStorage);
        addLog({
            id: crypto.randomUUID(),
            timestamp: new Date().toISOString(),
            type: 'storage',
            category: 'storage.local',
            summary: 'localStorage cleared',
            details: {
                oldState,
                action: 'clear'
            }
        });
    };

    // Track sessionStorage changes (similar to localStorage)
    const originalSessionSetItem = sessionStorage.setItem;
    sessionStorage.setItem = function (key: string, value: string) {
        const oldValue = sessionStorage.getItem(key);
        originalSessionSetItem.apply(sessionStorage, [key, value]);
        addLog({
            id: crypto.randomUUID(),
            timestamp: new Date().toISOString(),
            type: 'storage',
            category: 'storage.session',
            summary: `sessionStorage.${key} changed`,
            details: {
                key,
                oldValue,
                newValue: value,
                action: 'set'
            }
        });
    };

    // Track cookie changes
    let lastCookieState = parseCookies();
    const cookieObserver = setInterval(() => {
        const currentCookies = parseCookies();
        const changes: Array<{ key: string; oldValue?: string; newValue?: string }> = [];

        // Check for added or modified cookies
        Object.entries(currentCookies).forEach(([key, value]) => {
            if (!(key in lastCookieState) || lastCookieState[key] !== value) {
                changes.push({
                    key,
                    oldValue: lastCookieState[key],
                    newValue: value
                });
            }
        });

        // Check for removed cookies
        Object.keys(lastCookieState).forEach(key => {
            if (!(key in currentCookies)) {
                changes.push({
                    key,
                    oldValue: lastCookieState[key]
                });
            }
        });

        if (changes.length > 0) {
            addLog({
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                type: 'storage',
                category: 'storage.cookies',
                summary: `Cookie changes detected (${changes.length})`,
                details: {
                    changes,
                    currentState: currentCookies
                }
            });
        }

        lastCookieState = currentCookies;
    }, 1000);

    // Add storage event listener for cross-tab changes
    window.addEventListener('storage', (e) => {
        if (e.storageArea === localStorage) {
            addLog({
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                type: 'storage',
                category: 'storage.local',
                summary: `localStorage.${e.key} changed (from another tab)`,
                details: {
                    key: e.key,
                    oldValue: e.oldValue,
                    newValue: e.newValue,
                    url: e.url,
                    action: 'external'
                }
            });
        }
    });

    // Add button to debug controls for viewing storage state
    const addViewStorageButton = () => {
        const controls = document.querySelector('.debug-controls');
        if (controls) {
            const button = document.createElement('button');
            button.className = 'debug-button';
            button.id = 'view-storage';
            button.textContent = 'View Storage';
            button.addEventListener('click', logStorageState);
            controls.appendChild(button);
        }
    };

    // Initialize
    addViewStorageButton();
    logStorageState(); // Log initial state

    return {
        destroy: () => {
            clearInterval(cookieObserver);
            localStorage.setItem = originalSetItem;
            localStorage.removeItem = originalRemoveItem;
            localStorage.clear = originalClear;
            sessionStorage.setItem = originalSessionSetItem;
        }
    };
};