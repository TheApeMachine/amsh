import gsap from 'gsap';
import ScrollTrigger from 'gsap/ScrollTrigger';
import { Observer } from 'gsap/Observer';

gsap.registerPlugin(ScrollTrigger, Observer)

class YouiZlide extends HTMLElement {
    private template: HTMLTemplateElement;
    private zScaler: number = 0.5;
    private tl: gsap.core.Timeline = gsap.timeline();
    private panels: NodeListOf<Element> | null = null;
    private totalPanels: number = 0;
    private spacing: number = 50;
    private scaleRatio: number = 0.9;
    private rotationX: number = 0;
    private elevationStep: number = 50;
    private laidOut: boolean = false;

    constructor() {
        super();
        this.template = document.createElement('template');
        this.template.innerHTML = `
            <style>
                :host {
                    position: relative;
                    width: 100%;
                    height: 100%;
                    perspective: 500px;
                }
                .zlide {
                    position: absolute;
                    top: 0;
                    left: 0;
                    width: 100%;
                    height: 100%;
                    background-color: #000;
                }
            </style>
            <article class="zlide">
                <h1>Hello World</h1>
            </article>
            <article class="zlide">
                <h1>Hello World</h1>
            </article>
            <article class="zlide">
                <h1>Hello World</h1>
            </article>
            <article class="zlide">
                <h1>Hello World</h1>
            </article>
            <article class="zlide">
                <h1>Hello World</h1>
            </article>
            <article class="zlide">
                <h1>Hello World</h1>
            </article>
            <article class="zlide">
                <h1>Hello World</h1>
            </article>
        `;
        this.attachShadow({ mode: 'open' });
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
        this.panels = this.shadowRoot?.querySelectorAll(".zlide");
        this.totalPanels = this.panels?.length || 0;
        this.spacing = 50; // Adjust this value to increase/decrease spacing between panels
        this.scaleRatio = 0.9; // Adjust this value to control how quickly panels shrink
        this.rotationX = 0; // Adjust this value to change the "camera angle"
        this.elevationStep = 50; // The amount each panel is raised relative to the one in front
        Observer.create({
            target: window, // can be any element (selector text is fine)
            type: "wheel,touch", // comma-delimited list of what to listen for
            onUp: () => this.animate("+="),
            onDown: () => this.animate("-="),
          });
    }

    animate(direction: string = "+=") {
        this.tl.clear()
        this.tl.to(gsap.utils.toArray(this.panels), {
            z: direction + "1000",
            scale: direction + "0.1",
            y: direction + "100",
            duration: 0.1,
            ease: "power2.inOut",
            stagger: 0.01
        });
        this.tl.play()
    }
}

customElements.define("youi-zlide", YouiZlide);