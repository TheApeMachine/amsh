import * as THREE from 'three';
import * as CANNON from 'cannon';
import { IsometricScene } from './IsometricScene';
import { Layer } from './Layer';
import { NavigationControls } from './NavigationControls';
import { DataFlow } from './DataFlow';
import { AMSHNode } from './Node';
import { Connection } from './Connection';


export class ApeMachineShell extends HTMLElement {
    private scene!: IsometricScene;
    private physicsWorld!: CANNON.World;
    private layers!: Layer[];
    private navigationControls!: NavigationControls;
    private dataFlow!: DataFlow;
    private interLayerConnections: Connection[] = [];
    private allNodes: AMSHNode[] = [];
    private infoPanel!: HTMLElement;
    private raycaster: THREE.Raycaster;
    private mouse: THREE.Vector2;
    private rootNode: AMSHNode = new AMSHNode(0x4a148c, true, null, 0);
    private lastTime: number = 0;
    private selectedNode: AMSHNode | null = null;

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.raycaster = new THREE.Raycaster();
        this.mouse = new THREE.Vector2();
    }

    connectedCallback() {
        this.render();
        this.initializeScene();
        this.initializeRootNode();
        this.initializeLayers();
        this.initializeNavigation();
        this.initializeDataFlow();
        this.initializeInfoPanel();
        this.initializeSearch();
        this.initializeClickHandler();
        this.initializePhysics();  // Add this line
        this.updateFrame(performance.now());
    }

    private render() {
        this.shadowRoot!.innerHTML = `
            <style>
                :host {
                    display: block;
                    width: 100%;
                    height: 100%;
                }
                #scene-container {
                    width: 100%;
                    height: 100%;
                }
                #info-panel {
                    position: absolute;
                    top: 10px;
                    right: 10px;
                    width: 300px;
                    height: 400px;
                    background-color: rgba(0, 0, 0, 0.7);
                    color: white;
                    padding: 10px;
                    overflow-y: auto;
                    font-family: Arial, sans-serif;
                }
                #search-container {
                    position: absolute;
                    top: 10px;
                    left: 10px;
                }
                #search-input {
                    width: 200px;
                    padding: 5px;
                }
            </style>
            <div id="scene-container"></div>
            <div id="info-panel"></div>
            <div id="search-container">
                <input type="text" id="search-input" placeholder="Search nodes...">
            </div>
        `;
    }

    private initializeScene() {
        const container = this.shadowRoot!.getElementById('scene-container') as HTMLElement;
        this.scene = new IsometricScene(container);
    }

    private initializeLayers() {
        this.layers = [];
        const subsystems = ['processing', 'memory', 'communication'];
        for (let i = 0; i < 3; i++) {
            const layer = new Layer(i, 10, subsystems[i]);
            this.scene.scene.add(layer.group);
            this.layers.push(layer);
            this.allNodes.push(...layer.nodes);
        }
        this.createInterLayerConnections();

        this.layers.forEach(layer => {
            layer.nodes.forEach(node => {
                node.mesh.userData.node = node;
            });
        });
    }

    private createInterLayerConnections() {
        for (let i = 0; i < this.layers.length - 1; i++) {
            const sourceLayer = this.layers[i];
            const targetLayer = this.layers[i + 1];
            for (let j = 0; j < sourceLayer.nodes.length; j++) {
                if (Math.random() < 0.1) {  // 10% chance of inter-layer connection
                    const sourceNode = sourceLayer.nodes[j];
                    const targetNode = targetLayer.nodes[Math.floor(Math.random() * targetLayer.nodes.length)];
                    const connection = new Connection(sourceNode, targetNode);
                    this.scene.scene.add(connection.line);
                    this.scene.scene.add(connection.particles);
                    this.interLayerConnections.push(connection);
                }
            }
        }
    }

    private initializeNavigation() {
        this.navigationControls = new NavigationControls(
            this.layers.length,
            (layer: number) => this.focusLayer(layer)
        );
    }

    private initializeDragControls() {
        const container = this.shadowRoot!.getElementById('scene-container') as HTMLElement;
        container.addEventListener('mousedown', (event) => this.onMouseDown(event));
        container.addEventListener('mousemove', (event) => this.onMouseMove(event));
        container.addEventListener('mouseup', () => this.onMouseUp());
    }

    private onMouseDown(event: MouseEvent) {
        // Perform raycasting to select a node
        // Set this.selectedNode if a node is selected
    }
    
    private onMouseMove(event: MouseEvent) {
        if (this.selectedNode) {
            // Update the position of the selected node based on mouse movement
        }
    }
    
    private onMouseUp() {
        this.selectedNode = null;
    }

    private initializeDataFlow() {
        this.dataFlow = new DataFlow(this.scene.scene);
    }

    private initializeInfoPanel() {
        this.infoPanel = this.shadowRoot!.getElementById('info-panel') as HTMLElement;
    }

    private initializeSearch() {
        const searchInput = this.shadowRoot!.getElementById('search-input') as HTMLInputElement;
        searchInput.addEventListener('input', () => this.performSearch(searchInput.value));
    }

    private initializeClickHandler() {
        const container = this.shadowRoot!.getElementById('scene-container') as HTMLElement;
        container.addEventListener('click', (event) => this.onContainerClick(event));
    }

    private updateScene() {
        // Update connections and other elements that might be affected by node expansion
        this.layers.forEach(layer => layer.updateConnections());
        this.interLayerConnections.forEach(connection => connection.update());
    }

    private performSearch(query: string) {
        const lowerCaseQuery = query.toLowerCase();
        const matchingNodes = this.allNodes.filter(node => node.textSprite.sprite.name.toLowerCase().includes(lowerCaseQuery));
    
        if (matchingNodes.length > 0) {
            // Focus on the first matching node
            this.focusOnNode(matchingNodes[0]);
    
            // Optionally, highlight all matching nodes
            matchingNodes.forEach(node => this.scene.highlightNode(node));
        } else {
            // Provide feedback or reset highlights
            console.log('No matching nodes found.');
        }
    }    

    private focusOnNode(node: Node) {
        const nodePosition = new THREE.Vector3();
        node.mesh.getWorldPosition(nodePosition);
        this.scene.camera.position.set(nodePosition.x + 5, nodePosition.y + 5, nodePosition.z + 5);
        this.scene.camera.lookAt(nodePosition);
    }

    private focusLayer(index: number) {
        const layerPosition = new THREE.Vector3(0, -index * 2, 0);
        this.scene.camera.position.y = layerPosition.y + 10;
        this.scene.camera.lookAt(layerPosition);

        this.layers.forEach((layer, i) => {
            layer.setOpacity(i === index ? 1 : 0.3);
        });
    }

    private initializeRootNode() {
        this.rootNode = new AMSHNode(0x4a148c, true, null, 0);
        this.rootNode.expand(); // This will create the initial three layers
        if (this.scene) {
            this.scene.scene.add(this.rootNode.mesh);
        } else {
            console.error('Scene is not initialized');
        }
    }

    private onContainerClick(event: MouseEvent) {
        const container = this.shadowRoot!.getElementById('scene-container') as HTMLElement;
        const rect = container.getBoundingClientRect();
        
        // Calculate mouse position in normalized device coordinates
        this.mouse.x = ((event.clientX - rect.left) / container.clientWidth) * 2 - 1;
        this.mouse.y = -((event.clientY - rect.top) / container.clientHeight) * 2 + 1;

        // Update the picking ray with the camera and mouse position
        this.raycaster.setFromCamera(this.mouse, this.scene.camera);

        // Calculate objects intersecting the picking ray
        const intersects = this.raycaster.intersectObjects(this.scene.scene.children, true);

        for (let intersect of intersects) {
            let object: THREE.Object3D | null = intersect.object;
            while (object) {
                if (object.userData.node instanceof AMSHNode) {
                    object.userData.node.handleClick();
                    this.scene.highlightNode(object.userData.node);
                    this.updateScene();
                    break;
                }
                object = object.parent;
            }
            if (object) break;
        }
    }

    private initializePhysics() {
        this.physicsWorld = new CANNON.World();
        this.physicsWorld.gravity.set(0, -9.82, 0); // Earth's gravity
    
        // Add all node bodies to the physics world
        this.allNodes.forEach(node => this.physicsWorld.addBody(node.body));
    }

    private updateFrame(time: number) {
        requestAnimationFrame((t) => this.updateFrame(t));
        const delta = (time - this.lastTime) / 1000;
        this.lastTime = time;

        if (this.physicsWorld) {
            this.physicsWorld.step(1 / 60, delta, 3);

            // Update node meshes based on physics bodies
            //this.allNodes.forEach(node => node.updatePhysics());
        }

        this.updateNodes(this.rootNode, delta);
        this.scene.controls.update();
        this.dataFlow.animate();
        this.updateInfoPanel();
        this.scene.render();
    }

    private updateNodes(node: AMSHNode, delta: number) {
        if (node.mesh.userData.update) {
            node.mesh.userData.update(delta);
        }
        node.updateConnections();
        node.children.forEach(child => this.updateNodes(child, delta));
        node.layers.forEach(layer => layer.nodes.forEach(n => this.updateNodes(n, delta)));
    }    

    private updateInfoPanel() {
        // Example of updating info panel with random data
        this.infoPanel.innerHTML = `
            <h3>System Status</h3>
            <p>Active Nodes: ${this.allNodes.length}</p>
            <p>Data Processed: ${Math.floor(Math.random() * 1000000)} bytes</p>
            <p>Current Task: ${['Processing', 'Learning', 'Analyzing', 'Idle'][Math.floor(Math.random() * 4)]}</p>
            <h4>Recent Activity</h4>
            <ul>
                ${Array(5).fill(0).map(() => `<li>${this.generateRandomActivity()}</li>`).join('')}
            </ul>
        `;
    }

    private generateRandomActivity(): string {
        const activities = [
            "Processed natural language query",
            "Updated knowledge base",
            "Optimized neural network",
            "Generated response to user input",
            "Performed data analysis on incoming stream"
        ];
        return activities[Math.floor(Math.random() * activities.length)];
    }
}

customElements.define('ape-machine-shell', ApeMachineShell);
