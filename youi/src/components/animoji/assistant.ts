import { AnimojiStates } from "./types";
import gsap from "gsap";

class AnimojiAssistant extends HTMLElement {
    private baseDir = '/src/assets/animoji/';
    private template = document.createElement('template');
    private shadow: ShadowRoot;
    private animojiContainer: HTMLDivElement;
    private currentState: string = 'idle';
    private currentSet: (index: number) => Animoji;
    private currentIndex: number = 0;
    private isTransitioning: boolean = false;
    private lastSwitchTime: number = 0;
    private players: dotlottie[] = [];
    private wrapper;

    constructor() {
        super();
        this.template.innerHTML = `
            <style>
                :host {
                    display: block;
                    position: absolute;
                }
                dotlottie-player {
                    position: absolute;
                    width: 100%;
                    height: 100%;
                    opacity: 0;
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
                    filter: drop-shadow(0px 0px 8px rgba(0, 0, 0, 0.8));
                }
            </style>
            <div id="animoji-container"></div>
        `;


        this.shadow = this.attachShadow({ mode: 'open' });
        this.shadow.appendChild(this.template.content.cloneNode(true));
        this.animojiContainer = this.shadow.getElementById('animoji-container') as HTMLDivElement;
        
        // We only need two players to pull of the crosfade transition. We can
        // manage the src of the players to dynamically switch in and out Animojis.
        this.players.push(this.createPlayer());
        this.players.push(this.createPlayer());
    }

    createPlayer() {
        const player = document.createElement('dotlottie-player');
        player.setAttribute('background', 'transparent');
        player.setAttribute('speed', '1');
        player.setAttribute('loop', 'true');
        player.setAttribute('autoplay', 'false');
        this.animojiContainer.appendChild(player);
        return player;
    }

    loadNextAnimoji() {
        const index = this.currentIndex % this.currentSet.length;
        const animoji = this.currentSet(index);
        this.players[this.currentIndex % 2].setAttribute(
            'src', `${this.baseDir}${animoji}/lottie.json`
        );
    }

    playNextAnimoji() {
        this.isTransitioning = true;
        const tl = gsap.timeline({
            onComplete: () => {
                this.isTransitioning = false;
                this.lastSwitchTime = Date.now();
            }
        });
        tl.to(this.players[1], { opacity: 0, duration: 0.25, ease: 'power2.inOut' });
        tl.to(this.players[0], { opacity: 1, duration: 0.25, ease: 'power2.inOut' }, "<");
        tl.play();
        this.isTransitioning = false;
    }

    setState(newState: string) {
        if (
            this.currentState !== newState 
            && !this.isTransitioning 
            && Date.now() - this.lastSwitchTime >= 3000
        ) {
            this.currentState = newState;
            this.currentSet = AnimojiStates[this.currentState];
            this.currentIndex = 0;
            this.playNextAnimoji();
        }
    }

}

customElements.define('animoji-assistant', AnimojiAssistant);