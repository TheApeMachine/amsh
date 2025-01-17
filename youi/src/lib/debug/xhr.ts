export const setupXHRTracking = ({ addLog }: { addLog: (entry: any) => void }) => {
    const XHR = XMLHttpRequest;
    (window as any).XMLHttpRequest = function () {
        const xhr = new XHR();
        const id = crypto.randomUUID();

        // Track request details
        const requestDetails = {
            method: 'GET',
            url: '',
            startTime: 0,
            headers: new Map<string, string>()
        };

        // Override open to capture method and url
        const originalOpen = xhr.open;
        xhr.open = function (method: string, url: string, ...args: any[]) {
            requestDetails.method = method;
            requestDetails.url = url;
            return originalOpen.call(xhr, method, url, args[0] ?? true, ...args.slice(1));
        };

        // Override setRequestHeader to capture headers
        const originalSetRequestHeader = xhr.setRequestHeader;
        xhr.setRequestHeader = function (header: string, value: string) {
            requestDetails.headers.set(header, value);
            return originalSetRequestHeader.apply(xhr, [header, value]);
        };

        // Track request start
        xhr.addEventListener('loadstart', () => {
            requestDetails.startTime = performance.now();
            addLog({
                id,
                timestamp: new Date().toISOString(),
                type: 'xhr',
                category: 'xhr.start',
                summary: `${requestDetails.method} ${requestDetails.url}`,
                details: {
                    method: requestDetails.method,
                    url: requestDetails.url,
                    headers: Object.fromEntries(requestDetails.headers)
                },
                stack: new Error().stack
            });
        });

        // Track successful completion
        xhr.addEventListener('load', () => {
            const duration = performance.now() - requestDetails.startTime;

            // Try to parse response if it's JSON
            let parsedResponse = xhr.response;
            try {
                if (xhr.getResponseHeader('content-type')?.includes('application/json')) {
                    parsedResponse = JSON.parse(xhr.response);
                }
            } catch (e) {
                // Keep original response if parsing fails
            }

            addLog({
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                type: 'xhr',
                category: 'xhr.complete',
                summary: `${xhr.status} ${xhr.statusText} (${duration.toFixed(2)}ms)`,
                details: {
                    duration,
                    status: xhr.status,
                    statusText: xhr.statusText,
                    response: parsedResponse,
                    headers: xhr.getAllResponseHeaders(),
                    requestDetails: {
                        method: requestDetails.method,
                        url: requestDetails.url,
                        headers: Object.fromEntries(requestDetails.headers)
                    }
                }
            });
        });

        // Track failures
        xhr.addEventListener('error', () => {
            const duration = performance.now() - requestDetails.startTime;
            addLog({
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                type: 'xhr',
                category: 'xhr.error',
                summary: `Failed: ${requestDetails.method} ${requestDetails.url}`,
                details: {
                    duration,
                    requestDetails: {
                        method: requestDetails.method,
                        url: requestDetails.url,
                        headers: Object.fromEntries(requestDetails.headers)
                    }
                },
                stack: new Error().stack
            });
        });

        // Track aborted requests
        xhr.addEventListener('abort', () => {
            const duration = performance.now() - requestDetails.startTime;
            addLog({
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                type: 'xhr',
                category: 'xhr.abort',
                summary: `Aborted: ${requestDetails.method} ${requestDetails.url}`,
                details: {
                    duration,
                    requestDetails: {
                        method: requestDetails.method,
                        url: requestDetails.url,
                        headers: Object.fromEntries(requestDetails.headers)
                    }
                }
            });
        });

        // Track upload progress
        xhr.upload.addEventListener('progress', (event) => {
            if (event.lengthComputable) {
                const percent = (event.loaded / event.total * 100).toFixed(2);
                addLog({
                    id: crypto.randomUUID(),
                    timestamp: new Date().toISOString(),
                    type: 'xhr',
                    category: 'xhr.upload',
                    summary: `Upload progress: ${percent}%`,
                    details: {
                        loaded: event.loaded,
                        total: event.total,
                        percent: parseFloat(percent)
                    }
                });
            }
        });

        return xhr;
    };
};