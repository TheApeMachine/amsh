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
    private scrollPosition: number = 0;
    private scrollSpeed: number = 2; // Reduced for smoother movement
    private isScrolling: boolean = false;
    private zDistance: number = 500; // Distance between panels on z-axis
    private activeIndex: number = 0;
    private isAtRest: boolean = true;
    private scrollTimeout: number | null = null;

    constructor() {
        super();
        this.template = document.createElement('template');
        this.template.innerHTML = `
            <style>
                :host {
                    display: block;
                    width: 100vw;
                    height: 100vh;
                    perspective: 1500px;
                    overflow: hidden;
                }
                .zlide {
                    position: absolute;
                    top: 0;
                    left: 0;
                    width: 100%;
                    height: 100%;
                    background-color: #f0f0f0;
                    box-shadow: 0 10px 20px rgba(0,0,0,0.2);
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    font-size: 24px;
                    color: #333;
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
            target: window,
            type: "wheel,touch,pointer",
            onWheel: (e) => this.handleScroll(e.deltaY),
            onDrag: (e) => this.handleScroll(e.deltaY),
            onStop: () => this.stopScrolling(),
        });

        this.updatePanelPositions(true); // Initial positioning
    }

    handleScroll(deltaY: number) {
        this.isAtRest = false;
        this.scrollPosition += deltaY * this.scrollSpeed;
        this.updatePanelPositions(false);
    }

    stopScrolling() {
        this.isScrolling = false;
        this.resetToActivePanel();
    }

    resetToActivePanel() {
        this.isAtRest = true;
        this.activeIndex = Math.round(this.scrollPosition / this.zDistance) % this.totalPanels;
        if (this.activeIndex < 0) this.activeIndex += this.totalPanels;
        this.scrollPosition = this.activeIndex * this.zDistance;
        this.updatePanelPositions(true);
    }

    updatePanelPositions(smooth: boolean) {
        if (!this.panels) return;
        this.tl.clear();

        let zPos = ((1 * this.zDistance + this.scrollPosition) % (this.totalPanels * this.zDistance)) - (this.zDistance * (this.totalPanels - 1) / 2);
        let scale = Math.max(0.5, 1 - Math.abs(zPos) / (this.zDistance * this.totalPanels));
        let yPos = this.isAtRest ? 0 : -Math.abs(zPos) * 0.1 + 600;
        let opacity = scale;
        let xPos = 0; // Center horizontally
        const panels = gsap.utils.toArray(this.panels) as Element[];

        this.tl.to(panels, {
            x: xPos,
            y: yPos,
            z: zPos,
            scale: scale,
            opacity: opacity,
            duration: smooth ? 0.5 : 0.3,
            ease: smooth ? "power2.out" : "power1.out",
        });
    }
}

customElements.define("youi-zlide", YouiZlide);
