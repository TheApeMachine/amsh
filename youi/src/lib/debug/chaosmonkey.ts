import { setupScriptChaos, setupDOMChaos } from './chaosmonkey/domscriptwindow';
import { setupMemoryChaos, setupCPUChaos, setupStorageChaos } from './chaosmonkey/cpumemstore';
import { setupCommunicationChaos } from './chaosmonkey/communication';
import { setupEventChaos } from './chaosmonkey/eventsystem';
import { setupWebSocketChaos } from './chaosmonkey/websocket';
import { setupURLChaos } from './chaosmonkey/urlsearch';
import { setupHistoryLocationChaos } from './chaosmonkey/history';
import { setupCookieChaos } from './chaosmonkey/cookie';
import { setupResourceTimingChaos } from './chaosmonkey/resourcetiming';

interface ChaosConfig {
    enabled: boolean;
    probability: number;
    safeMode: boolean;
    allowedChaos: Set<string>;
}

interface ChaosEvent {
    id: string;
    timestamp: number;
    type: string;
    description: string;
    duration: number;
    impact: 'low' | 'medium' | 'high';
    recoverable: boolean;
}

export const setupChaosMonkey = ({ addLog, overlay }: { addLog: Function; overlay: HTMLElement }) => {
    const config: ChaosConfig = {
        enabled: false,
        probability: 0.2,
        safeMode: true,
        allowedChaos: new Set([
            'network',
            'memory',
            'cpu',
            'storage',
            'dom',
            'script',
            'worker',
            'event',
            'websocket',
            'cookie',
            'history',
            'url'
        ])
    };

    const chaosHistory: ChaosEvent[] = [];
    const activeNetwork: Map<string, Response> = new Map();

    // Create container within the provided overlay
    const container = document.createElement('div');
    container.className = 'debug-chaos';
    overlay.appendChild(container);

    // Update the container HTML with all chaos types
    container.innerHTML = `
        <div class="chaos-header">
            <div class="chaos-header-content">
                <h3>🐒 Chaos Monkey</h3>
                <button class="debug-button" id="toggle-chaos">Enable</button>
            </div>
            <div class="chaos-controls">
                <label class="chaos-slider">
                    Chaos Probability:
                    <input type="range" id="chaos-probability" min="0" max="100" value="20" />
                    <span id="probability-value">20%</span>
                </label>
                <label class="chaos-checkbox">
                    <input type="checkbox" id="safe-mode" checked />
                    Safe Mode (Recoverable Only)
                </label>
                <div class="chaos-types">
                    <div class="chaos-types-column">
                        <label><input type="checkbox" value="network" checked /> Network</label>
                        <label><input type="checkbox" value="memory" checked /> Memory</label>
                        <label><input type="checkbox" value="cpu" checked /> CPU</label>
                        <label><input type="checkbox" value="storage" checked /> Storage</label>
                        <label><input type="checkbox" value="dom" checked /> DOM</label>
                        <label><input type="checkbox" value="script" checked /> Scripts</label>
                    </div>
                    <div class="chaos-types-column">
                        <label><input type="checkbox" value="worker" checked /> Workers</label>
                        <label><input type="checkbox" value="event" checked /> Events</label>
                        <label><input type="checkbox" value="websocket" checked /> WebSocket</label>
                        <label><input type="checkbox" value="cookie" checked /> Cookies</label>
                        <label><input type="checkbox" value="history" checked /> History</label>
                        <label><input type="checkbox" value="url" checked /> URL</label>
                    </div>
                </div>
            </div>
        </div>
        <div class="chaos-content"></div>
    `;

    // Network chaos implementation
    const setupNetworkChaos = () => {
        const originalFetch = window.fetch;
        window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
            if (!shouldCreateChaos('network')) {
                return originalFetch(input, init);
            }

            // Pick a random network chaos type
            const chaosType = pickOne([
                createNetworkDelay,
                createNetworkError,
                createNetworkCorruption
            ]);

            return chaosType(input, originalFetch, init);
        };
    };

    // Network chaos types
    const createNetworkDelay = async (input: RequestInfo | URL, originalFetch: Function, init?: RequestInit) => {
        const delay = Math.random() * 2000 + 1000; // 1-3 second delay

        logChaos({
            type: 'network.delay',
            description: `Added ${delay.toFixed(0)}ms delay to ${input.toString()}`,
            duration: delay,
            impact: 'low',
            recoverable: true
        });

        await new Promise(resolve => setTimeout(resolve, delay));
        return originalFetch(input, init);
    };

    const createNetworkError = async (input: RequestInfo | URL, originalFetch: Function, init?: RequestInit) => {
        if (config.safeMode) {
            // In safe mode, just delay instead of failing
            return createNetworkDelay(input, originalFetch, init);
        }

        const errorType = pickOne([
            { status: 404, statusText: 'Not Found' },
            { status: 500, statusText: 'Internal Server Error' },
            { status: 503, statusText: 'Service Unavailable' }
        ]);

        logChaos({
            type: 'network.error',
            description: `Injected ${errorType.status} error for ${input.toString()}`,
            duration: 0,
            impact: 'high',
            recoverable: false
        });

        throw new Error(`Chaos Monkey: ${errorType.statusText}`);
    };

    const createNetworkCorruption = async (input: RequestInfo | URL, originalFetch: Function, init?: RequestInit) => {
        const response = await originalFetch(input, init);

        if (response.headers.get('content-type')?.includes('application/json')) {
            const originalJson = response.json;
            response.json = async () => {
                const data = await originalJson.call(response);
                return corruptJson(data);
            };
        }

        logChaos({
            type: 'network.corruption',
            description: `Corrupted response data for ${input.toString()}`,
            duration: 0,
            impact: 'medium',
            recoverable: true
        });

        return response;
    };

    // Helper to corrupt JSON data (safely)
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

    // Utility functions
    const shouldCreateChaos = (type: string): boolean => {
        return config.enabled &&
            config.allowedChaos.has(type) &&
            Math.random() < config.probability;
    };

    const pickOne = <T>(arr: T[]): T => {
        return arr[Math.floor(Math.random() * arr.length)];
    };

    const logChaos = (event: Omit<ChaosEvent, 'id' | 'timestamp'>) => {
        const chaosEvent: ChaosEvent = {
            id: crypto.randomUUID(),
            timestamp: Date.now(),
            ...event
        };

        chaosHistory.push(chaosEvent);
        updateChaosUI();

        addLog({
            id: crypto.randomUUID(),
            timestamp: new Date().toISOString(),
            type: 'chaos',
            category: `chaos.${event.type}`,
            summary: event.description,
            details: chaosEvent
        });
    };

    // Update UI
    const updateChaosUI = () => {
        const content = container.querySelector('.chaos-content');
        if (!content) return;

        content.innerHTML = chaosHistory
            .slice(-10)
            .reverse()
            .map(event => `
                <div class="chaos-event ${event.impact}-impact ${event.recoverable ? 'recoverable' : 'non-recoverable'}">
                    <div class="chaos-event-description">
                        ${event.description}
                    </div>
                    <div class="chaos-event-meta">
                        Impact: ${event.impact} | 
                        ${event.recoverable ? '♻️ Recoverable' : '⚠️ Non-recoverable'} |
                        ${new Date(event.timestamp).toLocaleTimeString()}
                    </div>
                </div>
            `).join('');
    };

    // Initialize
    setupNetworkChaos();
    setupDOMChaos({ logChaos, shouldCreateChaos, config });
    const scriptChaos = setupScriptChaos({ logChaos, shouldCreateChaos, config });
    const cpuChaos = setupCPUChaos({ logChaos, shouldCreateChaos, config });
    const memoryChaos = setupMemoryChaos({ logChaos, shouldCreateChaos, config });
    const storageChaos = setupStorageChaos({ logChaos, shouldCreateChaos, config });
    const communicationChaos = setupCommunicationChaos({ logChaos, shouldCreateChaos, config });
    const eventChaos = setupEventChaos({ logChaos, shouldCreateChaos, config });
    const websocketChaos = setupWebSocketChaos({ logChaos, shouldCreateChaos, config });
    const urlChaos = setupURLChaos({ logChaos, shouldCreateChaos, config });
    const historyLocationChaos = setupHistoryLocationChaos({ logChaos, shouldCreateChaos, config });
    const cookieChaos = setupCookieChaos({ logChaos, shouldCreateChaos, config });
    const resourceTimingChaos = setupResourceTimingChaos({ logChaos, shouldCreateChaos, config });

    // Setup event listeners
    document.getElementById('toggle-chaos')?.addEventListener('click', () => {
        config.enabled = !config.enabled;
        const button = document.getElementById('toggle-chaos');
        if (button) {
            button.textContent = config.enabled ? 'Disable' : 'Enable';
            button.style.color = config.enabled ? '#ff4444' : '';
        }
    });

    document.getElementById('chaos-probability')?.addEventListener('input', (e) => {
        const value = parseInt((e.target as HTMLInputElement).value);
        config.probability = value / 100;
        const display = document.getElementById('probability-value');
        if (display) {
            display.textContent = `${value}%`;
        }
    });

    document.getElementById('safe-mode')?.addEventListener('change', (e) => {
        config.safeMode = (e.target as HTMLInputElement).checked;
    });

    container.querySelectorAll('.chaos-types input[type="checkbox"]').forEach(checkbox => {
        checkbox.addEventListener('change', (e) => {
            const target = e.target as HTMLInputElement;
            if (target.checked) {
                config.allowedChaos.add(target.value);
            } else {
                config.allowedChaos.delete(target.value);
            }
        });
    });

    return {
        getConfig: () => ({ ...config }),
        getHistory: () => [...chaosHistory],
        clearHistory: () => {
            chaosHistory.length = 0;
            updateChaosUI();
        },
        destroy: () => {
            config.enabled = false;
            scriptChaos.cleanup();
            memoryChaos.cleanup();
            storageChaos.cleanup();
            communicationChaos.cleanup();
            eventChaos.cleanup();
            websocketChaos.cleanup();
            urlChaos.cleanup();
            historyLocationChaos.cleanup();
            cookieChaos.cleanup();
            resourceTimingChaos.cleanup();
        }
    };
};