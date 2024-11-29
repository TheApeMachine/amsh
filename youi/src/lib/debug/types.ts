export interface DebugEntry {
    id: string;
    timestamp: string;
    type: string;
    category: string;
    summary: string;
    details: any;
    stack?: string;
}

export type ConsoleMethod = 'log' | 'warn' | 'error' | 'info' | 'debug'; 