import { gsap } from 'gsap';
import { Conversation } from './agent';

class TeamConversationView extends HTMLElement {
    private team: number;
    private conversations: Conversation[] = [];
    private currentIndex = 0;
    private zoomLevel = 0;
    private container: HTMLElement | null = null;

    static get observedAttributes() { return ['team']; }

    constructor() {
        super();
        this.team = 1;
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.render();
        this.addEventListeners();
    }

    attributeChangedCallback(name: string, oldValue: string, newValue: string) {
        if (name === 'team' && oldValue !== newValue) {
            this.team = parseInt(newValue);
            this.render();
        }
    }

    private render() {
        if (!this.shadowRoot) return;
        this.shadowRoot.innerHTML = `
        <style>
            :host { 
                display: block; 
                width: 30%; 
                background-color: white;
                border-radius: 8px;
                box-shadow: 0 2px 10px rgba(0,0,0,0.1);
                padding: 15px;
            }
            h2 {
                color: #2c3e50;
                text-align: center;
                margin-top: 0;
            }
            .conversation-container { 
                height: 300px; 
                position: relative;
                perspective: 1000px;
                overflow: visible;
            }
            .conversation { 
                position: absolute; 
                width: 90%;
                left: 5%;
                padding: 10px;
                border-radius: 8px;
                box-shadow: 0 1px 3px rgba(0,0,0,0.12);
                transition: all 0.5s ease-in-out;
                background-color: white;
            }
            .controls { 
                display: flex; 
                justify-content: space-between;
                margin-top: 15px;
            }
            button {
                padding: 5px 10px;
                background-color: #3498db;
                color: white;
                border: none;
                border-radius: 4px;
                cursor: pointer;
            }
            button:hover {
                background-color: #2980b9;
            }
        </style>
        <h2>Team ${this.team}</h2>
        <div class="conversation-container"></div>
        <div class="controls">
            <button class="cycle-left">←</button>
            <button class="zoom-out">-</button>
            <button class="reset">Reset</button>
            <button class="zoom-in">+</button>
            <button class="cycle-right">→</button>
        </div>
        `;
        this.container = this.shadowRoot.querySelector('.conversation-container');
    }

    private addEventListeners() {
        if (!this.shadowRoot) return;
        this.shadowRoot.querySelector('.cycle-left')?.addEventListener('click', () => this.cycle(-1));
        this.shadowRoot.querySelector('.cycle-right')?.addEventListener('click', () => this.cycle(1));
        this.shadowRoot.querySelector('.zoom-in')?.addEventListener('click', () => this.zoom(1));
        this.shadowRoot.querySelector('.zoom-out')?.addEventListener('click', () => this.zoom(-1));
        this.shadowRoot.querySelector('.reset')?.addEventListener('click', () => this.reset());

        window.addEventListener('data-updated', (event: CustomEvent) => {
            const newConversations = event.detail.conversations[this.team] || [];
            this.updateConversations(newConversations);
        });
    }

    private updateConversations(newConversations: Conversation[]) {
        if (!this.container) return;

        const newMessages = newConversations.filter(conv => 
            !this.conversations.some(existing => existing.id === conv.id)
        );

        this.conversations = [...newMessages, ...this.conversations].slice(0, 10);

        newMessages.forEach((conv, index) => {
            const elem = document.createElement('div');
            elem.className = 'conversation';
            elem.textContent = conv.message;
            elem.style.backgroundColor = conv.sentiment === 'positive' ? 'lightgreen' : 'lightpink';
            this.container!.prepend(elem);

            gsap.fromTo(elem, 
                { opacity: 0, y: -50, rotationX: -90 },
                { opacity: 1, y: 0, rotationX: 0, duration: 0.5, delay: index * 0.1 }
            );
        });

        this.updatePositions();
    }

    private updatePositions() {
        if (!this.container) return;
        const elements = this.container.querySelectorAll('.conversation');
        elements.forEach((elem, index) => {
            const distanceFromCenter = index - this.currentIndex;
            gsap.to(elem, {
                z: distanceFromCenter * -100,
                y: distanceFromCenter * 5,
                rotationX: distanceFromCenter * 5,
                opacity: 1 - Math.abs(distanceFromCenter) * 0.2,
                scale: 1 + this.zoomLevel * 0.1 - Math.abs(distanceFromCenter) * 0.05,
                duration: 0.1,
                zIndex: 100 - Math.abs(distanceFromCenter)
            });
        });
    }

    private cycle(direction: number) {
        this.currentIndex = Math.max(0, Math.min(this.conversations.length - 1, this.currentIndex + direction));
        this.updatePositions();
    }

    private zoom(direction: number) {
        this.zoomLevel = Math.max(-3, Math.min(3, this.zoomLevel + direction));
        this.updatePositions();
    }

    private reset() {
        this.currentIndex = 0;
        this.zoomLevel = 0;
        this.updatePositions();
    }
}

customElements.define('team-conversation-view', TeamConversationView);