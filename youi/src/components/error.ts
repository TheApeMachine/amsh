class ErrorBoundary extends HTMLElement {
    template: HTMLTemplateElement;

    constructor() {
        super();
        this.template = document.createElement("template");
        this.template.innerHTML = `
            <style>
                :host {
                    display: flex;
                    width: 100%;
                    height: 100%;
                }
            </style>
            <dialog>
                <slot name="error"></slot>
            </dialog>    
        `;
        this.attachShadow({mode: "open"});
    }    

    connectedCallback() {
        this.shadowRoot!.appendChild(this.template.content.cloneNode(true));
    }

    public render(error: string) {
        this.shadowRoot!.querySelector("dialog")!.innerHTML = `<pre>${error}</pre>`;
    }
}

customElements.define("error-boundary", ErrorBoundary);
