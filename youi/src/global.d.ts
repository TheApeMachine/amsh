declare module '@uwdata/vgplot';
declare module 'affine';
declare module 'three/examples/jsm/controls/OrbitControls';
declare module 'three/examples/jsm/renderers/CSS2DRenderer';
declare module 'three/examples/jsm/postprocessing/EffectComposer';
declare module 'three/examples/jsm/postprocessing/RenderPass';
declare module 'three/examples/jsm/postprocessing/UnrealBloomPass';

declare module '*.css' {
    const styles: string;
    export default styles;
}

/*
 * global.d.ts
 * Extends the Window interface to include stateManager and eventManager.
 */
import { StateManager } from './lib/state';
import { EventManager } from './lib/event';

declare global {
    interface Window {
        stateManager: StateManager;
        eventManager: EventManager;
    }
}