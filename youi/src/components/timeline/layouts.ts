import * as THREE from 'three';

interface Agent {
    id: string;
    position: THREE.Vector3;
    connections: string[];
}

export class Layout3D {
    private agents: Agent[];

    constructor(agents: Agent[]) {
        this.agents = agents;
    }

    private getRandomPosition(scale: number = 100): THREE.Vector3 {
        return new THREE.Vector3(
            (Math.random() - 0.5) * scale,
            (Math.random() - 0.5) * scale,
            (Math.random() - 0.5) * scale
        );
    }

    private validatePositions(): void {
        this.agents.forEach(agent => {
            if (!agent.position || 
                isNaN(agent.position.x) || 
                isNaN(agent.position.y) || 
                isNaN(agent.position.z)) {
                agent.position = this.getRandomPosition();
            }
        });
    }

    public applyForceDirected(iterations: number = 50, springLength: number = 30, springStrength: number = 0.1, repulsionStrength: number = 300): Agent[] {
        for (let i = 0; i < iterations; i++) {
            this.agents.forEach(agent => {
                let force = new THREE.Vector3(0, 0, 0);

                // Apply spring forces
                agent.connections.forEach(connectedId => {
                    const connected = this.agents.find(a => a.id === connectedId);
                    if (connected) {
                        const diff = connected.position.clone().sub(agent.position);
                        const displacement = diff.length() - springLength;
                        force.add(diff.normalize().multiplyScalar(springStrength * displacement));
                    }
                });

                // Apply repulsion forces
                this.agents.forEach(other => {
                    if (other !== agent) {
                        const diff = agent.position.clone().sub(other.position);
                        const distance = diff.length() || 0.1; // Avoid division by zero
                        force.add(diff.normalize().multiplyScalar(repulsionStrength / (distance * distance)));
                    }
                });

                // Update position
                agent.position.add(force);
            });
        }

        this.validatePositions();
        return this.agents;
    }

    public applySphereLayout(radius: number = 50): Agent[] {
        const phi = Math.PI * (3 - Math.sqrt(5)); // Golden angle in radians

        this.agents.forEach((agent, i) => {
            const y = 1 - (i / (this.agents.length - 1)) * 2;
            const radiusAtY = Math.sqrt(1 - y * y) * radius;

            const theta = phi * i;

            const x = Math.cos(theta) * radiusAtY;
            const z = Math.sin(theta) * radiusAtY;

            agent.position.set(x, y * radius, z);
        });

        this.validatePositions();
        return this.agents;
    }

    public applyGridLayout(gridSize: number = 5): Agent[] {
        const spacing = 20;
        const offset = (gridSize - 1) * spacing / 2;

        this.agents.forEach((agent, i) => {
            const x = (i % gridSize) * spacing - offset;
            const y = Math.floor((i / gridSize) % gridSize) * spacing - offset;
            const z = Math.floor(i / (gridSize * gridSize)) * spacing - offset;

            agent.position.set(x, y, z);
        });

        this.validatePositions();
        return this.agents;
    }

    public applyRandomLayout(scale: number = 100): Agent[] {
        this.agents.forEach(agent => {
            agent.position.copy(this.getRandomPosition(scale));
        });

        this.validatePositions();
        return this.agents;
    }

    public applyCircularLayout(radius: number = 50, height: number = 20): Agent[] {
        const angleStep = (2 * Math.PI) / this.agents.length;

        this.agents.forEach((agent, i) => {
            const angle = i * angleStep;
            const x = Math.cos(angle) * radius;
            const z = Math.sin(angle) * radius;
            const y = (i / this.agents.length) * height - height / 2;

            agent.position.set(x, y, z);
        });

        this.validatePositions();
        return this.agents;
    }
}