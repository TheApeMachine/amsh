import "https://cdn.jsdelivr.net/gh/jerosoler/Drawflow/dist/drawflow.min.js"
import "@/components/ui/button.ts"
import "@/components/ui/input.ts"
import "@/components/ui/blocksuite.ts"
import gsap from "gsap"
import { div } from "@blocksuite/blocks/dist/surface-block/perfect-freehand/vec.js"

class ProductDesigner extends HTMLElement {

    private template: HTMLTemplateElement
    private zScaler: number = 0.5;

    constructor() {
        super()
        this.template = document.createElement('template')
        this.template.innerHTML = `
            <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/jerosoler/Drawflow@0.0.48/dist/drawflow.min.css">
            <style>
               ${this.loadStyles()}
                #drawflow {
                    display: block;
                    position: relative;
                    width: 100%;
                    height: 100%;
                }
                .panel {
                    display: block;
                    position: absolute;
                    top: 0;
                    left: 0;
                    width: 100%;
                    height: 100%;
                    background: rgba(255, 255, 255, 0.5);
                }
            </style>
            <div id="drawflow" class="panel"></div>
        `
        this.attachShadow({ mode: 'open' })
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true))
        const container = this.shadowRoot?.getElementById('drawflow');
        const editor = new Drawflow(container);
        editor.addModule("something")

        if (!container) {
            throw new Error("drawflow element not found")
        }

        if (container) {
            container.style.perspective = "1000px";
            container.style.perspectiveOrigin = "50% 30%"; // Adjust these values to fine-tune the effect
        }

        editor.curvature = 0;

        editor.start();

        const data = {
            name: ''
        };

        editor.addNode('foo', 1, 1, 100, 200, 'foo', data, `
            <youi-input></youi-input>
            <youi-button>Hello</youi-button>
        `);
        editor.addNode('bar', 1, 1, 400, 100, 'bar', data, '<youi-blocksuite></youi-blocksuite>');
        editor.addNode('bar', 1, 1, 400, 300, 'bar', data, 'Bar B');

        editor.addConnection(1, 2, "output_1", "input_1");
        editor.addConnection(1, 3, "output_1", "input_1");

        for (let i = 0; i < 10; i++) {
            const div = document.createElement("div")
            div.classList.add("panel")
            div.innerHTML = `
                <h1>Panel ${i + 1}</h1>
            `
            div.style.transform = `translate3d(0, 0, 0) scale(1) rotateX(-20deg)`
            div.style.transformOrigin = "center center"
            div.style.opacity = "1"
            this.shadowRoot?.appendChild(div)
        }

        this.animate()
    }

    animate() {
        const tl = gsap.timeline();
        const panels = this.shadowRoot?.querySelectorAll(".panel");
        const totalPanels = panels?.length || 0;
        const spacing = 50; // Adjust this value to increase/decrease spacing between panels
        const scaleRatio = 0.9; // Adjust this value to control how quickly panels shrink
        const rotationX = 0; // Adjust this value to change the "camera angle"
        const elevationStep = 50; // The amount each panel is raised relative to the one in front

        tl.to(panels, {
            z: (index) => -(index + 0.5) * spacing * this.zScaler,
            scale: (index) => Math.pow(scaleRatio, index),
            y: (index) => -(index * elevationStep) +600, // Each panel is raised by elevationStep pixels
            rotationX: rotationX,
            opacity: (index) => Math.max(0.8 - index * 0.15, 0.01),
            duration: 1,
            ease: "power2.inOut",
            stagger: 0.05
        });
        tl.play();
    }

    private loadStyles() {
        return `
        :host {
            --dfBackgroundColor: rgba(52, 55, 65, 1);
            --dfBackgroundSize: 20px;
            --dfBackgroundImage: radial-gradient(rgba(190, 190, 190, 1) 1px, transparent 1px);
        
            --dfNodeType: flex;
            --dfNodeTypeFloat: none;
            --dfNodeBackgroundColor: #ffffff;
            --dfNodeTextColor: #000000;
            --dfNodeBorderSize: 2px;
            --dfNodeBorderColor: #000000;
            --dfNodeBorderRadius: 4px;
            --dfNodeMinHeight: 40px;
            --dfNodeMinWidth: 160px;
            --dfNodePaddingTop: 15px;
            --dfNodePaddingBottom: 15px;
            --dfNodeBoxShadowHL: 0px;
            --dfNodeBoxShadowVL: 0px;
            --dfNodeBoxShadowBR: 0px;
            --dfNodeBoxShadowS: 0px;
            --dfNodeBoxShadowColor: #000000;
        
            --dfNodeHoverBackgroundColor: #ffffff;
            --dfNodeHoverTextColor: #000000;
            --dfNodeHoverBorderSize: 2px;
            --dfNodeHoverBorderColor: #000000;
            --dfNodeHoverBorderRadius: 4px;
        
            --dfNodeHoverBoxShadowHL: 0px;
            --dfNodeHoverBoxShadowVL: 2px;
            --dfNodeHoverBoxShadowBR: 15px;
            --dfNodeHoverBoxShadowS: 2px;
            --dfNodeHoverBoxShadowColor: #4ea9ff;
        
            --dfNodeSelectedBackgroundColor: rgba(110, 149, 247, 1);
            --dfNodeSelectedTextColor: #ffffff;
            --dfNodeSelectedBorderSize: 2px;
            --dfNodeSelectedBorderColor: #000000;
            --dfNodeSelectedBorderRadius: 4px;
        
            --dfNodeSelectedBoxShadowHL: 0px;
            --dfNodeSelectedBoxShadowVL: 0px;
            --dfNodeSelectedBoxShadowBR: 0px;
            --dfNodeSelectedBoxShadowS: 2px;
            --dfNodeSelectedBoxShadowColor: rgba(110, 82, 255, 1);
        
            --dfInputBackgroundColor: rgba(247, 185, 110, 1);
            --dfInputBorderSize: 2px;
            --dfInputBorderColor: #000000;
            --dfInputBorderRadius: 50px;
            --dfInputLeft: -27px;
            --dfInputHeight: 20px;
            --dfInputWidth: 20px;
        
            --dfInputHoverBackgroundColor: rgba(255, 0, 0, 1);
            --dfInputHoverBorderSize: 2px;
            --dfInputHoverBorderColor: #000000;
            --dfInputHoverBorderRadius: 50px;
        
            --dfOutputBackgroundColor: rgba(247, 117, 110, 1);
            --dfOutputBorderSize: 2px;
            --dfOutputBorderColor: rgba(0, 0, 0, 1);
            --dfOutputBorderRadius: 0px;
            --dfOutputRight: -5px;
            --dfOutputHeight: 10px;
            --dfOutputWidth: 20px;
        
            --dfOutputHoverBackgroundColor: #ffffff;
            --dfOutputHoverBorderSize: 2px;
            --dfOutputHoverBorderColor: #000000;
            --dfOutputHoverBorderRadius: 50px;
        
            --dfLineWidth: 6px;
            --dfLineColor: rgba(125, 125, 125, 1);
            --dfLineHoverColor: #4682b4;
            --dfLineSelectedColor: #43b993;
        
            --dfRerouteBorderWidth: 8px;
            --dfRerouteBorderColor: rgba(255, 0, 0, 1);
            --dfRerouteBackgroundColor: #ffffff;
        
            --dfRerouteHoverBorderWidth: 2px;
            --dfRerouteHoverBorderColor: #000000;
            --dfRerouteHoverBackgroundColor: #ffffff;
        
            --dfDeleteDisplay: block;
            --dfDeleteColor: #ffffff;
            --dfDeleteBackgroundColor: #000000;
            --dfDeleteBorderSize: 2px;
            --dfDeleteBorderColor: #ffffff;
            --dfDeleteBorderRadius: 50px;
            --dfDeleteTop: -15px;
        
            --dfDeleteHoverColor: #000000;
            --dfDeleteHoverBackgroundColor: #ffffff;
            --dfDeleteHoverBorderSize: 2px;
            --dfDeleteHoverBorderColor: #000000;
            --dfDeleteHoverBorderRadius: 50px;
        
            display: flex;
            perspective: 500px;
        }
        
        #drawflow {
            background: var(--dfBackgroundColor);
            background-size: var(--dfBackgroundSize) var(--dfBackgroundSize);
            background-image: var(--dfBackgroundImage);
        }
        
        .drawflow .drawflow-node {
            display: var(--dfNodeType);
            background: var(--dfNodeBackgroundColor);
            color: var(--dfNodeTextColor);
            border: var(--dfNodeBorderSize)  solid var(--dfNodeBorderColor);
            border-radius: var(--dfNodeBorderRadius);
            min-height: var(--dfNodeMinHeight);
            width: auto;
            min-width: var(--dfNodeMinWidth);
            padding-top: var(--dfNodePaddingTop);
            padding-bottom: var(--dfNodePaddingBottom);
            -webkit-box-shadow: var(--dfNodeBoxShadowHL) var(--dfNodeBoxShadowVL) var(--dfNodeBoxShadowBR) var(--dfNodeBoxShadowS) var(--dfNodeBoxShadowColor);
            box-shadow:  var(--dfNodeBoxShadowHL) var(--dfNodeBoxShadowVL) var(--dfNodeBoxShadowBR) var(--dfNodeBoxShadowS) var(--dfNodeBoxShadowColor);
        }
        
        .drawflow .drawflow-node:hover {
            background: var(--dfNodeHoverBackgroundColor);
            color: var(--dfNodeHoverTextColor);
            border: var(--dfNodeHoverBorderSize)  solid var(--dfNodeHoverBorderColor);
            border-radius: var(--dfNodeHoverBorderRadius);
            -webkit-box-shadow: var(--dfNodeHoverBoxShadowHL) var(--dfNodeHoverBoxShadowVL) var(--dfNodeHoverBoxShadowBR) var(--dfNodeHoverBoxShadowS) var(--dfNodeHoverBoxShadowColor);
            box-shadow:  var(--dfNodeHoverBoxShadowHL) var(--dfNodeHoverBoxShadowVL) var(--dfNodeHoverBoxShadowBR) var(--dfNodeHoverBoxShadowS) var(--dfNodeHoverBoxShadowColor);
        }
        
        .drawflow .drawflow-node.selected {
            background: var(--dfNodeSelectedBackgroundColor);
            color: var(--dfNodeSelectedTextColor);
            border: var(--dfNodeSelectedBorderSize)  solid var(--dfNodeSelectedBorderColor);
            border-radius: var(--dfNodeSelectedBorderRadius);
            -webkit-box-shadow: var(--dfNodeSelectedBoxShadowHL) var(--dfNodeSelectedBoxShadowVL) var(--dfNodeSelectedBoxShadowBR) var(--dfNodeSelectedBoxShadowS) var(--dfNodeSelectedBoxShadowColor);
            box-shadow:  var(--dfNodeSelectedBoxShadowHL) var(--dfNodeSelectedBoxShadowVL) var(--dfNodeSelectedBoxShadowBR) var(--dfNodeSelectedBoxShadowS) var(--dfNodeSelectedBoxShadowColor);
        }
        
        .drawflow .drawflow-node .input {
            left: var(--dfInputLeft);
            background: var(--dfInputBackgroundColor);
            border: var(--dfInputBorderSize)  solid var(--dfInputBorderColor);
            border-radius: var(--dfInputBorderRadius);
            height: var(--dfInputHeight);
            width: var(--dfInputWidth);
        }
        
        .drawflow .drawflow-node .input:hover {
            background: var(--dfInputHoverBackgroundColor);
            border: var(--dfInputHoverBorderSize)  solid var(--dfInputHoverBorderColor);
            border-radius: var(--dfInputHoverBorderRadius);
        }
        
        .drawflow .drawflow-node .outputs {
            float: var(--dfNodeTypeFloat);
        }
        
        .drawflow .drawflow-node .output {
            right: var(--dfOutputRight);
            background: var(--dfOutputBackgroundColor);
            border: var(--dfOutputBorderSize)  solid var(--dfOutputBorderColor);
            border-radius: var(--dfOutputBorderRadius);
            height: var(--dfOutputHeight);
            width: var(--dfOutputWidth);
        }
        
        .drawflow .drawflow-node .output:hover {
            background: var(--dfOutputHoverBackgroundColor);
            border: var(--dfOutputHoverBorderSize)  solid var(--dfOutputHoverBorderColor);
            border-radius: var(--dfOutputHoverBorderRadius);
        }
        
        .drawflow .connection .main-path {
            stroke-width: var(--dfLineWidth);
            stroke: var(--dfLineColor);
        }
        
        .drawflow .connection .main-path:hover {
            stroke: var(--dfLineHoverColor);
        }
        
        .drawflow .connection .main-path.selected {
            stroke: var(--dfLineSelectedColor);
        }
        
        .drawflow .connection .point {
            stroke: var(--dfRerouteBorderColor);
            stroke-width: var(--dfRerouteBorderWidth);
            fill: var(--dfRerouteBackgroundColor);
        }
        
        .drawflow .connection .point:hover {
            stroke: var(--dfRerouteHoverBorderColor);
            stroke-width: var(--dfRerouteHoverBorderWidth);
            fill: var(--dfRerouteHoverBackgroundColor);
        }
        
        .drawflow-delete {
            display: var(--dfDeleteDisplay);
            color: var(--dfDeleteColor);
            background: var(--dfDeleteBackgroundColor);
            border: var(--dfDeleteBorderSize) solid var(--dfDeleteBorderColor);
            border-radius: var(--dfDeleteBorderRadius);
        }
        
        .parent-node .drawflow-delete {
            top: var(--dfDeleteTop);
        }
        
        .drawflow-delete:hover {
            color: var(--dfDeleteHoverColor);
            background: var(--dfDeleteHoverBackgroundColor);
            border: var(--dfDeleteHoverBorderSize) solid var(--dfDeleteHoverBorderColor);
            border-radius: var(--dfDeleteHoverBorderRadius);
        }
          `
    }

}

customElements.define('product-designer', ProductDesigner)