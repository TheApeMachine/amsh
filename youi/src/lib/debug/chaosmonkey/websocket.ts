import { Log } from './types';
import { pickOne } from './utils';

export const setupWebSocketChaos = ({
    logChaos,
    shouldCreateChaos,
    config
}: {
    logChaos: (log: Log) => void;
    shouldCreateChaos: (type: string) => boolean;
    config: { safeMode: boolean };
}) => {
    const originalWebSocket = window.WebSocket;
    const sockets = new Set<WebSocket>();

    class ChaosWebSocket extends WebSocket {
        private readonly originalSend: typeof WebSocket.prototype.send;
        private messageQueue: any[] = [];
        private readonly fragmentSize = 16384; // 16KB default fragment size
        private pingInterval?: number;

        constructor(url: string | URL, protocols?: string | string[]) {
            super(url, protocols);
            sockets.add(this);
            this.originalSend = this.send;
            this.send = this.chaosSend;
            this.setupChaos();
        }

        private setupChaos() {
            // Random connection drops
            if (shouldCreateChaos('websocket')) {
                setInterval(() => {
                    if (Math.random() < 0.05) { // 5% chance every interval
                        this.close(1001, 'Chaos Monkey: Connection reset');

                        logChaos({
                            type: 'websocket.drop',
                            description: 'Dropped WebSocket connection',
                            duration: 0,
                            impact: 'high',
                            recoverable: true
                        });

                        // Maybe reconnect
                        if (Math.random() < 0.7) {
                            setTimeout(() => {
                                this.reconnect();
                            }, Math.random() * 2000);
                        }
                    }
                }, 10000);
            }

            // Manipulate ping/pong
            this.pingInterval = window.setInterval(() => {
                if (shouldCreateChaos('websocket') && Math.random() < 0.2) {
                    const chaos = pickOne(['delay', 'drop', 'duplicate']);

                    switch (chaos) {
                        case 'delay':
                            setTimeout(() => this.sendPing(), Math.random() * 1000);
                            break;
                        case 'drop':
                            // Don't send ping at all
                            break;
                        case 'duplicate':
                            for (let i = 0; i < 3; i++) {
                                setTimeout(() => this.sendPing(), i * 100);
                            }
                            break;
                    }
                } else {
                    this.sendPing();
                }
            }, 30000);

            // Message batching
            this.addEventListener('message', (event) => {
                if (shouldCreateChaos('websocket') && Math.random() < 0.15) {
                    this.messageQueue.push(event.data);

                    if (this.messageQueue.length > 5) {
                        this.flushMessageQueue();
                    }

                    logChaos({
                        type: 'websocket.batch',
                        description: `Batched ${this.messageQueue.length} messages`,
                        duration: 0,
                        impact: 'medium',
                        recoverable: true
                    });
                }
            });
        }

        private chaosSend(data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
            if (!shouldCreateChaos('websocket')) {
                this.originalSend(data);
                return;
            }

            const chaos = pickOne([
                'fragment',
                'delay',
                'corrupt',
                'reorder',
                'duplicate'
            ]);

            switch (chaos) {
                case 'fragment':
                    void this.sendFragmented(data);
                    break;

                case 'delay':
                    void this.sendDelayed(data);
                    break;

                case 'corrupt':
                    void this.sendCorrupted(data);
                    break;

                case 'reorder':
                    void this.sendReordered(data);
                    break;

                case 'duplicate':
                    void this.sendDuplicated(data);
                    break;

                default:
                    this.originalSend(data);
            }
        }

        private async sendFragmented(data: any) {
            if (data instanceof Blob || data instanceof ArrayBuffer) {
                const chunks = this.fragmentData(data);

                for (const chunk of chunks) {
                    await new Promise(resolve => setTimeout(resolve, Math.random() * 100));
                    this.originalSend(chunk);
                }

                logChaos({
                    type: 'websocket.fragment',
                    description: `Fragmented message into ${chunks.length} chunks`,
                    duration: chunks.length * 100,
                    impact: 'low',
                    recoverable: true
                });
            } else {
                this.originalSend(data);
            }
        }

        private async sendDelayed(data: any) {
            const delay = Math.random() * 2000;
            await new Promise(resolve => setTimeout(resolve, delay));

            logChaos({
                type: 'websocket.delay',
                description: `Delayed WebSocket message by ${delay.toFixed(0)}ms`,
                duration: delay,
                impact: 'medium',
                recoverable: true
            });

            this.originalSend(data);
        }

        private async sendCorrupted(data: any) {
            if (config.safeMode) {
                this.originalSend(data);
                return;
            }

            let corrupted = data;
            if (typeof data === 'string') {
                corrupted = this.corruptString(data);
            } else if (data instanceof ArrayBuffer) {
                corrupted = this.corruptArrayBuffer(data);
            }

            logChaos({
                type: 'websocket.corrupt',
                description: 'Corrupted WebSocket message',
                duration: 0,
                impact: 'high',
                recoverable: false
            });

            this.originalSend(corrupted);
        }

        private async sendReordered(data: any): Promise<void> {
            this.messageQueue.push(data);

            if (this.messageQueue.length > 3) {
                const shuffled = [...this.messageQueue];
                shuffled.sort(() => Math.random() - 0.5);

                for (const msg of shuffled) {
                    await new Promise(resolve => setTimeout(resolve, Math.random() * 100));
                    this.originalSend(msg);
                }

                logChaos({
                    type: 'websocket.reorder',
                    description: `Reordered ${shuffled.length} messages`,
                    duration: shuffled.length * 100,
                    impact: 'medium',
                    recoverable: true
                });

                this.messageQueue.length = 0;
            }
        }

        private async sendDuplicated(data: any) {
            const count = Math.floor(Math.random() * 3) + 2;

            for (let i = 0; i < count; i++) {
                await new Promise(resolve => setTimeout(resolve, Math.random() * 200));
                this.originalSend(data);
            }

            logChaos({
                type: 'websocket.duplicate',
                description: `Duplicated message ${count} times`,
                duration: count * 200,
                impact: 'medium',
                recoverable: true
            });
        }

        private fragmentData(data: Blob | ArrayBuffer): Array<Blob | ArrayBuffer> {
            const fragments = [];
            let offset = 0;
            const total = data instanceof Blob ? data.size : data.byteLength;

            while (offset < total) {
                const chunk = data.slice(offset, offset + this.fragmentSize);
                fragments.push(chunk);
                offset += this.fragmentSize;
            }

            return fragments;
        }

        private corruptString(str: string): string {
            const chars = str.split('');
            const pos = Math.floor(Math.random() * chars.length);
            chars[pos] = String.fromCharCode(chars[pos].charCodeAt(0) + 1);
            return chars.join('');
        }

        private corruptArrayBuffer(buffer: ArrayBuffer): ArrayBuffer {
            const view = new Uint8Array(buffer);
            const pos = Math.floor(Math.random() * view.length);
            view[pos] = view[pos] + 1;
            return view.buffer;
        }

        private sendPing() {
            if (this.readyState === WebSocket.OPEN) {
                this.originalSend(new Uint8Array([0x9])); // WebSocket ping frame
            }
        }

        private flushMessageQueue() {
            for (const msg of this.messageQueue) {
                this.originalSend(msg);
            }
            this.messageQueue.length = 0;
        }

        private reconnect() {
            const newSocket = new originalWebSocket(this.url);
            newSocket.onopen = this.onopen;
            newSocket.onclose = this.onclose;
            newSocket.onerror = this.onerror;
            newSocket.onmessage = this.onmessage;
            sockets.delete(this);
            sockets.add(newSocket);
        }

        close(code?: number, reason?: string) {
            clearInterval(this.pingInterval);
            sockets.delete(this);
            super.close(code, reason);
        }
    }

    window.WebSocket = ChaosWebSocket;

    return {
        getSockets: () => sockets,
        cleanup: () => {
            window.WebSocket = originalWebSocket;
            sockets.forEach(socket => socket.close());
        }
    };
};
