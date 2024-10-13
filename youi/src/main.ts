import { StateManager } from "@/lib/state";
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
    }
}

window.addEventListener("DOMContentLoaded", () => {
    window.stateManager = StateManager();
    window.stateManager.init();

    router(document.body!);
});