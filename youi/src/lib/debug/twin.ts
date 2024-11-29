// digital-twin.ts

interface PredictionResult {
    id: string;
    timestamp: number;
    type: 'warning' | 'error' | 'info';
    probability: number;
    description: string;
    impact: 'low' | 'medium' | 'high';
    context: any;
}

interface StateSnapshot {
    dom: string;
    storage: Record<string, any>;
    network: {
        pending: Map<string, Request>;
        cache: Map<string, Response>;
    };
    userActions: string[];
}

interface SerializableStateSnapshot {
    dom: string;
    storage: Record<string, any>;
    network: {
        pending: Array<{
            url: string;
            method: string;
            headers: Record<string, string>;
        }>;
        cache: Array<{
            url: string;
            status: number;
            headers: Record<string, string>;
            body?: string;
        }>;
    };
    userActions: string[];
}

export const setupDigitalTwin = ({ addLog, overlay }: { addLog: Function; overlay: HTMLElement }) => {
    let isActive = false;
    let userPatterns: Map<string, number> = new Map();
    let predictedStates: Map<string, StateSnapshot> = new Map();
    let currentSimulation: Worker | null = null;

    // Create UI within the provided overlay
    const container = document.createElement('div');
    container.className = 'debug-twin';
    overlay.appendChild(container);

    // Add header and controls
    container.innerHTML = `
        <div class="twin-header">
            <div class="twin-header-content">
                <h3>Digital Twin</h3>
                <div class="twin-controls">
                    <button class="debug-button" id="start-prediction">▶️ Start</button>
                    <button class="debug-button" id="clear-patterns">🧹 Clear</button>
                </div>
            </div>
            <div class="twin-stats"></div>
        </div>
        <div class="twin-content"></div>
    `;

    // Create simulation worker
    const createSimulationWorker = () => {
        const workerCode = `
            self.onmessage = function(e) {
                const { state, patterns, config } = e.data;
                
                // Internal shouldCreateChaos function
                const shouldCreateChaos = () => {
                    return config.chaosEnabled && Math.random() < config.chaosProbability;
                };
                
                // Convert patterns back to a usable format
                const patternMap = new Map(
                    patterns.map(p => [p.pattern, p.frequency])
                );
                
                // Simulate user actions based on patterns
                const simulateUserActions = (state, patterns) => {
                    const predictions = [];
                    
                    patterns.forEach((frequency, pattern) => {
                        if (shouldCreateChaos()) {
                            // Predict likely next actions
                            const nextActions = predictNextActions(pattern, frequency);
                            
                            // Simulate each action
                            nextActions.forEach(action => {
                                try {
                                    const result = simulateAction(state, action);
                                    if (result.warning) {
                                        predictions.push({
                                            type: 'warning',
                                            probability: result.probability,
                                            description: result.warning,
                                            impact: result.impact
                                        });
                                    }
                                } catch (error) {
                                    predictions.push({
                                        type: 'error',
                                        probability: 0.8,
                                        description: error.message,
                                        impact: 'high'
                                    });
                                }
                            });
                        }
                    });
                    
                    return predictions;
                };
                
                const predictNextActions = (pattern, frequency) => {
                    // Simple Markov chain-like prediction
                    const actions = pattern.split('->');
                    const lastAction = actions[actions.length - 1];
                    
                    // Predict based on frequency and pattern
                    return [
                        { type: 'click', target: lastAction, probability: frequency / 100 },
                        { type: 'input', target: lastAction, probability: frequency / 200 },
                        { type: 'navigate', target: '/' + lastAction, probability: frequency / 150 }
                    ];
                };
                
                const simulateAction = (state, action) => {
                    // Simulate action and detect potential issues
                    const result = { probability: action.probability };
                    
                    switch (action.type) {
                        case 'click':
                            if (state.network.pending.size > 3) {
                                result.warning = 'Potential race condition detected';
                                result.impact = 'medium';
                            }
                            break;
                            
                        case 'input':
                            if (state.userActions.length > 10) {
                                result.warning = 'High frequency user input detected';
                                result.impact = 'low';
                            }
                            break;
                            
                        case 'navigate':
                            if (state.network.pending.size > 0) {
                                result.warning = 'Navigation during pending requests';
                                result.impact = 'high';
                            }
                            break;
                    }
                    
                    return result;
                };
                
                // Run simulation with converted data
                const predictions = simulateUserActions(state, patternMap);
                self.postMessage(predictions);
            };
        `;

        const blob = new Blob([workerCode], { type: 'application/javascript' });
        return new Worker(URL.createObjectURL(blob));
    };

    // Track user actions to build patterns
    const trackUserAction = (action: string) => {
        if (!isActive) return;

        // Build pattern from recent actions
        const recentActions = Array.from(userPatterns.keys()).slice(-2);
        recentActions.push(action);
        const pattern = recentActions.join('->');

        // Update pattern frequency
        userPatterns.set(pattern, (userPatterns.get(pattern) ?? 0) + 1);

        // Trigger prediction
        predictFuture();
    };

    // Capture current state
    const captureState = (): StateSnapshot => ({
        dom: document.documentElement.outerHTML,
        storage: {
            local: { ...localStorage },
            session: { ...sessionStorage }
        },
        network: {
            pending: new Map(),
            cache: new Map()
        },
        userActions: Array.from(userPatterns.keys())
    });

    // Run prediction simulation
    const predictFuture = async () => {
        if (!isActive || userPatterns.size === 0) return;

        // Capture the current state first
        const currentState = captureState();

        // Create a serializable version of the state
        const state: SerializableStateSnapshot = {
            dom: currentState.dom,
            storage: currentState.storage,
            network: {
                pending: Array.from(currentState.network.pending.entries()).map(([url, request]) => ({
                    url,
                    method: request.method,
                    headers: Object.fromEntries(request.headers.entries())
                })),
                cache: Array.from(currentState.network.cache.entries()).map(([url, response]) => ({
                    url,
                    status: response.status,
                    headers: Object.fromEntries(response.headers.entries())
                }))
            },
            userActions: currentState.userActions
        };

        try {
            // Create new worker for this simulation
            if (currentSimulation) {
                currentSimulation.terminate();
            }

            currentSimulation = createSimulationWorker();
            currentSimulation.onmessage = (e: MessageEvent) => {
                const predictions: PredictionResult[] = e.data;
                updatePredictions(predictions);
            };

            // Send only serializable data
            currentSimulation.postMessage({
                state: JSON.parse(JSON.stringify(state)),
                patterns: Array.from(userPatterns.entries()).map(([key, value]) => ({
                    pattern: key,
                    frequency: value
                })),
                // Add configuration instead of functions
                config: {
                    chaosEnabled: true,
                    chaosProbability: 0.2
                }
            });
        } catch (error) {
            console.error('Failed to send message to worker:', error);
        }
    };

    // Update UI with predictions
    const updatePredictions = (predictions: PredictionResult[]) => {
        const content = container.querySelector('.twin-content');
        if (!content) return;

        const getTypeColor = (type: string) => {
            if (type === 'error') return '#ff4444';
            if (type === 'warning') return '#ffaa00';
            return '#4444ff';
        };

        content.innerHTML = predictions.map(prediction => `
            <div style="
                margin-bottom: 8px;
                padding: 8px;
                background: rgba(255, 255, 255, 0.05);
                border-radius: 4px;
                border-left: 3px solid ${getTypeColor(prediction.type)};
            ">
                <div style="display: flex; justify-content: space-between;">
                    <span style="color: #fff;">${prediction.description}</span>
                    <span style="color: #666;">${(prediction.probability * 100).toFixed(1)}%</span>
                </div>
                <div style="color: #666; font-size: 11px; margin-top: 4px;">
                    Impact: ${prediction.impact}
                </div>
            </div>
        `).join('');

        // Update stats
        const stats = container.querySelector('.twin-stats');
        if (stats) {
            stats.innerHTML = `
                Patterns: ${userPatterns.size} | 
                Predictions: ${predictions.length} |
                Confidence: ${predictions.reduce((acc, p) => acc + p.probability, 0) / predictions.length * 100
                }%
            `;
        }
    };

    // Setup event listeners for user actions
    const setupActionTracking = () => {
        document.addEventListener('click', (e) => {
            const element = e.target as HTMLElement;
            trackUserAction(`click:${element.tagName}:${element.className}`);
        });

        document.addEventListener('input', (e) => {
            const element = e.target as HTMLElement;
            trackUserAction(`input:${element.tagName}:${element.id || element.className}`);
        });

        window.addEventListener('popstate', () => {
            trackUserAction(`navigate:${location.pathname}`);
        });
    };

    // Setup UI controls
    document.getElementById('toggle-twin')?.addEventListener('click', () => {
        container.style.display = container.style.display === 'none' ? 'block' : 'none';
    });

    document.getElementById('start-prediction')?.addEventListener('click', () => {
        isActive = !isActive;
        const button = document.getElementById('start-prediction');
        if (button) {
            button.textContent = isActive ? '⏸ Pause' : '▶️ Start';
            button.style.color = isActive ? '#ffaa00' : '';
        }
    });

    document.getElementById('clear-patterns')?.addEventListener('click', () => {
        userPatterns.clear();
        predictedStates.clear();
        const content = container.querySelector('.twin-content');
        if (content) {
            content.innerHTML = '<div style="color: #666; text-align: center;">No predictions yet</div>';
        }
    });

    // Initialize
    setupActionTracking();

    return {
        getPredictions: () => predictedStates,
        getPatterns: () => userPatterns,
        destroy: () => {
            if (currentSimulation) {
                currentSimulation.terminate();
            }
            container.remove();
        }
    };
};