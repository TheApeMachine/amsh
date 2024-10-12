import * as THREE from 'three';
import { Layer } from './Layer';
import { Connection } from './Connection';
import { TextSprite } from './TextSprite';
import * as CANNON from 'cannon';
import { Node as ThreeNode } from 'three';

export class AMSHNode {
    mesh: THREE.Mesh;
    children: AMSHNode[];
    layers: Layer[];
    connections: Connection[];
    isExpanded: boolean;
    canExpand: boolean;
    parent: AMSHNode | null;
    depth: number;
    textSprite: TextSprite;
    body!: CANNON.Body;
    lod: THREE.LOD;

    constructor(color: number = 0x4a148c, canExpand: boolean = true, parent: AMSHNode | null = null, depth: number = 0) {
        const geometry = new THREE.BoxGeometry(0.8, 1, 0.8);
        const material = new THREE.MeshStandardMaterial({
            color: color,
            metalness: 0.3,
            roughness: 0.7,
            emissive: new THREE.Color(color).multiplyScalar(0.2)
        });
        this.lod = new THREE.LOD();
        // High-detail mesh
        const highDetailMesh = new THREE.Mesh(geometry, material);
        this.lod.addLevel(highDetailMesh, 0);
        // Low-detail mesh
        const lowDetailMesh = new THREE.Mesh(geometry, material);
        this.lod.addLevel(lowDetailMesh, 50);
        this.mesh = new THREE.Mesh(geometry, material);;
        this.children = [];
        this.layers = [];
        this.connections = [];
        this.isExpanded = false;
        this.canExpand = canExpand;
        this.parent = parent;
        this.depth = depth;
        this.textSprite = new TextSprite('Node');
        this.mesh.add(this.textSprite.sprite);

        this.createPhysicsBody();
        this.setupAnimation();
        this.updateColor();
    }

    private createPhysicsBody() {
        const shape = new CANNON.Box(new CANNON.Vec3(0.4, 0.5, 0.4));
        this.body = new CANNON.Body({ mass: 1, shape });
        this.body.position.copy(this.mesh.position as unknown as CANNON.Vec3);
    }

    public updatePhysics() {
        this.mesh.position.copy(this.body.position as unknown as THREE.Vector3);
        this.mesh.quaternion.copy(this.body.quaternion as unknown as THREE.Quaternion);
    }

    handleClick() {
        if (this.canExpand) {
            this.toggleExpansion();
            this.updateColor();
            // Recursively update colors of all children
            this.children.forEach(child => child.updateColor());
        }
    }

    public findPathTo(target: AMSHNode): AMSHNode[] {
        // Implement A* or Dijkstra's algorithm to find the shortest path
        return [];
    }    

    setupAnimation() {
        const pulseAnimation = new THREE.AnimationClip('pulse', 1, [
            new THREE.KeyframeTrack('.scale', [0, 0.5, 1], [
                1, 1, 1,
                1.05, 1.05, 1.05,
                1, 1, 1
            ])
        ]);
        const mixer = new THREE.AnimationMixer(this.mesh);
        const action = mixer.clipAction(pulseAnimation);
        action.play();

        this.mesh.userData.update = (delta: number) => {
            mixer.update(delta);
        };
    }

    expand() {
        if (this.isExpanded || !this.canExpand) return;
        this.isExpanded = true;

        // Create child nodes
        for (let i = 0; i < 4; i++) {
            const childNode = new AMSHNode(0x7c4dff, Math.random() > 0.3, this, this.depth + 1);
            childNode.mesh.scale.set(0.5, 0.5, 0.5);
            childNode.mesh.position.set(
                (i % 2 - 0.5) * 0.6,
                0.75,
                (Math.floor(i / 2) - 0.5) * 0.6
            );
            this.mesh.add(childNode.mesh);
            this.children.push(childNode);
        }

        // Create layers
        for (let i = 0; i < 3; i++) {
            const layer = new Layer(i, 5, `Subsystem ${this.depth}-${i}`);
            layer.group.position.set(0, (i + 1) * 1.5, 0);
            layer.group.scale.set(0.3, 0.3, 0.3);
            this.mesh.add(layer.group);
            this.layers.push(layer);
        }

        this.createConnections();
        this.updateColor();
    }

    collapse() {
        if (!this.isExpanded) return;
        this.isExpanded = false;
        this.children.forEach(child => {
            child.collapse();
            this.mesh.remove(child.mesh);
        });
        this.layers.forEach(layer => {
            this.mesh.remove(layer.group);
        });
        this.children = [];
        this.layers = [];
        this.connections = [];
        this.updateColor();
    }

    createConnections() {
        // Connect child nodes
        for (let i = 0; i < this.children.length; i++) {
            for (let j = i + 1; j < this.children.length; j++) {
                if (Math.random() < 0.5) {
                    const connection = new Connection(this.children[i], this.children[j]);
                    this.connections.push(connection);
                    this.mesh.add(connection.line);
                }
            }
        }

        // Connect layers
        this.layers.forEach(layer => {
            layer.nodes.forEach(node => {
                if (Math.random() < 0.3) {
                    const targetLayer = this.layers[Math.floor(Math.random() * this.layers.length)];
                    const targetNode = targetLayer.nodes[Math.floor(Math.random() * targetLayer.nodes.length)];
                    const connection = new Connection(node, targetNode);
                    this.connections.push(connection);
                    this.mesh.add(connection.line);
                }
            });
        });
    }

    updateColor() {
        const material = this.mesh.material as THREE.MeshStandardMaterial;
        if (this.isExpanded) {
            material.color.setHex(0x7c4dff);  // Lighter purple for expanded nodes
            material.emissive.setHex(0x3c1d99);
        } else if (this.canExpand) {
            material.color.setHex(0x4a148c);  // Dark purple for expandable nodes
            material.emissive.setHex(0x250a46);
        } else {
            material.color.setHex(0x9e9e9e);  // Grey for non-expandable nodes
            material.emissive.setHex(0x4f4f4f);
        }
    
        // Update children colors recursively
        this.children.forEach(child => child.updateColor());
    
        // Update colors of nodes in layers
        this.layers.forEach(layer => layer.nodes.forEach(node => node.updateColor()));
    }    

    updateConnections() {
        this.connections.forEach(connection => connection.update());
        this.children.forEach(child => child.updateConnections());
        this.layers.forEach(layer => layer.updateConnections());
    }

    toggleExpansion() {
        if (this.isExpanded) {
            this.collapse();
        } else {
            this.expand();
        }
    }

    addConnection(target: AMSHNode): Connection {
        const connection = new Connection(this, target);
        this.connections.push(connection);
        return connection;
    }

    updateText(text: string) {
        this.textSprite.updateText(text);
    }
}
