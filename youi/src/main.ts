import "@/styles.css";
import { createRouter } from '@/lib/router';
import { stateManager } from "@/lib/state";
import { EventManager } from "@/lib/event";

// Initialize the state manager and event manager
(async () => {
    await stateManager.init();
    console.log("State manager initialized");

    const eventManager = EventManager();
    eventManager.init();
    console.log("Event manager initialized");

    // Now these are globally available
    window.stateManager = stateManager;
    window.eventManager = eventManager;
})();

document.addEventListener("DOMContentLoaded", async () => {
    const main = document.getElementById('app');
    if (main) {
        const { router } = await createRouter();
        await router(main);
    }
});