import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';
import { EffectComposer } from 'three/examples/jsm/postprocessing/EffectComposer';
import { RenderPass } from 'three/examples/jsm/postprocessing/RenderPass';
import { UnrealBloomPass } from 'three/examples/jsm/postprocessing/UnrealBloomPass';
import { BokehPass } from 'three/examples/jsm/postprocessing/BokehPass';
import { ShaderPass } from 'three/examples/jsm/postprocessing/ShaderPass';
import { FirstPersonControls } from 'three/examples/jsm/controls/FirstPersonControls';
import { Node } from './Node';

const myShader = {
    uniforms: {
        time: { value: 0 }
    },
    vertexShader: `
        varying vec2 vUv;
        void main() {
            vUv = uv;
            gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
        }
    `,
    fragmentShader: `
        uniform float time;
        varying vec2 vUv;
        void main() {
            vec3 color = 0.5 + 0.5 * cos(time + vUv.xyx + vec3(0, 2, 4));
            gl_FragColor = vec4(color, 0.2); // Reduced opacity
        }
    `,
};

export class IsometricScene {
    scene: THREE.Scene;
    camera: THREE.OrthographicCamera;
    renderer: THREE.WebGLRenderer;
    controls: OrbitControls;
    composer: EffectComposer;
    highlightLight!: THREE.SpotLight;

    constructor(container: HTMLElement) {
        this.scene = new THREE.Scene();
        this.camera = new THREE.OrthographicCamera(-10, 10, 10, -10, 0.1, 1000);
        this.renderer = new THREE.WebGLRenderer({ antialias: true });
        this.renderer.setSize(container.clientWidth, container.clientHeight);
        this.renderer.toneMapping = THREE.ACESFilmicToneMapping;
        this.renderer.toneMappingExposure = 1;
        container.appendChild(this.renderer.domElement);

        this.camera.position.set(10, 10, 10);
        this.camera.lookAt(0, 0, 0);

        this.controls = new OrbitControls(this.camera, this.renderer.domElement);
        this.controls.enableRotate = true;
        this.controls.enableZoom = true;

        this.controls = new FirstPersonControls(this.camera, this.renderer.domElement);
        this.controls.lookSpeed = 0.1;
        this.controls.movementSpeed = 5;

        this.setupLighting();
        this.setupPostProcessing();

        window.addEventListener('resize', () => this.onWindowResize(), false);
    }

    setupLighting() {
        const ambientLight = new THREE.AmbientLight(0x404040, 0.2);
        this.scene.add(ambientLight);

        const directionalLight = new THREE.DirectionalLight(0xffffff, 0.5);
        directionalLight.position.set(5, 10, 7.5);
        this.scene.add(directionalLight);

        this.highlightLight = new THREE.SpotLight(0xffffff, 1);
        this.highlightLight.position.set(0, 10, 0);
        this.highlightLight.angle = Math.PI / 6;
        this.highlightLight.penumbra = 0.3;
        this.highlightLight.decay = 2;
        this.highlightLight.distance = 50;
        this.highlightLight.target.position.set(0, 0, 0);
        this.scene.add(this.highlightLight);
        this.scene.add(this.highlightLight.target);
    }

    highlightNode(node: Node) {
        const worldPosition = new THREE.Vector3();
        node.mesh.getWorldPosition(worldPosition);
        this.highlightLight.position.set(worldPosition.x, worldPosition.y + 5, worldPosition.z);
        this.highlightLight.target.position.set(worldPosition.x, worldPosition.y, worldPosition.z);
    }

    setupPostProcessing() {
        this.composer = new EffectComposer(this.renderer);
        
        const renderPass = new RenderPass(this.scene, this.camera);
        this.composer.addPass(renderPass);

        const bloomPass = new UnrealBloomPass(
            new THREE.Vector2(window.innerWidth, window.innerHeight),
            1.5, // strength
            0.4, // radius
            0.85 // threshold
        );
        this.composer.addPass(bloomPass);

        const bokehPass = new BokehPass(this.scene, this.camera, {
            focus: 10.0,
            aperture: 0.025,
            maxblur: 1.0,
        });
        this.composer.addPass(bokehPass);

        const shaderPass = new ShaderPass(myShader);
        this.composer.addPass(shaderPass);

        // Instead, add the shader as a background:
        const planeGeometry = new THREE.PlaneGeometry(100, 100);
        const planeMaterial = new THREE.ShaderMaterial(myShader);
        planeMaterial.transparent = true;
        const backgroundPlane = new THREE.Mesh(planeGeometry, planeMaterial);
        backgroundPlane.position.z = -10;
        backgroundPlane.name = 'background';
        this.scene.add(backgroundPlane);
    }

    onWindowResize() {
        const container = this.renderer.domElement.parentElement;
        if (container) {
            const width = container.clientWidth;
            const height = container.clientHeight;
            const aspect = width / height;
            const frustumSize = 20;
    
            this.camera.left = (-frustumSize * aspect) / 2;
            this.camera.right = (frustumSize * aspect) / 2;
            this.camera.top = frustumSize / 2;
            this.camera.bottom = -frustumSize / 2;
            this.camera.updateProjectionMatrix();
            this.renderer.setSize(width, height);
            this.composer.setSize(width, height);
        }
    }    

    render() {
        const backgroundPlane = this.scene.getObjectByName('background') as THREE.Mesh;
        if (backgroundPlane && backgroundPlane.material instanceof THREE.ShaderMaterial) {
            backgroundPlane.material.uniforms.time.value += 0.01;
        }
        this.composer.render();
    }
}
