import gsap from "gsap";
import { TextPlugin } from "gsap/TextPlugin";

gsap.registerPlugin(TextPlugin)

class TypeWriter extends HTMLElement {
    private template = document.createElement('template');
    
    constructor() {
        super();
        this.template.innerHTML = `
            <style>
                :host {
                    display: block;
                    width: 100%;
                    height: 100%;
                }

                .typewriter {
                    width: 100%;
                    height: 100%;
                    color: #333;
                }
            </style>
            <div class="typewriter"></div>
        `;
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
        this.render();
    }

    render() {
        gsap.fromTo(this.shadowRoot?.querySelector(".typewriter")!, {
            opacity: 0
        }, {
            opacity: 1,
            duration: 1,
            text: "Connecting to the worker pool..."
        });
    }
}

customElements.define("type-writer", TypeWriter);