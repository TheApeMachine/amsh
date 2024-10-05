import { onMount, onUnmount } from "@/lib/lifecycle";
import { getWorker } from "@/workers/utils";
import { gsap } from "gsap";
import { Draggable } from "gsap/Draggable";

gsap.registerPlugin(Draggable);

class DashboardEditor extends HTMLElement {
    private template = document.createElement("template");
    private shadow: ShadowRoot;
    private cellSize: number = 20; // Default cell size
    private draggables: HTMLElement[] = [];
    private worker = getWorker();

    constructor() {
        super();
        this.template.innerHTML = `
            <style>
                :host {
                    display: flex;
                    width: 100%;
                    height: 100%;
                }
                .dashboard {
                    display: flex;
                    width: 100%;
                    height: 100%;
                    position: relative;
                }
                .grid {
                    flex-grow: 1;
                    position: relative;
                    background: #ddd;
                    padding: 1px;
                    display: grid;
                    grid-template-columns: repeat(auto-fill, minmax(min(${this.cellSize}px, 100%), 1fr));
                    grid-template-rows: repeat(auto-fill, minmax(min(${this.cellSize}px, 100%), 1fr));
                    gap: 1px;
                    --c: rgba(0, 0, 0, 0.1);
                    --cell-size: ${this.cellSize}px;
                    background-image:
                        linear-gradient(to right, var(--c) 1px, transparent 1px),
                        linear-gradient(to bottom, var(--c) 1px, transparent 1px);
                    background-size: var(--cell-size) var(--cell-size);
                }
                .draggable {
                    position: absolute;
                    background: #EEE;
                    border: 1px solid #AAA;
                    z-index: 10;
                    cursor: move;
                }
                .resize-handle {
                    position: absolute;
                    bottom: 0;
                    right: 0;
                    width: 0px;
                    height: 0px;
                    margin-top: -17px;
                    margin-left: -17px;
                    border-style: solid;
                    border-width: 0 0 16px 16px;
                    border-color: transparent transparent #AAA transparent;
                    cursor: nw-resize !important;
                }
                .empty-message {
                    position: absolute;
                    top: 50%;
                    left: 50%;
                    transform: translate(-50%, -50%);
                    text-align: center;
                    color: #666;
                }
                .empty-message button {
                    margin-top: 10px;
                    padding: 10px 20px;
                    background: #007bff;
                    color: #fff;
                    border: none;
                    cursor: pointer;
                }
                .empty-message button:hover {
                    background: #0056b3;
                }
            </style>
            <div class="dashboard">
                <div class="grid" id="grid"></div>
                <div class="empty-message" id="emptyMessage">
                    <p>No blocks yet. Click the button below to add your first block.</p>
                    <button id="addFirstBlock" data-event="click" data-topic="state" data-effect="add-block">Add First Block</button>
                </div>
            </div>
        `;

        this.shadow = this.attachShadow({ mode: 'open' });
        this.shadow.appendChild(this.template.content.cloneNode(true));
    }

    connectedCallback() {
        onMount(this, () => {
            window.state.register('add-block', this.addBlockToGrid.bind(this));
            this.renderGrid();    
            this.updateEmptyMessage();
        });
    }

    disconnectedCallback() {
        onUnmount(this, () => {
            console.log('DashboardEditor unmounted');
            // Perform any cleanup here
        });
    }

    private renderGrid() {
        const gridElement = this.shadow.getElementById('grid');
        if (!gridElement) return;

        gridElement.style.setProperty('--cell-size', `${this.cellSize}px`);

        // Clear existing grid cells
        gridElement.innerHTML = '';

        // We don't need to create grid cells manually anymore
        // The grid will be created automatically by the CSS grid properties
    }

    addBlockToGrid() {
        const gridElement = this.shadow.getElementById('grid');
        if (!gridElement) return;

        const block = document.createElement('div');
        block.classList.add('draggable');
        block.style.width = `${this.cellSize * 10}px`;
        block.style.height = `${this.cellSize * 10}px`;
        block.style.left = `${this.cellSize}px`;
        block.style.top = `${this.cellSize}px`;
        block.setAttribute('data-id', this.draggables.length.toString());

        const chart = document.createElement('bar-chart');
        chart.style.width = `${this.cellSize * 10}px`;
        chart.style.height = `${this.cellSize * 10}px`;
        chart.style.left = `${this.cellSize}px`;
        chart.style.top = `${this.cellSize}px`;
        chart.setAttribute('data-id', this.draggables.length.toString());

        const handle = document.createElement('div');
        handle.classList.add('resize-handle');

        block.appendChild(handle);
        gridElement.appendChild(block);

        this.draggables.push(block);
        this.initializeDraggable(block);
        this.updateEmptyMessage();
    }

    private initializeDraggable(element: HTMLElement) {
        const handle = element.querySelector('.resize-handle') as HTMLElement;

        Draggable.create(element, {
            type: "top,left",
            bounds: this.shadow.getElementById('grid') as HTMLElement,
            onDragEnd: () => {
                // Snap to grid on drag end
                gsap.set(element, {
                    top: Math.round(element.offsetTop / this.cellSize) * this.cellSize,
                    left: Math.round(element.offsetLeft / this.cellSize) * this.cellSize
                });
            }
        });

        const cellSize = this.cellSize; // Capture cellSize from the class instance

        Draggable.create(handle, {
            type: "top,left",
            onPress: (e: Event) => {
                e.stopPropagation(); // cancel drag
            },
            onDrag: function(this: Draggable) {
                // Use the captured cellSize value
                gsap.set(this.target.parentNode, { 
                    width: this.x,
                    height: this.y
                });
            },
            liveSnap: function(endValue: number) {
                return Math.round(endValue / cellSize) * cellSize;
            }
        });
    }

    private updateEmptyMessage() {
        const emptyMessage = this.shadow.getElementById('emptyMessage');
        if (emptyMessage) {
            if (this.draggables.length === 0) {
                emptyMessage.style.display = 'block';
            } else {
                emptyMessage.style.display = 'none';
            }
        }
    }
}

customElements.define('dashboard-editor', DashboardEditor);