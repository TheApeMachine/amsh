import { Log } from './types';
import { pickOne } from '../utils';

export const setupHistoryLocationChaos = ({ logChaos, shouldCreateChaos, config }: { logChaos: (log: Log) => void, shouldCreateChaos: (type: string) => boolean, config: { safeMode: boolean } }) => {
    const originalPushState = window.history.pushState;
    const originalReplaceState = window.history.replaceState;
    const originalBack = window.history.back;
    const originalForward = window.history.forward;
    const originalGo = window.history.go;

    const historyStack = new Map<string, { state: any, title: string }>();
    let chaosTimeout: number | null = null;

    // Modify pushState to introduce chaos
    window.history.pushState = function (state, title, url) {
        if (shouldCreateChaos('history')) {
            const chaos = pickOne([
                'modifyUrl',
                'duplicateState',
                'randomJump',
                'delayedPush',
                'fragmentChaos'
            ]);

            const jumpSteps = Math.floor(Math.random() * 3) - 1;

            switch (chaos) {
                case 'modifyUrl':
                    if (url) {
                        const urlObj = new URL(url.toString(), window.location.href);
                        urlObj.searchParams.append('chaos', Date.now().toString());
                        url = urlObj.toString();

                        logChaos({
                            type: 'history.url',
                            description: 'Modified history URL',
                            duration: 0,
                            impact: 'low',
                            recoverable: true
                        });
                    }
                    break;

                case 'duplicateState':
                    originalPushState.call(history, state, title, url);
                    setTimeout(() => {
                        originalPushState.call(history,
                            { ...state, duplicate: true },
                            title + ' (Duplicate)',
                            url
                        );
                    }, 100);
                    return;

                case 'randomJump':
                    setTimeout(() => history.go(jumpSteps), 500);
                    break;

                case 'delayedPush':
                    setTimeout(() => {
                        originalPushState.call(history, state, title, url);
                    }, Math.random() * 1000);
                    return;

                case 'fragmentChaos':
                    if (url) {
                        const urlObj = new URL(url.toString(), window.location.href);
                        urlObj.hash = `chaos-${Math.random().toString(36).slice(2)}`;
                        url = urlObj.toString();
                    }
                    break;
            }
        }

        historyStack.set(url?.toString() ?? window.location.href, { state, title });
        return originalPushState.call(this, state, title, url);
    };

    // Modify replaceState for chaos
    window.history.replaceState = function (state, title, url) {
        if (shouldCreateChaos('history')) {
            const chaos = pickOne([
                'pushInstead',
                'queryParams',
                'stateCorruption'
            ]);

            switch (chaos) {
                case 'pushInstead':
                    return window.history.pushState(state, title, url);

                case 'queryParams':
                    if (url) {
                        const urlObj = new URL(url.toString(), window.location.href);
                        urlObj.searchParams.forEach((value, key) => {
                            if (Math.random() < 0.3) {
                                urlObj.searchParams.set(key, `chaos_${value}`);
                            }
                        });
                        url = urlObj.toString();
                    }
                    break;

                case 'stateCorruption':
                    if (state && typeof state === 'object') {
                        state = {
                            ...state,
                            _chaosModified: true,
                            _timestamp: Date.now()
                        };
                    }
                    break;
            }
        }

        return originalReplaceState.call(this, state, title, url);
    };

    // Create navigation chaos
    const createNavigationChaos = ({ logChaos, shouldCreateChaos, config }: { logChaos: (log: Log) => void, shouldCreateChaos: (type: string) => boolean, config: { safeMode: boolean } }) => {
        if (!chaosTimeout && shouldCreateChaos('history')) {
            const chaos = pickOne([
                'randomNavigation',
                'navigationLoop',
                'stateFlutter',
                'locationMutation'
            ]);

            switch (chaos) {
                case 'randomNavigation':
                    chaosTimeout = window.setTimeout(() => {
                        const steps = Math.floor(Math.random() * 5) - 2;
                        history.go(steps);
                        logChaos({
                            type: 'history.navigation',
                            description: `Random navigation: ${steps} steps`,
                            duration: 0,
                            impact: 'medium',
                            recoverable: true
                        });
                        chaosTimeout = null;
                    }, Math.random() * 2000);
                    break;

                case 'navigationLoop': {
                    let count = 0;
                    const interval = setInterval(() => {
                        if (count++ < 3) {
                            history.back();
                            logChaos({
                                type: 'history.navigation',
                                description: 'Navigation loop',
                                duration: 1000,
                                impact: 'medium',
                                recoverable: true
                            });
                            setTimeout(() => history.forward(), 500);
                        } else {
                            clearInterval(interval);
                        }
                    }, 1000);
                    break;
                }

                case 'stateFlutter': {
                    const currentUrl = window.location.href;
                    const states = Array.from(historyStack.entries())
                        .filter(([url]) => url !== currentUrl)
                        .slice(-3);

                    states.forEach(([url, { state, title }], index) => {
                        setTimeout(() => {
                            history.pushState(state, title, url);
                        }, index * 200);
                    });
                    break;
                }

                case 'locationMutation':
                    if (!config.safeMode) {
                        const originalHref = window.location.href;
                        Object.defineProperty(window.location, 'href', {
                            get: () => originalHref,
                            set: (value) => {
                                const chaosUrl = new URL(value);
                                chaosUrl.searchParams.append('chaos', 'true');
                                window.location.href = chaosUrl.toString();
                            },
                            configurable: true
                        });

                        setTimeout(() => {
                            delete (window.location as any).href;
                            Object.defineProperty(window.location, 'href', {
                                value: originalHref,
                                writable: true
                            });
                        }, 5000);
                    }
                    break;
            }
        }
    };

    // Monitor navigation events
    window.addEventListener('popstate', (event) => {
        if (shouldCreateChaos('history')) {
            const chaos = pickOne(['prevent', 'modify', 'duplicate']);

            switch (chaos) {
                case 'prevent':
                    if (!config.safeMode) {
                        event.preventDefault();
                        event.stopImmediatePropagation();

                        logChaos({
                            type: 'history.prevent',
                            description: 'Prevented navigation',
                            duration: 0,
                            impact: 'high',
                            recoverable: false
                        });
                    }
                    break;

                case 'modify':
                    if (event.state && typeof event.state === 'object') {
                        Object.defineProperty(event, 'state', {
                            get: () => ({
                                ...event.state,
                                chaosModified: true
                            })
                        });
                    }
                    break;

                case 'duplicate':
                    setTimeout(() => {
                        window.dispatchEvent(new PopStateEvent('popstate', {
                            state: event.state
                        }));
                    }, 100);
                    break;
            }
        }
    });

    // Setup periodic chaos
    const chaosInterval = setInterval(() => createNavigationChaos({ logChaos, shouldCreateChaos, config }), 10000);

    // Track scroll restoration
    let originalScrollRestoration = history.scrollRestoration;
    Object.defineProperty(history, 'scrollRestoration', {
        get: () => {
            if (shouldCreateChaos('history')) {
                return Math.random() < 0.5 ? 'auto' : 'manual';
            }
            return originalScrollRestoration;
        },
        set: (value) => {
            if (shouldCreateChaos('history')) {
                value = Math.random() < 0.5 ? 'auto' : 'manual';
            }
            originalScrollRestoration = value;
        }
    });

    return {
        getHistoryStack: () => historyStack,
        cleanup: () => {
            window.history.pushState = originalPushState;
            window.history.replaceState = originalReplaceState;
            window.history.back = originalBack;
            window.history.forward = originalForward;
            window.history.go = originalGo;
            if (chaosTimeout) {
                clearTimeout(chaosTimeout);
            }
            clearInterval(chaosInterval);
            historyStack.clear();
            Object.defineProperty(history, 'scrollRestoration', {
                value: originalScrollRestoration,
                writable: true
            });
        }
    };
};
