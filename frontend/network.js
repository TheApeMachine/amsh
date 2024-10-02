class NetworkGraphVisualization extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.networkNodes = new vis.DataSet();
        this.networkEdges = new vis.DataSet();
        this.network = null;
        this.thoughtConnections = [];
        this.maxConnections = 3;
        this.decayFactor = 0.9;
    }

    connectedCallback() {
        this.render();
        this.initNetworkGraph();
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
            #networkGraph {
                width: 100%;
                height: 100%;
                border: 1px solid #ddd;
            }
            #controls {
                margin-top: 10px;
            }
        </style>
        <div id="graphContainer">
            <div id="networkGraph"></div>
            <div id="controls">
                <label>
                    Max Connections:
                    <input type="number" id="maxConnections" min="1" max="10" value="3">
                </label>
                <label>
                    Min Similarity:
                    <input type="range" id="minSimilarity" min="0" max="1" step="0.1" value="0.5">
                </label>
                <button id="applyFilters">Apply Filters</button>
            </div>
        </div>
        `;

        this.shadowRoot.getElementById('applyFilters').addEventListener('click', () => this.applyFilters());
    }

    initNetworkGraph() {
        const container = this.shadowRoot.getElementById('networkGraph');
        const data = {
            nodes: this.networkNodes,
            edges: this.networkEdges
        };
        const options = {
            nodes: {
                shape: 'dot',
                size: 16
            },
            edges: {
                width: 0.5,
                color: { inherit: 'both' },
                smooth: {
                    type: 'continuous'
                }
            },
            physics: {
                forceAtlas2Based: {
                    gravitationalConstant: -26,
                    centralGravity: 0.005,
                    springLength: 230,
                    springConstant: 0.18
                },
                maxVelocity: 146,
                solver: 'forceAtlas2Based',
                timestep: 0.35,
                stabilization: { iterations: 150 }
            },
            groups: {
                reasoner: { color: { background: '#FF6384', border: '#FF6384' } },
                verifier: { color: { background: '#36A2EB', border: '#36A2EB' } },
                learning: { color: { background: '#FFCE56', border: '#FFCE56' } },
                metacognition: { color: { background: '#4BC0C0', border: '#4BC0C0' } },
                context_manager: { color: { background: '#9966FF', border: '#9966FF' } }
            }
        };
        this.network = new vis.Network(container, data, options);
    }

    addThought(thoughtData) {
        const { id, agentName, content, iteration, step, confidence } = thoughtData;
        
        this.thoughtConnections.push({ id, agentName, content, iteration, step, confidence });

        const relatedThoughts = this.findRelatedThoughts(content, confidence, iteration, id);
        
        if (relatedThoughts.length > 0) {
            this.networkNodes.add({
                id: id,
                label: `${agentName}\n${content.substring(0, 20)}...`,
                group: agentName,
                title: content,
                value: confidence
            });

            relatedThoughts.forEach(relatedThought => {
                this.networkEdges.add({
                    from: relatedThought.id,
                    to: id,
                    arrows: 'to',
                    dashes: true,
                    width: relatedThought.similarity * 3
                });

                const relatedThoughtData = this.thoughtConnections.find(t => t.id === relatedThought.id);
                if (relatedThoughtData && !this.networkNodes.get(relatedThought.id)) {
                    this.networkNodes.add({
                        id: relatedThought.id,
                        label: `${relatedThoughtData.agentName}\n${relatedThoughtData.content.substring(0, 20)}...`,
                        group: relatedThoughtData.agentName,
                        title: relatedThoughtData.content,
                        value: relatedThoughtData.confidence
                    });
                }
            });
        }
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
        // In a real implementation, you might use techniques like TF-IDF or word embeddings
        const words1 = new Set(text1.toLowerCase().split(/\W+/));
        const words2 = new Set(text2.toLowerCase().split(/\W+/));
        const intersection = new Set([...words1].filter(x => words2.has(x)));
        return intersection.size / Math.sqrt(words1.size * words2.size);
    }

    applyFilters() {
        this.maxConnections = parseInt(this.shadowRoot.getElementById('maxConnections').value);
        this.updateGraph();
    }

    updateGraph() {
        // Recalculate all edges based on current settings
        this.networkEdges.clear();
        this.thoughtConnections.forEach(thought => {
            const relatedThoughts = this.findRelatedThoughts(thought.content, thought.confidence, thought.iteration);
            relatedThoughts.forEach(relatedThought => {
                this.networkEdges.add({
                    from: relatedThought.id,
                    to: thought.id,
                    arrows: 'to',
                    dashes: true,
                    width: relatedThought.similarity * 3
                });
            });
        });
    }

    clear() {
        this.networkNodes.clear();
        this.networkEdges.clear();
        this.thoughtConnections = [];
    }
}

customElements.define('network-graph-visualization', NetworkGraphVisualization);