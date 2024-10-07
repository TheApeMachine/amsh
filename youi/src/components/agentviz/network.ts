import cytoscape from 'cytoscape';

class NetworkGraphView extends HTMLElement {
    private cy: any; // Cytoscape instance

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.render();
        this.initializeCytoscape();
    }

    private render() {
        if (!this.shadowRoot) return;
        this.shadowRoot.innerHTML = `
            <style>
                :host { 
                    display: block; 
                    width: 100%; 
                    height: 600px;
                    background-color: white;
                    border-radius: 8px;
                    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
                }
                #cy { 
                    width: 100%; 
                    height: 100%; 
                    border-radius: 8px;
                }
            </style>
            <div id="cy"></div>
        `;
    }

    private initializeCytoscape() {
        const container = this.shadowRoot?.querySelector('#cy');
        if (!container) return;

        this.cy = cytoscape({
            container: container as HTMLElement,
            elements: [],
            style: [
                {
                    selector: 'node',
                    style: {
                        'background-color': '#3498db',
                        'label': 'data(id)',
                        'color': '#fff',
                        'text-outline-color': '#2980b9',
                        'text-outline-width': 2,
                        'font-size': 12
                    }
                },
                {
                    selector: 'edge',
                    style: {
                        'width': 3,
                        'line-color': '#bdc3c7',
                        'target-arrow-color': '#bdc3c7',
                        'target-arrow-shape': 'triangle',
                        'curve-style': 'bezier'
                    }
                }
            ],
            layout: {
                name: 'cose',
                animate: false,
                randomize: true,
                componentSpacing: 100,
                nodeRepulsion: function(node) { return 400000; },
                nodeOverlap: 20,
                idealEdgeLength: function(edge) { return 100; },
                edgeElasticity: function(edge) { return 100; },
                nestingFactor: 5,
                gravity: 80,
                numIter: 1000,
                initialTemp: 200,
                coolingFactor: 0.95,
                minTemp: 1.0
            }
        });
        
        // Listen for data updates from parent
        window.addEventListener('data-updated', (event: CustomEvent) => {
            console.log("data-updated", event);
            this.updateGraph(event.detail.networkData);
        });
    }

    private updateGraph(networkData: { nodes: any[], edges: any[] }) {
        this.cy.elements().remove();
        this.cy.add(networkData.nodes);
        this.cy.add(networkData.edges);
        this.cy.layout({ name: 'cose', animate: false }).run();
    }
}

customElements.define('network-graph-view', NetworkGraphView);