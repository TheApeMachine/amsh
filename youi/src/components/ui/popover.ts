class YouiPopover extends HTMLElement {
    template: HTMLTemplateElement;

    constructor() {
        super();
        this.template = document.createElement("template");
        this.template.innerHTML = `
            <style>
                .tooltip {
                    height: 200px;
                    width: 300px;
                    position: absolute;
                    top: 50%;
                    left: 50%;
                    transform: translate(-50%, -50%);
                    background-color: rgba(255, 255, 255, 0.8);
                    box-shadow: 0px 5px 15px 0px rgba(0, 0, 0, 0.3);
                    border-radius: 0.25rem;
                    border: 4px solid #000;
                }
                .tooltip__arrow {
                    width: 50px;
                    height: 25px;
                    position: absolute;
                    top: 100%;
                    left: 50%;
                    transform: translateX(-50%);
                    overflow: hidden;
                }
                .tooltip__arrow::after {
                    content: "";
                    position: absolute;
                    width: 20px;
                    height: 20px;
                    background-color: rgba(255, 255, 255, 0.8);
                    transform: translateX(-50%) translateY(-50%) rotate(45deg);
                    top: 0;
                    left: 50%;
                    box-shadow: 1px 1px 20px 0px rgba(0, 0, 0, 0.6);
                    border: 4px solid #000;
                }
            </style>
            <div class="tooltip">
                <div class="tooltip__arrow"></div>
            </div>
        `;
        this.attachShadow({ mode: "open" });
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
    }
}

customElements.define("youi-popover", YouiPopover);