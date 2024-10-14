class YouiFormRow extends HTMLElement {
    public shadowRoot: ShadowRoot;

    constructor() {
        super();
        this.shadowRoot = this.attachShadow({ mode: "open" });
    }

    connectedCallback() {
        this.shadowRoot.innerHTML = `
        <style>
        .youi-form-row {
            display: flex;
            flex-direction: column;
            gap: 0.25rem;
        }
        </style>
        <div class="youi-form-row">
            <slot name="label"></slot>
            <slot name="input"></slot>
            <slot name="error"></slot>
            <slot name="hint"></slot>
        </div>
        `;
    }
}

customElements.define("youi-form-row", YouiFormRow);
