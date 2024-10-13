import { Chunk, Team, Agent } from './agent';
import './team';
import './network';

class ConversationVisualizer extends HTMLElement {
    private conversations: Record<string, Chunk[]> = {};
    private teams: Record<string, Team> = {};
    private networkData: { nodes: any[]; edges: any[] } = { nodes: [], edges: [] };
    private ws: WebSocket | null = null;
    private reconnectAttempts: number = 0;
    private maxReconnectAttempts: number = 5;
    private reconnectInterval: number = 5000; // 5 seconds
    private template: HTMLTemplateElement;

    constructor() {
        super();
        this.template = document.createElement('template');
        this.template.innerHTML = `
            <style>
                :host { 
                    display: block;
                    font-family: Arial, sans-serif;
                    color: #333;
                    background-color: #f0f0f0;
                    padding: 20px;
                }
                h1 {
                    text-align: center;
                    color: #2c3e50;
                }
                .container { 
                    display: flex; 
                    justify-content: space-around;
                    margin-top: 20px;
                }
                #toggle-view {
                    display: block;
                    margin: 20px auto;
                    padding: 10px 20px;
                    font-size: 16px;
                    background-color: #3498db;
                    color: white;
                    border: none;
                    border-radius: 5px;
                    cursor: pointer;
                }
                #toggle-view:hover {
                    background-color: #2980b9;
                }
                .hidden { display: none; }
                #notification {
                    position: fixed;
                    top: 20px;
                    right: 20px;
                    padding: 10px;
                    border-radius: 5px;
                    color: white;
                    font-weight: bold;
                    z-index: 1000;
                }
                .error { background-color: #e74c3c; }
                .warning { background-color: #f39c12; }
                .prompt-container {
                    margin-top: 20px;
                    display: flex;
                    justify-content: center;
                    align-items: center;
                }
                #prompt-input {
                    width: 60%;
                    padding: 10px;
                    font-size: 16px;
                    border: 1px solid #bdc3c7;
                    border-radius: 5px;
                    margin-right: 10px;
                }
                #send-prompt {
                    padding: 10px 20px;
                    font-size: 16px;
                    background-color: #2ecc71;
                    color: white;
                    border: none;
                    border-radius: 5px;
                    cursor: pointer;
                }
                #send-prompt:hover {
                    background-color: #27ae60;
                }
            </style>
            <h1>Multi-Team LLM Conversation Visualizer</h1>
            <div class="prompt-container">
                <input type="text" id="prompt-input" placeholder="Enter your prompt here...">
                <button id="send-prompt">Send Prompt</button>
            </div>
            <button id="toggle-view">Toggle View</button>
            <div class="container">
                <!-- Team views will be dynamically added here -->
            </div>
            <network-graph-view class="hidden"></network-graph-view>
            <div id="notification" class="hidden"></div>
        `;
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
        this.setupWebSocket();
        this.shadowRoot?.querySelector('#toggle-view')?.addEventListener('click', () => this.toggleView());
        this.shadowRoot?.querySelector('#send-prompt')?.addEventListener('click', () => this.sendPrompt());
    }

    setupWebSocket() {
        this.ws = new WebSocket("ws://localhost:8567/ws");
        this.ws.onmessage = (event) => {
            try {
                const chunk = JSON.parse(event.data);
                this.processChunk(chunk);
            } catch (e) {
                console.error('Failed to parse WebSocket message:', e);
            }
        };
        this.ws.onerror = (error) => {
            console.error("WebSocket error:", error);
            this.showNotification('WebSocket error occurred. Please check the connection.', 'error');
        };
        this.ws.onclose = () => {
            console.warn("WebSocket connection closed, attempting to reconnect...");
            this.showNotification('WebSocket connection lost. Reconnecting...', 'warning');
            setTimeout(() => this.setupWebSocket(), 1000);
        };
    }

    private sendPrompt() {
        const promptInput = this.shadowRoot?.querySelector('#prompt-input') as HTMLInputElement;
        const prompt = promptInput.value.trim();
        if (prompt && this.ws && this.ws.readyState === WebSocket.OPEN) {
            // Adjust the message format to match what your Go server expects
            this.ws.send(prompt);
            this.log("Sent prompt:", prompt);
            promptInput.value = ''; // Clear the input after sending
            this.showNotification('Prompt sent successfully!', 'success');
        } else if (!prompt) {
            this.showNotification('Please enter a prompt before sending.', 'warning');
        } else {
            this.showNotification('Unable to send prompt. WebSocket is not connected.', 'error');
            this.log("Failed to send prompt. WebSocket state:", this.ws?.readyState);
        }
    }

    private log(...args: any[]) {
        console.log(...args);
        const debugLog = this.shadowRoot?.querySelector('#debug-log');
        if (debugLog) {
            const logEntry = document.createElement('div');
            logEntry.textContent = args.map(arg => 
                typeof arg === 'object' ? JSON.stringify(arg) : arg.toString()
            ).join(' ');
            debugLog.appendChild(logEntry);
            debugLog.scrollTop = debugLog.scrollHeight;
        }
    }

    private processChunk(chunk: Chunk) {
        if (chunk.team) {
            this.updateTeam(chunk.team);
        }
        this.updateConversations(chunk);
        this.updateNetworkData(chunk);
        this.dispatchUpdateEvent();
    }

    private updateTeam(team: Team) {
        if (!this.teams[team.id]) {
            this.teams[team.id] = team;
            this.updateTeamView(team);
        }
    }

    private updateConversations(chunk: Chunk) {
        const teamId = chunk.team?.id || 'unknown';
        if (!this.conversations[teamId]) {
            this.conversations[teamId] = [];
        }
        this.conversations[teamId].unshift(chunk);
        this.conversations[teamId] = this.conversations[teamId].slice(0, 10);
    }

    private updateNetworkData(chunk: Chunk) {
        if (chunk.agent && chunk.team) {
            const nodeId = `${chunk.team.id}-${chunk.agent.id}`;
            if (!this.networkData.nodes.some(node => node.data.id === nodeId)) {
                this.networkData.nodes.push({ data: { id: nodeId, label: `${chunk.team.name} - ${chunk.agent.type}` } });
            }
            // You might want to add edges based on agent interactions here
        }
    }

    private updateTeamView(team: Team) {
        const container = this.shadowRoot?.querySelector('.container');
        let teamView = this.shadowRoot?.querySelector(`team-conversation-view[team="${team.id}"]`);
        if (!teamView) {
            teamView = document.createElement('team-conversation-view');
            teamView.setAttribute('team', team.id);
            container?.appendChild(teamView);
        }
        (teamView as any).updateTeamData(team);
    }

    private dispatchUpdateEvent() {
        window.dispatchEvent(new CustomEvent('data-updated', { 
            detail: { 
                conversations: this.conversations, 
                teams: this.teams,
                networkData: this.networkData 
            } 
        }));
    }

    private toggleView() {
        this.shadowRoot?.querySelector('.container')?.classList.toggle('hidden');
        this.shadowRoot?.querySelector('network-graph-view')?.classList.toggle('hidden');
    }

    private showNotification(message: string, type: 'error' | 'warning' | 'success') {
        const notification = this.shadowRoot?.querySelector('#notification');
        if (notification) {
            notification.textContent = message;
            notification.className = type;
            notification.classList.remove('hidden');
            setTimeout(() => notification.classList.add('hidden'), 5000);
        }
        this.log(`${type.toUpperCase()}: ${message}`);
    }
}

customElements.define('conversation-visualizer', ConversationVisualizer);