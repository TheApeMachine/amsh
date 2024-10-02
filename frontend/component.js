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
        this.ws = new WebSocket("ws://localhost:8080/ws");
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
        if (this.currentBlock) {
            const lineElement = document.createElement('div');
            lineElement.innerHTML = marked.parse(this.content);
            this.content = ""
            lineElement.style.opacity = '0';
            this.currentBlock.appendChild(lineElement);
    
            setTimeout(() => {
                lineElement.style.transition = 'opacity 0.5s ease-in-out';
                lineElement.style.opacity = '1';
            }, 10);
    
            const outputDiv = this.shadowRoot.getElementById('output');
            outputDiv.scrollTop = outputDiv.scrollHeight;    
        }

        const outputDiv = this.shadowRoot.getElementById('output');
        const blockElement = document.createElement('fieldset');
        blockElement.classList.add('block');
        const legend = document.createElement('legend');
        legend.textContent = chunk.agent_type || chunk.agent || 'Unknown Agent';
        blockElement.appendChild(legend);
        const loader = document.createElement('script');
        loader.src = `https://unpkg.com/@dotlottie/player-component@latest/dist/dotlottie-player.mjs` 
        loader.setAttribute("type", "module")
        const lottie = document.createElement('dotlottie-player');
        lottie.src = "https://lottie.host/11756a88-ccc6-402b-a472-7be2542bad28/UVq9Y6hMdS.json"
        lottie.setAttribute("background", "transparent");
        lottie.setAttribute("speed", "1");
        lottie.setAttribute("style", "width: 300px; height: 300px;");
        lottie.setAttribute("loop", "true");
        lottie.setAttribute("autoplay", "true");
        blockElement.appendChild(loader);
        blockElement.appendChild(lottie);
        outputDiv.appendChild(blockElement);
        this.currentBlock = blockElement;
        outputDiv.scrollTop = outputDiv.scrollHeight;    
    }

    addLineToOutput(chunk) {
        this.content += chunk.response.replace("```markdown", "").replace("```", "")
    }

    render() {
        this.shadowRoot.innerHTML = `
        <style>
            :host {
                display: flex;
                font-family: Arial, sans-serif;
                height: 100vh;
                width: 100vw;
                margin: 0;
                --primary-color: #3498db;
                --secondary-color: #2ecc71;
                --background-color: #ecf0f1;
                --text-color: #34495e;
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
                background-color: white;
                width: 50%;
                overflow-y: auto;
                border-left: 1px solid var(--primary-color);
            }
            
            h2 {
                color: var(--primary-color);
                border-bottom: 2px solid var(--primary-color);
                padding-bottom: 10px;
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

            fieldset.block {
                border: 1px solid var(--primary-color);
                border-radius: 4px;
                margin-bottom: 15px;
                padding: 10px;
                background-color: #f8f9fa;
            }

            fieldset.block legend {
                color: var(--primary-color);
                font-weight: bold;
                padding: 0 5px;
            }

            network-graph-visualization {
                width: 100%;
                height: 400px;
                margin-top: 20px;
            }
        </style>
        <div class="container">
            <div class="visualization">
                <div class="navigation">
                    <button id="prevBtn">Previous</button>
                    <span id="iteration">Iteration: ${this.iteration + 1} of ${this.maxIterations || 0}</span>
                    <button id="nextBtn">Next</button>
                </div>
                <div class="confidence-graph">
                    <canvas id="confidenceChart"></canvas>
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