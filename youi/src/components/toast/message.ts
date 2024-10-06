import { gsap } from "gsap"

class ToastMessage extends HTMLElement {
    message: string;
    type: string;
    index: number;
    total: number;
    isHovered: boolean;

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.message = '';
        this.type = '';
        this.index = 0;
        this.total = 0;
        this.isHovered = false;
    }

    static get observedAttributes() {
        return ['message', 'type', 'index', 'total', 'is-hovered'];
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (oldValue !== newValue) {
            this[name.replace(/-./g, x => x[1].toUpperCase())] = newValue;
            this.render();
        }
    }

    connectedCallback() {
        this.render();
        this.animateIn();
    }

    disconnectedCallback() {
        this.animateOut();
    }

    getPosition() {
        if (this.isHovered) {
            const angle = (this.index / this.total) * 360;
            const radius = 150;
            return {
                x: Math.sin(angle * Math.PI / 180) * radius,
                y: Math.cos(angle * Math.PI / 180) * radius,
                z: 0,
                rotateY: -angle,
                scale: 1,
            };
        } else {
            return {
                x: 0,
                y: this.index * 10,
                z: -this.index * 50,
                rotateY: 0,
                scale: 1 - this.index * 0.05,
            };
        }
    }

    animateIn() {
        gsap.fromTo(this.shadowRoot.host, {
            x: 300,
            opacity: 0,
        }, {
            x: this.getPosition().x,
            y: this.getPosition().y,
            z: this.getPosition().z,
            opacity: 1,
            duration: 0.5,
            ease: 'power4.out',
        });
    }

    animateOut() {
        gsap.to(this.shadowRoot.host, {
            x: 300,
            opacity: 0,
            duration: 0.5,
            ease: 'power4.in',
            onComplete: () => this.remove(),
        });
    }

    handleClose() {
        this.dispatchEvent(new CustomEvent('close-toast', { bubbles: true, composed: true }));
    }

    render() {
        this.shadowRoot.innerHTML = `
        <style>
          :host {
            position: absolute;
            width: 250px;
            background-color: white;
            border-radius: 8px;
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
            overflow: hidden;
            z-index: 99999;
          }
          .toast-content {
            padding: 16px;
            display: flex;
            align-items: center;
            gap: 12px;
          }
          .close-button {
            cursor: pointer;
            margin-left: auto;
            background: none;
            border: none;
            color: gray;
          }
          .close-button:hover {
            color: darkgray;
          }
        </style>
        <div class="toast-content">
          <span class="icon">Bla</span>
          <p>${this.message}</p>
          <button class="close-button" @click="${() => this.handleClose()}">X</button>
        </div>
      `;
    }
}

customElements.define('toast-message', ToastMessage);
