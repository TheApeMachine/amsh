class LoaderComponent extends HTMLElement {
    template: HTMLTemplateElement;

    constructor() {
        super();
        this.template = document.createElement("template");
        this.template.innerHTML = `
            <div class="loader">Loading...</div>
        `;
        this.attachShadow({mode: "open"});
    }

    connectedCallback() {
        this.shadowRoot!.appendChild(this.template.content.cloneNode(true));
    }
}

customElements.define("loader-component", LoaderComponent);