import { gsap } from "gsap";

class LayersManager extends HTMLElement {
    private template: HTMLTemplateElement;
    private layers: Record<string, HTMLElement> = {};
    private currentLayer: number = 0;
    private targetLayer: number = 0;
    private positions: string[] = [];
    private distance: number = 1000;
    private duration: number = 10;
    private easy: number = 1;

    constructor() {
        super();
        this.template = document.createElement("template");
        this.template.innerHTML = `
            <style>
                :host {
                    display: flex;
                    width: 100%;
                    height: 100%;
                    transform-style: preserve-3d;
                    perspective: 1500px;
                    position: absolute;
                    top: 0;
                    left: 0;
                }
                #layers {
                    display: flex;
                    width: 100%;
                    height: 100%;
                    transform-style: preserve-3d;
                    perspective: 1500px;
                    position: absolute;
                    top: 0;
                    left: 0;
                }
            </style>
            <div id="layers">
                <slot></slot>
            </div>
        `;
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        // Append template content to shadowRoot
        this.shadowRoot!.appendChild(this.template.content.cloneNode(true));

        // Access the slot and its assigned elements
        const slot = this.shadowRoot!.querySelector('slot') as HTMLSlotElement;
        const assignedElements = slot.assignedElements({ flatten: true }) as HTMLElement[];
        console.log("Assigned Elements:", assignedElements);

        this.positions = assignedElements.map((element) => element.getAttribute("data-effect") || "");
        console.log("Positions Array:", this.positions);

        // Populate layers from assigned elements
        assignedElements.forEach((element, index) => {
            if (element instanceof HTMLElement) {
                const effect = element.getAttribute("data-effect");
                if (effect) {
                    console.log("Setting GSAP properties for:", effect);
                    gsap.set(element, {
                        position: "absolute",
                        width: "100%",
                        height: "100%",
                        transformStyle: "preserve-3d",
                        backfaceVisibility: "hidden",
                        z: index * -this.distance,
                        zIndex: this.positions.length - index
                    });
    
                    this.layers[effect] = element;
                }
            }
        });
        console.log("Layers Object:", this.layers);

        this.currentLayer = 0;
        this.targetLayer = 0;
        window.stateManager.register("dashboard", this.goto.bind(this, "dashboard"));
        window.stateManager.register("slides", this.goto.bind(this, "slides"));
        window.stateManager.register("table", this.goto.bind(this, "table"));
    }

    public goto(position: string) {
        this.targetLayer = this.positions.indexOf(position);
        if (this.targetLayer === -1) {
            console.error(`Target layer not found: ${position}`);
            return;
        }
        console.log("Goto target layer:", position, this.targetLayer);
        this.transition();
    }

    private transition() {
        console.log("Transition from currentLayer to targetLayer:", this.currentLayer, this.targetLayer);
        if (this.currentLayer === this.targetLayer) return;

        const steps = this.targetLayer - this.currentLayer;
        const tl = gsap.timeline({ onComplete: () => {
            console.log("Animation complete. Updating currentLayer to targetLayer:", this.targetLayer);
            this.currentLayer = this.targetLayer;
        }});
        const animationOrder = steps > 0 ? this.positions : [...this.positions].reverse();
        const reverse = steps < 0;

        animationOrder.forEach((position: string, index: number) => {
            console.log("Animating position:", position);
            const zPosition = (index - this.targetLayer + 1) * -this.distance;

            if (this.layers[position]) {
                tl.to(this.layers[position], {
                    z: zPosition,
                    duration: this.duration,
                    ease: `back.inOut(${this.easy})`,
                }, 0);

                if (index !== this.targetLayer) {
                    tl.to(this.layers[position], {
                        opacity: 0.5,
                        filter: "blur(10px)",
                        rotationX: reverse ? 20 : -20,
                        duration: this.duration / 2,
                        ease: `back.inOut(${this.easy})`,
                    }, this.duration / 4);

                    tl.to(this.layers[position], {
                        opacity: 1,
                        filter: "blur(0px)",
                        rotationX: 0,
                        duration: this.duration / 4,
                        ease: `back.inOut(${this.easy})`,
                    }, this.duration / 2);
                }
            } else {
                console.error(`Layer not found for position: ${position}`);
            }
        });

        // Adding a "camera" effect during the animation
        tl.to(this.shadowRoot!.host, {
            rotationY: reverse ? -45 : 45,
            transformOrigin: "50% 50% -500px",
            duration: this.duration,
            ease: `power2.inOut`,
        }, 0);

        tl.to(this.shadowRoot!.host, {
            rotationY: 0,
            duration: this.duration / 2,
            ease: `power2.inOut`,
        }, this.duration / 2);

        tl.play();
    }
}

customElements.define("layers-manager", LayersManager);