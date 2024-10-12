import Drawflow from 'drawflow'

class ProductDesigner extends HTMLElement {

    private template: HTMLTemplateElement
    private drawflow!: Drawflow

    constructor() {
        super()
        this.template = document.createElement('template')
        this.template.innerHTML = `
            <style>
                :host {
                    display: block;
                    width: 100%;
                    height: 100%;
                }
                #drawflow {
                    width: 100%;
                    height: 100%;
                }
            </style>
            <drawflow id="drawflow"></drawflow>
        `
        this.attachShadow({ mode: 'open' })
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true))
        var id = this.shadowRoot?.getElementById("drawflow")

        if (!id) {
            throw new Error("drawflow element not found")
        }

        this.drawflow = new Drawflow(id)
        this.drawflow.start()
    }

}

customElements.define('product-designer', ProductDesigner)