import { ConsoleMethod, DebugEntry } from './types';

// Extract console logger to standalone function
export const createConsoleLogger = (originalMethod: Function, method: ConsoleMethod, addLogFn: (entry: DebugEntry) => void) => {
    return (...args: any[]) => {
        originalMethod.apply(console, args);

        addLogFn({
            id: crypto.randomUUID(),
            timestamp: new Date().toISOString(),
            type: 'console',
            category: `console.${method}`,
            summary: args.map(arg =>
                typeof arg === 'object' ? JSON.stringify(arg) : String(arg)
            ).join(' ').slice(0, 100),
            details: args,
            stack: new Error().stack
        });
    };
};

// Update setupConsoleTracking
export const setupConsoleTracking = (originalConsole: Record<ConsoleMethod, Function>, addLog: (entry: DebugEntry) => void) => {
    const methods: ConsoleMethod[] = ['log', 'warn', 'error', 'info', 'debug'];

    methods.forEach(method => {
        originalConsole[method] = console[method];
        console[method] = createConsoleLogger(originalConsole[method], method, addLog);
    });
};