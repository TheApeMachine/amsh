import { jsx } from '@/lib/template';
import cytoscape from 'cytoscape';

export const NetworkGraphView = () => {
    const container = document.querySelector('#cy');
    if (!container) return;

    const cy = cytoscape({
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
            nodeRepulsion: () => 400000,
            nodeOverlap: 20,
            idealEdgeLength: () => 100,
            edgeElasticity: () => 100,
            nestingFactor: 5,
            gravity: 80,
            numIter: 1000,
            initialTemp: 200,
            coolingFactor: 0.95,
            minTemp: 1.0
        }
    });

    const updateGraph = (networkData: { nodes: any[], edges: any[] }) => {
        cy.elements().remove();
        cy.add(networkData.nodes);
        cy.add(networkData.edges);
        cy.layout({ name: 'cose', animate: false }).run();
    }

    return (
        <div id="cy"></div>
    );
}