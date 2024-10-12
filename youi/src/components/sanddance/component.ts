import * as vega from 'vega';
import * as SandDance from '@msrvida/sanddance';

SandDance.use(vega);

class SandDanceComponent extends HTMLElement {
    private data: any[];
    private shadow: ShadowRoot;
    private viewer: SandDance.Viewer;
    private insight: SandDance.Insight;
    constructor() {
        super();
        this.data = [
            { myX: 0, myY: 0, myZ: 0 },
            { myX: 1, myY: 1, myZ: 1 },
            { myX: 2, myY: 2, myZ: 2 },
        ];

        this.shadow = this.attachShadow({ mode: "open" });
        this.shadow.innerHTML = `
            <style>
                :host {
                    display: block;
                }
                #vis {
                    width: 100%;
                    height: 100%;
                }
            </style>
            <div id="vis"></div>
        `;

        this.viewer = new SandDance.Viewer(this.shadow.querySelector('#vis') as HTMLElement);
        this.insight = {
            columns: {
                x: 'myX',
                y: 'myY',
                z: 'myZ'
            },
            size: {
                height: 700,
                width: 700
            },
            chart: 'scatterplot'
        };
    }

    connectedCallback() {
        this.viewer.render({ insight: this.insight }, this.data);
    }
}

customElements.define("sanddance-component", SandDanceComponent);