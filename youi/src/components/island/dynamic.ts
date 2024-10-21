import { gsap } from "gsap";
import { Flip } from "gsap/Flip";
import "@/components/ui/button";

gsap.registerPlugin(Flip);

export class DynamicIsland extends HTMLElement {
    private template = document.createElement('template');
    private island: HTMLElement | null = null;
    private state: any = null;

    constructor() {
        super();
        this.template.innerHTML = `
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
                    box-shadow: rgba(0, 0, 0, 0.4) 0px 2px 4px, rgba(0, 0, 0, 0.3) 0px 7px 13px -3px, rgba(0, 0, 0, 0.2) 0px -3px 0px inset;

                    &.button {
                        display: inline-grid;
                        width: unset;
                        height: unset;

                        > main {
                            box-shadow: 0px 0px 0px 3px #FFF;
                            padding: 0.25rem 1rem;
                            font-size: 1rem;
                            cursor: pointer;
                            border-radius: .125rem;
                            outline: 2px solid rgba(0, 0, 0, 1);
                            background: #FFF;
                            color: #333;
                        }
                    }
                    
                    &.card {
                        display: inline-grid;
                        width: 50%;
                        height: 50%;
                        
                        > main {
                            height: 100%;
                            box-shadow: 0px 0px 0px 3px #FFF;
                            padding: 0.25rem 1rem;
                            font-size: 1rem;
                            cursor: pointer;
                            border-radius: .125rem;
                            outline: 2px solid rgba(0, 0, 0, 1);
                            background: #FFF;
                            color: #333;
                        }
                    }
                    &.modal {
                        display: inline-grid;
                        width: 50%;
                        height: 50%;
                        box-shadow: rgba(220, 220, 220, 0.2) 0px 60px 40px -7px;
                        
                        > main {
                            height: 100%;
                            box-shadow: 0px 0px 0px 3px #FFF;
                            padding: 0.25rem 1rem;
                            font-size: 1rem;
                            cursor: pointer;
                            border-radius: .125rem;
                            outline: 2px solid rgba(0, 0, 0, 1);
                            background: #FFF;
                            color: #333;
                        }
                    }
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

                    > span {
                        padding: 0.25rem 1rem;
                        border-radius: 0.125rem;
                        font-size: 1rem;
                        box-shadow: rgba(0, 0, 0, 0.4) 0px 2px 4px, rgba(0, 0, 0, 0.3) 0px 7px 13px -3px, rgba(0, 0, 0, 0.2) 0px -3px 0px inset;
                    }
                }
                footer {
                    grid-area: footer;
                }
                article {
                    grid-area: flyout;
                }
            </style>
            <div id="island" class="button">
                <header></header>
                <aside></aside>
                <main><span>test</span></main>
                <article></article>
                <footer></footer>
            </div>
        `;

        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
        this.island = this.shadowRoot!.querySelector('#island') as HTMLElement;
        
        // Ensure the element is in the DOM before animating
        if (this.island) {
            // Delay the animation slightly to ensure DOM is ready
            setTimeout(() => {
                this.animateIsland();
            }, 0);
        } else {
            console.error('Island element not found');
        }
    }

    private animateIsland() {
        this.state = Flip.getState(this.island);
        this.island!.classList.remove('button');
        this.island!.classList.add('card');
        Flip.from(this.state, {
            duration: 1,
            ease: "power.inOut",
            absolute: true,
            onComplete: () => {
                console.log('complete');
            },
        });
        this.state = Flip.getState(this.island);
        this.island!.classList.remove('card');
        this.island!.classList.add('modal');
        Flip.from(this.state, {
            duration: 1,
            ease: "power.inOut",
            absolute: true,
            onComplete: () => {
                console.log('complete');
            },
        });
    }

}

customElements.define('dynamic-island', DynamicIsland);
