class MoodLight extends HTMLElement {
    private template = document.createElement('template');

    constructor() {
        super();

        this.template.innerHTML = `
            <style>
                :host {
                    display: flex;
                    justify-content: center;
                    align-items: center;
                }

                .light {
                    width: 2rem;
                    height: 1rem;
                    background: #FFF;
                }
            </style>
            <div></div>
        `;
    }

    static get observedAttributes() {
        return ['color'];
    }

    attributeChangedCallback(name: string, oldValue: string, newValue: string) {
        if (name === 'color') {
            this.shadowRoot!.querySelector('.light')!.style.backgroundColor = newValue;
        }
    }

    connectedCallback() {
        this.shadowRoot!.querySelector('.light')!.style.backgroundColor = this.getAttribute('color') || '#FFF';
    }
}

customElements.define("mood-light", MoodLight);
