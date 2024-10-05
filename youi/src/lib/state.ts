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
            getState(effect)();
        }
    };

    const registry: Record<string, any> = {};

    const getState = (key: string) => {
        const stateValue = state[key];
        const registryValue = stateValue ? undefined : registry[key];

        if (!stateValue) {
            console.error("lib.StateManager.get", "No state value found for key", key);
        }

        return stateValue ?? registryValue;
    };

    const setState = (stateFragment: Record<string, any>) => {
        Object.assign(state, stateFragment);
        persist();
    };

    const register = (key: string, value: any) => {
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
        init
    };
}

export default StateManager;