import * as THREE from 'three';
import { AMSHNode } from './Node';

export class Connection {
    line: THREE.Line;
    particles: THREE.Points;

    constructor(source: AMSHNode, target: AMSHNode) {
        const material = new THREE.LineBasicMaterial({ color: 0x64ffda });
        const geometry = new THREE.BufferGeometry();
        const points = [source.mesh.position, target.mesh.position];
        geometry.setFromPoints(points);
        this.line = new THREE.Line(geometry, material);

        // Create particles for data flow effect
        const particleGeometry = new THREE.BufferGeometry();
        const particlePositions = new Float32Array(30 * 3); // 30 particles per connection
        for (let i = 0; i < particlePositions.length; i += 3) {
            const t = i / particlePositions.length;
            particlePositions[i] = source.mesh.position.x * (1 - t) + target.mesh.position.x * t;
            particlePositions[i + 1] = source.mesh.position.y * (1 - t) + target.mesh.position.y * t;
            particlePositions[i + 2] = source.mesh.position.z * (1 - t) + target.mesh.position.z * t;
        }
        particleGeometry.setAttribute('position', new THREE.BufferAttribute(particlePositions, 3));
        const particleMaterial = new THREE.PointsMaterial({ color: 0x00ffff, size: 0.05 });
        this.particles = new THREE.Points(particleGeometry, particleMaterial);

        this.line.userData = { source, target };
        this.particles.userData = { source, target };
    }

    update() {
        const positions = this.line.geometry.attributes.position.array as Float32Array;
        const particlePositions = this.particles.geometry.attributes.position.array as Float32Array;
        const sourcePos = (this.line.userData.source as AMSHNode).mesh.position;
        const targetPos = (this.line.userData.target as AMSHNode).mesh.position;

        // Update line positions
        positions[0] = sourcePos.x;
        positions[1] = sourcePos.y;
        positions[2] = sourcePos.z;
        positions[3] = targetPos.x;
        positions[4] = targetPos.y;
        positions[5] = targetPos.z;

        // Update particle positions
        for (let i = 0; i < particlePositions.length; i += 3) {
            const t = i / particlePositions.length;
            particlePositions[i] = sourcePos.x * (1 - t) + targetPos.x * t;
            particlePositions[i + 1] = sourcePos.y * (1 - t) + targetPos.y * t;
            particlePositions[i + 2] = sourcePos.z * (1 - t) + targetPos.z * t;

            // Animate particles
            particlePositions[i + 1] += 0.01; // Move particles up
            if (particlePositions[i + 1] > targetPos.y) {
                particlePositions[i + 1] = sourcePos.y; // Reset to start
            }
        }

        this.line.geometry.attributes.position.needsUpdate = true;
        this.particles.geometry.attributes.position.needsUpdate = true;
    }
}