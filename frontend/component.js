// pipeline-visualization.js

class PipelineVisualization extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: "open" });
        this.iteration = 0;
        this.step = 0;
        this.confidences = [];
        this.agents = {};
        this.currentAgent = null;
        this.currentBlock = null;
        this.chart = null;
        this.networkGraph = null;
        this.ws = null;
        this.content = "";
        this.colors = {
            reasoner: "#6E95F7",
            verifier: "#F7746D",
            learning: "#F7B96D",
            metacognition: "#06C26F",
            context_manager: "#F76D95"
        };
        this.maxIterations = 0;
        this.sessionData = [];
    }

    connectedCallback() {
        this.render();
        this.currentBlock = this.shadowRoot.getElementById("output");
        this.networkGraph = this.shadowRoot.querySelector("network-graph-visualization");
        this.setupWebSocket();
        this.updateConfidenceGraph();
        this.attachEventListeners();
    }

    disconnectedCallback() {
        // Close WebSocket connection
        if (this.ws) {
            this.ws.close();
        }

        // Remove event listeners
        this.removeEventListeners();
    }

    setupWebSocket() {
        this.ws = new WebSocket("ws://localhost:8567/ws");
        this.ws.onmessage = (event) => {
            try {
                const chunk = JSON.parse(event.data);
                this.process(chunk);
            } catch (e) {
                console.error('Failed to parse WebSocket message:', e);
            }
        };
        this.ws.onerror = (error) => {
            console.error("WebSocket error:", error);
            this.showNotification('WebSocket error occurred. Please check the connection.', 'error');
        };
        this.ws.onclose = () => {
            console.warn("WebSocket connection closed, attempting to reconnect...");
            this.showNotification('WebSocket connection lost. Reconnecting...', 'warning');
            setTimeout(() => this.setupWebSocket(), 1000);
        };
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        this.shadowRoot.appendChild(notification);
        setTimeout(() => {
            notification.remove();
        }, 5000);
    }

    process(chunk) {
        this.iteration = chunk.iteration;
        this.maxIterations = chunk.maxIterations || this.maxIterations;

        if (!this.agents[chunk.agent]) {
            this.agents[chunk.agent] = {
                ...chunk,
                agent_type: chunk.agent_type,
                confidences: [],
                color: this.colors[chunk.agent_type] || this.getRandomColor()
            };
        }

        if (!this.currentAgent || this.currentAgent.agent !== chunk.agent) {
            this.currentAgent = this.agents[chunk.agent];
            this.addNewBlock(chunk);
        }

        this.addLineToOutput(chunk);

        if (chunk.response.includes('confidence:')) {
            const confidence = chunk.response.includes('high') ? 1 : chunk.response.includes('medium') ? 0 : -1;
            this.currentAgent.confidences.push(confidence);
            this.updateConfidenceGraph();
        }

        this.updateNetworkGraph(chunk);

        const iterationSpan = this.shadowRoot.getElementById('iteration');
        if (iterationSpan) {
            iterationSpan.textContent = `Iteration: ${this.iteration} of ${this.maxIterations}`;
        }

        // Save session data
        this.sessionData.push(chunk);
    }

    updateConfidenceGraph() {
        const chartContainer = this.shadowRoot.querySelector('.confidence-graph');
        const canvas = this.shadowRoot.getElementById('confidenceChart');

        if (!canvas) {
            console.error('Cannot find confidence chart canvas');
            return;
        }

        const data = {
            labels: this.getLabels(),
            datasets: this.createDatasets()
        };

        const options = {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    min: -1,
                    max: 1,
                    title: {
                        display: true,
                        text: 'Confidence'
                    }
                }
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    callbacks: {
                        label: (context) => {
                            const agent = context.dataset.label;
                            const confidence = context.parsed.y;
                            return `${agent}: ${confidence}`;
                        }
                    }
                }
            },
            animation: {
                duration: 500
            }
        };

        if (this.chart) {
            // Update existing chart data
            this.chart.data.labels = this.getLabels();
            this.chart.data.datasets = this.createDatasets();
            this.chart.update();
        } else {
            // Initialize chart
            this.chart = new Chart(canvas, {
                type: 'bar',
                data: data,
                options: options
            });
        }
    }

    sendPrompt(prompt) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(prompt);
            this.resetVisualization();
        } else {
            console.error('WebSocket is not open');
            this.showNotification('WebSocket is not connected. Please try again later.', 'error');
            this.setupWebSocket(); // Try to reconnect
        }
    }

    resetVisualization() {
        this.agents = {};
        this.currentAgent = null;
        this.confidences = [];
        this.iteration = 0;
        this.maxIterations = 0;
        this.step = 0;
        this.sessionData = [];
        this.shadowRoot.getElementById('output').innerHTML = '';

        const iterationSpan = this.shadowRoot.getElementById('iteration');
        if (iterationSpan) {
            iterationSpan.textContent = `Iteration: 0 of 0`;
        }

        if (this.chart) {
            this.chart.destroy();
            this.chart = null;
        }

        this.updateConfidenceGraph();
        if (this.networkGraph) {
            this.networkGraph.clear();
        }
    }

    getLabels() {
        const maxLength = Math.max(...Object.values(this.agents).map(a => a.confidences.length));
        return Array.from({ length: maxLength }, (_, i) => `Step ${i + 1}`);
    }

    createDatasets() {
        const datasets = Object.values(this.agents).map(agent => ({
            label: agent.agent_type,
            data: agent.confidences,
            backgroundColor: agent.color,
            type: 'bar'
        }));

        const totalConfidences = this.calculateTotalConfidences();
        const lineDataset = {
            label: 'Total Confidence',
            data: totalConfidences,
            borderColor: 'rgba(0, 0, 0, 0.7)',
            backgroundColor: 'rgba(0, 0, 0, 0.7)',
            type: 'line',
            fill: false,
            tension: 0.1
        };

        return [...datasets, lineDataset];
    }

    calculateTotalConfidences() {
        const maxLength = Math.max(...Object.values(this.agents).map(a => a.confidences.length), 0);
        const totalConfidences = new Array(maxLength).fill(0);

        Object.values(this.agents).forEach(agent => {
            agent.confidences.forEach((confidence, index) => {
                totalConfidences[index] += confidence;
            });
        });

        return totalConfidences;
    }

    updateNetworkGraph(chunk) {
        if (this.networkGraph) {
            this.networkGraph.addThought({
                id: `thought_${this.iteration}_${this.step}_${chunk.agent}`,
                agentName: chunk.agent_type,
                content: chunk.response,
                iteration: this.iteration,
                step: this.step,
                confidence: this.currentAgent.confidences[this.currentAgent.confidences.length - 1] || 0
            });
            this.step++;
        }
    }

    addNewBlock(chunk) {
        if (this.content && this.currentBlock) {
            const contentElement = document.createElement('div');
            contentElement.innerHTML = marked.parse(this.content);
            this.currentBlock.appendChild(contentElement);
            this.content = "";
        }

        const outputDiv = this.shadowRoot.getElementById('output');
        const blockElement = document.createElement('details');
        blockElement.classList.add('card');
        blockElement.style.setProperty('--agent-color', this.colors[chunk.agent_type] || '#000');

        const summary = document.createElement('summary');
        summary.textContent = chunk.agent_type || chunk.agent || 'Unknown Agent';
        blockElement.appendChild(summary);

        const content = document.createElement('div');
        content.classList.add('card-content');

        blockElement.appendChild(content);
        outputDiv.appendChild(blockElement);
        this.currentBlock = content;

        // Trigger the animation
        setTimeout(() => blockElement.classList.add('show'), 10);

        outputDiv.scrollTop = outputDiv.scrollHeight;
    }

    addLineToOutput(chunk) {
        this.content += chunk.response.replace("```markdown", "").replace("```", "");
    }

    attachEventListeners() {
        this.prevBtn = this.shadowRoot.getElementById('prevBtn');
        this.nextBtn = this.shadowRoot.getElementById('nextBtn');
        const sendPromptButton = this.shadowRoot.getElementById('sendPrompt');
        const iterationsInput = this.shadowRoot.getElementById('iterationsInput');
        const promptInput = this.shadowRoot.getElementById('promptInput');
        const saveSessionBtn = this.shadowRoot.getElementById('saveSession');
        const loadSessionBtn = this.shadowRoot.getElementById('loadSession');

        this.prevBtn.addEventListener('click', () => this.navigate(-1));
        this.nextBtn.addEventListener('click', () => this.navigate(1));
        sendPromptButton.addEventListener('click', () => {
            const iterations = parseInt(iterationsInput.value, 10);
            const prompt = promptInput.value.trim();

            if (isNaN(iterations) || iterations <= 0) {
                alert('Please enter a valid number of iterations.');
                return;
            }

            if (!prompt) {
                alert('Please enter a prompt.');
                return;
            }

            this.sendPrompt(`${iterations}<:>${prompt}`);
        });

        saveSessionBtn.addEventListener('click', () => this.saveSession());
        loadSessionBtn.addEventListener('click', () => this.loadSession());
    }

    removeEventListeners() {
        // Remove event listeners added in attachEventListeners
        // Example:
        this.prevBtn.removeEventListener('click', () => this.navigate(-1));
        // ... remove other event listeners similarly
    }

    navigate(direction) {
        const newIteration = this.iteration + direction;

        if (newIteration >= 0 && newIteration <= this.maxIterations) {
            this.iteration = newIteration;
            this.updateVisualization();
            this.updateNavButtons();
        }
    }

    updateVisualization() {
        // Implement logic to update the visualization based on the current iteration
        // This may involve updating the confidence graph, output blocks, etc.
    }

    updateNavButtons() {
        this.prevBtn.disabled = this.iteration <= 0;
        this.nextBtn.disabled = this.iteration >= this.maxIterations;
    }

    saveSession() {
        const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(this.sessionData, null, 2));
        const downloadAnchorNode = document.createElement('a');
        downloadAnchorNode.setAttribute("href", dataStr);
        downloadAnchorNode.setAttribute("download", "session_data.json");
        document.body.appendChild(downloadAnchorNode);
        downloadAnchorNode.click();
        downloadAnchorNode.remove();
    }

    loadSession() {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.json';
        input.onchange = (event) => {
            const file = event.target.files[0];
            if (!file) return;
            const reader = new FileReader();
            reader.onload = (e) => {
                try {
                    const data = JSON.parse(e.target.result);
                    this.loadSessionData(data);
                } catch (error) {
                    alert('Failed to load session data. Invalid file format.');
                }
            };
            reader.readAsText(file);
        };
        input.click();
    }

    loadSessionData(data) {
        this.resetVisualization();
        data.forEach(chunk => this.process(chunk));
        this.showNotification('Session loaded successfully.', 'success');
    }

    getRandomColor() {
        const letters = '0123456789ABCDEF';
        let color = '#';
        for (let i = 0; i < 6; i++) {
            color += letters[Math.floor(Math.random() * 16)];
        }
        return color;
    }

    render() {
        this.shadowRoot.innerHTML = `
        <style>
            /* Styles omitted for brevity */
            /* Add your styles here */
        </style>
        <div class="container">
            <div class="visualization">
                <div class="navigation">
                    <button id="prevBtn">Previous</button>
                    <span id="iteration">Iteration: ${this.iteration} of ${this.maxIterations || 0}</span>
                    <button id="nextBtn">Next</button>
                </div>
                <div class="confidence-graph" style="position: relative; height:30vh; width:100%">
                    <canvas id="confidenceChart" style="position: relative; height: 100%; width:100%"></canvas>
                </div>
                <network-graph-visualization></network-graph-visualization>
                <div class="input-group">
                    <input type="number" id="iterationsInput" min="1" max="10" value="3" placeholder="Iterations">
                    <input type="text" id="promptInput" value="Solve the riddle..." placeholder="Enter your prompt">
                    <button id="sendPrompt">Send Prompt</button>
                </div>
                <div class="session-controls">
                    <button id="saveSession">Save Session</button>
                    <button id="loadSession">Load Session</button>
                </div>
            </div>
            <div id="output" class="output"></div>
        </div>
        <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
        <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
        `;

        // Initialize the chart if Chart.js is loaded
        if (typeof Chart !== 'undefined') {
            this.updateConfidenceGraph();
        } else {
            console.error('Chart.js is not loaded. Ensure it is included before the component script.');
        }
    }
}

// Register the custom element
customElements.define('pipeline-visualization', PipelineVisualization);
