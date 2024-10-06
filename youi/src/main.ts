import { StateManager } from "@/lib/state";
import { EventManager } from "@/lib/event";
import {router} from "@/router.ts";
import "@/components/root.ts";
import "@/components/ui/gridspace";
import "@/components/error";
import "@/components/loader";
import "@/components/toast/container";
import "@/components/toast/message";

declare global {
    interface Window {
        stateManager: ReturnType<typeof StateManager>;
        eventManager: ReturnType<typeof EventManager>;
    }
}

window.addEventListener("DOMContentLoaded", () => {
    window.stateManager = StateManager();
    window.stateManager.init();

    window.eventManager = EventManager();
    window.eventManager.init();

    router(document.body!);
});