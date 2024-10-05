import { getWorker } from "../workers/utils";

export const EventManager = () => {
    const init = () => {[
        "click", 
        "wheel",
        "keydown", 
        "drag"
    ].forEach(eventType => {
        console.log("lib.EventManager.init", eventType);
        document.addEventListener(eventType, on);
    })};

    const on = (event: any) => {
        console.log("lib.EventManager.on", event);
        const path = event.path || (event.composedPath && event.composedPath());

        if (!path) {
            console.error("lib.EventManager.on", "No path found for event", event);
            return;
        }

        for (const element of path) {
            console.log("lib.EventManager.on", "element", element);
            if (
                element instanceof HTMLElement 
                && element.hasAttribute('data-event') 
                && element.hasAttribute('data-effect')
            ) {
                getWorker().postMessage({
                    event: element.getAttribute('data-event'),
                    effect: element.getAttribute('data-effect'),
                    topic: element.getAttribute('data-topic')
                });

                break;
            }
        }
    }

    return { init }
}

export default EventManager;