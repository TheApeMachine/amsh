class Approot extends HTMLElement {
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
                main {
                    display: flex;
                    flex-grow: 1;
                }
            </style>
            <main></main>
        `;
        this.attachShadow({mode: "open"});
    }

    connectedCallback() {
        this.shadowRoot!.appendChild(this.template.content.cloneNode(true));
    }

    public render(content: DocumentFragment) {
        this.shadowRoot!.querySelector("main")!.innerHTML = "";
        this.shadowRoot!.querySelector("main")!.appendChild(content);
    }
}

customElements.define("app-root", Approot);