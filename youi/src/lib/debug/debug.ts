import "@/lib/debug.css";
import { setupXHRTracking } from "@/lib/debug/xhr";
import { setupMutationObserver } from "./debug/mutation";
import { setupEventListeners } from "./debug/event";
import { setupConsoleTracking } from "./debug/console";
import { setupFetchTracking } from "./debug/fetch";
import { setupStorageTracking } from "./debug/storage";
import { createJsonViewer } from "./debug/jsonview";
import { setupTimeTravel } from './debug/timetravel';
import { setup3DNetworkView } from './debug/networkgraph';
import { setupCodeSmellDetector } from './debug/codesmells';
import { setupDigitalTwin } from './debug/twin';
import { setupChaosMonkey } from './debug/chaosmonkey';
//import { setupDebugChoreographer } from "./debug/choreograph";
interface DebugEntry {
    id: string;
    timestamp: string;
    type: string;
    category: string;
    summary: string;
    details: any;
    stack?: string;
}

type ConsoleMethod = keyof Console;

export const createDebugOverlay = () => {
    let logs: DebugEntry[] = [];
    let expandedEntries = new Set<string>();
    let observer: MutationObserver;
    const originalConsole: Record<string, any> = {};
    let isMinimized = false;

    const addLog = (entry: DebugEntry) => {
        logs.unshift(entry);
        if (logs.length > 200) logs.pop();
        updateDisplay();
    };

    // Create main container
    const container = document.createElement('div');
    container.className = 'debug-overlay';

    // Create header with controls
    const header = document.createElement('div');
    header.className = 'debug-header';

    const title = document.createElement('span');
    title.textContent = 'Debug Tools';
    title.className = 'debug-header-title';

    const controls = document.createElement('div');
    controls.className = 'debug-header-controls';
    controls.innerHTML = `
        <button class="debug-button" id="clear-logs">Clear</button>
        <button class="debug-button" id="expand-all">Expand All</button>
        <button class="debug-button" id="minimize-debug">−</button>
    `;

    header.appendChild(title);
    header.appendChild(controls);
    container.appendChild(header);

    // Create content grid
    const contentGrid = document.createElement('div');
    contentGrid.className = 'debug-content-grid';
    container.appendChild(contentGrid);

    // Create sections for each feature
    const mainLogsSection = document.createElement('div');
    mainLogsSection.className = 'debug-section logs';

    const networkSection = document.createElement('div');
    networkSection.className = 'debug-section network';

    const timeTravelSection = document.createElement('div');
    timeTravelSection.className = 'debug-section timetravel';

    // Add new sections for the unused features
    const codeSmellSection = document.createElement('div');
    codeSmellSection.className = 'debug-section codesmells';

    const digitalTwinSection = document.createElement('div');
    digitalTwinSection.className = 'debug-section digitaltwin';

    const chaosMonkeySection = document.createElement('div');
    chaosMonkeySection.className = 'debug-section chaosmonkey';

    // Update grid layout to include all sections
    contentGrid.appendChild(mainLogsSection);
    contentGrid.appendChild(networkSection);
    contentGrid.appendChild(timeTravelSection);
    contentGrid.appendChild(codeSmellSection);
    contentGrid.appendChild(digitalTwinSection);
    contentGrid.appendChild(chaosMonkeySection);

    document.body.appendChild(container);

    // Setup minimize functionality
    const minimizeButton = document.getElementById('minimize-debug');
    // if (minimizeButton) {
    //     minimizeButton.onclick = () => {
    //         isMinimized = !isMinimized;
    //         container.classList.toggle('minimized', isMinimized);
    //         contentGrid.classList.toggle('hidden', isMinimized);
    //         minimizeButton.textContent = isMinimized ? '+' : '−';
    //     };
    // }

    const renderStackTrace = (entry: DebugEntry) => {
        return entry.stack
            ? `<div class="debug-stack-trace">${entry.stack}</div>`
            : '';
    };

    const getDetailsContent = (entry: DebugEntry) => {
        if (typeof entry.details === 'object') {
            return `<div class="debug-json" data-json='${JSON.stringify(entry.details)}'></div>`;
        }
        return `<pre>${entry.details}</pre>`;
    };

    const updateDisplay = () => {
        const html = logs.map(entry => {
            const time = entry.timestamp
                ? entry.timestamp.split('T')[1].split('.')[0]
                : 'unknown';

            const isExpanded = expandedEntries.has(entry.id);

            return `
            <div class="debug-entry">
                <div class="debug-entry-header" data-id="${entry.id}">
                    <span class="debug-timestamp">[${time}]</span>
                    <span class="debug-category" data-type="${entry.type}">${entry.category}</span>
                    <span class="debug-summary">${entry.summary}</span>
                </div>
                ${isExpanded ? `
                    <div class="debug-details">
                        ${getDetailsContent(entry)}
                        ${renderStackTrace(entry)}
                    </div>
                ` : ''}
            </div>
        `;
        }).join('');

        mainLogsSection.innerHTML = html;

        // Initialize JSON viewers
        mainLogsSection.querySelectorAll('.debug-json').forEach(jsonContainer => {
            const jsonData = JSON.parse(jsonContainer.getAttribute('data-json') ?? '{}');
            createJsonViewer(jsonData, {
                initialExpandDepth: 1,
                showTypes: true,
                theme: 'dark'
            })(jsonContainer as HTMLElement);
        });
    };

    // // Move event delegation outside updateDisplay
    // mainLogsSection.addEventListener('click', (e) => {
    //     const header = (e.target as Element).closest('.debug-entry-header');
    //     if (header) {
    //         const id = header.getAttribute('data-id');
    //         if (id) {
    //             if (expandedEntries.has(id)) {
    //                 expandedEntries.delete(id);
    //             } else {
    //                 expandedEntries.add(id);
    //             }
    //             updateDisplay();
    //         }
    //     }
    // });

    // Initialize features
    setupConsoleTracking(originalConsole, addLog);
    setupXHRTracking({ addLog });
    setupFetchTracking({ addLog });
    setupMutationObserver({ addLog });
    setupEventListeners({ logs, expandedEntries, updateDisplay });
    setupStorageTracking({ addLog });
    const networkView = setup3DNetworkView({ addLog, overlay: networkSection });
    const timeTravel = setupTimeTravel({ addLog, overlay: timeTravelSection });

    // Initialize new features
    const codeSmells = setupCodeSmellDetector({ addLog, overlay: codeSmellSection });
    const digitalTwin = setupDigitalTwin({ addLog, overlay: digitalTwinSection });
    const chaosMonkey = setupChaosMonkey({ addLog, overlay: chaosMonkeySection });
    //const debugChoreographer = setupDebugChoreographer({ addLog, overlay: timeTravelSection });

    return {
        clear: () => {
            logs = [];
            updateDisplay();
        },
        addLog,
        destroy: () => {
            observer?.disconnect();
            networkView.destroy();
            timeTravel.destroy();
            codeSmells.destroy();
            digitalTwin.destroy();
            chaosMonkey.destroy();
            container.remove();
            Object.keys(originalConsole).forEach((method) => {
                console[method as ConsoleMethod] = originalConsole[method];
            });
        }
    };
};