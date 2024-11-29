import { DebugEntry } from "./types";

export const setupEventListeners = ({ logs, expandedEntries, updateDisplay }: { logs: DebugEntry[], expandedEntries: Set<string>, updateDisplay: () => void }) => {
    document.getElementById('clear-logs')?.addEventListener('click', () => {
        logs.length = 0;
        updateDisplay();
    });

    document.getElementById('expand-all')?.addEventListener('click', () => {
        if (expandedEntries.size === logs.length) {
            expandedEntries.clear();
        } else {
            logs.forEach(entry => expandedEntries.add(entry.id));
        }
        updateDisplay();
    });
};