// network-3d.ts
import * as THREE from 'three';
import { CSS2DRenderer, CSS2DObject } from 'three/addons/renderers/CSS2DRenderer.js';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';

interface NetworkNode {
    url: string;
    type: 'endpoint' | 'resource';
    requests: number;
    avgDuration: number;
    lastStatus: number;
    position: THREE.Vector3;
    velocity: THREE.Vector3;
    object: THREE.Mesh;
    label: CSS2DObject;
}

interface NetworkLink {
    source: NetworkNode;
    target: NetworkNode;
    requests: number;
    object: THREE.Line;
}

export const setup3DNetworkView = ({ addLog, overlay }: { addLog: Function; overlay: HTMLElement }) => {
    // Create container - now using the passed overlay element instead of creating new one
    const container = overlay;
    container.classList.add('debug-section');

    // Add controls inside container
    const controls = document.createElement('div');
    controls.className = 'debug-controls';
    container.appendChild(controls);

    // Create renderer container
    const rendererContainer = document.createElement('div');
    rendererContainer.className = 'renderer-container';
    container.appendChild(rendererContainer);

    // Setup Three.js
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(75, rendererContainer.clientWidth / rendererContainer.clientHeight, 0.1, 1000);
    camera.position.z = 5;

    // Setup renderers
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(rendererContainer.clientWidth, rendererContainer.clientHeight);
    rendererContainer.appendChild(renderer.domElement);

    const labelRenderer = new CSS2DRenderer();
    labelRenderer.setSize(rendererContainer.clientWidth, rendererContainer.clientHeight);
    labelRenderer.domElement.className = 'renderer-label';
    rendererContainer.appendChild(labelRenderer.domElement);

    // Add orbital controls
    const orbitControls = new OrbitControls(camera, labelRenderer.domElement);
    orbitControls.enableDamping = true;
    orbitControls.dampingFactor = 0.05;
    orbitControls.rotateSpeed = 0.5;
    orbitControls.zoomSpeed = 0.8;
    orbitControls.panSpeed = 0.8;
    orbitControls.minDistance = 2;
    orbitControls.maxDistance = 20;
    orbitControls.screenSpacePanning = true;

    // Track nodes and links
    const nodes = new Map<string, NetworkNode>();
    const links = new Map<string, NetworkLink>();

    // Materials
    const nodeMaterial = new THREE.MeshPhongMaterial({
        color: 0x00ff00,
        emissive: 0x00ff00,
        emissiveIntensity: 0.5
    });
    const linkMaterial = new THREE.LineBasicMaterial({
        color: 0x00ff00,
        opacity: 0.5,
        transparent: true
    });

    // Add lighting
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.5);
    scene.add(ambientLight);

    const pointLight = new THREE.PointLight(0xffffff, 1);
    pointLight.position.set(5, 5, 5);
    scene.add(pointLight);

    // Create node mesh
    const createNodeMesh = () => {
        const geometry = new THREE.SphereGeometry(0.2, 32, 32);
        const mesh = new THREE.Mesh(geometry, nodeMaterial.clone());
        return mesh;
    };

    // Create or update node
    const updateNode = (url: string, type: 'endpoint' | 'resource', status: number, duration: number) => {
        let node = nodes.get(url);

        if (!node) {
            const mesh = createNodeMesh();
            const label = new CSS2DObject(createLabel(url));

            node = {
                url,
                type,
                requests: 0,
                avgDuration: 0,
                lastStatus: status,
                position: new THREE.Vector3(
                    (Math.random() - 0.5) * 10,
                    (Math.random() - 0.5) * 10,
                    (Math.random() - 0.5) * 10
                ),
                velocity: new THREE.Vector3(),
                object: mesh,
                label
            };

            mesh.position.copy(node.position);
            scene.add(mesh);
            scene.add(label);
            nodes.set(url, node);
        }

        // Update node properties
        node.requests++;
        node.avgDuration = (node.avgDuration * (node.requests - 1) + duration) / node.requests;
        node.lastStatus = status;

        // Update visual properties
        const scale = Math.min(1 + (node.requests * 0.1), 3);
        node.object.scale.setScalar(scale);

        // Color based on status
        let color: number;
        if (status >= 400) {
            color = 0xff0000;
        } else if (status >= 300) {
            color = 0xffff00;
        } else {
            color = 0x00ff00;
        }
        const material = node.object.material as THREE.MeshPhongMaterial;
        material.color.setHex(color);
        material.emissive.setHex(color);

        // Emit particles for activity
        createActivityParticle(node.position.clone(), status);

        return node;
    };

    // Create activity particles
    const createActivityParticle = (position: THREE.Vector3, status: number) => {
        const geometry = new THREE.BufferGeometry();
        geometry.setAttribute('position', new THREE.Float32BufferAttribute([0, 0, 0], 3));

        let color: number;
        if (status >= 400) {
            color = 0xff0000;
        } else if (status >= 300) {
            color = 0xffff00;
        } else {
            color = 0x00ff00;
        }
        const material = new THREE.PointsMaterial({
            color,
            size: 0.1,
            transparent: true,
            opacity: 1
        });

        const particle = new THREE.Points(geometry, material);
        particle.position.copy(position);

        scene.add(particle);

        // Animate particle
        const animate = () => {
            material.opacity -= 0.02;
            particle.scale.multiplyScalar(1.05);

            if (material.opacity <= 0) {
                scene.remove(particle);
                geometry.dispose();
                material.dispose();
                return;
            }

            requestAnimationFrame(animate);
        };

        animate();
    };

    // Create or update link
    const updateLink = (sourceUrl: string, targetUrl: string) => {
        const linkId = `${sourceUrl}-${targetUrl}`;
        let link = links.get(linkId);

        if (!link) {
            const sourceNode = nodes.get(sourceUrl)!;
            const targetNode = nodes.get(targetUrl)!;

            const geometry = new THREE.BufferGeometry().setFromPoints([
                sourceNode.position,
                targetNode.position
            ]);

            const line = new THREE.Line(geometry, linkMaterial.clone());
            scene.add(line);

            link = {
                source: sourceNode,
                target: targetNode,
                requests: 0,
                object: line
            };

            links.set(linkId, link);
        }

        link.requests++;
        const material = link.object.material as THREE.LineBasicMaterial;
        material.opacity = Math.min(0.2 + (link.requests * 0.1), 1);

        return link;
    };

    // Create label element
    const createLabel = (text: string) => {
        const div = document.createElement('div');
        div.textContent = decodeURIComponent(text);
        div.className = 'node-label';
        return div;
    };

    // Force-directed layout simulation
    const simulateForces = () => {
        const REPULSION = 0.5;
        const ATTRACTION = 0.005;
        const DAMPING = 0.85;

        // Apply forces between nodes
        nodes.forEach(node1 => {
            nodes.forEach(node2 => {
                if (node1 === node2) return;

                const diff = node1.position.clone().sub(node2.position);
                const distance = diff.length();

                if (distance === 0) return;

                const force = diff.normalize().multiplyScalar(REPULSION / (distance * distance));
                node1.velocity.add(force);
                node2.velocity.sub(force);
            });
        });

        // Apply spring forces along links
        links.forEach(link => {
            const diff = link.target.position.clone().sub(link.source.position);
            const distance = diff.length();
            const force = diff.normalize().multiplyScalar((distance - 2) * ATTRACTION);

            link.source.velocity.add(force);
            link.target.velocity.sub(force);
        });

        // Update positions
        nodes.forEach(node => {
            node.velocity.multiplyScalar(DAMPING);
            node.position.add(node.velocity);
            node.object.position.copy(node.position);
            node.label.position.copy(node.position);
        });

        // Update link geometries
        links.forEach(link => {
            const positions = new Float32Array([
                link.source.position.x, link.source.position.y, link.source.position.z,
                link.target.position.x, link.target.position.y, link.target.position.z
            ]);
            link.object.geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
            link.object.geometry.attributes.position.needsUpdate = true;
        });
    };

    // Animation loop
    const animate = () => {
        requestAnimationFrame(animate);
        orbitControls.update();
        simulateForces();
        renderer.render(scene, camera);
        labelRenderer.render(scene, camera);
    };
    animate();

    // Handle window resize
    const onWindowResize = () => {
        const width = rendererContainer.clientWidth;
        const height = rendererContainer.clientHeight;

        camera.aspect = width / height;
        camera.updateProjectionMatrix();

        renderer.setSize(width, height);
        labelRenderer.setSize(width, height);
    };
    const resizeObserver = new ResizeObserver(() => {
        onWindowResize();
    });
    resizeObserver.observe(rendererContainer);

    // Track network requests
    const originalFetch = window.fetch;
    window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
        const startTime = performance.now();
        const url = input instanceof Request ? input.url : input.toString();

        try {
            const response = await originalFetch(input, init);
            const duration = performance.now() - startTime;

            // Update visualization
            const sourceNode = updateNode('origin', 'endpoint', 200, 0);
            const targetNode = updateNode(url, 'endpoint', response.status, duration);
            updateLink('origin', url);

            // Log node details for debugging
            addLog?.({
                type: 'network',
                message: `Request from ${sourceNode.url} to ${targetNode.url}`,
                data: { status: response.status, duration }
            });

            return response;
        } catch (error) {
            const duration = performance.now() - startTime;
            updateNode(url, 'endpoint', 500, duration);
            throw error;
        }
    };

    // Handle controls
    let isVisible = true;
    document.getElementById('toggle-3d')?.addEventListener('click', () => {
        isVisible = !isVisible;
        container.style.display = isVisible ? 'block' : 'none';
        const button = document.getElementById('toggle-3d');
        if (button) {
            button.textContent = isVisible ? '👁 Hide View' : '👁 Show View';
        }
    });

    document.getElementById('reset-camera')?.addEventListener('click', () => {
        camera.position.set(0, 0, 5);
        camera.lookAt(0, 0, 0);
        orbitControls.reset();
    });

    let showLabels = true;
    document.getElementById('toggle-labels')?.addEventListener('click', () => {
        showLabels = !showLabels;
        nodes.forEach(node => {
            node.label.visible = showLabels;
        });
        const button = document.getElementById('toggle-labels');
        if (button) {
            button.textContent = showLabels ? '🏷 Hide Labels' : '🏷 Show Labels';
        }
    });

    return {
        getNodes: () => nodes,
        getLinks: () => links,
        destroy: () => {
            window.fetch = originalFetch;
            window.removeEventListener('resize', onWindowResize);
            container.remove();
            renderer.dispose();
            scene.traverse(object => {
                if (object instanceof THREE.Mesh) {
                    object.geometry.dispose();
                    (object.material as THREE.Material).dispose();
                }
            });
        }
    };
};