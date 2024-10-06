import * as d3 from 'd3';

interface Node {
    id: number;
    x: number;
    y: number;
    radius: number;
    category: string;
    details: string;
    elevation: number;
    targetElevation: number;
    vx: number;
    vy: number;
    creationTime: number;
    hidden: boolean;
}

interface Edge {
    source: Node;
    target: Node;
    creationTime: number;
}

interface RingSegment {
    category: string;
    start: number;
    end: number;
    current: number;
}

class LookingGlassLens extends HTMLElement {
    private template: HTMLTemplateElement;
    private canvas: HTMLCanvasElement;
    private ctx: CanvasRenderingContext2D;
    private tooltip: HTMLDivElement;
    private narrationBox: HTMLDivElement;
    private startButton: HTMLButtonElement;
    private nextStepButton: HTMLButtonElement;
    private prevStepButton: HTMLButtonElement;
    private searchInput: HTMLInputElement;
    private filterSelect: HTMLSelectElement;
    private timeSlider: HTMLInputElement;
    private timeDisplay: HTMLSpanElement;
    private analyticsPanel: HTMLDivElement;
    private width: number;
    private height: number;
    private nodes: Node[];
    private edges: Edge[];
    private categories: string[];
    private lens: { x: number; y: number; radius: number; snapping: boolean };
    private ringSegments: RingSegment[];
    private lastTime: number;
    private scale: number;
    private offsetX: number;
    private offsetY: number;
    private isPanning: boolean;
    private draggingNode: Node | null;
    private storySteps: any[];
    private currentStoryStep: number;
    private highlightedEdges: Edge[];
    private timeSteps: number;
    private currentTimeStep: number;
    private simulation: d3.Simulation<Node, Edge>;
    private isPlaying: boolean;

    constructor() {
        super();
        this.attachShadow({ mode: "open" });
        this.template = document.createElement("template");
        this.template.innerHTML = `
            <style>
                :host {
                    display: block;
                    width: 100%;
                    height: 100%;
                }

                canvas {
                    display: block;
                }

                #tooltip,
                #narrationBox,
                #controlPanel {
                    position: absolute;
                    background: rgba(0, 0, 0, 0.7);
                    color: #fff;
                    border-radius: 5px;
                    font-size: 12px;
                    z-index: 1000;
                }

                #tooltip {
                    padding: 5px 10px;
                    pointer-events: none;
                    display: none;
                }

                #narrationBox {
                    bottom: 20px;
                    left: 20px;
                    padding: 10px;
                    max-width: 300px;
                    display: none;
                }

                #controlPanel {
                    top: 20px;
                    right: 20px;
                    padding: 10px;
                    width: 220px;
                }

                #controlPanel>* {
                    width: 100%;
                    margin-bottom: 10px;
                }

                #analyticsPanel {
                    position: absolute;
                    top: 20px;
                    left: 20px;
                    background: rgba(0, 0, 0, 0.7);
                    color: #fff;
                    padding: 10px;
                    border-radius: 5px;
                    font-size: 12px;
                    max-width: 300px;
                }
            </style>
            <canvas id="networkCanvas"></canvas>
            <div id="tooltip"></div>
            <div id="narrationBox"></div>
            <div id="controlPanel">
                <button id="startButton">Start Story Mode</button>
                <button id="nextStepButton" style="display: none;">Next</button>
                <button id="prevStepButton" style="display: none;">Back</button>
                <input type="text" id="searchInput" placeholder="Search nodes...">
                <select id="filterSelect">
                    <option value="all">All Categories</option>
                </select>
                <input type="range" id="timeSlider" min="0" max="100" value="100">
                <span id="timeDisplay">Current Time: 100%</span>
            </div>
            <div id="analyticsPanel"></div>
        `;

        this.width = window.innerWidth;
        this.height = window.innerHeight;
        this.nodes = [];
        this.edges = [];
        this.categories = [];
        this.lens = { x: this.width / 2, y: this.height / 2, radius: 150, snapping: false };
        this.ringSegments = [];
        this.lastTime = 0;
        this.scale = 1;
        this.offsetX = 0;
        this.offsetY = 0;
        this.isPanning = false;
        this.draggingNode = null;
        this.storySteps = [];
        this.currentStoryStep = 0;
        this.highlightedEdges = [];
        this.timeSteps = 100;
        this.currentTimeStep = this.timeSteps;
        this.simulation = d3.forceSimulation<Node, Edge>();
        this.isPlaying = false;

        // Bind methods that will be used as event handlers
        this.onMouseMove = this.onMouseMove.bind(this);
        this.onWheel = this.onWheel.bind(this);
        this.onMouseDown = this.onMouseDown.bind(this);
        this.onMouseUp = this.onMouseUp.bind(this);
        this.onNodeHover = this.onNodeHover.bind(this);
        this.onDoubleClick = this.onDoubleClick.bind(this);
        this.onSearch = this.onSearch.bind(this);
        this.onFilter = this.onFilter.bind(this);
        this.onTimeChange = this.onTimeChange.bind(this);
    }

    connectedCallback() {
        this.shadowRoot!.appendChild(this.template.content.cloneNode(true));
        this.canvas = this.shadowRoot!.querySelector('#networkCanvas')!;
        this.ctx = this.canvas.getContext('2d')!;
        this.tooltip = this.shadowRoot!.querySelector('#tooltip')!;
        this.narrationBox = this.shadowRoot!.querySelector('#narrationBox')!;
        this.startButton = this.shadowRoot!.querySelector('#startButton')!;
        this.nextStepButton = this.shadowRoot!.querySelector('#nextStepButton')!;
        this.prevStepButton = this.shadowRoot!.querySelector('#prevStepButton')!;
        this.searchInput = this.shadowRoot!.querySelector('#searchInput')!;
        this.filterSelect = this.shadowRoot!.querySelector('#filterSelect')!;
        this.timeSlider = this.shadowRoot!.querySelector('#timeSlider')!;
        this.timeDisplay = this.shadowRoot!.querySelector('#timeDisplay')!;
        this.analyticsPanel = this.shadowRoot!.querySelector('#analyticsPanel')!;

        this.init();
    }

    init() {
        this.canvas.width = this.width;
        this.canvas.height = this.height;

        this.categories = ['Workstation', 'Server', 'Router', 'Database'];
        this.nodes = Array.from({ length: 100 }, (_, i) => ({
            id: i,
            x: Math.random() * this.width,
            y: Math.random() * this.height,
            radius: 3 + Math.random() * 3,
            category: this.categories[Math.floor(Math.random() * this.categories.length)],
            details: "IP: " + Array.from({ length: 4 }, () => Math.floor(Math.random() * 256)).join('.'),
            elevation: 0,
            targetElevation: 0,
            vx: 0,
            vy: 0,
            creationTime: Math.floor(Math.random() * this.timeSteps),
            hidden: false
        }));

        this.edges = [];
        for (let i = 0; i < 200; i++) {
            const sourceIndex = Math.floor(Math.random() * this.nodes.length);
            const targetIndex = Math.floor(Math.random() * this.nodes.length);
            if (sourceIndex !== targetIndex) {
                this.edges.push({
                    source: this.nodes[sourceIndex],
                    target: this.nodes[targetIndex],
                    creationTime: Math.floor(Math.random() * this.timeSteps)
                });
            }
        }

        this.categories.forEach(cat => {
            this.ringSegments.push({ category: cat, start: 0, end: 0, current: 0 });
            this.filterSelect.innerHTML += `<option value="${cat}">${cat}</option>`;
        });

        this.canvas.addEventListener('mousemove', this.onMouseMove);
        this.canvas.addEventListener('wheel', this.onWheel);
        this.canvas.addEventListener('mousedown', this.onMouseDown);
        this.canvas.addEventListener('mouseup', this.onMouseUp);
        this.canvas.addEventListener('mousemove', this.onNodeHover);
        this.canvas.addEventListener('mouseleave', () => { this.tooltip.style.display = 'none'; });
        this.canvas.addEventListener('mouseout', () => { this.isPanning = false; });
        this.canvas.addEventListener('dblclick', this.onDoubleClick);

        this.startButton.addEventListener('click', () => this.startStoryMode());
        this.nextStepButton.addEventListener('click', () => this.nextStoryStep());
        this.prevStepButton.addEventListener('click', () => this.prevStoryStep());
        this.searchInput.addEventListener('input', this.onSearch);
        this.filterSelect.addEventListener('change', this.onFilter);
        this.timeSlider.addEventListener('input', this.onTimeChange);

        this.createStorySteps();
        this.updateAnalytics();
        this.setupSimulation();
        requestAnimationFrame(this.animate.bind(this));
    }

    setupSimulation() {
        this.simulation = d3.forceSimulation(this.nodes)
            .force('charge', d3.forceManyBody().strength(-30))
            .force('center', d3.forceCenter(this.width / 2, this.height / 2))
            .force('collision', d3.forceCollide().radius(d => d.radius * 2))
            .force('link', d3.forceLink(this.edges).id((d: any) => d.id).distance(50).strength(0.1))
            .on('tick', () => {
                this.ctx.clearRect(0, 0, this.width, this.height);
                this.drawEdges();
                this.nodes.forEach(node => this.drawNode(node));
            });
    }

    onMouseMove(event: MouseEvent) {
        const rect = this.canvas.getBoundingClientRect();
        const mouseX = (event.clientX - rect.left - this.offsetX) / this.scale;
        const mouseY = (event.clientY - rect.top - this.offsetY) / this.scale;

        if (this.draggingNode) {
            this.draggingNode.x = mouseX;
            this.draggingNode.y = mouseY;
        } else if (this.isPanning) {
            this.offsetX += event.movementX;
            this.offsetY += event.movementY;
        }
    }

    onWheel(event: WheelEvent) {
        event.preventDefault();
        if (event.ctrlKey) {
            const zoomIntensity = 0.1;
            const wheel = event.deltaY < 0 ? 1 : -1;
            const zoom = Math.exp(wheel * zoomIntensity);

            this.scale *= zoom;

            const rect = this.canvas.getBoundingClientRect();
            const mouseX = (event.clientX - rect.left - this.offsetX) / this.scale;
            const mouseY = (event.clientY - rect.top - this.offsetY) / this.scale;

            this.offsetX -= (mouseX * zoom - mouseX) * this.scale;
            this.offsetY -= (mouseY * zoom - mouseY) * this.scale;
        } else {
            this.lens.radius = Math.max(50, Math.min(300, this.lens.radius - event.deltaY * 0.5));
        }
    }

    onMouseDown(event: MouseEvent) {
        const rect = this.canvas.getBoundingClientRect();
        const mouseX = (event.clientX - rect.left - this.offsetX) / this.scale;
        const mouseY = (event.clientY - rect.top - this.offsetY) / this.scale;

        for (let i = 0; i < this.nodes.length; i++) {
            const node = this.nodes[i];
            const dx = mouseX - node.x;
            const dy = mouseY - node.y;
            const distance = Math.sqrt(dx * dx + dy * dy);

            if (distance < node.radius * 2) {
                this.draggingNode = node;
                break;
            }
        }

        if (!this.draggingNode && (event.button === 1 || event.button === 0)) {
            this.isPanning = true;
        }
    }

    onMouseUp() {
        this.isPanning = false;
        this.draggingNode = null;
    }

    onDoubleClick(event: MouseEvent) {
        const rect = this.canvas.getBoundingClientRect();
        const mouseX = (event.clientX - rect.left - this.offsetX) / this.scale;
        const mouseY = (event.clientY - rect.top - this.offsetY) / this.scale;

        this.nodes.forEach(node => {
            const dx = node.x - mouseX;
            const dy = node.y - mouseY;
            const distance = Math.sqrt(dx * dx + dy * dy);
            if (distance < node.radius * 2) {
                this.highlightNodeEdges(node);
            }
        });
    }

    highlightNodeEdges(node: Node) {
        this.highlightedEdges = this.edges.filter(edge => edge.source === node || edge.target === node);
        setTimeout(() => { this.highlightedEdges = []; }, 3000);
    }

    drawNode(node: Node) {
        if (this.ctx) {
            this.ctx.beginPath();
            this.ctx.arc(node.x, node.y, node.radius, 0, Math.PI * 2);
            this.ctx.fillStyle = this.getCategoryColor(node.category);
            this.ctx.fill();
        }
    }

    drawEdges() {
        if (this.ctx) {
            this.edges.forEach(edge => {
                this.ctx.beginPath();
                this.ctx.moveTo(edge.source.x, edge.source.y);
                this.ctx.lineTo(edge.target.x, edge.target.y);
                this.ctx.strokeStyle = 'rgba(200, 200, 200, 0.1)';
                this.ctx.lineWidth = 0.5;
                this.ctx.stroke();
            });
        }
    }

    getCategoryColor(category: string): string {
        const colors: { [key: string]: string } = {
            'Workstation': 'rgba(100, 200, 255, 0.7)',
            'Server': 'rgba(255, 150, 100, 0.7)',
            'Router': 'rgba(100, 255, 150, 0.7)',
            'Database': 'rgba(255, 255, 100, 0.7)'
        };
        return colors[category] || 'rgba(200, 200, 200, 0.7)';
    }

    onSearch() {
        const searchTerm = this.searchInput.value.toLowerCase();
        this.nodes.forEach(node => {
            node.hidden = !node.details.toLowerCase().includes(searchTerm) &&
                !node.category.toLowerCase().includes(searchTerm);
        });
    }

    onFilter() {
        const selectedCategory = this.filterSelect.value;
        this.nodes.forEach(node => {
            node.hidden = selectedCategory !== 'all' && node.category !== selectedCategory;
        });
    }

    onTimeChange() {
        this.currentTimeStep = parseInt(this.timeSlider.value);
        this.timeDisplay.textContent = `Current Time: ${this.currentTimeStep}%`;
        this.updateAnalytics();
    }

    updateAnalytics() {
        const visibleNodes = this.nodes.filter(n => n.creationTime <= this.currentTimeStep && !n.hidden);
        const visibleEdges = this.edges.filter(e => e.creationTime <= this.currentTimeStep && !e.source.hidden && !e.target.hidden);

        const degreeDistribution: { [key: number]: number } = {};
        visibleNodes.forEach(node => {
            const degree = visibleEdges.filter(e => e.source === node || e.target === node).length;
            degreeDistribution[degree] = (degreeDistribution[degree] || 0) + 1;
        });

        const averageDegree = visibleEdges.length * 2 / visibleNodes.length;
        const maxDegree = Math.max(...Object.keys(degreeDistribution).map(Number));

        this.analyticsPanel.innerHTML = `
            <h3>Network Analytics</h3>
            <p>Nodes: ${visibleNodes.length}</p>
            <p>Edges: ${visibleEdges.length}</p>
            <p>Avg Degree: ${averageDegree.toFixed(2)}</p>
            <p>Max Degree: ${maxDegree}</p>
        `;
    }

    animate(time: number) {
        const deltaTime = time - this.lastTime;
        this.lastTime = time;

        this.ctx.clearRect(0, 0, this.width, this.height);

        this.simulation.alpha(0.1).restart();
        this.drawEdges();

        this.nodes.forEach(node => {
            if (node.hidden || node.creationTime > this.currentTimeStep) return;
            this.drawNode(node);
        });

        requestAnimationFrame(this.animate.bind(this));
    }

    createStorySteps() {
        this.storySteps = [
            {
                targetNode: this.nodes[0],
                narration: "Welcome to the network visualization. Let's start by focusing on this central node.",
                action: () => {
                    this.moveLensToTarget(this.storySteps[0].targetNode);
                }
            },
            {
                targetNode: this.nodes[1],
                narration: "Notice how the nodes gather under the lens, and the pie chart updates accordingly.",
                action: () => {
                    this.moveLensToTarget(this.storySteps[1].targetNode);
                }
            }
        ];
    }

    moveLensToTarget(node: Node, duration: number = 2000) {
        const startX = this.lens.x;
        const startY = this.lens.y;
        const targetX = node.x;
        const targetY = node.y;
        const startTime = performance.now();

        const animateMove = (time: number) => {
            const elapsed = time - startTime;
            const progress = Math.min(elapsed / duration, 1);
            const easeProgress = this.easeInOutCubic(progress);

            this.lens.x = startX + (targetX - startX) * easeProgress;
            this.lens.y = startY + (targetY - startY) * easeProgress;

            if (progress < 1) {
                requestAnimationFrame(animateMove);
            }
        }

        requestAnimationFrame(animateMove);
    }

    easeInOutCubic(t: number): number {
        return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
    }

    startStoryMode() {
        this.startButton.style.display = 'none';
        this.nextStepButton.style.display = 'inline';
        this.narrationBox.style.display = 'block';
        this.currentStoryStep = 0;
        this.nextStoryStep();
    }

    nextStoryStep() {
        if (this.currentStoryStep < this.storySteps.length) {
            const step = this.storySteps[this.currentStoryStep];
            step.action();
            this.updateNarration(step.narration);
            this.currentStoryStep++;
            if (this.currentStoryStep > 0) {
                this.prevStepButton.style.display = 'inline';
            }
            if (this.currentStoryStep === this.storySteps.length) {
                this.nextStepButton.style.display = 'none';
            }
        }
    }

    prevStoryStep() {
        if (this.currentStoryStep > 1) {
            this.currentStoryStep -= 2;
            this.nextStoryStep();
        } else {
            this.prevStepButton.style.display = 'none';
        }
    }

    updateNarration(text: string) {
        this.narrationBox.innerHTML = text;
    }
}

customElements.define("looking-glass-lens", LookingGlassLens);