// NavigationControls.ts
export class NavigationControls {
    private currentLayer: number;
    private maxLayers: number;
    private onLayerChange: (layer: number) => void;

    constructor(maxLayers: number, onLayerChange: (layer: number) => void) {
        this.currentLayer = 0;
        this.maxLayers = maxLayers;
        this.onLayerChange = onLayerChange;
        window.addEventListener('keydown', this.onKeyDown.bind(this));
    }

    private onKeyDown(event: KeyboardEvent): void {
        if (event.key === 'ArrowUp' && this.currentLayer < this.maxLayers - 1) {
            this.currentLayer++;
            this.onLayerChange(this.currentLayer);
        } else if (event.key === 'ArrowDown' && this.currentLayer > 0) {
            this.currentLayer--;
            this.onLayerChange(this.currentLayer);
        }
    }
}
