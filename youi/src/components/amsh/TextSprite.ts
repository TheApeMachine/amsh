import * as THREE from 'three';

export class TextSprite {
    sprite: THREE.Sprite;
    canvas: HTMLCanvasElement;
    context: CanvasRenderingContext2D;

    constructor(text: string) {
        this.canvas = document.createElement('canvas');
        this.context = this.canvas.getContext('2d')!;
        const texture = new THREE.Texture(this.canvas);
        const material = new THREE.SpriteMaterial({ map: texture });
        this.sprite = new THREE.Sprite(material);
        this.updateText(text);
    }

    updateText(text: string) {
        const fontSize = 24;
        this.context.font = `${fontSize}px Arial`;
        this.canvas.width = this.context.measureText(text).width;
        this.canvas.height = fontSize;
        this.context.font = `${fontSize}px Arial`;
        this.context.fillStyle = 'white';
        this.context.fillText(text, 0, fontSize);
        (this.sprite.material as THREE.SpriteMaterial).map!.needsUpdate = true;
    }
}