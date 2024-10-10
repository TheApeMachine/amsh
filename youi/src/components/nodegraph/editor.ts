import cytoscape from 'cytoscape';
import edgehandles from 'cytoscape-edgehandles';

cytoscape.use(edgehandles);

class NodegraphEditor extends HTMLElement {
    private cy: cytoscape.Core | null = null;
    private template: HTMLTemplateElement;


    constructor() {
        super();
        this.template = document.createElement('template');
        this.template.innerHTML = `
            <style>
                :host { 
                    display: block; 
                    width: 100%; 
                    height: 100%;
                    background-color: #1e1e1e;
                    border-radius: 8px;
                    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
                }
                #cy { 
                    width: 100%; 
                    height: calc(100% - 40px); 
                    border-radius: 8px;
                }
                .controls {
                    height: 40px;
                    display: flex;
                    justify-content: flex-start;
                    align-items: center;
                    padding: 0 10px;
                }
                button {
                    margin-right: 10px;
                    padding: 5px 10px;
                    background-color: #3498db;
                    color: white;
                    border: none;
                    border-radius: 4px;
                    cursor: pointer;
                }
                .port {
                    width: 12px;
                    height: 12px;
                    background-color: #f39c12;
                    border-radius: 50%;
                    position: absolute;
                }
                .popper-handle {
                    width: 20px;
                    height: 20px;
                    background: red;
                    border-radius: 20px;
                    z-index: 9999;
                }
            </style>
            <div class="controls">
                <button id="addCompound">Add Compound</button>
                <button id="addController">Add Controller</button>
                <button id="addWorker">Add Worker</button>
            </div>
            <div id="cy"></div>
        `;
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
        this.initializeCytoscape();
        this.setupEventListeners();
    }

    private initializeCytoscape() {
        this.cy = cytoscape({
            container: document.getElementById('cy'),

            boxSelectionEnabled: false,
            autounselectify: true,

            style: cytoscape.stylesheet()
                .selector('node')
                .css({
                    'height': 80,
                    'width': 80,
                    'background-fit': 'cover',
                    'border-color': '#000',
                    'border-width': 3,
                    'border-opacity': 0.5
                })
                .selector('.eating')
                .css({
                    'border-color': 'red'
                })
                .selector('.eater')
                .css({
                    'border-width': 9
                })
                .selector('edge')
                .css({
                    'curve-style': 'bezier',
                    'width': 6,
                    'target-arrow-shape': 'triangle',
                    'line-color': '#ffaaaa',
                    'target-arrow-color': '#ffaaaa'
                })
                .selector('#bird')
                .css({
                    'background-image': 'https://live.staticflickr.com/7272/7633179468_3e19e45a0c_b.jpg'
                })
                .selector('#cat')
                .css({
                    'background-image': 'https://live.staticflickr.com/1261/1413379559_412a540d29_b.jpg'
                })
                .selector('#ladybug')
                .css({
                    'background-image': 'https://live.staticflickr.com/3063/2751740612_af11fb090b_b.jpg'
                })
                .selector('#aphid')
                .css({
                    'background-image': 'https://live.staticflickr.com/8316/8003798443_32d01257c8_b.jpg'
                })
                .selector('#rose')
                .css({
                    'background-image': 'https://live.staticflickr.com/5109/5817854163_eaccd688f5_b.jpg'
                })
                .selector('#grasshopper')
                .css({
                    'background-image': 'https://live.staticflickr.com/6098/6224655456_f4c3c98589_b.jpg'
                })
                .selector('#plant')
                .css({
                    'background-image': 'https://live.staticflickr.com/3866/14420309584_78bf471658_b.jpg'
                })
                .selector('#wheat')
                .css({
                    'background-image': 'https://live.staticflickr.com/2660/3715569167_7e978e8319_b.jpg'
                }),

            elements: {
                nodes: [
                    { data: { id: 'cat' } },
                    { data: { id: 'bird' } },
                    { data: { id: 'ladybug' } },
                    { data: { id: 'aphid' } },
                    { data: { id: 'rose' } },
                    { data: { id: 'grasshopper' } },
                    { data: { id: 'plant' } },
                    { data: { id: 'wheat' } }
                ],
                edges: [
                    { data: { source: 'cat', target: 'bird' } },
                    { data: { source: 'bird', target: 'ladybug' } },
                    { data: { source: 'bird', target: 'grasshopper' } },
                    { data: { source: 'grasshopper', target: 'plant' } },
                    { data: { source: 'grasshopper', target: 'wheat' } },
                    { data: { source: 'ladybug', target: 'aphid' } },
                    { data: { source: 'aphid', target: 'rose' } }
                ]
            },

            layout: {
                name: 'breadthfirst',
                directed: true,
                padding: 10
            }
        }); // cy init

        this.cy.on('tap', 'node', function () {
            var nodes = this;
            var food = [];

            nodes.addClass('eater');

            for (; ;) {
                var connectedEdges = nodes.connectedEdges(function (el) {
                    return !el.target().anySame(nodes);
                });

                var connectedNodes = connectedEdges.targets();

                Array.prototype.push.apply(food, connectedNodes);

                nodes = connectedNodes;

                if (nodes.empty()) { break; }
            }

            var delay = 0;
            var duration = 500;
            for (var i = food.length - 1; i >= 0; i--) {
                (function () {
                    var thisFood = food[i];
                    var eater = thisFood.connectedEdges(function (el) {
                        return el.target().same(thisFood);
                    }).source();

                    thisFood.delay(delay, function () {
                        eater.addClass('eating');
                    }).animate({
                        position: eater.position(),
                        css: {
                            'width': 10,
                            'height': 10,
                            'border-width': 0,
                            'opacity': 0
                        }
                    }, {
                        duration: duration,
                        complete: function () {
                            thisFood.remove();
                        }
                    });

                    delay += duration;
                })();
            }
        });
    }

    private setupEventListeners() {
        if (!this.cy) return;
        
        this.shadowRoot?.querySelector('#addCompound')?.addEventListener('click', () => this.addNode('compound'));
        this.shadowRoot?.querySelector('#addController')?.addEventListener('click', () => this.addNode('controller'));
        this.shadowRoot?.querySelector('#addWorker')?.addEventListener('click', () => this.addNode('worker'));

        this.cy.on('tap', 'node', (event) => {
            const node = event.target;
            this.centerOnNode(node);
            this.dispatchEvent(new CustomEvent('node-selected', { detail: { id: node.id(), type: node.classes()[0] } }));
        });
    }

    private addNode(type: string) {
        if (!this.cy) return;

        const id = `${type}${this.cy.nodes().length + 1}`;
        const node = this.cy.add({ data: { id }, classes: type });

        if (type === 'compound') {
            this.addPortsToCompound(node);
        } else if (this.cy.nodes('.compound').length > 0) {
            const compound = this.cy.nodes('.compound').last();
            node.move({ parent: compound.id() });
        }

        this.cy.layout({ name: 'cose', animate: true }).run();
    }

    private addPortsToCompound(compound) {
        if (!this.cy) return;

        const inputPort = this.cy.add({
            data: { id: `${compound.id()}-input`, parent: compound.id(), type: 'port' },
            position: { x: compound.position('x') - 10, y: compound.position('y') }
        });

        const outputPort = this.cy.add({
            data: { id: `${compound.id()}-output`, parent: compound.id(), type: 'port' },
            position: { x: compound.position('x') + 10, y: compound.position('y') }
        });

        this.addPortVisual(inputPort, 'left');
        this.addPortVisual(outputPort, 'right');
    }

    private addPortVisual(portNode, position) {
        if (!this.cy) return;

        const popper = portNode.popper({
            content: () => {
                const div = document.createElement('div');
                div.classList.add('port');
                div.style.position = position === 'left' ? 'absolute' : 'absolute';
                div.style.left = position === 'left' ? '-6px' : 'auto';
                div.style.right = position === 'right' ? '-6px' : 'auto';
                div.style.top = '50%';
                div.style.transform = 'translateY(-50%)';
                document.body.appendChild(div);
                return div;
            },
            popper: {}
        });

        const update = () => {
            popper.update();
        };

        portNode.on('position', update);
        this.cy.on('pan zoom resize', update);
    }

    private centerOnNode(node) {
        if (!this.cy) return;
        
        const neighborhood = node.neighborhood().add(node);
        this.cy.animate({
            fit: {
                eles: neighborhood,
                padding: 50
            },
            duration: 500,
            easing: 'ease-out'
        });
    }

    public updateGraph(networkData: { nodes: any[], edges: any[] }) {
        if (!this.cy) return;
        
        this.cy.elements().remove();
        this.cy.add(networkData.nodes);
        this.cy.add(networkData.edges);
        this.cy.nodes('.compound').forEach(compound => this.addPortsToCompound(compound));
        this.cy.layout({ name: 'cose', animate: true }).run();
    }
}

customElements.define('nodegraph-editor', NodegraphEditor);