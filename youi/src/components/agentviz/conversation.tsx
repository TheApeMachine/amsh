import { jsx } from '@/lib/template';
import { Chunk, Team, Agent } from './agent';
import './team';
import { NetworkGraphView } from './network';

export const ConversationVisualizer = () => {
    const conversations: Record<string, Chunk[]> = {};
    const teams: Record<string, Team> = {};
    const networkData: { nodes: any[]; edges: any[] } = { nodes: [], edges: [] };
    let ws: WebSocket | null = null;
    const template: HTMLTemplateElement = document.createElement('template');

    const showNotification = (message: string, type: 'error' | 'warning' | 'success') => {
        const notification = template.querySelector('#notification');
        if (notification) {
            notification.textContent = message;
            notification.className = type;
            notification.classList.remove('hidden');
            setTimeout(() => notification.classList.add('hidden'), 5000);
        }
        log(`${type.toUpperCase()}: ${message}`);
    };

    const setupWebSocket = () => {
        ws = new WebSocket("ws://localhost:8567/ws");
        ws.onmessage = (event) => {
            // Handle incoming messages
        };
        
        ws.onopen = () => {
            setInterval(() => {
                if (ws?.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify({ type: 'ping' }));
                }
            }, 30000);
        };

        ws.onerror = (error) => {
            console.error("WebSocket error:", error);
            showNotification('WebSocket error occurred. Please check the connection.', 'error');
        };
    };

    const sendPrompt = () => {
        const promptInput = template.querySelector('#prompt-input') as HTMLInputElement;
        const prompt = promptInput.value.trim();
        if (prompt && ws && ws.readyState === WebSocket.OPEN) {
            ws.send(prompt);
            log("Sent prompt:", prompt);
            promptInput.value = '';
            showNotification('Prompt sent successfully!', 'success');
        } else if (!prompt) {
            showNotification('Please enter a prompt before sending.', 'warning');
        } else {
            showNotification('Unable to send prompt. WebSocket is not connected.', 'error');
            log("Failed to send prompt. WebSocket state:", ws?.readyState);
        }
    };

    const log = (...args: any[]) => {
        console.log(...args);
        const debugLog = template.querySelector('#debug-log');
        if (debugLog) {
            const logEntry = document.createElement('div');
            logEntry.textContent = args.map(arg => 
                typeof arg === 'object' ? JSON.stringify(arg) : arg.toString()
            ).join(' ');
            debugLog.appendChild(logEntry);
            debugLog.scrollTop = debugLog.scrollHeight;
        }
    };

    const processChunk = (chunk: Chunk) => {
        if (chunk.team) {
            updateTeam(chunk.team);
        }
        updateConversations(chunk);
        updateNetworkData(chunk);
        dispatchUpdateEvent();
    };

    const updateTeam = (team: Team) => {
        if (!teams[team.id]) {
            teams[team.id] = team;
            updateTeamView(team);
        }
    };

    const updateConversations = (chunk: Chunk) => {
        const teamId = chunk.team?.id || 'unknown';
        if (!conversations[teamId]) {
            conversations[teamId] = [];
        }
        conversations[teamId].unshift(chunk);
        conversations[teamId] = conversations[teamId].slice(0, 10);
    };

    const updateNetworkData = (chunk: Chunk) => {
        if (chunk.agent && chunk.team) {
            const nodeId = `${chunk.team.id}-${chunk.agent.id}`;
            if (!networkData.nodes.some((node: any) => node.data.id === nodeId)) {
                networkData.nodes.push({ 
                    data: { 
                        id: nodeId, 
                        label: `${chunk.team.name} - ${chunk.agent.type}` 
                    } 
                });
            }
        }
    };

    const updateTeamView = (team: Team) => {
        const container = template.querySelector('.container');
        let teamView = template.querySelector(`team-conversation-view[team="${team.id}"]`);
        if (!teamView) {
            teamView = document.createElement('team-conversation-view');
            teamView.setAttribute('team', team.id);
            container?.appendChild(teamView);
        }
        (teamView as any).updateTeamData(team);
    };

    const dispatchUpdateEvent = () => {
        window.dispatchEvent(new CustomEvent('data-updated', { 
            detail: { 
                conversations: conversations, 
                teams: teams,
                networkData: networkData 
            } 
        }));
    };

    const toggleView = () => {
        template.querySelector('.container')?.classList.toggle('hidden');
        template.querySelector('network-graph-view')?.classList.toggle('hidden');
    };

    setupWebSocket();

    return (
        <div className="visualizer">
            <h1>Multi-Team LLM Conversation Visualizer</h1>
            <div className="prompt-container">
                <input 
                    type="text" 
                    id="prompt-input" 
                    placeholder="Enter your prompt here..."
                />
                <button id="send-prompt">Send Prompt</button>
            </div>
            <button id="toggle-view">Toggle View</button>
            <div className="container">
                {/* Team views will be dynamically added here */}
            </div>
            <NetworkGraphView />
            <div id="notification" className="hidden"></div>
        </div>
    );
};
