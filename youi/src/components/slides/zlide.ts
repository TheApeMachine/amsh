import gsap from 'gsap';
import { Observer } from 'gsap/Observer';
import { Flip } from 'gsap/Flip';
gsap.registerPlugin(Flip, Observer);

class YouiZlide extends HTMLElement {
    private template: HTMLTemplateElement;
    private panels: NodeListOf<Element> | null = null;
    private totalPanels: number = 0;
    private scaleRatio: number = 0.9;
    private zDistance: number = 500; // Distance between panels on z-axis
    private scrollPosition: number = 0;
    private scrollSpeed: number = 1; // Reduced for smoother movement
    private tl: gsap.core.Timeline = gsap.timeline();
    private scrollTimeout: number | null = null;
    private isScrolling: boolean = false;

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
            <div id="zlides">
            </div>
        `;
        this.attachShadow({ mode: 'open' });
    }

    onConnectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
        this.panels = this.shadowRoot?.querySelectorAll(".zlide") ?? null;
        this.totalPanels = this.panels?.length ?? 0;

        Observer.create({
            target: window,
            type: "wheel,touch,pointer",
            onWheel: (e) => {
                console.log("onWheel");
                this.onScroll();
            },
            onDrag: (e) => {
                console.log("onDrag");
                this.onScroll();
            }
        });
    }

    moveCard() {
        const lastItem = this.shadowRoot?.querySelector("#zlides:last-child");

        if (this.shadowRoot && lastItem) {
            lastItem.style.display = "none"; // Hide the last item
            const newItem = document.createElement("article");
            newItem.className = lastItem.className; // Set the same class name
            newItem.textContent = lastItem.textContent; // Copy the text content
            this.shadowRoot?.insertBefore(newItem, this.shadowRoot?.firstChild); // Insert the new item at the beginning of the slider
        }
    }

    onScroll() {
        console.log("onScroll");
        let state = Flip.getState(".item");

        this.moveCard();

        Flip.from(state, {
            targets: ".zlide",
            ease: "sine.inOut",
            absolute: true,
            onEnter: (elements) => {
                return gsap.from(elements, {
                    yPercent: 20,
                    opacity: 0,
                    ease: "sine.out"
                });
            },
            onLeave: (element) => {
                return gsap.to(element, {
                    yPercent: 20,
                    xPercent: -20,
                    transformOrigin: "bottom left",
                    opacity: 0,
                    ease: "sine.out",
                    onComplete() {
                        console.log("logging", element[0])
                        this.shadowRoot?.removeChild(element[0]);
                    }
                });
            }
        });
    }
}

customElements.define("youi-zlide", YouiZlide);