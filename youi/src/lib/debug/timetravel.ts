// time-travel.ts
interface TimelineSnapshot {
    id: string;
    timestamp: number;
    dom: string;
    storage: {
        local: Record<string, string>;
        session: Record<string, string>;
        cookies: Record<string, string>;
    };
    network: {
        requests: Array<{
            url: string;
            method: string;
            status?: number;
            timestamp: number;
            state: 'active' | 'completed';
        }>;
    };
    logs: any[];
    state: any;
}

export const setupTimeTravel = ({ addLog, overlay }: { addLog: Function; overlay: HTMLElement }) => {
    const timeline: TimelineSnapshot[] = [];
    let isRecording = false;
    let currentSnapshotId: string | null = null;

    // Create timeline UI
    const timelineUI = document.createElement('div');
    timelineUI.className = 'debug-timeline';
    timelineUI.innerHTML = `
        <div class="debug-timeline-controls">
            <button class="debug-button" id="record-toggle">⏺ Record</button>
            <button class="debug-button" id="snapshot-now">📸 Snapshot</button>
            <input type="range" id="timeline-scrubber" min="0" max="100" value="100" disabled />
            <span id="timeline-time">Now</span>
        </div>
        <div class="debug-timeline-markers"></div>
    `;
    overlay.insertBefore(timelineUI, overlay.firstChild);

    // Capture current application state
    const captureState = (): TimelineSnapshot => {
        // Capture network state from Performance API
        const networkEntries = performance.getEntriesByType('resource')
            .map(entry => ({
                url: entry.name,
                method: 'GET', // Default to GET as PerformanceEntry doesn't provide method
                timestamp: entry.startTime,
                state: entry.duration === 0 ? 'active' as const : 'completed' as const
            }));

        return {
            id: crypto.randomUUID(),
            timestamp: Date.now(),
            dom: document.documentElement.outerHTML,
            storage: {
                local: { ...localStorage },
                session: { ...sessionStorage },
                cookies: Object.fromEntries(
                    document.cookie.split(';').map(c => {
                        const [key, value] = c.trim().split('=');
                        return [key, value];
                    })
                )
            },
            network: {
                requests: networkEntries
            },
            logs: [],
            state: captureGlobalState()
        };
    };

    // Capture global state by monitoring window properties
    const captureGlobalState = () => {
        const globals = new Set<string>();
        let obj = window;
        do {
            Object.getOwnPropertyNames(obj).forEach(prop => globals.add(prop));
        } while (obj = Object.getPrototypeOf(obj));

        const state: Record<string, any> = {};
        globals.forEach(prop => {
            try {
                const value = (window as any)[prop];
                if (typeof value !== 'function' && prop !== 'localStorage' && prop !== 'sessionStorage') {
                    state[prop] = value;
                }
            } catch (e) {
                // Skip properties we can't access
            }
        });
        return state;
    };

    // Restore application to a specific snapshot
    const restoreSnapshot = (snapshot: TimelineSnapshot) => {
        try {
            // Store debug overlay reference
            const debugOverlay = document.querySelector('.debug-overlay');
            const timelineUI = document.querySelector('.debug-timeline');

            if (!debugOverlay || !timelineUI) {
                throw new Error('Debug overlay elements not found');
            }

            // Parse the snapshot DOM
            const tempDoc = new DOMParser().parseFromString(snapshot.dom, 'text/html');

            // Remove the debug overlay from the snapshot DOM if it exists there
            const snapshotOverlay = tempDoc.querySelector('.debug-overlay');
            snapshotOverlay?.remove();

            // Replace the entire document except the debug overlay
            const currentOverlay = debugOverlay.cloneNode(true);
            document.documentElement.innerHTML = tempDoc.documentElement.innerHTML;

            // Re-add the debug overlay
            document.body.appendChild(currentOverlay);

            // Restore storage
            localStorage.clear();
            sessionStorage.clear();
            Object.entries(snapshot.storage.local).forEach(([k, v]) => localStorage.setItem(k, v));
            Object.entries(snapshot.storage.session).forEach(([k, v]) => sessionStorage.setItem(k, v));

            // Restore cookies
            Object.entries(snapshot.storage.cookies).forEach(([k, v]) => {
                document.cookie = `${k}=${v}`;
            });

            // Notify about the time travel
            addLog({
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                type: 'time-travel',
                category: 'debug.timeTravel',
                summary: `Restored snapshot from ${new Date(snapshot.timestamp).toLocaleTimeString()}`,
                details: {
                    snapshotId: snapshot.id,
                    timestamp: snapshot.timestamp,
                    changes: {
                        dom: true,
                        storage: true,
                        network: snapshot.network.requests.length
                    }
                }
            });

            currentSnapshotId = snapshot.id;

            // Reinitialize event listeners for the timeline UI
            initializeTimelineControls();
        } catch (error: unknown) {
            addLog({
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                type: 'error',
                category: 'debug.timeTravel',
                summary: 'Failed to restore snapshot',
                details: {
                    error: error instanceof Error ? error.message : 'Unknown error'
                }
            });
        }
    };

    // Update timeline UI
    const updateTimelineUI = () => {
        const markers = document.querySelector('.debug-timeline-markers');
        if (!markers) return;

        markers.innerHTML = timeline.map(snapshot => {
            const percent = ((snapshot.timestamp - timeline[0].timestamp) /
                (timeline[timeline.length - 1].timestamp - timeline[0].timestamp)) * 100;
            return `
                <div class="timeline-marker" 
                     style="left: ${percent}%" 
                     data-snapshot-id="${snapshot.id}"
                     title="${new Date(snapshot.timestamp).toLocaleTimeString()}"
                     ${snapshot.id === currentSnapshotId ? 'data-current="true"' : ''}
                ></div>
            `;
        }).join('');

        // Update scrubber
        const scrubber = document.getElementById('timeline-scrubber') as HTMLInputElement;
        if (scrubber) {
            scrubber.max = (timeline.length - 1).toString();
            scrubber.value = timeline.findIndex(s => s.id === currentSnapshotId).toString();
            scrubber.disabled = timeline.length < 2;
        }
    };

    // Add this new function to initialize timeline controls
    const initializeTimelineControls = () => {
        const recordToggle = document.getElementById('record-toggle');
        const snapshotNow = document.getElementById('snapshot-now');
        const timelineScrubber = document.getElementById('timeline-scrubber') as HTMLInputElement;

        if (recordToggle) {
            recordToggle.addEventListener('click', () => {
                isRecording = !isRecording;
                recordToggle.textContent = isRecording ? '⏹ Stop' : '⏺ Record';
                recordToggle.style.color = isRecording ? '#ff4444' : '';

                if (isRecording) {
                    const recordingInterval = setInterval(() => {
                        if (timeline.length >= 100) { // Prevent memory issues with too many snapshots
                            isRecording = false;
                            recordToggle.textContent = '⏺ Record';
                            recordToggle.style.color = '';
                            clearInterval(recordingInterval);
                            addLog({
                                id: crypto.randomUUID(),
                                timestamp: new Date().toISOString(),
                                type: 'warning',
                                category: 'debug.timeTravel',
                                summary: 'Recording stopped: maximum snapshots reached'
                            });
                            return;
                        }
                        timeline.push(captureState());
                        updateTimelineUI();
                    }, 1000);

                    (recordToggle as any).recordingInterval = recordingInterval;
                } else {
                    clearInterval((recordToggle as any).recordingInterval);
                }
            });
        }

        if (snapshotNow) {
            snapshotNow.addEventListener('click', () => {
                timeline.push(captureState());
                updateTimelineUI();
            });
        }

        if (timelineScrubber) {
            timelineScrubber.addEventListener('input', (e) => {
                const index = parseInt((e.target as HTMLInputElement).value);
                if (Number.isInteger(index) && index >= 0 && index < timeline.length) {
                    restoreSnapshot(timeline[index]);
                    updateTimelineUI();
                }
            });
        }
    };

    // Initialize timeline with current state
    timeline.push(captureState());
    updateTimelineUI();
    initializeTimelineControls();

    return {
        getTimeline: () => timeline,
        getCurrentSnapshot: () => timeline.find(s => s.id === currentSnapshotId),
        goToSnapshot: (id: string) => {
            const snapshot = timeline.find(s => s.id === id);
            if (snapshot) {
                restoreSnapshot(snapshot);
                updateTimelineUI();
            }
        },
        destroy: () => {
            const recordToggle = document.getElementById('record-toggle');
            if (recordToggle) {
                clearInterval((recordToggle as any).recordingInterval);
            }
            timelineUI.remove();
        }
    };
};