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
        this.content = ""
        this.colors = {
            reasoner: "#6E95F7",
            verifier: "#F7746D",
            learning: "#F7B96D",
            metacognition: "#06C26F",
            context_manager: "#F76D95"
        }
    }

    connectedCallback() {
        this.render();
        this.currentBlock = this.shadowRoot.getElementById("output");
        this.networkGraph = this.shadowRoot.querySelector("network-graph-visualization");
        this.setupWebSocket();
        this.updateConfidenceGraph();
    }

    setupWebSocket() {
        this.ws = new WebSocket("ws://localhost:8567/ws");
        this.ws.onmessage = (event) => {
            const chunk = JSON.parse(event.data);
            this.process(chunk);
        };
        this.ws.onerror = (error) => {
            console.error("WebSocket error:", error);
        };
        this.ws.onclose = () => {
            console.warn("WebSocket connection closed");
        };
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

    createDatasets() {
        const barDatasets = Object.values(this.agents).map(agent => ({
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

        return [...barDatasets, lineDataset];
    }

    updateConfidenceGraph() {
        this.shadowRoot.getElementById('confidenceChart').style.width = '100%';

        const ctx = this.shadowRoot.getElementById('confidenceChart');
        if (!ctx) {
            console.error('Cannot find confidence chart canvas');
            return;
        }

        const data = {
            labels: this.getLabels(),
            datasets: this.createDatasets()
        };

        const options = {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true,
                    suggestedMin: -1,
                    suggestedMax: 1,
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
                }
            },
            animation: {
                duration: 500
            }
        };

        if (this.chart) {
            this.chart.data = data;
            this.chart.options = options;
            this.chart.update('none');
        } else {
            this.chart = new Chart(ctx, {
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

    process(chunk) {
        console.log('Processing chunk:', chunk);
        this.iteration = chunk.iteration;
        this.maxIterations = chunk.maxIterations || this.maxIterations;

        if (!this.agents[chunk.agent]) {
            this.agents[chunk.agent] = {
                ...chunk,
                agent_type: chunk.agent_type,
                confidences: [],
                color: this.colors[chunk.agent_type]
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
    }

    getLabels() {
        const maxLength = Math.max(...Object.values(this.agents).map(a => a.confidences.length));
        return Array.from({ length: maxLength }, (_, i) => `Step ${i + 1}`);
    }

    updateNetworkGraph(chunk) {
        if (this.networkGraph) {
            this.networkGraph.addThought({
                id: `thought_${this.iteration}_${this.step}_${chunk.agent}`,
                agentName: chunk.agentType,
                content: chunk.response,
                iteration: this.iteration,
                step: this.step,
                confidence: this.currentAgent.confidences[this.currentAgent.confidences.length - 1] || 0
            });
            this.step++;
        }
    }

    getRandomColor() {
        const letters = '0123456789ABCDEF';
        let color = '#';
        for (let i = 0; i < 6; i++) {
            color += letters[Math.floor(Math.random() * 16)];
        }
        return color;
    }

    processThought(line) {
        const thoughtId = `thought_${this.iteration}_${this.step}_${this.currentAgent.agent_type}`;
        const agentName = this.currentAgent.agent_type;
        const [thoughtContent, confidenceStr] = line.split('(confidence:');
        const confidence = confidenceStr.includes('high') ? 1 : confidenceStr.includes('medium') ? 0 : -1;
        
        this.networkGraph.addThought({
            id: thoughtId,
            agentName: agentName,
            content: thoughtContent.trim(),
            iteration: this.currentIteration,
            step: this.currentStep,
            confidence: confidence
        });

        this.currentStep++;
    }

    updateNavButtons() {
        const prevBtn = this.shadowRoot.getElementById('prevBtn');
        const nextBtn = this.shadowRoot.getElementById('nextBtn');
        
        prevBtn.disabled = this.currentIteration === 0;
        nextBtn.disabled = this.currentIteration === this.data.iterations.length - 1;
    }

    navigate(direction) {
        const newIteration = this.currentIteration + direction;
        if (newIteration >= 0 && newIteration < this.data.iterations.length) {
            this.currentIteration = newIteration;
            this.updateConfidenceGraph();
            this.updateNavButtons();
        }
    }

    addNewBlock(chunk) {
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
        // Get position of the new block
        const newBlockPosition = this.currentBlock.getBoundingClientRect();
        const targetY = newBlockPosition.top + (newBlockPosition.height / 2) - 32;
        const targetX = newBlockPosition.right + 64;
        // Position the animoji loader at the vertical center, right of the new block.
        this.currentAnimojiLoader = this.shadowRoot.getElementById('animojiLoader');
        this.currentAnimojiLoader.setPosition(targetX, targetY);
        this.currentAnimojiLoader.setState('thinking');
    }

    addLineToOutput(chunk) {
        if (!this.currentBlock) return;

        if (this.content) {
            const contentElement = document.createElement('div');
            contentElement.innerHTML = marked.parse(this.content);
            this.currentBlock.appendChild(contentElement);
            this.content = "";
        }

        this.content += chunk.response.replace("```markdown", "").replace("```", "");

        // Update animoji state based on content
        if (this.currentAnimojiLoader) {
            if (chunk.response.includes('high')) {
                this.currentAnimojiLoader.setState('high');
            } else if (chunk.response.includes('medium')) {
                this.currentAnimojiLoader.setState('medium');
            } else if (chunk.response.includes('low')) {
                this.currentAnimojiLoader.setState('low');
            } else {
                this.currentAnimojiLoader.setState('thinking');
            }
        }
    }

    moveToNextCard() {
        const nextCard = this.currentBlock.closest('.card').nextElementSibling;
        if (nextCard && nextCard.classList.contains('card')) {
            this.currentBlock = nextCard.querySelector('.card-content');
        }
    }

    render() {
        this.shadowRoot.innerHTML = `
        <style>
            :host {
                display: flex;
                font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
                height: 100vh;
                width: 100vw;
                margin: 0;
                --primary-color: #3498db;
                --secondary-color: #2ecc71;
                --background-color: #f7f9fc;
                --text-color: #34495e;
                --card-background: #ffffff;
            }
            
            .container {
                display: flex;
                flex-grow: 1;
                flex-direction: row;
                justify-content: center;
                align-items: stretch;
                width: 100%;
                background-color: var(--background-color);
            }
            
            .visualization {
                flex-grow: 1;
                padding: 20px;
                gap: 20px;
                display: flex;
                width: 50%;
                flex-direction: column;
                align-items: stretch;
                justify-content: flex-start;
            }
            
            .output {
                flex-grow: 1;
                padding: 20px;
                background-color: var(--card-background);
                width: 50%;
                overflow-y: auto;
                border-left: 1px solid var(--primary-color);
            }
            
            .card {
                background-color: var(--card-background);
                border-radius: 8px;
                box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
                margin-bottom: 16px;
                overflow: hidden;
                transition: all 0.3s ease;
                opacity: 0;
                transform: translateY(20px);
                position: relative;
            }

            .card.show {
                opacity: 1;
                transform: translateY(0);
            }

            .card summary {
                padding: 16px;
                cursor: pointer;
                font-weight: bold;
                background-color: var(--agent-color);
                color: white;
                transition: background-color 0.3s ease;
            }

            .card summary:hover {
                background-color: color-mix(in srgb, var(--agent-color) 80%, white 20%);
            }

            .card[open] summary {
                border-bottom: 1px solid rgba(0, 0, 0, 0.1);
            }

            .card-content {
                padding: 16px;
                display: flex;
                flex-direction: column;
                gap: 12px;
            }

            .confidence-graph {
                flex-grow: 1;
                margin-top: 20px;
                min-height: 200px;
            }
            
            .navigation {
                display: flex;
                justify-content: space-between;
                margin-bottom: 15px;
            }
            
            button {
                background-color: var(--primary-color);
                color: white;
                border: none;
                padding: 10px 15px;
                border-radius: 4px;
                cursor: pointer;
                transition: background-color 0.3s ease;
            }
            
            button:hover {
                background-color: #2980b9;
            }
            
            button:disabled {
                background-color: #bdc3c7;
                cursor: not-allowed;
            }

            .input-group {
                display: flex;
                gap: 10px;
                margin-bottom: 15px;
            }

            input[type="number"], input[type="text"] {
                flex-grow: 1;
                padding: 10px;
                border: 1px solid var(--primary-color);
                border-radius: 4px;
                font-size: 16px;
            }

            input[type="number"] {
                width: 60px;
            }

            network-graph-visualization {
                width: 100%;
                height: 400px;
                margin-top: 20px;
            }

            @keyframes fadeIn {
                from { opacity: 0; transform: translateY(20px); }
                to { opacity: 1; transform: translateY(0); }
            }

            .fade-in {
                animation: fadeIn 0.5s ease-out;
            }
        </style>
        <div class="container">
            <div class="visualization">
                <div class="navigation">
                    <button id="prevBtn">Previous</button>
                    <span id="iteration">Iteration: ${this.iteration + 1} of ${this.maxIterations || 0}</span>
                    <button id="nextBtn">Next</button>
                </div>
                <div class="confidence-graph" style="position: relative; height:20vh; width:100%">
                    <canvas id="confidenceChart" style="position: relative; height: 100%; width:100%"></canvas>
                </div>
                <network-graph-visualization></network-graph-visualization>
                <div class="input-group">
                    <input type="number" id="iterationsInput" min="1" max="10" value="3" placeholder="Iterations">
                    <input type="text" id="promptInput" value="Solve the riddle: In a fruit's sweet name, I'm hidden three, A triple threat within its juicy spree. Find me and you'll discover a secret delight." placeholder="Enter your prompt">
                    <button id="sendPrompt">Send Prompt</button>
                </div>
            </div>
            <div id="output" class="output"></div>
        </div>
        <animoji-loader id="animojiLoader"></animoji-loader>
        <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
        <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
        `;
    
        // Initialize buttons and event listeners
        this.prevBtn = this.shadowRoot.getElementById('prevBtn');
        this.nextBtn = this.shadowRoot.getElementById('nextBtn');
        const sendPromptButton = this.shadowRoot.getElementById('sendPrompt');
        const iterationsInput = this.shadowRoot.getElementById('iterationsInput');
        const promptInput = this.shadowRoot.getElementById('promptInput');

        this.prevBtn.addEventListener('click', () => this.navigate(-1));
        this.nextBtn.addEventListener('click', () => this.navigate(1));
        sendPromptButton.addEventListener('click', () => {
            const iterations = iterationsInput.value;
            const prompt = promptInput.value;
            this.sendPrompt(`${iterations}<:>${prompt}`);
        });

        // Initialize the chart only if Chart.js is already loaded
        if (typeof Chart !== 'undefined') {
            this.updateConfidenceGraph();
        } else {
            console.error('Chart.js is not loaded. Ensure it is included before the component script.');
        }
    }

    navigate(direction) {
        const newIteration = this.currentIteration + direction;

        if (newIteration >= 0 && newIteration < this.data.iterations.length) {
            this.currentIteration = newIteration;
            this.updateConfidenceGraph();
            this.updateNavButtons();
        }
    }
}

customElements.define('pipeline-visualization', PipelineVisualization);