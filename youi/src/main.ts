import { stateManager } from "@/lib/state";
import {router} from "@/router.ts";
import "@/components/root.ts";
import "@/components/ui/gridspace";
import "@/components/error";
import "@/components/loader";
import "@/components/toast/container";
import "@/components/toast/message";

declare global {
    interface Window {
        stateManager: typeof stateManager;
    }
}

window.addEventListener("DOMContentLoaded", () => {
    window.stateManager = stateManager;
    window.stateManager.init();

    router(document.body!);
});