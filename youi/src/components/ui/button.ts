import { YouiVariant } from "./types";
import { gsap } from "gsap";

export class YouiButton extends HTMLElement {
    private shadowRoot: ShadowRoot;
    private variant: YouiVariant = "default";
    private size: 'small' | 'medium' | 'large' = 'medium';
    private loading: boolean = false;
    private animationTimeline: gsap.core.Timeline | null = null;

    constructor() {
        super();
        this.shadowRoot = this.attachShadow({ mode: "open" });
    }

    static get observedAttributes() {
        return ["variant", "size", "disabled", "loading"];
    }

    connectedCallback() {
        this.render();
        this.setupEventListeners();
        this.playEnterAnimation();
    }

    private playEnterAnimation() {
        const button = this.shadowRoot!.querySelector('.youi-button') as HTMLElement;
        this.animationTimeline = gsap.timeline()
            .from(button, {
                scale: 0.5,
                opacity: 0,
                duration: 0.3,
                ease: "back.out(1.7)"
            });
    }

    attributeChangedCallback(name: string, oldValue: string, newValue: string) {
        if (oldValue !== newValue) {
            switch (name) {
                case "variant":
                    this.variant = newValue as YouiVariant;
                    this.updateVariant();
                    break;
                case "size":
                    this.size = newValue as 'small' | 'medium' | 'large';
                    this.updateSize();
                    break;
                case "disabled":
                    this.updateDisabledState();
                    break;
                case "loading":
                    this.loading = newValue !== null;
                    this.updateLoadingState();
                    break;
            }
        }
    }

    private render() {
        const style = document.createElement('style');
        style.textContent = `
            :host {
                --youi-primary: #3498db;
                --youi-secondary: #2ecc71;
                --youi-danger: #e74c3c;
                --youi-unit: 1rem;
                --youi-radius: 0.25rem;
                display: inline-block;
            }
            .youi-button {
                font-family: 'Arial', sans-serif;
                border: none;
                border-radius: var(--youi-radius);
                cursor: pointer;
                transition: all 0.3s ease;
                display: inline-flex;
                align-items: center;
                justify-content: center;
                position: relative;
                overflow: hidden;
            }
            .youi-button:focus {
                outline: none;
                box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.5);
            }
            .youi-button[disabled] {
                opacity: 0.6;
                cursor: not-allowed;
            }
            .youi-button.default {
                background-color: #f1f1f1;
                color: #333;
            }
            .youi-button.primary {
                background-color: var(--youi-primary);
                color: white;
            }
            .youi-button.secondary {
                background-color: var(--youi-secondary);
                color: white;
            }
            .youi-button.danger {
                background-color: var(--youi-danger);
                color: white;
            }
            .youi-button.small {
                font-size: 0.8rem;
                padding: 0.4rem 0.8rem;
            }
            .youi-button.medium {
                font-size: 1rem;
                padding: 0.6rem 1.2rem;
            }
            .youi-button.large {
                font-size: 1.2rem;
                padding: 0.8rem 1.6rem;
            }
            .youi-button .icon {
                margin-right: 0.5rem;
            }
            .youi-button .loading {
                position: absolute;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%);
                width: 20px;
                height: 20px;
                border: 2px solid rgba(255,255,255,0.3);
                border-radius: 50%;
                border-top-color: #fff;
                animation: spin 1s ease-in-out infinite;
                display: none;
            }
            @keyframes spin {
                to { transform: translate(-50%, -50%) rotate(360deg); }
            }
            .ripple {
                position: absolute;
                border-radius: 50%;
                transform: scale(0);
                animation: ripple 0.6s linear;
                background-color: rgba(255, 255, 255, 0.7);
            }
            @keyframes ripple {
                to {
                    transform: scale(4);
                    opacity: 0;
                }
            }
        `;

        const button = document.createElement('button');
        button.className = `youi-button ${this.variant} ${this.size}`;
        button.setAttribute('role', 'button');
        button.innerHTML = `
            <span class="icon"></span>
            <span class="content"><slot></slot></span>
            <span class="loading"></span>
        `;

        this.shadowRoot.appendChild(style);
        this.shadowRoot.appendChild(button);

        this.updateVariant();
        this.updateSize();
        this.updateDisabledState();
        this.updateLoadingState();
    }

    private setupEventListeners() {
        this.addEventListener('click', this.handleClick.bind(this));
        this.addEventListener('keydown', this.handleKeydown.bind(this));
        this.addEventListener('mousedown', this.createRipple.bind(this));
    }

    private updateVariant() {
        const button = this.shadowRoot.querySelector('.youi-button');
        if (button) {
            button.className = `youi-button ${this.variant} ${this.size}`;
        }
    }

    private updateSize() {
        const button = this.shadowRoot.querySelector('.youi-button');
        if (button) {
            button.classList.remove('small', 'medium', 'large');
            button.classList.add(this.size);
        }
    }

    private updateDisabledState() {
        const button = this.shadowRoot.querySelector('.youi-button');
        if (button) {
            if (this.hasAttribute('disabled')) {
                button.setAttribute('disabled', '');
            } else {
                button.removeAttribute('disabled');
            }
        }
    }

    private updateLoadingState() {
        const button = this.shadowRoot.querySelector('.youi-button');
        const loadingSpinner = this.shadowRoot.querySelector('.loading');
        const content = this.shadowRoot.querySelector('.content');
        if (button && loadingSpinner && content) {
            if (this.loading) {
                button.setAttribute('disabled', '');
                loadingSpinner.setAttribute('style', 'display: block;');
                content.setAttribute('style', 'visibility: hidden;');
            } else {
                button.removeAttribute('disabled');
                loadingSpinner.setAttribute('style', 'display: none;');
                content.setAttribute('style', 'visibility: visible;');
            }
        }
    }

    private handleClick(event: MouseEvent) {
        if (this.hasAttribute('disabled') || this.loading) {
            event.preventDefault();
            event.stopPropagation();
            return;
        }

        const button = this.shadowRoot!.querySelector('.youi-button') as HTMLElement;
        gsap.to(button, {
            scale: 0.95,
            duration: 0.1,
            yoyo: true,
            repeat: 1,
            ease: "power1.inOut"
        });

        this.dispatchEvent(new CustomEvent('youi-button-click', { bubbles: true, composed: true }));
    }

    public playExitAnimation(): Promise<void> {
        return new Promise((resolve) => {
            const button = this.shadowRoot!.querySelector('.youi-button') as HTMLElement;
            gsap.to(button, {
                scale: 0.5,
                opacity: 0,
                duration: 0.3,
                ease: "back.in(1.7)",
                onComplete: resolve
            });
        });
    }

    private handleKeydown(event: KeyboardEvent) {
        if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            this.click();
        }
    }

    private createRipple(event: MouseEvent) {
        const button = this.shadowRoot.querySelector('.youi-button');
        if (button) {
            const circle = document.createElement('span');
            const diameter = Math.max(button.clientWidth, button.clientHeight);
            const radius = diameter / 2;

            circle.style.width = circle.style.height = `${diameter}px`;
            circle.style.left = `${event.clientX - (button.offsetLeft + radius)}px`;
            circle.style.top = `${event.clientY - (button.offsetTop + radius)}px`;
            circle.classList.add('ripple');

            const ripple = button.getElementsByClassName('ripple')[0];
            if (ripple) {
                ripple.remove();
            }

            button.appendChild(circle);
        }
    }

    set icon(value: string) {
        const iconElement = this.shadowRoot.querySelector('.icon');
        if (iconElement) {
            iconElement.textContent = value;
        }
    }
}

customElements.define("youi-button", YouiButton);
