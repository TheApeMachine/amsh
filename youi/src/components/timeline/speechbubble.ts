import * as THREE from 'three';
import { CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer';

export class SpeechBubble {
    private bubble: THREE.Group;
    private textObject: CSS2DObject;
    private lifespan: number;
    private creationTime: number;

    constructor(text: string, position: THREE.Vector3, scene: THREE.Scene) {
        this.bubble = new THREE.Group();
        this.lifespan = 5000; // 5 seconds lifespan
        this.creationTime = Date.now();

        // Create bubble background
        const bubbleGeometry = new THREE.SphereGeometry(2, 32, 32, 0, Math.PI * 2, 0, Math.PI / 2);
        const bubbleMaterial = new THREE.MeshBasicMaterial({
            color: 0x00ff00,
            transparent: true,
            opacity: 0.3
        });
        const bubbleMesh = new THREE.Mesh(bubbleGeometry, bubbleMaterial);
        this.bubble.add(bubbleMesh);

        // Create text
        const textElement = document.createElement('div');
        textElement.className = 'speech-bubble-text';
        textElement.textContent = text;
        textElement.style.color = 'white';
        textElement.style.padding = '10px';
        textElement.style.backgroundColor = 'rgba(0, 0, 0, 0.7)';
        textElement.style.borderRadius = '5px';
        textElement.style.fontSize = '12px';
        textElement.style.maxWidth = '150px';
        textElement.style.overflow = 'hidden';
        textElement.style.textOverflow = 'ellipsis';

        this.textObject = new CSS2DObject(textElement);
        this.textObject.position.set(0, 2, 0);
        this.bubble.add(this.textObject);

        // Position the bubble
        this.bubble.position.copy(position);
        this.bubble.position.y += 3; // Offset above the agent

        scene.add(this.bubble);
    }

    update() {
        const age = Date.now() - this.creationTime;
        const opacity = Math.max(0, 1 - age / this.lifespan);
        (this.bubble.children[0] as THREE.Mesh).material.opacity = opacity * 0.3;
        this.textObject.element.style.opacity = opacity.toString();

        // Slowly float upwards
        this.bubble.position.y += 0.01;

        return age < this.lifespan;
    }

    remove(scene: THREE.Scene) {
        scene.remove(this.bubble);
    }
}