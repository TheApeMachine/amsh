import { gsap } from "gsap";
import { Flip } from "gsap/Flip";
import "@/components/ui/button";
import "@/components/animoji/assistant";
import "@/components/island/moodlight";
import "@/components/animoji/typewriter";

gsap.registerPlugin(Flip);

export class DynamicIsland extends HTMLElement {
    private template = document.createElement('template');
    private island: HTMLElement | null = null;
    private tl: gsap.core.Timeline | null = gsap.timeline();

    private content: {
        [key: string]: {
            element: string,
            target: string
        }
    } = {
        island: {
            element: "mood-light",
            target: "main"
        },
        animoji: {
            element: "animoji-assistant",
            target: "aside"
        }
    }

    private animations: {
        [key: string]: (selector: HTMLElement) => gsap.core.Timeline
    } = {
        island: (selector: HTMLElement) => this.tl!.to(selector, {
            boxShadow: '0px 0px 0px 3px #999',
            outline: '2px solid rgba(0, 0, 0, 1)',
            borderRadius: '1rem',
            background: '#FFF',
            color: '#333',
            transform: 'scale(1)',
            ease: 'back.inOut',
        }),
        hover: (selector: HTMLElement) => this.tl!.to(selector, {
            boxShadow: '0px 0px 0px 3px #999',
            outline: '2px solid rgba(0, 0, 0, 1)',
            borderRadius: '1rem',
            background: '#FFF',
            color: '#333',
            transform: 'scale(2)',
            ease: 'back.inOut',
        }),
        button: (selector: HTMLElement) => this.tl!.to(selector, {
            boxShadow: '0px 0px 0px 3px #FFF',
            padding: '0.25rem 1rem',
            fontSize: '1rem',
            ease: 'back.inOut',
        }),
        animoji: (selector: HTMLElement) => this.tl!.to(selector, {
            boxShadow: '0px 0px 0px 3px #FFF',
            ease: 'back.inOut',
            borderRadius: '0.125rem',
            width: "25%",
            height: 64
        }),
    };

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
                    display: inline-grid; 
                    grid-template-columns: auto 1fr auto; 
                    grid-template-rows: auto 1fr auto; 
                    grid-template-areas: 
                        "header header header"
                        "sidebar main flyout"
                        "sidebar footer flyout";
                    box-shadow: 0px 0px 0px 3px #999;
                    padding: 0.25rem 1rem;
                    font-size: 1rem;
                    cursor: pointer;
                    border-radius: 1rem;
                    outline: 2px solid rgba(0, 0, 0, 1);
                    background: #FFF;
                    color: #333;
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
                        box-shadow: 
                            rgba(0, 0, 0, 0.4) 0px 2px 4px, 
                            rgba(0, 0, 0, 0.3) 0px 7px 13px -3px, 
                            rgba(0, 0, 0, 0.2) 0px -3px 0px inset;
                    }
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

        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
        this.island = this.shadowRoot!.querySelector('#island') as HTMLElement;
        this.animations.island(this.island);

        this.island!.addEventListener('mouseenter', () => this.render('hover'));
        this.island!.addEventListener('mouseleave', () => this.render('island'));

        setTimeout(() => {
            this.render("animoji");
            this.shadowRoot?.querySelector("main")!.appendChild(document.createElement("type-writer"));
        }, 3000);
    }

    render(state: string) {
        this.tl?.clear();

        this.shadowRoot!.querySelector(
            this.content[state].target
        )!.appendChild(
            document.createElement(
                this.content[state].element
            )
        );
        
        this.animations[state](this.island!).play();
    }

}

customElements.define('dynamic-island', DynamicIsland);
