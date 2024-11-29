import { DebugEntry } from "./types";

export const setupFetchTracking = ({ addLog }: { addLog: (entry: DebugEntry) => void }) => {
    const originalFetch = window.fetch;

    window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
        const id = crypto.randomUUID();
        const startTime = performance.now();

        // Parse request details
        const url = input instanceof Request ? input.url : input.toString();
        const method = init?.method ?? (input instanceof Request ? input.method : 'GET');
        const headers = new Headers(init?.headers || (input instanceof Request ? input.headers : {}));

        // Log request start
        addLog({
            id,
            timestamp: new Date().toISOString(),
            type: 'fetch',
            category: 'fetch.start',
            summary: `${method} ${url}`,
            details: {
                method,
                url,
                headers: Object.fromEntries(headers.entries()),
                body: init?.body
            },
            stack: new Error().stack
        });

        try {
            const response = await originalFetch(input, init);
            const duration = performance.now() - startTime;

            // Clone the response so we can read the body
            const clonedResponse = response.clone();
            let responseData: any;

            try {
                // Try to parse as JSON first
                if (response.headers.get('content-type')?.includes('application/json')) {
                    responseData = await clonedResponse.json();
                } else {
                    // Fall back to text for other content types
                    responseData = await clonedResponse.text();
                }
            } catch (e) {
                responseData = 'Unable to parse response body';
            }

            addLog({
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                type: 'fetch',
                category: 'fetch.complete',
                summary: `${response.status} ${response.statusText} (${duration.toFixed(2)}ms)`,
                details: {
                    duration,
                    status: response.status,
                    statusText: response.statusText,
                    headers: Object.fromEntries(response.headers.entries()),
                    response: responseData,
                    requestDetails: {
                        method,
                        url,
                        headers: Object.fromEntries(headers.entries())
                    }
                }
            });

            return response;
        } catch (error) {
            const duration = performance.now() - startTime;

            addLog({
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                type: 'fetch',
                category: 'fetch.error',
                summary: `Failed: ${method} ${url}`,
                details: {
                    duration,
                    error: error instanceof Error ? error.message : 'Unknown error',
                    requestDetails: {
                        method,
                        url,
                        headers: Object.fromEntries(headers.entries())
                    }
                },
                stack: error instanceof Error ? error.stack : new Error().stack
            });

            throw error;
        }
    };
};