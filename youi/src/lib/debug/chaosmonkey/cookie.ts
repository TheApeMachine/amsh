import { Log } from './types';
import { pickOne } from './utils';

export const setupCookieChaos = ({ logChaos, shouldCreateChaos }: { logChaos: (log: Log) => void, shouldCreateChaos: (type: string) => boolean, config: { safeMode: boolean } }) => {
    const originalCookie = Object.getOwnPropertyDescriptor(Document.prototype, 'cookie');
    const cookieJar = new Map<string, string>();
    let lastModified = Date.now();

    // Hijack cookie getter/setter
    Object.defineProperty(document, 'cookie', {
        get() {
            if (shouldCreateChaos('cookie')) {
                const chaos = pickOne(['expire', 'reorder', 'duplicate', 'malform']);

                switch (chaos) {
                    case 'expire':
                        return simulateExpiredCookies();
                    case 'reorder':
                        return reorderCookies();
                    case 'duplicate':
                        return duplicateCookies();
                    case 'malform':
                        return malformCookies();
                }
            }
            return originalCookie?.get?.call(document) || '';
        },

        set(value: string) {
            if (shouldCreateChaos('cookie')) {
                const chaos = pickOne([
                    'domain',
                    'sameSite',
                    'timeshift',
                    'httpOnly',
                    'attributes',
                    'inject'
                ]);

                switch (chaos) {
                    case 'domain':
                        value = modifyDomain(value);
                        break;
                    case 'sameSite':
                        value = modifySameSite(value);
                        break;
                    case 'timeshift':
                        value = modifyExpiration(value);
                        break;
                    case 'httpOnly':
                        value = toggleHttpOnly(value);
                        break;
                    case 'attributes':
                        value = shuffleAttributes(value);
                        break;
                    case 'inject':
                        injectChaoCookie();
                        break;
                }
            }

            cookieJar.set(getCookieName(value), value);
            lastModified = Date.now();
            originalCookie?.set?.call(document, value);
        }
    });

    const simulateExpiredCookies = () => {
        const cookies = originalCookie?.get?.call(document).split('; ') || [];
        return cookies
            .filter(() => Math.random() > 0.3) // Randomly "expire" some cookies
            .join('; ');
    };

    const reorderCookies = () => {
        const cookies = originalCookie?.get?.call(document).split('; ') || [];
        return cookies
            .sort(() => Math.random() - 0.5)
            .join('; ');
    };

    const duplicateCookies = () => {
        const cookies = originalCookie?.get?.call(document).split('; ') || [];
        return [...cookies, ...cookies.map((c: string) => c + '_dupe')].join('; ');
    };

    const malformCookies = () => {
        const cookies = originalCookie?.get?.call(document).split('; ') || [];
        return cookies
            .map((cookie: string) => {
                if (Math.random() < 0.2) {
                    return corruptCookie(cookie);
                }
                return cookie;
            })
            .join('; ');
    };

    const modifyDomain = (cookie: string): string => {
        const parts = cookie.split(';');
        const domainIndex = parts.findIndex(p => p.trim().startsWith('domain='));

        if (domainIndex > -1) {
            // Modify existing domain
            const domain = parts[domainIndex].split('=')[1];
            parts[domainIndex] = `domain=chaos.${domain}`;
        } else {
            // Add new domain
            parts.push(' domain=chaos.local');
        }

        logChaos({
            type: 'cookie.domain',
            description: 'Modified cookie domain',
            duration: 0,
            impact: 'medium',
            recoverable: true
        });

        return parts.join(';');
    };

    const modifySameSite = (cookie: string): string => {
        const values = ['Strict', 'Lax', 'None'];
        const parts = cookie.split(';');
        const sameIndex = parts.findIndex(p => p.trim().startsWith('SameSite='));

        if (sameIndex > -1) {
            parts[sameIndex] = `SameSite=${pickOne(values)}`;
        } else {
            parts.push(` SameSite=${pickOne(values)}`);
        }

        return parts.join(';');
    };

    const modifyExpiration = (cookie: string): string => {
        const parts = cookie.split(';');
        const expiresIndex = parts.findIndex(p =>
            p.trim().startsWith('expires=') || p.trim().startsWith('max-age=')
        );

        if (expiresIndex > -1) {
            const timeShift = Math.random() < 0.5 ?
                1000 * 60 * 60 * 24 : // 24 hours in the future
                -1000 * 60 * 30;      // 30 minutes in the past

            const date = new Date(Date.now() + timeShift);
            parts[expiresIndex] = `expires=${date.toUTCString()}`;

            logChaos({
                type: 'cookie.expiry',
                description: timeShift > 0 ? 'Extended cookie lifetime' : 'Shortened cookie lifetime',
                duration: Math.abs(timeShift),
                impact: 'medium',
                recoverable: true
            });
        }

        return parts.join(';');
    };

    const toggleHttpOnly = (cookie: string): string => {
        const parts = cookie.split(';');
        const httpOnlyIndex = parts.findIndex(p => p.trim() === 'HttpOnly');

        if (httpOnlyIndex > -1) {
            parts.splice(httpOnlyIndex, 1);
        } else {
            parts.push(' HttpOnly');
        }

        return parts.join(';');
    };

    const shuffleAttributes = (cookie: string): string => {
        const [nameValue, ...attributes] = cookie.split(';');
        const shuffledAttrs = [...attributes].sort(() => Math.random() - 0.5);
        return [
            nameValue,
            ...shuffledAttrs.map(attr => attr.trim())
        ].join(';');
    };

    const injectChaoCookie = () => {
        const chaoCookies = [
            'chaos_tracking=true',
            `chaos_timestamp=${Date.now()}`,
            'chaos_session=active'
        ];

        chaoCookies.forEach(cookie => {
            originalCookie?.set?.call(document, cookie);
        });

        logChaos({
            type: 'cookie.inject',
            description: 'Injected chaos cookies',
            duration: 0,
            impact: 'low',
            recoverable: true
        });
    };

    const corruptCookie = (cookie: string): string => {
        const [name, value] = cookie.split('=');
        if (!value) return cookie;

        try {
            // Try to parse as JSON
            const parsed = JSON.parse(value);
            if (typeof parsed === 'object') {
                return `${name}=${JSON.stringify({
                    ...parsed,
                    _chaos: true,
                    _timestamp: Date.now()
                })}`;
            }
        } catch {
            // Not JSON, corrupt the string
            return `${name}=${value.split('').reverse().join('')}`;
        }

        return cookie;
    };

    const getCookieName = (cookie: string): string => {
        return cookie.split(';')[0].split('=')[0].trim();
    };

    // Monitor cookie changes
    const observer = new MutationObserver(() => {
        const currentCookies = document.cookie;
        if (lastModified < Date.now() - 1000) { // Avoid recursion
            if (shouldCreateChaos('cookie')) {
                // Randomly remove or modify cookies
                const cookies = currentCookies.split('; ');
                cookies.forEach(cookie => {
                    if (Math.random() < 0.1) {
                        const name = getCookieName(cookie);
                        document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT`;
                    }
                });
            }
        }
    });

    observer.observe(document, {
        subtree: true,
        childList: true,
        attributes: true,
        characterData: true
    });

    // Setup periodic cookie chaos
    const chaosInterval = setInterval(() => {
        if (shouldCreateChaos('cookie') && Math.random() < 0.2) {
            const chaos = pickOne([
                'cleanup',
                'rotation',
                'multiplication'
            ]);

            let cookies: string[];
            let rotationCookies: string[];

            switch (chaos) {
                case 'cleanup':
                    cookies = document.cookie.split('; ');
                    cookies.forEach(cookie => {
                        if (Math.random() < 0.3) {
                            const name = getCookieName(cookie);
                            document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT`;
                        }
                    });
                    break;

                case 'rotation':
                    rotationCookies = Array.from(cookieJar.values());
                    if (rotationCookies.length > 1) {
                        for (let i = 0; i < rotationCookies.length; i++) {
                            const nextIndex = (i + 1) % rotationCookies.length;
                            const [name] = rotationCookies[i].split('=');
                            const [, value] = rotationCookies[nextIndex].split('=');
                            document.cookie = `${name}=${value}`;
                        }
                    }
                    break;

                case 'multiplication':
                    // Duplicate random cookies
                    document.cookie.split('; ').forEach(cookie => {
                        if (Math.random() < 0.2) {
                            const [name, value] = cookie.split('=');
                            document.cookie = `${name}_clone=${value}`;
                        }
                    });
                    break;
            }
        }
    }, 5000);

    return {
        getCookieJar: () => cookieJar,
        cleanup: () => {
            Object.defineProperty(document, 'cookie', originalCookie || {});
            observer.disconnect();
            clearInterval(chaosInterval);
            cookieJar.clear();
        }
    };
};
