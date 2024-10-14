import { YouiVariant } from "./types";

export class YouiForm extends HTMLElement {
    public shadowRoot: ShadowRoot;
    private error: string = "";

    constructor() {
        super();
        this.shadowRoot = this.attachShadow({ mode: "open" });
    }

    static get observedAttributes() {
        return ["error"];
    }

    connectedCallback() {
        this.render();
    }

    attributeChangedCallback(name: string, oldValue: string, newValue: string) {
        if (oldValue !== newValue) {
            switch (name) {
                case "error":
                    this.error = newValue as YouiVariant;
                    break;
            }
        }
    }

    private render() {
        const style = document.createElement('style');
        style.textContent = `
            :host {
                display: inline-block;
            }
            .youi-form {
            }
        `;

        const form = document.createElement('form');
        form.className = `youi-form`;
        form.setAttribute('role', 'form');
        form.innerHTML = `
        <slot></slot>
        `;

        this.shadowRoot.appendChild(style);
        this.shadowRoot.appendChild(form);
    }
}

customElements.define("youi-form", YouiForm);
