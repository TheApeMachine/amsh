// DataFlow.ts
import * as THREE from 'three';
import { GPUComputationRenderer } from 'three/examples/jsm/misc/GPUComputationRenderer';


export class DataFlow {
    private particles: THREE.Points;

    constructor(scene: THREE.Scene) {
        const particleGeometry = new THREE.BufferGeometry();
        const particleCount = 1000;

        const positions = new Float32Array(particleCount * 3);
        for (let i = 0; i < particleCount * 3; i++) {
            positions[i] = (Math.random() - 0.5) * 10;
        }

        particleGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));

        const particleMaterial = new THREE.PointsMaterial({
            color: 0x64ffda,
            size: 0.05,
            transparent: true
        });

        this.particles = new THREE.Points(particleGeometry, particleMaterial);
        scene.add(this.particles);
    }

    animate(): void {
        const positions = this.particles.geometry.attributes.position.array as Float32Array;
        for (let i = 0; i < positions.length; i += 3) {
            positions[i + 1] -= 0.01;
            if (positions[i + 1] < -5) {
                positions[i + 1] = 5;
            }
        }
        this.particles.geometry.attributes.position.needsUpdate = true;
    }
}