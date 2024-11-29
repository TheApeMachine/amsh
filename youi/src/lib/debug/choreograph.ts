interface Dancer {
    element: HTMLElement;
    role: 'lead' | 'support' | 'ensemble';
    energy: number;
    moves: string[];
    partners: Set<Dancer>;
    stage: { x: number; y: number; z: number };
}

interface ChoreographyMove {
    name: string;
    duration: number;
    keyframes: Keyframe[];
    timing: KeyframeAnimationOptions;
}

export const setupDebugChoreographer = ({ addLog, overlay }: { addLog: Function; overlay: HTMLElement }) => {
    const dancers = new Map<HTMLElement, Dancer>();
    let isPerforming = false;
    let musicContext: AudioContext | null = null;
    let danceObserver: MutationObserver | null = null;
    const danceBook = new Map<string, ChoreographyMove>();

    // Create stage container
    const stage = document.createElement('div');
    stage.className = 'debug-stage';
    stage.style.width = '100%';
    stage.style.height = '300px';

    // Create controls container
    const controls = document.createElement('div');
    controls.className = 'debug-header-controls';
    controls.innerHTML = `
        <button class="debug-button" id="choreo-toggle">🎭 Show</button>
        <button class="debug-button" id="choreo-start">💃 Start</button>
        <button class="debug-button" id="choreo-style">🎨 Style</button>
    `;
    stage.appendChild(controls);

    // Add floor for perspective
    const floor = document.createElement('div');
    floor.className = 'debug-stage-floor';
    stage.appendChild(floor);

    overlay.appendChild(stage);

    // Throttle function for performance
    const throttle = <T extends (...args: any[]) => void>(
        func: T,
        limit: number
    ): T => {
        let inThrottle = false;
        return ((...args: Parameters<T>) => {
            if (!inThrottle) {
                func(...args);
                inThrottle = true;
                setTimeout(() => (inThrottle = false), limit);
            }
        }) as T;
    };

    // Helper functions
    const createVisualElement = (dancer: Dancer) => {
        const visual = document.createElement('div');
        visual.className = 'dancer-visual';
        visual.dataset.dancerId = dancer.element.tagName +
            (dancer.element.id || dancer.element.className).slice(0, 32); // Limit ID length
        return visual;
    };

    const createDancer = (element: HTMLElement): Dancer => ({
        element,
        role: element === document.body ? 'lead' : 'ensemble',
        energy: Math.random(),
        moves: ['spin', 'jump', 'wave'],
        partners: new Set(),
        stage: {
            x: Math.random() * 300,
            y: Math.random() * 200,
            z: Math.random() * 100
        }
    });

    const updateDancerPosition = (dancer: Dancer, visual: HTMLElement) => {
        requestAnimationFrame(() => {
            visual.style.transform =
                `translate3d(${dancer.stage.x}px, ${dancer.stage.y}px, ${dancer.stage.z}px)`;
        });
    };

    // Batch process nodes
    const processMutations = throttle((mutations: MutationRecord[]) => {
        const addedNodes = new Set<Node>();
        const removedNodes = new Set<Node>();

        mutations.forEach(mutation => {
            if (mutation.type === 'childList') {
                mutation.addedNodes.forEach(node => addedNodes.add(node));
                mutation.removedNodes.forEach(node => removedNodes.add(node));
            }
        });

        // Process in batches
        requestAnimationFrame(() => {
            addedNodes.forEach(node => {
                if (node instanceof HTMLElement && !dancers.has(node)) {
                    addDancer(node);
                }
            });

            removedNodes.forEach(node => {
                if (node instanceof HTMLElement) {
                    removeDancer(node);
                }
            });
        });
    }, 100); // Throttle to 100ms

    const addDancer = (element: HTMLElement) => {
        if (!element || dancers.has(element)) return;

        try {
            const dancer = createDancer(element);
            dancers.set(element, dancer);
            const visual = createVisualElement(dancer);
            stage.appendChild(visual);
            updateDancerPosition(dancer, visual);
        } catch (error) {
            console.warn('Failed to add dancer:', error);
        }
    };

    const removeDancer = (element: HTMLElement) => {
        if (!element) return;

        const dancer = dancers.get(element);
        if (!dancer) return;

        const visual = stage.querySelector(
            `[data-dancer-id="${element.tagName}${(element.id || element.className).slice(0, 32)
            }"]`
        );
        visual?.remove();
        dancers.delete(element);
    };

    // Initialize controls
    const initializeControls = () => {
        const toggleBtn = controls.querySelector('#choreo-toggle');
        const startBtn = controls.querySelector('#choreo-start');
        const styleBtn = controls.querySelector('#choreo-style');

        toggleBtn?.addEventListener('click', () => {
            const isHidden = stage.style.display === 'none';
            stage.style.display = isHidden ? 'block' : 'none';
            (toggleBtn as HTMLElement).textContent = isHidden ? '🎭 Hide' : '🎭 Show';
        });

        startBtn?.addEventListener('click', () => {
            isPerforming = !isPerforming;
            (startBtn as HTMLElement).textContent = isPerforming ? '⏸ Pause' : '💃 Start';
        });

        let currentStyle = 0;
        const styles = ['matrix', 'underwater', 'disco'];

        styleBtn?.addEventListener('click', () => {
            currentStyle = (currentStyle + 1) % styles.length;
            requestAnimationFrame(() => {
                stage.querySelectorAll('.dancer-visual').forEach(dancer => {
                    styles.forEach(style => dancer.classList.remove(`${style}-style`));
                    dancer.classList.add(`${styles[currentStyle]}-style`);
                });
            });
        });
    };

    // Initialize observer with performance optimizations
    const initializeObserver = () => {
        danceObserver = new MutationObserver(processMutations);

        try {
            danceObserver.observe(document.body, {
                childList: true,
                subtree: true,
                attributes: false, // Ignore attribute changes
                characterData: false // Ignore text changes
            });
        } catch (error) {
            console.warn('Failed to initialize observer:', error);
        }
    };

    // Defer initialization
    queueMicrotask(() => {
        initializeControls();
        initializeObserver();
    });

    return {
        getStage: () => stage,
        getDancers: () => dancers,
        addMove: (name: string, choreography: ChoreographyMove) => {
            danceBook.set(name, choreography);
        },
        destroy: () => {
            isPerforming = false;
            danceObserver?.disconnect();
            stage.remove();
            musicContext?.close();
            dancers.clear();
        }
    };
};