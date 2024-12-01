import { pickOne } from '../../utils';

interface Log {
    type: string;
    description: string;
    duration: number;
    impact: 'low' | 'medium' | 'high';
    recoverable: boolean;
}

export const setupDOMChaos = ({ logChaos, shouldCreateChaos, config }: { logChaos: (log: Log) => void, shouldCreateChaos: (type: string) => boolean, config: { safeMode: boolean } }) => {
    const createDOMChange = () => {
        const changes = [
            shuffleChildren,
            toggleVisibility,
            tweakStyles,
            injectPlaceholder,
            swapElements
        ];

        pickOne(changes)();
    };

    const shuffleChildren = () => {
        const parent = pickRandomElement();
        if (!parent?.children?.length) return;

        const children = Array.from(parent.children);
        for (let i = children.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            parent.appendChild(children[j]);
        }

        logChaos({
            type: 'dom.shuffle',
            description: `Shuffled ${children.length} children in ${elementPath(parent)}`,
            duration: 0,
            impact: 'medium',
            recoverable: true
        });
    };

    const toggleVisibility = () => {
        const element = pickRandomElement();
        if (!element) return;

        const originalDisplay = element.style.display;
        element.style.display = originalDisplay === 'none' ? '' : 'none';

        setTimeout(() => {
            element.style.display = originalDisplay;
        }, 2000);

        logChaos({
            type: 'dom.visibility',
            description: `Temporarily hidden ${elementPath(element)}`,
            duration: 2000,
            impact: 'medium',
            recoverable: true
        });
    };

    const tweakStyles = () => {
        const element = pickRandomElement();
        if (!element) return;

        const originalStyles = element.style.cssText;
        const changes = pickOne([
            () => { element.style.transform = `rotate(${Math.random() * 10 - 5}deg)`; },
            () => { element.style.opacity = '0.5'; },
            () => { element.style.filter = 'blur(2px)'; },
            () => { element.style.scale = '0.95'; }
        ]);

        changes();

        setTimeout(() => {
            element.style.cssText = originalStyles;
        }, 3000);

        logChaos({
            type: 'dom.style',
            description: `Modified styles of ${elementPath(element)}`,
            duration: 3000,
            impact: 'low',
            recoverable: true
        });
    };

    const injectPlaceholder = () => {
        const element = pickRandomElement();
        if (!element) return;

        const placeholder = document.createElement('div');
        placeholder.textContent = '🐒 Chaos was here';
        placeholder.style.cssText = `
            padding: 4px;
            background: rgba(255, 0, 0, 0.1);
            border: 1px dashed red;
            font-size: 10px;
            pointer-events: none;
        `;

        element.appendChild(placeholder);

        setTimeout(() => {
            placeholder.remove();
        }, 4000);

        logChaos({
            type: 'dom.inject',
            description: `Injected placeholder into ${elementPath(element)}`,
            duration: 4000,
            impact: 'low',
            recoverable: true
        });
    };

    const swapElements = () => {
        if (config.safeMode) return; // Too risky for safe mode

        const element1 = pickRandomElement();
        const element2 = pickRandomElement();
        if (!element1 || !element2 || element1 === element2) return;

        const parent1 = element1.parentNode;
        const parent2 = element2.parentNode;
        const next1 = element1.nextSibling;
        const next2 = element2.nextSibling;

        parent2?.insertBefore(element1, next2);
        parent1?.insertBefore(element2, next1);

        logChaos({
            type: 'dom.swap',
            description: `Swapped ${elementPath(element1)} with ${elementPath(element2)}`,
            duration: 0,
            impact: 'high',
            recoverable: false
        });
    };

    // Helper to pick random element
    const pickRandomElement = (): HTMLElement | null => {
        const elements = document.querySelectorAll('*');
        const safeElements = Array.from(elements).filter(el => {
            // Avoid chaos monkey UI and critical elements
            return !el.closest('.debug-chaos') &&
                !el.closest('script') &&
                !el.closest('style') &&
                el.tagName !== 'HTML' &&
                el.tagName !== 'BODY' &&
                el.tagName !== 'HEAD';
        });

        return safeElements[Math.floor(Math.random() * safeElements.length)] as HTMLElement;
    };

    // Helper to get element path
    const elementPath = (element: Element): string => {
        const path: string[] = [];
        let current = element;

        while (current && current.tagName !== 'HTML') {
            let selector = current.tagName.toLowerCase();
            if (current.id) {
                selector += `#${current.id}`;
            } else if (current.className) {
                selector += `.${current.className.split(' ')[0]}`;
            }
            path.unshift(selector);
            current = current.parentElement as Element;
        }

        return path.join(' > ');
    };

    setInterval(() => {
        if (shouldCreateChaos('dom')) {
            createDOMChange();
        }
    }, 5000);
};

export const setupScriptChaos = ({ logChaos, shouldCreateChaos, config }: { logChaos: (log: Log) => void, shouldCreateChaos: (type: string) => boolean, config: { safeMode: boolean } }) => {
    const handleChaosEffect = (prop: string, value: any, args: any[], context: any) => {
        const chaos = pickOne(['delay', 'error', 'modify', 'skip']);

        switch (chaos) {
            case 'delay':
                logChaos({
                    type: 'script.delay',
                    description: `Delayed execution of global function ${prop}`,
                    duration: 1000,
                    impact: 'low',
                    recoverable: true
                });
                return new Promise(resolve => {
                    setTimeout(() => {
                        resolve(value.apply(context, args));
                    }, 1000);
                });

            case 'error':
                if (!config.safeMode) {
                    logChaos({
                        type: 'script.error',
                        description: `Injected error into ${prop}`,
                        duration: 0,
                        impact: 'high',
                        recoverable: false
                    });
                    throw new Error(`Chaos Monkey: Function ${prop} failed`);
                }
                break;

            case 'modify':
                logChaos({
                    type: 'script.modify',
                    description: `Modified arguments for ${prop}`,
                    duration: 0,
                    impact: 'medium',
                    recoverable: true
                });
                return value.apply(context, args.map(arg =>
                    typeof arg === 'number' ? arg + (Math.random() * 0.1) : arg
                ));

            case 'skip':
                logChaos({
                    type: 'script.skip',
                    description: `Skipped execution of ${prop}`,
                    duration: 0,
                    impact: 'medium',
                    recoverable: true
                });
                return undefined;
        }
        return value.apply(context, args);
    };

    const interceptGlobals = () => {
        // Keep track of added globals
        const addedGlobals = new Set<string>();
        const originalValues = new Map<string, any>();

        // Watch for new globals
        const globalProxy = new Proxy(window, {
            set(target: any, prop: string, value: any) {
                if (!addedGlobals.has(prop) &&
                    !(prop in target) &&
                    typeof value === 'function') {

                    addedGlobals.add(prop);
                    originalValues.set(prop, value);

                    // Create wrapped function directly here
                    target[prop] = function (this: any, ...args: any[]) {
                        if (shouldCreateChaos('script')) {
                            return handleChaosEffect(prop, value, args, this);
                        }
                        return value.apply(this, args);
                    };
                    return true;
                }
                return Reflect.set(target, prop, value);
            }
        });

        // Replace window with proxy
        Object.getOwnPropertyNames(window).forEach(prop => {
            if (!(prop in globalProxy)) {
                Object.defineProperty(globalProxy, prop, Object.getOwnPropertyDescriptor(window, prop)!);
            }
        });

        return {
            cleanup: () => {
                addedGlobals.forEach(prop => {
                    if (originalValues.has(prop)) {
                        (window as any)[prop] = originalValues.get(prop);
                    }
                });
            }
        };
    };

    const scriptChaos = interceptGlobals();

    return {
        cleanup: () => scriptChaos.cleanup()
    };
};
