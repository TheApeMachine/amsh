import { Animoji, AnimojiStates } from "./types";
import gsap from "gsap";
import '@dotlottie/player-component';

class AnimojiAssistant extends HTMLElement {
    private template = document.createElement('template');
    private baseDir = '/src/assets/animoji/';
    private shadow: ShadowRoot;
    private animojiContainer!: HTMLDivElement;
    private currentState: keyof typeof AnimojiStates = 'idle';
    private currentIndex: number = 0;
    private isTransitioning: boolean = false;
    private lastSwitchTime: number = 0;
    private players: any
    private currentPlayer: HTMLElement | null = null;
    private cycleInterval: number | null = null;
    private chatContainer!: HTMLDivElement;
    private chatInput!: HTMLInputElement;
    private chatOutput!: HTMLDivElement;
    private boundKeydownHandler: (event: KeyboardEvent) => void;

    constructor() {
        super();
        this.template.innerHTML = `
            <style>
                :host {
                    display: block;
                    position: absolute;
                    z-index: 1000;
                    width: 100%;
                    height: 100%;
                }
                dotlottie-player {
                    position: absolute;
                    width: 100%;
                    height: 100%;
                    opacity: 1;
                }
                dotlottie-player.active {
                    opacity: 1;
                }
                #animoji-container {
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    width: 64px;
                    height: 64px;
                    position: absolute;
                    top: 0;
                    left: 0;
                    overflow: hidden;
                    filter: drop-shadow(0px 0px 1rem rgba(0, 0, 0, 0.1));
                    z-index: 2000;
                }
            </style>
            <div class="dynamic-island">
                <div id="animoji-container"></div>
            </div>
        `;

        this.shadow = this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.shadow.appendChild(this.template.content.cloneNode(true));
        this.animojiContainer = this.shadow.getElementById('animoji-container') as HTMLDivElement;
        this.lastSwitchTime = Date.now();
        
        this.players = {};
        // Create the animoji players.
        for (const state in AnimojiStates) {
            this.players[state as keyof typeof AnimojiStates] = [];
            const stateAnimojis = AnimojiStates[state as keyof typeof AnimojiStates];
            if (typeof stateAnimojis === 'function') {
                const animojis = new Set();
                for (let i = 0; i < 5; i++) {
                    const animoji = stateAnimojis(i);
                    if (!animojis.has(animoji)) {
                        animojis.add(animoji);
                        const player = this.createPlayer(animoji);
                        this.players[state as keyof typeof AnimojiStates].push(player);
                        this.animojiContainer.appendChild(player);
                    }
                }
            }
        }

        this.setState('idle');
        this.startCycling();
        console.log('AnimojiAssistant connected and initialized');
    }

    createPlayer(animoji: Animoji) {
        const player = document.createElement('dotlottie-player') as any;
        player.setAttribute('background', 'transparent');
        player.setAttribute('speed', '1');
        player.setAttribute('loop', 'true');
        player.setAttribute('autoplay', 'true');
        const src = `${this.baseDir}${animoji}/lottie.json`;
        player.setAttribute('src', src);
        player.style.opacity = '0'; // Set initial opacity to 0
        player.addEventListener('error', (e: any) => console.error('Lottie player error:', e));
        console.log(`Created player with src: ${src}`);
        return player;
    }

    startCycling() {
        if (this.cycleInterval) {
            clearInterval(this.cycleInterval);
        }
        this.cycleInterval = window.setInterval(() => {
            this.playNextAnimoji();
        }, 5000); // Cycle every 5 seconds
    }

    playNextAnimoji() {
        const currentPlayers = this.players[this.currentState];
        if (!currentPlayers || currentPlayers.length === 0) {
            console.error(`No players found for state ${this.currentState}`);
            return;
        }

        this.isTransitioning = true;
        this.currentIndex = (this.currentIndex + 1) % currentPlayers.length;
        const nextPlayer = currentPlayers[this.currentIndex];

        const tl = gsap.timeline({
            onComplete: () => {
                this.currentPlayer = nextPlayer;
                this.isTransitioning = false;
                this.lastSwitchTime = Date.now();
                console.log(`Transitioned to ${this.currentState} animation at index ${this.currentIndex}`);
            }
        });

        if (this.currentPlayer) {
            tl.to(this.currentPlayer, { opacity: 0, duration: 0.25, ease: 'power2.inOut' });
        }
        tl.to(nextPlayer, { opacity: 1, duration: 0.25, ease: 'power2.inOut' }, "<");
        tl.play();
    }

    setState(newState: keyof typeof AnimojiStates | 'chat') {
        console.log(`Setting state to ${newState}`);
        if (this.currentState !== newState && !this.isTransitioning) {
            if (newState === 'chat') {
                this.enterChatMode();
            } else {
                if (this.currentState === 'chat') {
                    this.exitChatMode();
                }
                this.currentState = newState as keyof typeof AnimojiStates;
                this.currentIndex = -1;
                this.playNextAnimoji();
            }
        }
    }

    enterChatMode() {
        this.currentState = 'chat';
        gsap.to(this.animojiContainer, {
            width: '300px',
            height: '400px',
            duration: 0.5,
            ease: 'power2.inOut',
            onComplete: () => {
                this.chatContainer.style.display = 'block';
                this.typewriterEffect("Hello! How can I assist you today?");
                this.chatInput.focus(); // Focus on the input field
            }
        });
    }

    disconnectedCallback() {
        if (this.cycleInterval) {
            clearInterval(this.cycleInterval);
        }

        // Remove the keydown event listener
        document.removeEventListener('keydown', this.boundKeydownHandler);
    }

}

customElements.define('animoji-assistant', AnimojiAssistant);
