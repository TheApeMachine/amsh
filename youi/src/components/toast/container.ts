class ToastContainer extends HTMLElement {
    private toasts: any;
    private isHovered: boolean;
    private template: HTMLTemplateElement;

    constructor() {
        super();
        this.template = document.createElement("template")
        this.template.innerHTML = `
        <style>
            :host {
                position: absolute;
                bottom: 16px;
                right: 16px;
                width: 300px;
                height: 400px;
                perspective: 1000px;
                z-index: 99999;
            }
        </style>
        <div
            @mouseenter="${() => this.handleMouseEnter()}"
            @mouseleave="${() => this.handleMouseLeave()}"
        >
            ${this.toasts?.map((toast, index) => `
                <toast-message
                    message="${toast.message}"
                    type="${toast.type}"
                    index="${index}"
                    total="${this.toasts.length}"
                    is-hovered="${this.isHovered}"
                    @close-toast="${() => this.removeToast(toast.id)}"
                ></toast-message>
            `).join('')}
        </div>
        `
        this.attachShadow({ mode: 'open' });
        this.toasts = [];
        this.isHovered = false;
    }

    connectedCallback() {
        this.shadowRoot!.appendChild(this.template.content.cloneNode(true));
    }

    addToast(message: string, type: string) {
        const id = Date.now();
        this.toasts = [{ id, message, type }, ...this.toasts];
        setTimeout(() => this.removeToast(id), 5000);
    }

    removeToast(id) {
        this.toasts = this.toasts.filter(toast => toast.id !== id);
    }

    handleMouseEnter() {
        this.isHovered = true;
    }

    handleMouseLeave() {
        this.isHovered = false;
    }
}

customElements.define('toast-container', ToastContainer);
