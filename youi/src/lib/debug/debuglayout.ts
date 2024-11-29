interface Position {
    x: number;
    y: number;
}

interface OverlayState {
    element: HTMLElement;
    isDragging: boolean;
    dragOffset: Position;
    cleanup?: () => void;
}

interface CreateOverlayOptions {
    id: string;
    title: string;
    element: HTMLElement;
    defaultPosition?: Position;
}

export const createDebugLayout = () => {
    const overlayStates = new Map<string, OverlayState>();
    const Z_INDEX_BASE = 10000;

    const loadPosition = (id: string): Position => {
        try {
            const saved = localStorage.getItem(`debug-position-${id}`);
            return saved ? JSON.parse(saved) : { x: 20, y: 20 };
        } catch {
            return { x: 20, y: 20 };
        }
    };

    const savePosition = (id: string, position: Position): void => {
        localStorage.setItem(`debug-position-${id}`, JSON.stringify(position));
    };

    const createOverlay = ({
        id,
        title,
        element,
        defaultPosition
    }: CreateOverlayOptions): HTMLElement => {
        const position = loadPosition(id);

        // Create container
        const container = document.createElement('div');
        container.className = 'debug-overlay-container';

        // Create header
        const header = document.createElement('div');
        header.className = 'debug-header';

        const titleElement = document.createElement('span');
        titleElement.textContent = title;

        const controls = document.createElement('div');
        controls.style.cssText = 'display: flex; gap: 8px;';

        // Add minimize/maximize button
        const toggleButton = document.createElement('button');
        toggleButton.className = 'debug-button';
        toggleButton.textContent = '−';
        // toggleButton.onclick = () => {
        //     const content = container.querySelector('.debug-content') as HTMLElement;
        //     if (content) {
        //         if (content.style.display === 'none') {
        //             content.style.display = 'block';
        //             toggleButton.textContent = '−';
        //         } else {
        //             content.style.display = 'none';
        //             toggleButton.textContent = '+';
        //         }
        //     }
        // };

        // Add reset position button
        const resetButton = document.createElement('button');
        resetButton.className = 'debug-button';
        resetButton.textContent = 'Reset';
        // resetButton.onclick = () => {
        //     const defaultPos = defaultPosition || { x: 20, y: 20 };
        //     container.style.left = `${defaultPos.x}px`;
        //     container.style.top = `${defaultPos.y}px`;
        //     savePosition(id, defaultPos);
        // };

        controls.appendChild(toggleButton);
        controls.appendChild(resetButton);
        header.appendChild(titleElement);
        header.appendChild(controls);

        // Add content
        const contentWrapper = document.createElement('div');
        contentWrapper.className = 'debug-content';
        contentWrapper.appendChild(element);

        container.appendChild(header);
        container.appendChild(contentWrapper);

        // Setup dragging and store state
        overlayStates.set(id, {
            element: container,
            isDragging: false,
            dragOffset: { x: 0, y: 0 }
        });

        // Store cleanup function
        const state = overlayStates.get(id);
        if (state) {
            state.cleanup = cleanup;
        }

        document.body.appendChild(container);
        return container;
    };

    const cleanup = () => {
        overlayStates.forEach((state, id) => {
            state.cleanup?.();
            state.element.remove();
        });
        overlayStates.clear();
    };

    return {
        createOverlay,
        cleanup
    };
};