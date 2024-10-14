import { YouiVariant } from "./types";

class YouiInput extends HTMLElement {
    shadowRoot: ShadowRoot;
    template: HTMLTemplateElement;
    variant: YouiVariant = "default";
    pressed: boolean = false;

    constructor() {
        super();
        this.template = document.createElement("template");
        this.template.innerHTML = `
        <style>
            :host {
                --youi-unit: 1rem;
                --youi-unit-2: 2rem;
                --youi-radius: 0.25rem;
                --youi-border-width: 2px;
                --youi-muted: #ccc;
                display: flex;
                align-items: center;
                justify-content: center;
                padding: var(--youi-unit, 1rem);
                border: var(--youi-border-width, 2px) solid var(--youi-muted, #ccc);
                border-radius: var(--youi-radius, 0.25rem);
                cursor: pointer;
            }
            input {
                all: unset;
            }
            .youi-input {
                background: var(--button-bg, #f0f0f0);
                color: var(--button-color, #000);
                padding: var(--youi-unit, 1rem) var(--youi-unit-2, 2rem);
                border-radius: var(--youi-radius);
                min-width: 44px;
                min-height: 44px;
            }
            .youi-input:focus {
                outline: var(--youi-border-width, 2px) solid var(--youi-brand, hsl(268, 100%, 50%));
            }
            .youi-input[disabled] {
                opacity: 0.6;
                pointer-events: none;
            }
        </style>
        <input class="youi-input" role="input" type="text" tabindex="0">
            <slot></slot>
        </input>
        `;
        this.shadowRoot = this.attachShadow({ mode: "open" });
    }

    static get observedAttributes() {
        return ["variant", "pressed", "disabled"];
    }

    connectedCallback() {
        this.variant = this.getAttribute("variant") as YouiVariant || "default";
        this.pressed = this.hasAttribute("pressed");
        this.shadowRoot.appendChild(this.template.content.cloneNode(true));
        this.updateVariant();
        this.setAttribute('role', 'input');
        this.setAttribute('tabindex', '0');
        this.updatePressedState();

        // Keyboard interaction for accessibility.
        this.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                this.click(); // Simulate a click on keyboard interaction.
            }
        });
    }

    attributeChangedCallback(name: string, oldValue: string, newValue: string) {
        if (name === "variant") {
            this.variant = newValue as YouiVariant;
            this.updateVariant();
        }
        if (name === "pressed") {
            this.pressed = newValue !== null;
            this.updatePressedState();
        }
    }

    updateVariant() {
        const input = this.shadowRoot?.querySelector(".youi-input");
        if (input) {
            input.className = "youi-input"; // Reset class to avoid stacking classes
            if (this.variant) {
                input.classList.add(this.variant);
            }
        }
    }

    updatePressedState() {
        if (this.pressed) {
            this.setAttribute('aria-pressed', 'true');
        } else {
            this.removeAttribute('aria-pressed');
        }
    }

    // Overriding the click method to add pressed state handling.
    click() {
        if (this.hasAttribute('disabled')) return; // Prevent action if disabled
        super.click();
        if (this.pressed !== undefined) {
            this.pressed = !this.pressed;
            if (this.pressed) {
                this.setAttribute('pressed', '');
            } else {
                this.removeAttribute('pressed');
            }
        }
    }
}

customElements.define("youi-input", YouiInput);