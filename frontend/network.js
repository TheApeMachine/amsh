class NetworkGraphVisualization extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.cy = null;
        this.thoughtConnections = [];
        this.maxConnections = 3;
        this.decayFactor = 0.9;
        this.minConnectionsForCluster = 3;
        this.clusterColors = ["#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF"];
        this.updateBatch = [];
        this.updateTimeout = null;
        this.layouts = {
            'cose': { name: 'cose', animate: 'end', animationDuration: 500, randomize: false, componentSpacing: 100 },
            'concentric': { name: 'concentric', animate: 'end', animationDuration: 500 },
            'breadthfirst': { name: 'breadthfirst', animate: 'end', animationDuration: 500 },
            'circle': { name: 'circle', animate: 'end', animationDuration: 500 },
            'grid': { name: 'grid', animate: 'end', animationDuration: 500 }
        };
        this.currentLayout = 'cose';
    }

    connectedCallback() {
        this.render();
        this.initCytoscapeGraph();
    }

    render() {
        this.shadowRoot.innerHTML = `
        <style>
            :host {
                display: block;
                width: 100%;
                height: 100%;
            }
            #graphContainer {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                width: 100%;
                height: 100%;
            }
            #cytoscapeGraph {
                width: 100%;
                height: 100%;
                border: 1px solid #ddd;
            }
            #controls {
                margin-top: 10px;
            }
        </style>
        <div id="graphContainer">
            <div id="cytoscapeGraph"></div>
            <div id="controls">
                <label>
                    Max Connections:
                    <input type="number" id="maxConnections" min="1" max="10" value="3">
                </label>
                <label>
                    Min Similarity:
                    <input type="range" id="minSimilarity" min="0" max="1" step="0.1" value="0.5">
                </label>
                <label>
                    Min Connections for Cluster:
                    <input type="number" id="minConnectionsForCluster" min="2" max="10" value="3">
                </label>
                <label>
                    Layout:
                    <select id="layoutSelect">
                        <option value="cose">CoSE</option>
                        <option value="concentric">Concentric</option>
                        <option value="breadthfirst">Breadth-first</option>
                        <option value="circle">Circle</option>
                        <option value="grid">Grid</option>
                    </select>
                </label>
                <button id="applyFilters">Apply Filters</button>
                <button id="clusterGraph">Cluster Graph</button>
            </div>
        </div>
        `;

        this.shadowRoot.getElementById('applyFilters').addEventListener('click', () => this.applyFilters());
        this.shadowRoot.getElementById('clusterGraph').addEventListener('click', () => this.clusterGraph());
        this.shadowRoot.getElementById('layoutSelect').addEventListener('change', (e) => this.changeLayout(e.target.value));
    }

    initCytoscapeGraph() {
        const container = this.shadowRoot.getElementById('cytoscapeGraph');
        this.cy = cytoscape({
            container: container,
            style: [
                {
                    selector: 'node',
                    style: {
                        'background-color': '#666',
                        'label': 'data(label)',
                        'text-valign': 'center',
                        'text-halign': 'center',
                        'text-wrap': 'wrap',
                        'text-max-width': '100px'
                    }
                },
                {
                    selector: 'edge',
                    style: {
                        'width': 3,
                        'line-color': '#ccc',
                        'target-arrow-color': '#ccc',
                        'target-arrow-shape': 'triangle',
                        'curve-style': 'bezier'
                    }
                },
                {
                    selector: ':parent',
                    style: {
                        'background-opacity': 0.333,
                        'label': 'data(label)',
                        'text-valign': 'top',
                        'text-halign': 'center'
                    }
                }
            ],
            layout: this.layouts[this.currentLayout]
        });

        this.cy.on('tap', 'node', (evt) => {
            const node = evt.target;
            if (node.isParent()) {
                this.expandCluster(node);
            }
        });
    }

    addThought(thoughtData) {
        const { id, agentName, content, iteration, step, confidence } = thoughtData;
        
        this.thoughtConnections.push({ id, agentName, content, iteration, step, confidence });

        const relatedThoughts = this.findRelatedThoughts(content, confidence, iteration, id);
        
        if (relatedThoughts.length > 0) {
            this.updateBatch.push({
                group: 'nodes',
                data: { 
                    id: id, 
                    label: `${agentName}\n${content.substring(0, 20)}...`,
                    fullContent: content,
                    agentName: agentName,
                    confidence: confidence
                }
            });

            relatedThoughts.forEach(relatedThought => {
                const relatedThoughtData = this.thoughtConnections.find(t => t.id === relatedThought.id);
                if (relatedThoughtData) {
                    this.updateBatch.push({
                        group: 'nodes',
                        data: {
                            id: relatedThought.id,
                            label: `${relatedThoughtData.agentName}\n${relatedThoughtData.content.substring(0, 20)}...`,
                            fullContent: relatedThoughtData.content,
                            agentName: relatedThoughtData.agentName,
                            confidence: relatedThoughtData.confidence
                        }
                    });

                    this.updateBatch.push({
                        group: 'edges',
                        data: {
                            id: `${relatedThought.id}-${id}`,
                            source: relatedThought.id,
                            target: id,
                            weight: relatedThought.similarity
                        }
                    });
                }
            });
        }

        this.scheduleUpdate();
    }

    scheduleUpdate() {
        if (this.updateTimeout) {
            clearTimeout(this.updateTimeout);
        }

        this.updateTimeout = setTimeout(() => {
            this.applyBatchUpdate();
        }, 100);  // Adjust this delay as needed
    }

    applyBatchUpdate() {
        if (this.updateBatch.length > 0) {
            this.cy.startBatch();
            this.cy.add(this.updateBatch);
            this.cy.endBatch();
            this.updateBatch = [];
            this.runLayout();
        }
    }

    runLayout() {
        const layout = this.cy.layout(this.layouts[this.currentLayout]);
        layout.run();
    }

    findRelatedThoughts(content, confidence, currentIteration, currentId) {
        const minSimilarity = parseFloat(this.shadowRoot.getElementById('minSimilarity').value) || 0.1;
        return this.thoughtConnections
            .filter(thought => thought.id !== currentId)
            .map(thought => ({
                id: thought.id,
                similarity: this.calculateSimilarity(content, thought.content),
                iterationDiff: currentIteration - thought.iteration
            }))
            .filter(thought => thought.similarity > minSimilarity)
            .sort((a, b) => b.similarity - a.similarity)
            .slice(0, this.maxConnections)
            .map(thought => ({
                ...thought,
                similarity: thought.similarity * Math.pow(this.decayFactor, thought.iterationDiff)
            }));
    }

    calculateSimilarity(text1, text2) {
        // This is a placeholder for a more sophisticated similarity calculation
        const words1 = new Set(text1.toLowerCase().split(/\W+/));
        const words2 = new Set(text2.toLowerCase().split(/\W+/));
        const intersection = new Set([...words1].filter(x => words2.has(x)));
        return intersection.size / Math.sqrt(words1.size * words2.size);
    }

    applyFilters() {
        this.maxConnections = parseInt(this.shadowRoot.getElementById('maxConnections').value);
        this.minConnectionsForCluster = parseInt(this.shadowRoot.getElementById('minConnectionsForCluster').value);
        this.updateGraph();
    }

    updateGraph() {
        this.cy.startBatch();
        this.cy.edges().remove();
        
        this.thoughtConnections.forEach(thought => {
            const relatedThoughts = this.findRelatedThoughts(thought.content, thought.confidence, thought.iteration, thought.id);
            relatedThoughts.forEach(relatedThought => {
                if (thought.id !== relatedThought.id) {
                    this.cy.add({
                        group: 'edges',
                        data: {
                            id: `${relatedThought.id}-${thought.id}`,
                            source: relatedThought.id,
                            target: thought.id,
                            weight: relatedThought.similarity
                        }
                    });
                }
            });
        });
        this.cy.endBatch();

        this.runLayout();
    }

    clear() {
        this.cy.elements().remove();
        this.thoughtConnections = [];
    }

    clusterGraph() {
        this.cy.startBatch();
        // Remove existing clusters
        this.cy.nodes().forEach(node => {
            if (node.isParent()) {
                this.expandCluster(node);
            }
        });

        // First level of clustering
        this.clusterByConnections();

        // Second level of clustering (clusters of clusters)
        this.clusterByConnections();
        this.cy.endBatch();

        this.runLayout();
    }

    clusterByConnections() {
        const nodesToCluster = this.cy.nodes().filter(node => 
            !node.isParent() && node.connectedEdges().length >= this.minConnectionsForCluster
        );

        nodesToCluster.forEach((node, index) => {
            const neighborhood = node.neighborhood().nodes().filter(n => !n.isParent());
            if (neighborhood.length >= this.minConnectionsForCluster) {
                const clusterId = 'cluster_' + this.cy.nodes().length;
                this.cy.add({
                    group: 'nodes',
                    data: {
                        id: clusterId,
                        label: 'Cluster ' + (index + 1)
                    }
                });
                neighborhood.move({ parent: clusterId });
                node.move({ parent: clusterId });
            }
        });
    }

    expandCluster(clusterNode) {
        this.cy.startBatch();
        const childNodes = clusterNode.children();
        childNodes.move({ parent: null });
        this.cy.remove(clusterNode);
        this.cy.endBatch();
        this.runLayout();
    }

    changeLayout(layoutName) {
        this.currentLayout = layoutName;
        this.runLayout();
    }
}

customElements.define('network-graph-visualization', NetworkGraphVisualization);