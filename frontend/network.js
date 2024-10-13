// network-graph-visualization.js

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
            'cose': { name: 'cose', animate: 'end', animationDuration: 1000, randomize: false, componentSpacing: 100 },
            'concentric': { name: 'concentric', animate: 'end', animationDuration: 1000 },
            'breadthfirst': { name: 'breadthfirst', animate: 'end', animationDuration: 1000 },
            'circle': { name: 'circle', animate: 'end', animationDuration: 1000 },
            'grid': { name: 'grid', animate: 'end', animationDuration: 1000 }
        };
        this.currentLayout = 'cose';
        this.searchTerm = '';
        this.highlightedNodes = new Set();
        this.debounceTimer = null;
        this.animationSpeed = 1000;
        this.showLabels = true;
    }

    connectedCallback() {
        this.render();
        if (typeof cytoscape === 'undefined' || typeof tippy === 'undefined' || typeof coseBilkent === 'undefined') {
            console.error('Cytoscape.js, Tippy.js, and necessary extensions must be included before this component.');
            return;
        }
        this.initCytoscapeGraph();
    }

    disconnectedCallback() {
        // Remove event listeners
        this.removeEventListeners();

        // Destroy Cytoscape instance
        if (this.cy) {
            this.cy.destroy();
        }
    }

    render() {
        this.shadowRoot.innerHTML = `
        <style>
            /* Styles omitted for brevity */
            /* Add your styles here */
        </style>
        <div id="graphContainer">
            <div id="cytoscapeGraph"></div>
            <div id="controls">
                <!-- Controls omitted for brevity -->
                <!-- Add your controls here -->
            </div>
            <div id="searchContainer">
                <input type="text" id="searchInput" placeholder="Search nodes...">
                <button id="searchButton">Search</button>
            </div>
            <div id="statsContainer"></div>
        </div>
        `;

        // Attach event listeners
        this.attachEventListeners();
    }

    attachEventListeners() {
        this.shadowRoot.getElementById('applyFilters').addEventListener('click', this.applyFilters.bind(this));
        this.shadowRoot.getElementById('clusterGraph').addEventListener('click', this.clusterGraph.bind(this));
        this.shadowRoot.getElementById('layoutSelect').addEventListener('change', (e) => this.changeLayout(e.target.value));
        this.shadowRoot.getElementById('exportGraph').addEventListener('click', () => this.exportGraph('png'));
        this.shadowRoot.getElementById('exportSvg').addEventListener('click', () => this.exportGraph('svg'));
        this.shadowRoot.getElementById('searchButton').addEventListener('click', this.searchNodes.bind(this));
        this.shadowRoot.getElementById('searchInput').addEventListener('input', (e) => {
            this.searchTerm = e.target.value;
            this.debounce(this.searchNodes.bind(this), 300);
        });
        this.shadowRoot.getElementById('animationSpeed').addEventListener('input', (e) => {
            this.animationSpeed = parseInt(e.target.value, 10);
            this.updateAnimationSpeed();
        });
        this.shadowRoot.getElementById('toggleLabels').addEventListener('change', (e) => {
            this.showLabels = e.target.checked;
            this.toggleLabels();
        });
    }

    removeEventListeners() {
        // Remove all attached event listeners to avoid memory leaks
        // Example:
        this.shadowRoot.getElementById('applyFilters').removeEventListener('click', this.applyFilters.bind(this));
        // ... remove other event listeners similarly
    }

    initCytoscapeGraph() {
        const container = this.shadowRoot.getElementById('cytoscapeGraph');
        this.cy = cytoscape({
            container: container,
            style: this.getCytoscapeStyles(),
            layout: this.layouts[this.currentLayout],
            wheelSensitivity: 0.1, // Improve zooming experience
        });

        // Initialize event handlers
        this.initEventHandlers();
    }

    getCytoscapeStyles() {
        return [
            {
                selector: 'node',
                style: {
                    'background-color': '#666',
                    'label': this.showLabels ? 'data(label)' : '',
                    'text-valign': 'center',
                    'text-halign': 'center',
                    'text-wrap': 'wrap',
                    'text-max-width': '100px',
                    'font-size': '10px',
                    'color': '#000'
                }
            },
            {
                selector: 'edge',
                style: {
                    'width': 'data(weight)',
                    'line-color': '#ccc',
                    'curve-style': 'bezier',
                    'opacity': 0.7
                }
            },
            {
                selector: ':parent',
                style: {
                    'background-opacity': 0.333,
                    'label': this.showLabels ? 'data(label)' : '',
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
                    'width': 4
                }
            },
            {
                selector: '.filtered',
                style: {
                    'display': 'none'
                }
            }
        ];
    }

    initEventHandlers() {
        this.cy.on('tap', 'node', (evt) => {
            const node = evt.target;
            if (node.isParent()) {
                this.expandCluster(node);
            } else {
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

        // Handle layout ready event for animations
        this.cy.on('layoutready', () => {
            // Implement any additional logic after layout is ready
        });
    }

    /**
     * Adds a new thought to the network graph.
     * @param {Object} thoughtData - The data for the thought to add.
     * @param {string} thoughtData.id - The unique identifier for the thought.
     * @param {string} thoughtData.agentName - The name of the agent.
     * @param {string} thoughtData.content - The content of the thought.
     * @param {number} thoughtData.iteration - The iteration number.
     * @param {number} thoughtData.step - The step number.
     * @param {number} thoughtData.confidence - The confidence level.
     */
    addThought({ id, agentName, content, iteration, step, confidence }) {
        const existingNode = this.cy.getElementById(id);
        if (existingNode.length > 0) {
            console.warn(`Node with id ${id} already exists.`);
            return;
        }

        this.thoughtConnections.push({ id, agentName, content, iteration, step, confidence });

        const relatedThoughts = this.findRelatedThoughts(content, confidence, iteration, id);

        if (relatedThoughts.length > 0) {
            const nodesToAdd = [];
            const edgesToAdd = [];

            nodesToAdd.push({
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
                if (relatedThoughtData && this.cy.getElementById(relatedThought.id).empty()) {
                    nodesToAdd.push({
                        group: 'nodes',
                        data: {
                            id: relatedThought.id,
                            label: `${relatedThoughtData.agentName}\n${relatedThoughtData.content.substring(0, 20)}...`,
                            fullContent: relatedThoughtData.content,
                            agentName: relatedThoughtData.agentName,
                            confidence: relatedThoughtData.confidence
                        }
                    });
                }

                const edgeId = `${relatedThought.id}-${id}`;
                if (this.cy.getElementById(edgeId).empty()) {
                    edgesToAdd.push({
                        group: 'edges',
                        data: {
                            id: edgeId,
                            source: relatedThought.id,
                            target: id,
                            weight: relatedThought.similarity * 5  // Scale up for visibility
                        }
                    });
                }
            });

            this.cy.add([...nodesToAdd, ...edgesToAdd]);
            this.runLayout();
            this.initTooltipsForElements(this.cy.$(nodesToAdd.map(item => `#${item.data.id}`)));
            this.updateStats();
        }
    }

    runLayout() {
        const layoutOptions = {
            ...this.layouts[this.currentLayout],
            animate: 'end',
            animationDuration: this.animationSpeed,
        };
        const layout = this.cy.layout(layoutOptions);
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
        const tf1 = this.getTermFrequency(text1);
        const tf2 = this.getTermFrequency(text2);
        return this.computeCosineSimilarity(tf1, tf2);
    }

    getTermFrequency(text) {
        const words = text.toLowerCase().split(/\W+/).filter(Boolean);
        const tf = {};
        words.forEach(word => {
            tf[word] = (tf[word] || 0) + 1;
        });
        return tf;
    }

    computeCosineSimilarity(tf1, tf2) {
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

        this.cy.batch(() => {
            this.cy.nodes().forEach(node => {
                const data = node.data();
                const matchesAgent = agentNameFilter === '' || data.agentName.toLowerCase().includes(agentNameFilter);
                const matchesConfidence = data.confidence >= minConfidence;
                node.style('display', matchesAgent && matchesConfidence ? 'element' : 'none');
            });

            this.cy.edges().forEach(edge => {
                const sourceVisible = edge.source().style('display') !== 'none';
                const targetVisible = edge.target().style('display') !== 'none';
                edge.style('display', sourceVisible && targetVisible ? 'element' : 'none');
            });
        });

        this.runLayout();
        this.updateStats();
    }

    clear() {
        this.cy.elements().remove();
        this.thoughtConnections = [];
        this.updateStats();
    }

    clusterGraph() {
        if (typeof cytoscapeLouvain === 'undefined') {
            console.error('Cytoscape.js Louvain extension must be included for clustering.');
            return;
        }

        const clusters = this.cy.elements().communityDetection();
        clusters.communities.forEach((community, index) => {
            const clusterId = `cluster_${index}`;
            const clusterColor = this.clusterColors[index % this.clusterColors.length];
            this.cy.add({
                group: 'nodes',
                data: { id: clusterId, label: `Cluster ${index + 1}` },
                style: {
                    'background-color': clusterColor,
                    'opacity': 0.3
                }
            });

            community.forEach(node => {
                node.move({ parent: clusterId });
            });
        });

        this.runLayout();
        this.updateStats();
    }

    expandCluster(clusterNode) {
        this.cy.batch(() => {
            const childNodes = clusterNode.children();
            childNodes.move({ parent: null });
            this.cy.remove(clusterNode);
        });
        this.runLayout();
        this.updateStats();
    }

    changeLayout(layoutName) {
        this.currentLayout = layoutName;
        this.runLayout();
    }

    highlightNeighbors(node) {
        const neighborhood = node.closedNeighborhood();

        this.cy.batch(() => {
            this.cy.elements().removeClass('highlighted');
            neighborhood.addClass('highlighted');
            this.cy.elements().difference(neighborhood).addClass('faded');
            neighborhood.removeClass('faded');
        });
    }

    unhighlightAll() {
        this.cy.batch(() => {
            this.cy.elements().removeClass('highlighted faded');
        });
    }

    showNodeDetailsModal(node) {
        const data = node.data();
        const modal = this.createModal();
        this.shadowRoot.appendChild(modal);

        this.shadowRoot.getElementById('modalAgentName').textContent = data.agentName;
        this.shadowRoot.getElementById('modalContent').textContent = data.fullContent;
        this.shadowRoot.getElementById('modalConfidence').textContent = data.confidence;
        this.shadowRoot.getElementById('detailsModal').style.display = 'flex';
    }

    createModal() {
        if (this.shadowRoot.getElementById('detailsModal')) {
            return this.shadowRoot.getElementById('detailsModal');
        }

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

        this.shadowRoot.getElementById('closeModal').addEventListener('click', () => {
            modal.style.display = 'none';
        });

        return modal;
    }

    searchNodes() {
        const searchTerm = this.searchTerm.toLowerCase();
        const regex = new RegExp(searchTerm, 'i');
        this.highlightedNodes.clear();

        if (searchTerm) {
            this.cy.nodes().forEach(node => {
                const nodeData = node.data();
                if (nodeData.fullContent && regex.test(nodeData.fullContent)) {
                    this.highlightedNodes.add(node);
                }
            });

            this.cy.batch(() => {
                this.cy.elements().addClass('faded');
                this.highlightedNodes.forEach(node => {
                    node.removeClass('faded');
                    node.neighborhood().removeClass('faded');
                });
            });
        } else {
            this.cy.batch(() => {
                this.cy.elements().removeClass('faded');
            });
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
            avgConnections: this.cy.nodes(':visible').averageDegree(false)
        };

        const statsContainer = this.shadowRoot.getElementById('statsContainer');
        statsContainer.innerHTML = `
            Nodes: ${stats.nodes} | 
            Edges: ${stats.edges} | 
            Clusters: ${stats.clusters} | 
            Avg. Connections: ${stats.avgConnections.toFixed(2)}
        `;
    }

    exportGraph(format) {
        let dataStr;
        if (format === 'json') {
            const graphData = {
                nodes: this.cy.nodes().map(node => node.data()),
                edges: this.cy.edges().map(edge => edge.data())
            };
            dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(graphData, null, 2));
        } else if (format === 'png' || format === 'svg') {
            const blob = this.cy[format]({ full: true });
            const url = URL.createObjectURL(blob);
            dataStr = url;
        } else {
            console.error('Unsupported export format:', format);
            return;
        }

        const downloadAnchorNode = document.createElement('a');
        downloadAnchorNode.setAttribute("href", dataStr);
        downloadAnchorNode.setAttribute("download", `network_graph_export.${format}`);
        document.body.appendChild(downloadAnchorNode);
        downloadAnchorNode.click();
        downloadAnchorNode.remove();
    }

    toggleLabels() {
        this.cy.style().selector('node').style('label', this.showLabels ? 'data(label)' : '').update();
        this.cy.style().selector(':parent').style('label', this.showLabels ? 'data(label)' : '').update();
    }

    updateAnimationSpeed() {
        this.runLayout();
    }

    initTooltipsForElements(elements) {
        elements.forEach(ele => {
            if (ele.isNode()) {
                this.createNodeTooltip(ele);
            } else if (ele.isEdge()) {
                this.createEdgeTooltip(ele);
            }
        });
    }

    createNodeTooltip(node) {
        const ref = node.popperRef();
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
            trigger: 'manual',
            placement: 'top',
            hideOnClick: false,
            interactive: true,
        });

        node.on('mouseover', () => node.tippy.show());
        node.on('mouseout', () => node.tippy.hide());
    }

    createEdgeTooltip(edge) {
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
    }
}

// Register the custom element
customElements.define('network-graph-visualization', NetworkGraphVisualization);
