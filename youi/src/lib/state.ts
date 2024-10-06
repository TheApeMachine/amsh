import localforage from 'localforage';
import { getWorker } from "@/workers/utils";

export interface StateModel {
    eventHandlers: Record<string, (event: any) => void>;
}

export const StateManager = () => {
    const state: Record<string, any> = {
        ready: false
    };

    const worker = getWorker();

    worker.onmessage = (event: MessageEvent) => {
        console.log("lib.StateManager.onmessage", event);
        const { topic, effect } = event.data;
        if (topic === "state" && effect) {
            const fn = getState(effect)
            console.log("lib.StateManager.onmessage", fn);
            fn();
        }
    };

    const registry: Record<string, any> = {};

    const getState = (key: string) => {
        const stateValue = state[key];
        const registryValue = stateValue ? undefined : registry[key];

        if (!stateValue && !registryValue) {
            console.error("lib.StateManager.get", "No state value found for key", key);
        }

        return stateValue ?? registryValue;
    };

    const setState = (stateFragment: Record<string, any>) => {
        Object.assign(state, stateFragment);
        persist();
    };

    const register = (key: string, value: any) => {
        console.log("lib.StateManager.register", key, value);
        registry[key] = value;
    };

    const persist = () => {
        localforage.setItem('state', state);
    };

    const init = async (): Promise<void> => {
        try {
            const value = await localforage.getItem('state');
            if (value) {
                Object.assign(state, value);
            }
            console.log("lib.StateManager.init", state);
            setState({ ready: true });
        } catch (error) {
            console.error("Error initializing StateManager:", error);
            setState({ ready: false });
            throw error;
        }
    };

    return {
        getState,
        setState,
        register,
        init,
        registry
    };
}

export default StateManager;