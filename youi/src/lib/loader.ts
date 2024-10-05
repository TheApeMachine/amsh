const cache: Record<string, { data: any, timestamp: number }> = {};
const cacheExpirationTime = 60000; // Cache for 60 seconds

// Define the loader that initially returns "loading" state
export const loader = async (
    requests: Record<string, { 
        url: string, 
        method?: string, 
        params?: Record<string, any>,
        headers?: Record<string, string>
    }>
): Promise<{ state: "loading" | "error" | "success", results: Record<string, any> }> => {
    console.debug("loader", "requests", requests)
    const now = Date.now();
    
    const results: Record<string, any> = {};
    try {
        for (const key in requests) {
            const { url, method = "GET", params, headers } = requests[key];
            const cacheKey = `${method}:${url}:${JSON.stringify(params || {})}`;
            
            console.debug("loader", "cacheKey", cacheKey)

            // Check cache and expiration
            if (cache[cacheKey] && now - cache[cacheKey].timestamp < cacheExpirationTime) {
                results[key] = cache[cacheKey].data;
                console.debug("loader", "cache hit", cacheKey, results[key])
            } else {
                const response = await fetchWithParams(url, method, params, headers);
                const data = await response.json();
                results[key] = data;
                cache[cacheKey] = { data, timestamp: now }; // Store with timestamp
                console.debug("loader", "cache miss", cacheKey, results[key])
            }
        }

        return { state: "success", results };
    } catch (error: any) {
        console.error("loader", "error", error)
        return { 
            state: "error", 
            results: error instanceof Error ? error : new Error('An unknown error occurred') 
        };
    }
};

// Utility function to handle query params or request body based on method
const fetchWithParams = (
    url: string, 
    method: string = "GET", 
    params?: Record<string, any>, 
    headers?: Record<string, string>
) => {
    if (method === "GET" && params) {
        const queryParams = new URLSearchParams(params).toString();
        return fetch(`${url}?${queryParams}`, { headers: { ...headers } });
    } else {
        console.debug("loader", "fetchWithParams", url, method, params, headers)
        return fetch(url, {
            method,
            headers: { "Content-Type": "application/json", ...headers },
            body: method !== "GET" ? JSON.stringify(params) : undefined,
        });
    }
};
