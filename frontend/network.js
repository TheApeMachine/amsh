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
        this.searchTerm = '';
        this.highlightedNodes = new Set();
        this.debounceTimer = null;
    }

    connectedCallback() {
        this.render();
        this.loadDependencies().then(() => {
            this.initCytoscapeGraph();
            this.initTippyTooltips();
        });
    }

    async loadDependencies() {
        // Load Cytoscape.js
        if (!window.cytoscape) {
            await this.loadScript('https://unpkg.com/cytoscape@3.21.1/dist/cytoscape.min.js');
        }

        // Load Tippy.js and its CSS
        if (!window.tippy) {
            await this.loadScript('https://unpkg.com/@popperjs/core@2/dist/umd/popper.min.js');
            await this.loadScript('https://unpkg.com/tippy.js@6/dist/tippy-bundle.umd.min.js');
            const tippyCss = document.createElement('link');
            tippyCss.rel = 'stylesheet';
            tippyCss.href = 'https://unpkg.com/tippy.js@6/dist/tippy.css';
            this.shadowRoot.appendChild(tippyCss);
        }
    }

    loadScript(src) {
        return new Promise((resolve, reject) => {
            const script = document.createElement('script');
            script.src = src;
            script.onload = resolve;
            script.onerror = reject;
            this.shadowRoot.appendChild(script);
        });
    }

    render() {
        this.shadowRoot.innerHTML = `
        <style>
            :host {
                display: block;
                width: 100%;
                height: 100%;
                font-family: Arial, sans-serif;
                box-sizing: border-box;
            }
            #graphContainer {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                width: 100%;
                height: 100%;
                padding: 10px;
                box-sizing: border-box;
            }
            #cytoscapeGraph {
                width: 100%;
                height: calc(100% - 180px);
                border: 1px solid #ddd;
                border-radius: 5px;
                overflow: hidden;
                background-color: #f9f9f9;
            }
            #controls {
                margin-top: 10px;
                display: flex;
                flex-wrap: wrap;
                gap: 10px;
                align-items: center;
                width: 100%;
                box-sizing: border-box;
            }
            label {
                display: flex;
                align-items: center;
                font-size: 0.9em;
            }
            input, select {
                margin-left: 5px;
                padding: 5px;
                border: 1px solid #ccc;
                border-radius: 3px;
                font-size: 0.9em;
            }
            button {
                padding: 5px 10px;
                background-color: #4CAF50;
                color: white;
                border: none;
                border-radius: 3px;
                cursor: pointer;
                transition: background-color 0.3s;
                font-size: 0.9em;
            }
            button:hover {
                background-color: #45a049;
            }
            #searchContainer {
                display: flex;
                align-items: center;
                margin-top: 10px;
                width: 100%;
                box-sizing: border-box;
            }
            #searchInput {
                flex: 1;
                padding: 5px;
                border: 1px solid #ccc;
                border-radius: 3px;
                font-size: 0.9em;
            }
            #searchButton {
                margin-left: 5px;
                padding: 5px 10px;
                background-color: #2196F3;
                color: white;
                border: none;
                border-radius: 3px;
                cursor: pointer;
                transition: background-color 0.3s;
                font-size: 0.9em;
            }
            #searchButton:hover {
                background-color: #0b7dda;
            }
            #statsContainer {
                margin-top: 10px;
                font-size: 0.9em;
                width: 100%;
                text-align: center;
            }
            @media (max-width: 600px) {
                #controls {
                    flex-direction: column;
                    align-items: stretch;
                }
                label {
                    width: 100%;
                    justify-content: space-between;
                }
                #searchContainer {
                    flex-direction: column;
                }
                #searchButton {
                    margin-left: 0;
                    margin-top: 5px;
                }
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
                    <span id="minSimilarityValue">0.5</span>
                </label>
                <label>
                    Min Connections for Cluster:
                    <input type="number" id="minConnectionsForCluster" min="2" max="10" value="3">
                </label>
                <label>
                    Agent Name:
                    <input type="text" id="agentNameFilter" placeholder="e.g., Agent A">
                </label>
                <label>
                    Min Confidence:
                    <input type="range" id="minConfidence" min="0" max="1" step="0.1" value="0.5">
                    <span id="minConfidenceValue">0.5</span>
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
                <button id="exportGraph">Export Graph</button>
            </div>
            <div id="searchContainer">
                <input type="text" id="searchInput" placeholder="Search nodes...">
                <button id="searchButton">Search</button>
            </div>
            <div id="statsContainer"></div>
        </div>
        `;

        // Event Listeners for Range Inputs to display current value
        this.shadowRoot.getElementById('minSimilarity').addEventListener('input', (e) => {
            this.shadowRoot.getElementById('minSimilarityValue').textContent = e.target.value;
        });

        this.shadowRoot.getElementById('minConfidence').addEventListener('input', (e) => {
            this.shadowRoot.getElementById('minConfidenceValue').textContent = e.target.value;
        });

        // Attach event listeners
        this.shadowRoot.getElementById('applyFilters').addEventListener('click', () => this.applyFilters());
        this.shadowRoot.getElementById('clusterGraph').addEventListener('click', () => this.clusterGraph());
        this.shadowRoot.getElementById('layoutSelect').addEventListener('change', (e) => this.changeLayout(e.target.value));
        this.shadowRoot.getElementById('exportGraph').addEventListener('click', () => this.exportGraph());
        this.shadowRoot.getElementById('searchButton').addEventListener('click', () => this.searchNodes());
        this.shadowRoot.getElementById('searchInput').addEventListener('input', (e) => {
            this.searchTerm = e.target.value;
            this.debounce(this.searchNodes.bind(this), 300);
        });
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
                        'text-max-width': '100px',
                        'font-size': '10px'
                    }
                },
                {
                    selector: 'edge',
                    style: {
                        'width': 'data(weight)',
                        'line-color': '#ccc',
                        'target-arrow-color': '#ccc',
                        'target-arrow-shape': 'triangle',
                        'curve-style': 'bezier',
                        'opacity': 0.7
                    }
                },
                {
                    selector: ':parent',
                    style: {
                        'background-opacity': 0.333,
                        'label': 'data(label)',
                        'text-valign': 'top',
                        'text-halign': 'center',
                        'font-size': '12px'
                    }
                },
                {
                    selector: 'node.highlighted',
                    style: {
                        'background-color': '#ff0',
                        'border-color': '#f00',
                        'border-width': '2px'
                    }
                },
                {
                    selector: 'edge.highlighted',
                    style: {
                        'line-color': '#f00',
                        'target-arrow-color': '#f00',
                        'width': 4
                    }
                },
                {
                    selector: '.filtered',
                    style: {
                        'display': 'none'
                    }
                }
            ],
            layout: this.layouts[this.currentLayout]
        });

        this.cy.on('tap', 'node', (evt) => {
            const node = evt.target;
            if (node.isParent()) {
                this.expandCluster(node);
            } else {
                // Using modal instead of alert
                this.showNodeDetailsModal(node);
            }
        });

        this.cy.on('mouseover', 'node', (event) => {
            const node = event.target;
            this.highlightNeighbors(node);
        });

        this.cy.on('mouseout', 'node', () => {
            this.unhighlightAll();
        });
    }

    initTippyTooltips() {
        const self = this;
        this.cy.nodes().forEach(node => {
            const ref = node.popperRef(); // used only for positioning
            const dummyDomEle = document.createElement('div');

            node.tippy = tippy(dummyDomEle, {
                getReferenceClientRect: ref.getBoundingClientRect,
                content: () => {
                    const content = document.createElement('div');
                    content.innerHTML = `<strong>Agent:</strong> ${node.data('agentName')}<br>
                                         <strong>Content:</strong> ${node.data('fullContent')}<br>
                                         <strong>Confidence:</strong> ${node.data('confidence')}`;
                    return content;
                },
                trigger: 'manual', // manual since we handle show/hide
                placement: 'top',
                hideOnClick: false,
                interactive: true,
            });

            node.on('mouseover', () => node.tippy.show());
            node.on('mouseout', () => node.tippy.hide());
        });

        this.cy.edges().forEach(edge => {
            const ref = edge.popperRef();
            const dummyDomEle = document.createElement('div');

            edge.tippy = tippy(dummyDomEle, {
                getReferenceClientRect: ref.getBoundingClientRect,
                content: () => {
                    const content = document.createElement('div');
                    content.innerHTML = `<strong>Weight:</strong> ${edge.data('weight')}`;
                    return content;
                },
                trigger: 'manual',
                placement: 'top',
                hideOnClick: false,
                interactive: true,
            });

            edge.on('mouseover', () => edge.tippy.show());
            edge.on('mouseout', () => edge.tippy.hide());
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
                            weight: relatedThought.similarity * 5  // Scale up for visibility
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
            this.initTippyTooltips(); // Reinitialize tooltips for new elements
            this.updateStats();
        }
    }

    runLayout() {
        const layout = this.cy.layout(this.layouts[this.currentLayout]);
        layout.run();
    }

    findRelatedThoughts(content, confidence, currentIteration, currentId) {
        const minSimilarity = parseFloat(this.shadowRoot.getElementById('minSimilarity').value) || 0.1;
        const minConfidence = parseFloat(this.shadowRoot.getElementById('minConfidence').value) || 0.1;
        const agentNameFilter = this.shadowRoot.getElementById('agentNameFilter').value.toLowerCase();

        return this.thoughtConnections
            .filter(thought => 
                thought.id !== currentId &&
                thought.confidence >= minConfidence &&
                (agentNameFilter === '' || thought.agentName.toLowerCase().includes(agentNameFilter))
            )
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
        // Basic TF calculation; for better results, integrate a more robust library or algorithm
        const getTermFrequency = (text) => {
            const words = text.toLowerCase().split(/\W+/).filter(Boolean);
            const tf = {};
            words.forEach(word => {
                tf[word] = (tf[word] || 0) + 1;
            });
            return tf;
        };

        const tf1 = getTermFrequency(text1);
        const tf2 = getTermFrequency(text2);

        const allTerms = new Set([...Object.keys(tf1), ...Object.keys(tf2)]);
        let dotProduct = 0;
        let magnitude1 = 0;
        let magnitude2 = 0;

        allTerms.forEach(term => {
            const tfidf1 = tf1[term] || 0;
            const tfidf2 = tf2[term] || 0;
            dotProduct += tfidf1 * tfidf2;
            magnitude1 += tfidf1 * tfidf1;
            magnitude2 += tfidf2 * tfidf2;
        });

        if (magnitude1 === 0 || magnitude2 === 0) return 0;

        return dotProduct / (Math.sqrt(magnitude1) * Math.sqrt(magnitude2));
    }

    applyFilters() {
        this.maxConnections = parseInt(this.shadowRoot.getElementById('maxConnections').value);
        this.minConnectionsForCluster = parseInt(this.shadowRoot.getElementById('minConnectionsForCluster').value);
        this.updateGraph();
    }

    updateGraph() {
        const agentNameFilter = this.shadowRoot.getElementById('agentNameFilter').value.toLowerCase();
        const minConfidence = parseFloat(this.shadowRoot.getElementById('minConfidence').value) || 0.1;
        const minSimilarity = parseFloat(this.shadowRoot.getElementById('minSimilarity').value) || 0.1;

        this.cy.startBatch();
        this.cy.elements().removeClass('filtered');

        this.cy.nodes().forEach(node => {
            const data = node.data();
            const matchesAgent = agentNameFilter === '' || data.agentName.toLowerCase().includes(agentNameFilter);
            const matchesConfidence = data.confidence >= minConfidence;
            if (!matchesAgent || !matchesConfidence) {
                node.addClass('filtered');
            } else {
                node.removeClass('filtered');
            }
        });

        this.cy.edges().forEach(edge => {
            const source = edge.source();
            const target = edge.target();
            if (source.hasClass('filtered') || target.hasClass('filtered')) {
                edge.addClass('filtered');
            } else {
                edge.removeClass('filtered');
            }
        });

        this.cy.endBatch();

        this.runLayout();
        this.updateStats();
    }

    clear() {
        this.cy.elements().remove();
        this.thoughtConnections = [];
        this.updateStats();
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
        this.updateStats();
    }

    clusterByConnections() {
        const nodesToCluster = this.cy.nodes().filter(node => 
            !node.isParent() && node.connectedEdges().length >= this.minConnectionsForCluster
        );

        nodesToCluster.forEach((node, index) => {
            const neighborhood = node.neighborhood().nodes().filter(n => !n.isParent());
            if (neighborhood.length >= this.minConnectionsForCluster) {
                const clusterId = 'cluster_' + this.cy.nodes().length;
                const clusterColor = this.clusterColors[index % this.clusterColors.length];
                this.cy.add({
                    group: 'nodes',
                    data: {
                        id: clusterId,
                        label: 'Cluster ' + (index + 1)
                    },
                    style: {
                        'background-color': clusterColor,
                        'opacity': 0.3
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
        this.updateStats();
    }

    changeLayout(layoutName) {
        this.currentLayout = layoutName;
        this.runLayout();
    }

    highlightNeighbors(node) {
        const neighborhood = node.neighborhood().add(node);
        
        this.cy.elements().removeClass('highlighted');
        neighborhood.addClass('highlighted');
        
        this.cy.elements().difference(neighborhood).style('opacity', 0.3);
        neighborhood.style('opacity', 1);
    }

    unhighlightAll() {
        this.cy.elements().removeClass('highlighted');
        this.cy.elements().style('opacity', 1);
    }

    showNodeDetailsModal(node) {
        const data = node.data();
        // Create modal if it doesn't exist
        if (!this.shadowRoot.getElementById('detailsModal')) {
            const modal = document.createElement('div');
            modal.id = 'detailsModal';
            modal.style.position = 'fixed';
            modal.style.top = '0';
            modal.style.left = '0';
            modal.style.width = '100%';
            modal.style.height = '100%';
            modal.style.backgroundColor = 'rgba(0,0,0,0.5)';
            modal.style.display = 'flex';
            modal.style.alignItems = 'center';
            modal.style.justifyContent = 'center';
            modal.style.zIndex = '1000';

            const modalContent = document.createElement('div');
            modalContent.style.backgroundColor = '#fff';
            modalContent.style.padding = '20px';
            modalContent.style.borderRadius = '5px';
            modalContent.style.width = '300px';
            modalContent.style.boxShadow = '0 5px 15px rgba(0,0,0,0.3)';
            modalContent.innerHTML = `
                <h3>Node Details</h3>
                <p><strong>Agent:</strong> <span id="modalAgentName"></span></p>
                <p><strong>Content:</strong> <span id="modalContent"></span></p>
                <p><strong>Confidence:</strong> <span id="modalConfidence"></span></p>
                <button id="closeModal">Close</button>
            `;

            modal.appendChild(modalContent);
            this.shadowRoot.appendChild(modal);

            this.shadowRoot.getElementById('closeModal').addEventListener('click', () => {
                modal.style.display = 'none';
            });
        }

        this.shadowRoot.getElementById('modalAgentName').textContent = data.agentName;
        this.shadowRoot.getElementById('modalContent').textContent = data.fullContent;
        this.shadowRoot.getElementById('modalConfidence').textContent = data.confidence;
        this.shadowRoot.getElementById('detailsModal').style.display = 'flex';
    }

    searchNodes() {
        const searchTerm = this.searchTerm.toLowerCase();
        this.highlightedNodes.clear();
        
        if (searchTerm) {
            this.cy.nodes().forEach(node => {
                const nodeData = node.data();
                if (nodeData.fullContent && nodeData.fullContent.toLowerCase().includes(searchTerm)) {
                    this.highlightedNodes.add(node);
                }
            });
            
            this.cy.elements().style('opacity', 0.3);
            this.highlightedNodes.forEach(node => {
                node.style('opacity', 1);
                node.neighborhood().style('opacity', 1);
            });
        } else {
            this.cy.elements().style('opacity', 1);
        }
    }

    debounce(func, wait) {
        clearTimeout(this.debounceTimer);
        this.debounceTimer = setTimeout(func, wait);
    }

    updateStats() {
        const stats = {
            nodes: this.cy.nodes(':visible').size(),
            edges: this.cy.edges(':visible').size(),
            clusters: this.cy.nodes(':parent').filter(node => node.isParent()).size(),
            avgConnections: this.cy.nodes(':visible').averageDegree()
        };

        const statsContainer = this.shadowRoot.getElementById('statsContainer');
        statsContainer.innerHTML = `
            Nodes: ${stats.nodes} | 
            Edges: ${stats.edges} | 
            Clusters: ${stats.clusters} | 
            Avg. Connections: ${stats.avgConnections.toFixed(2)}
        `;
    }

    exportGraph() {
        const graphData = {
            nodes: this.cy.nodes().map(node => node.data()),
            edges: this.cy.edges().map(edge => edge.data())
        };

        const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(graphData, null, 2));
        const downloadAnchorNode = document.createElement('a');
        downloadAnchorNode.setAttribute("href", dataStr);
        downloadAnchorNode.setAttribute("download", "network_graph_export.json");
        document.body.appendChild(downloadAnchorNode);
        downloadAnchorNode.click();
        downloadAnchorNode.remove();
    }
}

customElements.define('network-graph-visualization', NetworkGraphVisualization);
