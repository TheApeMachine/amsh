import * as THREE from 'three';
import { AMSHNode } from './Node';
import { Connection } from './Connection';

export class Layer {
    group: THREE.Group;
    nodes: AMSHNode[];
    connections: Connection[];

    constructor(index: number, size: number, subsystemType: string) {
        this.group = new THREE.Group();
        this.nodes = [];
        this.connections = [];

        const platformGeometry = new THREE.PlaneGeometry(size, size);
        const platformMaterial = new THREE.MeshBasicMaterial({ 
            color: this.getSubsystemColor(subsystemType), 
            transparent: true, 
            opacity: 0.7 
        });
        const platform = new THREE.Mesh(platformGeometry, platformMaterial);
        platform.rotation.x = -Math.PI / 2;
        platform.position.y = -index * 2;

        this.group.add(platform);

        this.createNodes(size, index);
        this.createConnections();
    }

    private createNodes(size: number, index: number) {
        for (let x = -size/2 + 1; x <= size/2 - 1; x += 2) {
            for (let z = -size/2 + 1; z <= size/2 - 1; z += 2) {
                const node = new AMSHNode();
                node.mesh.position.set(x, -index * 2 + 0.5, z);
                this.group.add(node.mesh);
                this.nodes.push(node);
            }
        }
    }

    private createConnections() {
        for (let i = 0; i < this.nodes.length; i++) {
            for (let j = i + 1; j < this.nodes.length; j++) {
                if (Math.random() < 0.2) {  // 20% chance of connection
                    const connection = this.nodes[i].addConnection(this.nodes[j]);
                    this.group.add(connection.line);
                    this.connections.push(connection);
                }
            }
        }
    }

    private getSubsystemColor(subsystemType: string): number {
        switch(subsystemType) {
            case 'processing': return 0x1a237e;
            case 'memory': return 0x004d40;
            case 'communication': return 0xb71c1c;
            default: return 0x1a237e;
        }
    }

    updateConnections() {
        this.connections.forEach(connection => connection.update());
    }
}