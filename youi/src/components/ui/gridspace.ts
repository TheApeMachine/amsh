import { gsap } from 'gsap';

class GridSpace extends HTMLElement {
  private tunnelDepth: number = 10;
  private timeline: gsap.core.Timeline;

  constructor() {
    super();
    this.attachShadow({ mode: 'open' });
    this.timeline = gsap.timeline();
  }

  connectedCallback() {
    this.render();
  }

  disconnectedCallback() {
    if (this.timeline) {
      this.timeline.kill();
    }
  }

  private render() {
    if (!this.shadowRoot) return;

    const style = document.createElement('style');
    style.textContent = `
      :host {
        display: block;
        position: fixed;
        inset: 0;
        background-color: black;
        overflow: hidden;
        z-index: -10;
        perspective: 1000px;
        transform-style: preserve-3d;
      }
      .grid {
        position: absolute;
        inset: 0;
        display: grid;
        grid-template-columns: repeat(8, 1fr);
        grid-template-rows: repeat(8, 1fr);
        transform-style: preserve-3d;
      }
      .cell {
        border: 1px solid rgba(100, 10, 255, 0.9);
        opacity: 0.5;
      }
      .cell:nth-child(even) {
        background-color: rgba(100, 10, 255, 0.5);
      }
    `;

    this.shadowRoot.appendChild(style);

    for (let i = 0; i < this.tunnelDepth; i++) {
      const fragment = document.createDocumentFragment();
      this.createGrid(fragment, 'rotateX(90deg)', 'translateY(-50vh)', i);
      this.createGrid(fragment, 'rotateX(-90deg)', 'translateY(50vh)', i);
      this.createGrid(fragment, 'rotateY(-90deg)', 'translateX(-50vw)', i);
      this.createGrid(fragment, 'rotateY(90deg)', 'translateX(50vw)', i);
      this.shadowRoot.appendChild(fragment);
    }
  }

  private createGrid(parent: DocumentFragment, rotation: string, translation: string, index: number) {
    const grid = document.createElement('div');
    grid.className = 'grid';
    grid.style.transform = `${rotation} ${translation} translateZ(${-50 - index * 50}vh)`;
    grid.style.zIndex = `${this.tunnelDepth - index}`;

    for (let i = 0; i < 64; i++) {
      const cell = document.createElement('div');
      cell.className = 'cell';
      grid.appendChild(cell);
    }

    parent.appendChild(grid);
  }
}

customElements.define('grid-space', GridSpace);