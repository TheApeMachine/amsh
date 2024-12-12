type SmellSeverity = 'low' | 'medium' | 'high';

interface SmellDetector {
    name: string;
    description: string;
    severity: SmellSeverity;
    check: () => SmellInstance[];
}

interface SmellInstance {
    type: string;
    location: string;
    description: string;
    severity: SmellSeverity;
    timestamp: number;
    context: any;
    suggestion: string;
}

interface PerformanceMetrics {
    renderTimes: number[];
    memoryUsage: number[];
    eventHandlers: Map<string, number>;
    networkDupes: Map<string, Set<string>>;
    stateChanges: Map<string, number>;
}

export const setupCodeSmellDetector = ({ addLog, overlay }: { addLog: Function; overlay: HTMLElement }) => {
    const metrics: PerformanceMetrics = {
        renderTimes: [],
        memoryUsage: [],
        eventHandlers: new Map(),
        networkDupes: new Map(),
        stateChanges: new Map()
    };

    // Create visualization container
    const container = document.createElement('div');
    container.className = 'debug-smells';
    overlay.appendChild(container);

    const header = document.createElement('div');
    header.innerHTML = `
        <div class="smells-header">
            <div class="smells-header-content">
                <h3>Code Smells Detector</h3>
                <button class="debug-button" id="toggle-smells">👃 Hide</button>
            </div>
            <div class="smell-metrics"></div>
        </div>
    `;
    container.appendChild(header);

    const content = document.createElement('div');
    content.className = 'smells-content';
    container.appendChild(content);

    // Define smell detectors
    const detectors: SmellDetector[] = [
        {
            name: 'Memory Leak Detection',
            description: 'Detects potential memory leaks from growing arrays/objects',
            severity: 'high',
            check: () => {
                const instances: SmellInstance[] = [];
                const usage = (performance as any).memory?.usedJSHeapSize || 0;
                metrics.memoryUsage.push(usage);

                if (metrics.memoryUsage.length > 10) {
                    const growth = metrics.memoryUsage[metrics.memoryUsage.length - 1] -
                        metrics.memoryUsage[metrics.memoryUsage.length - 10];
                    if (growth > 5000000) { // 5MB growth
                        instances.push({
                            type: 'memory-leak',
                            location: 'heap',
                            description: 'Significant memory growth detected',
                            severity: 'high',
                            timestamp: Date.now(),
                            context: {
                                growth: `${(growth / 1024 / 1024).toFixed(2)}MB`,
                                trend: metrics.memoryUsage.slice(-10)
                            },
                            suggestion: 'Check for unbounded arrays, event listeners, or cache structures'
                        });
                    }
                }
                return instances;
            }
        },
        {
            name: 'Duplicate Network Requests',
            description: 'Detects repeated network requests to the same endpoint',
            severity: 'medium',
            check: () => {
                const instances: SmellInstance[] = [];
                metrics.networkDupes.forEach((timestamps, url) => {
                    const recent = [...timestamps].filter(t => Date.now() - parseInt(t) < 5000);
                    if (recent.length > 2) {
                        instances.push({
                            type: 'network-spam',
                            location: url,
                            description: `${recent.length} duplicate requests in 5s`,
                            severity: 'medium',
                            timestamp: Date.now(),
                            context: {
                                url,
                                timestamps: recent
                            },
                            suggestion: 'Consider implementing request debouncing or caching'
                        });
                    }
                });
                return instances;
            }
        },
        {
            name: 'Event Handler Accumulation',
            description: 'Detects potential event handler leaks',
            severity: 'medium',
            check: () => {
                const instances: SmellInstance[] = [];
                metrics.eventHandlers.forEach((count, element) => {
                    if (count > 3) {
                        const [elementInfo, eventType] = element.split(':');
                        instances.push({
                            type: 'event-accumulation',
                            location: elementInfo,
                            description: `${count} '${eventType}' handlers on ${elementInfo}`,
                            severity: 'medium',
                            timestamp: Date.now(),
                            context: {
                                element: elementInfo,
                                eventType,
                                count
                            },
                            suggestion: 'Ensure event handlers are properly removed when components unmount'
                        });
                    }
                });
                return instances;
            }
        },
        {
            name: 'Rapid State Changes',
            description: 'Detects possibly unnecessary state updates',
            severity: 'low',
            check: () => {
                const instances: SmellInstance[] = [];
                metrics.stateChanges.forEach((count, key) => {
                    if (count > 5) {
                        instances.push({
                            type: 'state-churn',
                            location: key,
                            description: `${count} state updates in 1s`,
                            severity: 'low',
                            timestamp: Date.now(),
                            context: {
                                key,
                                count
                            },
                            suggestion: 'Consider debouncing or batching state updates'
                        });
                    }
                });
                return instances;
            }
        }
    ];

    // Track network requests
    const originalFetch = window.fetch;
    window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
        let url: string;
        if (typeof input === 'string') {
            url = input;
        } else if (input instanceof URL) {
            url = input.href;
        } else {
            url = input.url;
        }
        const timestamps = metrics.networkDupes.get(url) || new Set();
        timestamps.add(Date.now().toString());
        metrics.networkDupes.set(url, timestamps);
        return originalFetch(input, init);
    };

    // Track event listeners
    const originalAddEventListener = EventTarget.prototype.addEventListener;
    EventTarget.prototype.addEventListener = function (type: string, listener: EventListener) {
        let elementInfo = '';
        if (this instanceof Element) {
            const element = this as Element;

            // Get all possible identifying information
            const id = element.id ? `#${element.id}` : '';
            const classes = (typeof element.className === 'string' && element.className)
                ? `.${element.className.trim().split(/\s+/).join('.')}`
                : '';
            const tag = element.tagName?.toLowerCase() || 'unknown';

            // Add data attributes if available
            const dataAttrs = Array.from(element.attributes)
                .filter(attr => attr.name.startsWith('data-'))
                .map(attr => `[${attr.name}="${attr.value}"]`)
                .join('');

            // Add role if available
            const role = element.getAttribute('role') ? `[role="${element.getAttribute('role')}"]` : '';

            // Combine all identifiers
            elementInfo = `<${tag}${id}${classes}${role}${dataAttrs}>`;

            // Fallback if we still don't have any identifying information
            if (elementInfo === `<${tag}>`) {
                // Try to get some context from parent elements
                const parent = element.parentElement;
                let parentInfo = '';
                if (parent) {
                    const parentId = parent.id ? `#${parent.id}` : '';
                    const parentTag = parent.tagName.toLowerCase();
                    parentInfo = ` (child of <${parentTag}${parentId}>)`;
                }
                elementInfo += parentInfo;
            }
        } else {
            // For non-Element EventTargets (e.g., window, document)
            elementInfo = String(this) || this.constructor.name || 'unknown';
        }

        const key = `${elementInfo}:${type}`;
        metrics.eventHandlers.set(key, (metrics.eventHandlers.get(key) ?? 0) + 1);
        return originalAddEventListener.apply(this, [type, listener]);
    };

    // Track render times using requestAnimationFrame
    let lastFrameTime = performance.now();
    const measureFrame = () => {
        const now = performance.now();
        const frameDuration = now - lastFrameTime;
        metrics.renderTimes.push(frameDuration);
        if (metrics.renderTimes.length > 60) { // Keep last second
            metrics.renderTimes.shift();
        }
        lastFrameTime = now;
        requestAnimationFrame(measureFrame);
    };
    requestAnimationFrame(measureFrame);

    // Create heat map visualization
    const createHeatmap = (smells: SmellInstance[]) => {
        const heatmap = document.createElement('div');
        heatmap.style.cssText = `
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(20px, 1fr));
            gap: 2px;
            margin-top: 8px;
        `;

        smells.forEach(smell => {
            const cell = document.createElement('div');
            cell.style.cssText = `
                width: 20px;
                height: 20px;
                border-radius: 4px;
                background: ${getSeverityColor(smell.severity)};
                cursor: pointer;
            `;
            cell.title = `${smell.type}: ${smell.description}`;

            // cell.addEventListener('click', () => {
            //     addLog({
            //         id: crypto.randomUUID(),
            //         timestamp: new Date().toISOString(),
            //         type: 'code-smell',
            //         category: `smell.${smell.type}`,
            //         summary: smell.description,
            //         details: {
            //             ...smell,
            //             detected: new Date(smell.timestamp).toLocaleTimeString()
            //         }
            //     });
            // });

            heatmap.appendChild(cell);
        });

        return heatmap;
    };

    // Update metrics display
    const updateMetrics = () => {
        const metricsDiv = container.querySelector('.smell-metrics');
        if (!metricsDiv) return;

        const avgFrameTime = metrics.renderTimes.reduce((a, b) => a + b, 0) / metrics.renderTimes.length;
        const memoryTrend = metrics.memoryUsage.length > 1 ?
            (metrics.memoryUsage[metrics.memoryUsage.length - 1] - metrics.memoryUsage[metrics.memoryUsage.length - 2]) : 0;

        metricsDiv.innerHTML = `
            <div style="display: flex; gap: 16px; color: #fff; opacity: 0.8;">
                <div>FPS: ${(1000 / avgFrameTime).toFixed(1)}</div>
                <div>Memory: ${(memoryTrend / 1024 / 1024).toFixed(1)}MB/s</div>
                <div>Events: ${metrics.eventHandlers.size}</div>
                <div>Network: ${metrics.networkDupes.size}</div>
            </div>
        `;
    };

    // Update visualization
    let isVisible = true;
    const updateVisuals = () => {
        if (!isVisible) return;

        const smells = detectors.flatMap(detector => detector.check());

        // Update content
        const contentDiv = container.querySelector('.smells-content');
        if (contentDiv) {
            contentDiv.innerHTML = '';
            if (smells.length > 0) {
                contentDiv.appendChild(createHeatmap(smells));

                // Add details
                const details = document.createElement('div');
                details.style.marginTop = '16px';
                details.innerHTML = smells.map(smell => `
                    <div style="
                        margin-bottom: 8px;
                        padding: 8px;
                        background: rgba(255, 255, 255, 0.05);
                        border-radius: 4px;
                    ">
                        <div style="color: ${getSeverityTextColor(smell.severity)};">
                            ${smell.type}
                        </div>
                        <div style="color: #fff; margin: 4px 0;">
                            ${smell.description}
                        </div>
                        <div style="color: #666; font-size: 11px;">
                            ${smell.suggestion}
                        </div>
                    </div>
                `).join('');
                contentDiv.appendChild(details);
            } else {
                contentDiv.innerHTML = `
                    <div style="color: #666; text-align: center; padding: 20px;">
                        No code smells detected! 🌹
                    </div>
                `;
            }
        }

        updateMetrics();
    };

    // Setup update interval
    const updateInterval = setInterval(updateVisuals, 1000);

    // Handle visibility toggle
    document.getElementById('toggle-smells')?.addEventListener('click', () => {
        isVisible = !isVisible;
        container.style.display = isVisible ? 'block' : 'none';
        const button = document.getElementById('toggle-smells');
        if (button) {
            button.textContent = isVisible ? '👃 Hide' : '👃 Show';
        }
    });

    return {
        getMetrics: () => metrics,
        addDetector: (detector: SmellDetector) => detectors.push(detector),
        destroy: () => {
            clearInterval(updateInterval);
            container.remove();
            window.fetch = originalFetch;
            EventTarget.prototype.addEventListener = originalAddEventListener;
        }
    };
};

const getSeverityColor = (severity: 'high' | 'medium' | 'low') => {
    if (severity === 'high') return 'rgba(255, 0, 0, 0.5)';
    if (severity === 'medium') return 'rgba(255, 165, 0, 0.5)';
    return 'rgba(255, 255, 0, 0.5)';
};

const getSeverityTextColor = (severity: SmellSeverity) => {
    if (severity === 'high') return '#ff4444';
    if (severity === 'medium') return '#ffaa00';
    return '#ffff00';
};