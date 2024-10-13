declare global {
    interface Window {
        sharedWorker?: Worker;
    }
}

/*
getWorker retrieves the worker from the window object. If the worker is not yet initialized, 
it creates a new worker and assigns it to the window object. This ensures that the worker 
is a shared resource across different parts of the application, and it is initialized only once.
*/
export const getWorker = () => {
    if (!window.sharedWorker) {
        // The worker handles the state and event management. This is a lot more efficient, as
        // it can run in the background and not be blocked by, or block, the main thread.
        window.sharedWorker = new Worker(new URL('./worker.js', import.meta.url), { type: 'module' });
    }
    
    return window.sharedWorker;
};