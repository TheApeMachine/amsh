export function pickOne<T>(options: T[]): T {
    return options[Math.floor(Math.random() * options.length)];
} 