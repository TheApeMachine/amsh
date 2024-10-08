import gsap from "gsap"
import { Draggable } from "gsap/Draggable"

gsap.registerPlugin(Draggable) 

class NodegraphNode extends HTMLElement {
    private template = document.createElement("template")

    constructor() {
        super();
        this.template.innerHTML = `
            <style>
                :host {
                    display: block;
                    position: absolute;
                }

                .node {
                    position: absolute;
                    width: 160px;
                    height: 60px;
                    border-radius: 5px;
                    background: #1e1e1e;
                    border: 1px solid #434343;
                    color: white;
                    display: flex;
                    flex-direction: column;
                    justify-content: center;
                    align-items: center;
                    cursor: default;
                    z-index: 1;
                }

                .node__in, .node__out {
                    position: absolute;
                    width: 16px;
                    height: 100%;
                    top: 0;
                    display: flex;
                    flex-direction: column;
                    justify-content: center;
                    align-items: center;
                }

                .node__in {
                    left: -8px;
                }

                .node__out {
                    right: -8px;
                }

                .node__socket {
                    width: 16px;
                    height: 16px;
                    margin: 1px 0;
                    border-radius: 100%;
                    background: #1e1e1e;
                    border: 1px solid #434343;
                }

                .node__in > .node__socket--on {
                    background: #f77edd;
                }

                .node__out > .node__socket--on {
                    background: #bef77e;
                }

                .node__content {
                    padding: 5px 20px;
                    position: relative;
                    flex: 1 0 0;
                    display: flex;
                    width: 100%;
                    min-height: 60px;
                }

                .node__content:before {
                    position: absolute;
                    top: auto;
                    left: 0;
                    bottom: 0;
                    content: "";
                    display: block;
                    height: 30px;
                    width: 100%;
                    background-image: linear-gradient(to top, #1e1e1e, rgba(30, 30, 30, 0));
                    z-index: 2;
                }

                .node__holder {
                    display: block;
                    display: block;
                    overflow: hidden;
                    position: relative;
                    width: 100%;
                }

                .node__answer {
                    position: relative;
                    padding: 10px 20px;
                    width: 100%;
                    border-top: 1px solid #434343;
                }

                .node.ui-selected {
                    border-color: yellow;
                }

                .node.ui-selecting {
                    border-color: red;
                }
            </style>
            <div class="node" data-event="drag">
                <slot></slot>
            </div>
        `
        this.attachShadow({ mode: 'open' });
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true))
    }
}

customElements.define('nodegraph-node', NodegraphNode);