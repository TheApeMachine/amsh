import { Conversation } from './agent';
import './team';
import './network';

class ConversationVisualizer extends HTMLElement {
    private conversations: Record<number, Conversation[]> = {};
    private threads: Record<number, Conversation[]> = {};
    private networkData: { nodes: any[]; edges: any[] } = { nodes: [], edges: [] };

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.render();
        this.startSimulation();
    }

    private render() {
        if (!this.shadowRoot) return;
        this.shadowRoot.innerHTML = `
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
            </style>
            <h1>Multi-Team LLM Conversation Visualizer</h1>
            <button id="toggle-view">Toggle View</button>
            <div class="container">
                <team-conversation-view team="1"></team-conversation-view>
                <team-conversation-view team="2"></team-conversation-view>
                <team-conversation-view team="3"></team-conversation-view>
            </div>
            <network-graph-view class="hidden"></network-graph-view>
        `;

        this.shadowRoot.querySelector('#toggle-view')?.addEventListener('click', () => this.toggleView());
    }

    private startSimulation() {
        // setInterval(() => {
        //     const newConversation = this.generateNewConversation();
        //     console.log("newConversation", newConversation);
        //     this.updateData(newConversation);
        // }, 1000);
    }

    private generateNewConversation(): Conversation {
        return {
            id: Date.now(),
            from: { team: Math.floor(Math.random() * 3) + 1, agent: Math.floor(Math.random() * 3) + 1 },
            to: { team: Math.floor(Math.random() * 3) + 1, agent: Math.floor(Math.random() * 3) + 1 },
            message: `Message at ${new Date().toLocaleTimeString()}`,
            sentiment: Math.random() > 0.5 ? 'positive' : 'negative',
            threadId: Math.floor(Math.random() * 5) + 1
        };
    }

    private updateData(newConversation: Conversation) {
        // Update conversations
        const team = newConversation.from.team;
        this.conversations[team] = [newConversation, ...(this.conversations[team] || [])].slice(0, 10);

        // Update threads
        const thread = this.threads[newConversation.threadId] || [];
        this.threads[newConversation.threadId] = [...thread, newConversation];

        // Update network data
        const sourceId = `${newConversation.from.team}-${newConversation.from.agent}`;
        const targetId = `${newConversation.to.team}-${newConversation.to.agent}`;
        const existingEdge = this.networkData.edges.find(edge => edge.data.source === sourceId && edge.data.target === targetId);
        if (existingEdge) {
            existingEdge.data.weight++;
        } else {
            this.networkData.edges.push({ data: { id: `${sourceId}-${targetId}`, source: sourceId, target: targetId, weight: 1 } });
        }

        // Ensure nodes exist
        [sourceId, targetId].forEach(id => {
            if (!this.networkData.nodes.some(node => node.data.id === id)) {
                this.networkData.nodes.push({ data: { id } });
            }
        });

        // Dispatch custom event to update child components
        window.dispatchEvent(new CustomEvent('data-updated', { 
            detail: { 
                conversations: this.conversations, 
                threads: this.threads, 
                networkData: this.networkData 
            } 
        }));
    }

    private toggleView() {
        this.shadowRoot?.querySelector('.container')?.classList.toggle('hidden');
        this.shadowRoot?.querySelector('network-graph-view')?.classList.toggle('hidden');
    }
}

customElements.define('conversation-visualizer', ConversationVisualizer);