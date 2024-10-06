class ToastContainer extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.toasts = [];
        this.isHovered = false;
    }

    connectedCallback() {
        this.render();
    }

    addToast(message, type) {
        const id = Date.now();
        this.toasts = [{ id, message, type }, ...this.toasts];
        this.render();
        setTimeout(() => this.removeToast(id), 5000);
    }

    removeToast(id) {
        this.toasts = this.toasts.filter(toast => toast.id !== id);
        this.render();
    }

    handleMouseEnter() {
        this.isHovered = true;
        this.render();
    }

    handleMouseLeave() {
        this.isHovered = false;
        this.render();
    }

    render() {
        this.shadowRoot.innerHTML = `
        <style>
          :host {
            position: fixed;
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
          ${this.toasts.map((toast, index) => `
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
      `;
    }
}

customElements.define('toast-container', ToastContainer);
