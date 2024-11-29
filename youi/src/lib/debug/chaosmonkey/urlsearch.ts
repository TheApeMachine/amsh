import { Log } from './types';
import { pickOne } from './utils';

export const setupURLChaos = ({ logChaos, shouldCreateChaos, config }: { logChaos: (log: Log) => void, shouldCreateChaos: (type: string) => boolean, config: { safeMode: boolean } }) => {
    const originalURL = window.URL;
    const originalURLSearchParams = window.URLSearchParams;
    const urlHistory = new Set<string>();

    window.URL = class ChaosURL extends URL {
        constructor(url: string | URL, base?: string | URL) {
            if (shouldCreateChaos('url')) {
                url = applyChaosToURL(url.toString());
            }
            super(url, base);
            urlHistory.add(this.href);
        }

        get href() {
            const href = super.href;
            if (shouldCreateChaos('url')) {
                return modifyHref(href);
            }
            return href;
        }

        set href(value: string) {
            if (shouldCreateChaos('url')) {
                value = applyChaosToURL(value);
            }
            super.href = value;
            urlHistory.add(value);
        }

        get search() {
            const search = super.search;
            if (shouldCreateChaos('url')) {
                return modifySearch(search);
            }
            return search;
        }

        set search(value: string) {
            if (shouldCreateChaos('url')) {
                value = applyChaosToSearch(value);
            }
            super.search = value;
        }

        get hash() {
            const hash = super.hash;
            if (shouldCreateChaos('url') && Math.random() < 0.3) {
                return `#chaos-${Date.now()}`;
            }
            return hash;
        }

        set hash(value: string) {
            if (shouldCreateChaos('url')) {
                value = `${value}-chaos-${Math.random().toString(36).slice(2)}`;
            }
            super.hash = value;
        }
    };

    window.URLSearchParams = class ChaosURLSearchParams extends URLSearchParams {
        constructor(init?: string | URLSearchParams | string[][] | Record<string, string>) {
            if (shouldCreateChaos('url') && init) {
                init = modifySearchParamsInit(init);
            }
            super(init);
        }

        append(name: string, value: string) {
            if (shouldCreateChaos('url')) {
                const chaos = pickOne(['duplicate', 'modify', 'case']);
                switch (chaos) {
                    case 'duplicate':
                        super.append(name, value);
                        super.append(`${name}_chaos`, value);
                        return;
                    case 'modify':
                        value = `chaos_${value}_${Date.now()}`;
                        break;
                    case 'case':
                        name = Math.random() < 0.5 ? name.toLowerCase() : name.toUpperCase();
                        break;
                }
            }
            super.append(name, value);
        }

        set(name: string, value: string) {
            if (shouldCreateChaos('url')) {
                const chaos = pickOne(['array', 'encode', 'prefix']);
                switch (chaos) {
                    case 'array':
                        super.set(name, '');
                        for (let i = 0; i < 3; i++) {
                            super.append(name, `${value}_${i}`);
                        }
                        return;
                    case 'encode':
                        value = encodeURIComponent(value) + Math.random().toString(36).slice(2);
                        break;
                    case 'prefix':
                        name = `chaos_${name}`;
                        break;
                }
            }
            super.set(name, value);
        }

        delete(name: string) {
            if (shouldCreateChaos('url')) {
                // Sometimes don't delete, or delete similar params
                if (Math.random() < 0.3) {
                    return;
                }
                const allParams = Array.from(this.keys());
                const similar = allParams.filter(k =>
                    k.includes(name) || name.includes(k)
                );
                similar.forEach(k => super.delete(k));
                return;
            }
            super.delete(name);
        }

        sort() {
            if (shouldCreateChaos('url')) {
                // Random sort instead of alphabetical
                const entries = Array.from(this.entries());
                const shuffled = [...entries].sort(() => Math.random() - 0.5);
                Array.from(this.keys()).forEach(key => this.delete(key));
                shuffled.forEach(([k, v]) => this.append(k, v));
                return;
            }
            super.sort();
        }
    };

    const applyChaosToURL = (url: string): string => {
        const urlObj = new URL(url);
        const chaos = pickOne(['path', 'query', 'fragment', 'protocol']);

        switch (chaos) {
            case 'path': {
                const paths = urlObj.pathname.split('/');
                if (paths.length > 2) {
                    const randomIndex = Math.floor(Math.random() * (paths.length - 1)) + 1;
                    paths[randomIndex] = `chaos_${paths[randomIndex]}`;
                    urlObj.pathname = paths.join('/');
                }
                break;
            }

            case 'query':
                urlObj.searchParams.append('chaos', Date.now().toString());
                break;

            case 'fragment':
                urlObj.hash = urlObj.hash ?
                    `${urlObj.hash}-chaos` :
                    '#chaos';
                break;

            case 'protocol':
                if (!config.safeMode) {
                    urlObj.protocol = Math.random() < 0.5 ?
                        'http:' : 'https:';
                }
                break;
        }

        logChaos({
            type: 'url.modify',
            description: `Modified URL: ${chaos}`,
            duration: 0,
            impact: 'medium',
            recoverable: true
        });

        return urlObj.toString();
    };

    const modifyHref = (href: string): string => {
        if (Math.random() < 0.2) {
            const urlObj = new URL(href);
            urlObj.searchParams.sort();
            if (urlObj.hash) {
                urlObj.hash = `${urlObj.hash}-${Date.now()}`;
            }
            return urlObj.toString();
        }
        return href;
    };

    const modifySearch = (search: string): string => {
        if (!search) return search;
        const params = new URLSearchParams(search);
        if (Math.random() < 0.3) {
            params.append('chaos_timestamp', Date.now().toString());
        }
        return `?${params.toString()}`;
    };

    const applyChaosToSearch = (search: string): string => {
        if (!search) return search;
        const params = new URLSearchParams(search);
        const entries = Array.from(params.entries());

        if (entries.length > 0) {
            const randomEntry = entries[Math.floor(Math.random() * entries.length)];
            params.set(randomEntry[0], `chaos_${randomEntry[1]}`);
        }

        return params.toString();
    };

    const modifySearchParamsInit = (init: any): any => {
        if (typeof init === 'string') {
            const params = new URLSearchParams(init);
            params.append('chaos', 'true');
            return params.toString();
        } else if (init instanceof URLSearchParams) {
            init.append('chaos', 'true');
        } else if (Array.isArray(init)) {
            init.push(['chaos', 'true']);
        } else if (typeof init === 'object') {
            return { ...init, chaos: 'true' };
        }
        return init;
    };

    // Monitor URL changes
    const observer = new MutationObserver(() => {
        if (shouldCreateChaos('url') && Math.random() < 0.1) {
            const currentURL = window.location.href;
            const urlObj = new URL(currentURL);

            if (Math.random() < 0.5) {
                // Modify search params
                urlObj.searchParams.append(`chaos_${Date.now()}`, 'true');
            } else {
                // Modify hash
                urlObj.hash = urlObj.hash ?
                    `${urlObj.hash}-chaos` :
                    '#chaos';
            }

            history.replaceState(null, '', urlObj.toString());
        }
    });

    observer.observe(document, {
        subtree: true,
        childList: true,
        attributes: true,
        attributeFilter: ['href', 'src']
    });

    return {
        getURLHistory: () => urlHistory,
        cleanup: () => {
            window.URL = originalURL;
            window.URLSearchParams = originalURLSearchParams;
            observer.disconnect();
            urlHistory.clear();
        }
    };
};
