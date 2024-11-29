export interface Log {
    type: string;
    description: string;
    duration: number;
    impact: 'low' | 'medium' | 'high';
    recoverable: boolean;
} 