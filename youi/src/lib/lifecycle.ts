interface LifecycleHandler {
    onMount?: () => void;
    onUnmount?: () => void;
}

const lifecycleHandlers: WeakMap<Node, LifecycleHandler> = new WeakMap();

export const onMount = (element: HTMLElement, handler: () => void) => {
    const wrappedHandler = () => {
        handler();
        // If this element has a shadow root, observe it too
        if (element.shadowRoot) {
            observeShadowRoot(element.shadowRoot);
        }
    };
    
    if (element.isConnected) {
        wrappedHandler();
    } else {
        element.addEventListener('connected', wrappedHandler, { once: true });
    }
};

export const onUnmount = (element: HTMLElement, handler: () => void) => {
    element.addEventListener('disconnected', handler, { once: true });
};

const observeShadowRoot = (shadowRoot: ShadowRoot) => {
    const observer = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
            mutation.addedNodes.forEach((node) => {
                if (node instanceof HTMLElement) {
                    node.dispatchEvent(new CustomEvent('connected'));
                }
            });
            mutation.removedNodes.forEach((node) => {
                if (node instanceof HTMLElement) {
                    node.dispatchEvent(new CustomEvent('disconnected'));
                }
            });
        });
    });
    observer.observe(shadowRoot, { childList: true, subtree: true });
};

// Function to trigger the "onMount" lifecycle when a node is added to the DOM
const triggerMount = (node: Node) => {
    const handlers = lifecycleHandlers.get(node);
    if (handlers?.onMount) {
        handlers.onMount();
    }
};

// Function to trigger the "onUnmount" lifecycle when a node is removed from the DOM
const triggerUnmount = (node: Node) => {
    const handlers = lifecycleHandlers.get(node);
    if (handlers?.onUnmount) {
        handlers.onUnmount();
    }
};

const observer = new MutationObserver((mutationsList) => {
    mutationsList.forEach((mutation) => {
        mutation.addedNodes.forEach((node) => triggerMount(node));
        mutation.removedNodes.forEach((node) => triggerUnmount(node));
    });
});

// Start observing the document for added/removed nodes
observer.observe(
    document.body, 
    { childList: true, subtree: true }
);