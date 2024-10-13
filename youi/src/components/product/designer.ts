import "https://cdn.jsdelivr.net/gh/jerosoler/Drawflow/dist/drawflow.min.js"

class ProductDesigner extends HTMLElement {

    private template: HTMLTemplateElement

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
                    display: block;
                    position: relative;
                    width: 100%;
                    height: 800px;
                }
            </style>
            <div id="drawflow"></div>
            <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/jerosoler/Drawflow@0.0.48/dist/drawflow.min.css">
        `
        this.attachShadow({ mode: 'open' })
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true))
        const container = this.shadowRoot?.getElementById('drawflow');
        const editor = new Drawflow(container);

        if (!container) {
            throw new Error("drawflow element not found")
        }

        editor.reroute = true;
        editor.reroute_fix_curvature = true;

        editor.start();

        const data = {
            name: ''
        };

        editor.addNode('foo', 1, 1, 100, 200, 'foo', data, 'Foo');
        editor.addNode('bar', 1, 1, 400, 100, 'bar', data, 'Bar A');
        editor.addNode('bar', 1, 1, 400, 300, 'bar', data, 'Bar B');

        editor.addConnection(1, 2, "output_1", "input_1");
        editor.addConnection(1, 3, "output_1", "input_1");
    }

}

customElements.define('product-designer', ProductDesigner)