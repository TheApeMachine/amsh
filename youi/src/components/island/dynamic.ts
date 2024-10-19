import { gsap } from "gsap";
import { Flip } from "gsap/Flip";
import "@/components/ui/button";

gsap.registerPlugin(Flip);

export class DynamicIsland extends HTMLElement {
    private island: HTMLElement | null = null;
    private contentElement: HTMLElement | null = null;
    private currentState: string = 'closed';
    private states: Record<string, any> = {};

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.island = this.shadowRoot!.querySelector('#island') as HTMLElement;
        this.setupStates();

        this.render();
        ["closed", "button", "closed"].forEach((state) => {
            setTimeout(() => {
                this.morphTo(state);
            }, 3000);
        });
    }

    private render() {
        this.shadowRoot!.innerHTML = `
            <style>
                :host {
                    display: block;
                    width: 100%;
                    height: 100%;
                }
                #island {
                    display: grid; 
                    grid-template-columns: auto 1fr auto; 
                    grid-template-rows: auto 1fr auto; 
                    grid-template-areas: 
                        "header header header"
                        "sidebar main flyout"
                        "sidebar footer flyout";
                    width: 100%;
                    height: 100%;
                }
                header {
                    grid-area: header;
                }
                aside {
                    grid-area: sidebar;
                }
                main {
                    grid-area: main;
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    justify-content: center;
                }
                footer {
                    grid-area: footer;
                }
                article {
                    grid-area: flyout;
                }
            </style>
            <div id="island">
                <header></header>
                <aside></aside>
                <main></main>
                <article></article>
                <footer></footer>
            </div>
        `;
    }

    private setupStates() {
        // Define your states here
        this.states = {
            closed: { content: '' },
            button: { content: '<youi-button>Click me</youi-button>' },
            // ... other states ...
        };
    }

    public morphTo(state: string) {
        console.log('morphTo', state);
        if (this.states.hasOwnProperty(state) && this.island && this.contentElement) {
            const config = this.states[state];

            // Capture the current state
            const state = Flip.getState(this.island);

            // Update content
            this.contentElement.innerHTML = config.content;

            // Update styles
            Object.assign(this.island.style, {
                width: config.width,
                height: config.height,
                // ... other style properties ...
            });

            // Animate the change
            Flip.from(state, {
                duration: 0.5,
                ease: "power1.inOut",
                onComplete: () => {
                    this.currentState = state;
                    this.dispatchEvent(new CustomEvent('stateChanged', { detail: { newState: state }, bubbles: true, composed: true }));
                }
            });

            // If we're transitioning to a component with enter animation (like our button)
            const component = this.contentElement.firstElementChild as any;
            if (component && typeof component.playEnterAnimation === 'function') {
                component.playEnterAnimation();
            }
        }
    }

    public async transitionFrom(state: string) {
        if (this.currentState === state && this.contentElement) {
            const component = this.contentElement.firstElementChild as any;
            if (component && typeof component.playExitAnimation === 'function') {
                await component.playExitAnimation();
            }
        }
    }
}

customElements.define('dynamic-island', DynamicIsland);
