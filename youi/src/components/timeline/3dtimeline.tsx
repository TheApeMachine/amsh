// Import necessary modules
import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';
import { CSS2DRenderer, CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer';
import { EffectComposer } from 'three/examples/jsm/postprocessing/EffectComposer';
import { RenderPass } from 'three/examples/jsm/postprocessing/RenderPass';
import { UnrealBloomPass } from 'three/examples/jsm/postprocessing/UnrealBloomPass';
import { ShaderMaterial } from 'three';
import gsap from 'gsap';
import { SpeechBubble } from './speechbubble';
import { jsx } from '@/lib/template';

interface AgentData {
    id: string;
    position: THREE.Vector3;
    connections: string[];
    type: string;
    time: number;
    parentId?: string;
}

export const InteractiveAI3DTimelineMindMap = () => {
    // Scene and rendering variables
    const scene = new THREE.Scene();
    let camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 2000);
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    const labelRenderer = new CSS2DRenderer();
    const composer = new EffectComposer(renderer);
    const controls = new OrbitControls(camera, labelRenderer.domElement);
    const agents = new Map<string, THREE.Mesh>();
    let connections: THREE.Line[] = [];
    const timeline = new THREE.Object3D();
    const particles = new THREE.Points();
    let currentLayoutName = 'force';

    // Interaction variables
    const raycaster = new THREE.Raycaster();
    const mouse = new THREE.Vector2();
    const tooltip = document.createElement('div');
    let selectedAgent: THREE.Mesh | null = null;
    let agentControlPanel: HTMLDivElement | null = null;
    let isSimulationPaused = false;
    let speechBubbles: SpeechBubble[] = [];

    // Data variables
    let agentDataList: AgentData[] = [];
    const typeColors: { [key: string]: number } = {
        'Controller': 0xff5733, // Orange
        'Worker': 0x33ff57,     // Green
        'Analyzer': 0x3357ff,   // Blue
        'Scheduler': 0xff33a8,  // Pink
        'Temporary': 0xffff00   // Yellow
    };

    const initializeScene = () => {
        createTooltip();
        createTimeline();
        createParticleSystem();
        createControls();
        addEventListeners();
        loadData(); // Load initial data
        startSimulation();

        // Use GSAP's ticker for synchronized animations
        gsap.ticker.add(animate.bind(this));
        scene.background = new THREE.Color(getComputedStyle(document.documentElement).getPropertyValue('--background-color') || '#000011');

        camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 2000);
        camera.position.z = 100;

        renderer.setSize(window.innerWidth, window.innerHeight);
        document.body.appendChild(renderer.domElement);

        const labelRenderer = new CSS2DRenderer();
        labelRenderer.setSize(window.innerWidth, window.innerHeight);
        labelRenderer.domElement.style.position = 'absolute';
        labelRenderer.domElement.style.top = '0px';
        document.body.appendChild(labelRenderer.domElement);

        const controls = new OrbitControls(camera, labelRenderer.domElement);
        controls.enableDamping = true;
        controls.dampingFactor = 0.05;

        const ambientLight = new THREE.AmbientLight(0x333333);
        scene.add(ambientLight);

        const pointLight = new THREE.PointLight(0xffffff, 2);
        pointLight.position.set(0, 0, 50);
        scene.add(pointLight);

        // Set up post-processing
        const renderPass = new RenderPass(scene, camera);
        const bloomPass = new UnrealBloomPass(
            new THREE.Vector2(window.innerWidth, window.innerHeight),
            1.5,
            0.4,
            0.85
        );

        const composer = new EffectComposer(renderer);
        composer.addPass(renderPass);
        composer.addPass(bloomPass);
    }

    const createTooltip = () => {
        const tooltip = document.createElement('div');
        tooltip.style.position = 'absolute';
        tooltip.style.background = 'rgba(0, 0, 0, 0.7)';
        tooltip.style.color = '#fff';
        tooltip.style.padding = '5px';
        tooltip.style.borderRadius = '3px';
        tooltip.style.pointerEvents = 'none';
        tooltip.style.display = 'none';
        document.body.appendChild(tooltip);
    }

    const createTimeline = () => {
        const geometry = new THREE.BufferGeometry().setFromPoints([
            new THREE.Vector3(-50, -30, 0),
            new THREE.Vector3(50, -30, 0)
        ]);
        const material = new THREE.LineBasicMaterial({ color: 0x00ffff, linewidth: 2 });
        const timeline = new THREE.Line(geometry, material);
        scene.add(timeline);

        const timeLabel = document.createElement('div');
        timeLabel.className = 'label';
        timeLabel.textContent = 'Time';
        timeLabel.style.color = '#00ffff';
        timeLabel.style.fontSize = '16px';
        const timeLabelObject = new CSS2DObject(timeLabel);
        timeLabelObject.position.set(0, -33, 0);
        timeline.add(timeLabelObject);
    }

    const createParticleSystem = () => {
        const particleCount = 1000;
        const particles = new THREE.BufferGeometry();
        const positions = new Float32Array(particleCount * 3);

        for (let i = 0; i < particleCount * 3; i += 3) {
            positions[i] = (Math.random() - 0.5) * 200;
            positions[i + 1] = (Math.random() - 0.5) * 200;
            positions[i + 2] = (Math.random() - 0.5) * 200;
        }

        particles.setAttribute('position', new THREE.BufferAttribute(positions, 3));

        const particleMaterial = new THREE.PointsMaterial({
            color: 0x00ffff,
            size: 0.1,
            blending: THREE.AdditiveBlending,
            transparent: true,
            opacity: 0.8
        });

        const particlePoints = new THREE.Points(particles, particleMaterial);
        scene.add(particlePoints);
    }

    const addAgent = (agentData: AgentData) => {
        const { position, id, type } = agentData;

        // Ensure position is valid
        if (isNaN(position.x) || isNaN(position.y) || isNaN(position.z)) {
            position.set(Math.random() * 100 - 50, Math.random() * 100 - 50, Math.random() * 100 - 50);
        }

        // Choose geometry based on type
        let geometry: THREE.BufferGeometry;
        if (type === 'Temporary') {
            geometry = new THREE.TetrahedronGeometry(1, 0); // Different shape for temporary agents
        } else {
            geometry = new THREE.SphereGeometry(1, 32, 32);
        }

        const color = new THREE.Color(typeColors[type] || 0xffffff);
        const material = new ShaderMaterial({
            uniforms: {
                time: { value: 0 },
                color: { value: color }
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
                uniform vec3 color;
                varying vec2 vUv;
                void main() {
                    float glow = sin(time * 5.0 + length(vUv - 0.5) * 10.0) * 0.5 + 0.5;
                    gl_FragColor = vec4(color * glow, 1.0);
                }
            `
        });
        const agent = new THREE.Mesh(geometry, material);
        agent.position.copy(position);
        agent.name = id; // Set the name for identification

        const label = document.createElement('div');
        label.className = 'label';
        label.textContent = id;
        label.style.color = '#ffffff';
        label.style.fontSize = '14px';
        const labelObject = new CSS2DObject(label);
        labelObject.position.set(0, 1.5, 0);
        agent.add(labelObject);

        scene.add(agent);
        agents.set(id, agent);

        // Add a point on the timeline
        const timePoint = new THREE.Mesh(
            new THREE.SphereGeometry(0.5, 16, 16),
            new THREE.MeshBasicMaterial({ color: 0x00ffff })
        );
        timePoint.position.set(agentData.time * 2 - 50, -30, 0);
        timeline.add(timePoint);

        // Add a vertical line connecting the agent to the timeline
        const lineGeometry = new THREE.BufferGeometry().setFromPoints([
            new THREE.Vector3(agentData.time * 2 - 50, -30, 0),
            agent.position
        ]);
        const lineMaterial = new THREE.LineBasicMaterial({ color: 0x00ffff, opacity: 0.5, transparent: true });
        const line = new THREE.Line(lineGeometry, lineMaterial);
        scene.add(line);
    }

    const addConnection = (agentId1: string, agentId2: string) => {
        const agent1 = agents.get(agentId1);
        const agent2 = agents.get(agentId2);

        if (agent1 && agent2) {
            const points = [
                agent1.position.clone(),
                agent2.position.clone()
            ];

            const geometry = new THREE.BufferGeometry().setFromPoints(points);
            const material = new THREE.LineBasicMaterial({ color: 0x00ffff, opacity: 0.7, transparent: true });
            const connection = new THREE.Line(geometry, material);
            connection.userData = { start: agentId1, end: agentId2 }; // Store the connected agent IDs
            scene.add(connection);
            connections.push(connection);

            // Add data flow particles
            addDataFlowParticles(agent1.position, agent2.position, connection);
        }
    }

    const addDataFlowParticles = (start: THREE.Vector3, end: THREE.Vector3, connection: THREE.Line) => {
        const particleMaterial = new THREE.PointsMaterial({
            color: 0xffa500,
            size: 0.2,
            transparent: true,
            opacity: 0.6
        });

        const particleCount = 50;
        const particleGeometry = new THREE.BufferGeometry();
        const positions = new Float32Array(particleCount * 3);

        for (let i = 0; i < particleCount; i++) {
            positions[i * 3] = start.x;
            positions[i * 3 + 1] = start.y;
            positions[i * 3 + 2] = start.z;
        }

        particleGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
        const particles = new THREE.Points(particleGeometry, particleMaterial);
        connection.add(particles);

        // Animate particles along the line
        const direction = end.clone().sub(start);
        const length = direction.length();
        direction.normalize();

        gsap.to(particles.position, {
            x: `+=${direction.x * length}`,
            y: `+=${direction.y * length}`,
            z: `+=${direction.z * length}`,
            duration: 5,
            repeat: -1,
            ease: 'none'
        });
    }

    const applyLayout = (layoutName: string) => {
        console.log("Applying layout:", layoutName);

        switch (layoutName) {
            case 'force':
                applyForceLayout();
                break;
            case 'sphere':
                applySphereLayout();
                break;
            case 'grid':
                applyGridLayout();
                break;
            case 'circular':
                applyCircularLayout();
                break;
            default:
                applyRandomLayout();
        }

        currentLayoutName = layoutName;
    }

    const animateToNewLayout = () => {
        agentDataList.forEach(agentData => {
            const agent = agents.get(agentData.id);
            if (agent) {
                gsap.to(agent.position, {
                    x: agentData.position.x,
                    y: agentData.position.y,
                    z: agentData.position.z,
                    duration: 2,
                    ease: 'power1.inOut'
                });
            }
        });
    }

    const applyForceLayout = () => {
        // Simple force-directed layout using Coulomb's law and Hooke's law

        const iterations = 100;
        const k = 50; // Spring constant
        const repulsionStrength = 10000;

        for (let i = 0; i < iterations; i++) {
            // Reset forces
            const forces: { [id: string]: THREE.Vector3 } = {};
            agentDataList.forEach(agent => {
                forces[agent.id] = new THREE.Vector3();
            });

            // Calculate repulsion forces
            for (let j = 0; j < agentDataList.length; j++) {
                for (let k = j + 1; k < agentDataList.length; k++) {
                    const agentA = agentDataList[j];
                    const agentB = agentDataList[k];
                    const delta = agentA.position.clone().sub(agentB.position);
                    const distance = delta.length() + 0.1; // Avoid division by zero
                    const forceMagnitude = repulsionStrength / (distance * distance);
                    const force = delta.normalize().multiplyScalar(forceMagnitude);

                    forces[agentA.id].add(force);
                    forces[agentB.id].sub(force);
                }
            }

            // Calculate spring forces
            agentDataList.forEach(agent => {
                agent.connections.forEach(connId => {
                    const connectedAgent = agentDataList.find(a => a.id === connId);
                    if (connectedAgent) {
                        const delta = agent.position.clone().sub(connectedAgent.position);
                        const distance = delta.length();
                        const forceMagnitude = -k * (distance - 30);
                        const force = delta.normalize().multiplyScalar(forceMagnitude);

                        forces[agent.id].add(force);
                    }
                });
            });

            // Update positions
            agentDataList.forEach(agent => {
                agent.position.add(forces[agent.id].multiplyScalar(0.01));
            });
        }

        animateToNewLayout();
    }

    const applySphereLayout = () => {
        const radius = 50;
        const phi = Math.PI * (3 - Math.sqrt(5)); // Golden angle in radians

        agentDataList.forEach((agentData, i) => {
            const y = 1 - (i / (agentDataList.length - 1)) * 2;
            const radiusAtY = Math.sqrt(1 - y * y) * radius;

            const theta = phi * i;

            const x = Math.cos(theta) * radiusAtY;
            const z = Math.sin(theta) * radiusAtY;

            agentData.position.set(x, y * radius, z);
        });

        animateToNewLayout();
    }

    const applyGridLayout = () => {
        const gridSize = Math.ceil(Math.cbrt(agentDataList.length));
        const spacing = 20;
        const offset = (gridSize - 1) * spacing / 2;

        agentDataList.forEach((agentData, i) => {
            const x = (i % gridSize) * spacing - offset;
            const y = (Math.floor((i / gridSize) % gridSize)) * spacing - offset;
            const z = (Math.floor(i / (gridSize * gridSize))) * spacing - offset;

            agentData.position.set(x, y, z);
        });

        animateToNewLayout();
    }

    const applyCircularLayout = () => {
        const radius = 50;
        const angleStep = (2 * Math.PI) / agentDataList.length;

        agentDataList.forEach((agentData, i) => {
            const angle = i * angleStep;
            const x = Math.cos(angle) * radius;
            const z = Math.sin(angle) * radius;
            const y = 0;

            agentData.position.set(x, y, z);
        });

        animateToNewLayout();
    }

    const applyRandomLayout = () => {
        agentDataList.forEach(agentData => {
            agentData.position.set(
                (Math.random() - 0.5) * 100,
                (Math.random() - 0.5) * 100,
                (Math.random() - 0.5) * 100
            );
        });
        animateToNewLayout();
    }

    const createControls = () => {
        const controlsContainer = document.createElement('div');
        controlsContainer.style.position = 'absolute';
        controlsContainer.style.top = '10px';
        controlsContainer.style.left = '10px';
        controlsContainer.style.display = 'flex';
        controlsContainer.style.flexDirection = 'column';
        controlsContainer.style.gap = '10px';
        controlsContainer.style.background = 'var(--control-background, rgba(0, 0, 0, 0.7))';
        controlsContainer.style.padding = '10px';
        controlsContainer.style.borderRadius = '5px';

        // Layout buttons
        const layouts = ['force', 'sphere', 'grid', 'circular', 'random'];
        layouts.forEach(layout => {
            const button = document.createElement('button');
            button.textContent = `Apply ${layout} layout`;
            button.onclick = () => {
                console.log("Button clicked: Applying layout", layout);
                applyLayout(layout);
            };
            controlsContainer.appendChild(button);
        });

        // Simulation controls
        const simulationControls = document.createElement('div');
        simulationControls.style.display = 'flex';
        simulationControls.style.gap = '5px';

        const playPauseButton = document.createElement('button');
        playPauseButton.textContent = 'Pause';
        playPauseButton.onclick = () => {
            isSimulationPaused = !isSimulationPaused;
            playPauseButton.textContent = isSimulationPaused ? 'Play' : 'Pause';
            if (isSimulationPaused) {
                gsap.globalTimeline.pause();
            } else {
                gsap.globalTimeline.resume();
            }
        };
        simulationControls.appendChild(playPauseButton);

        const speedSlider = createSlider('Speed', 0.5, 2, 1, 0.1, (value) => {
            gsap.globalTimeline.timeScale(value);
        });
        simulationControls.appendChild(speedSlider);

        controlsContainer.appendChild(simulationControls);

        // Theme toggle
        const themeSelect = document.createElement('select');
        const themes = ['Default', 'Dark', 'Neon'];
        themes.forEach(theme => {
            const option = document.createElement('option');
            option.value = theme.toLowerCase();
            option.textContent = theme;
            themeSelect.appendChild(option);
        });

        themeSelect.onchange = () => {
            applyTheme(themeSelect.value);
        };
        controlsContainer.appendChild(themeSelect);

        // Time slider
        const timeSlider = createSlider('Time', 0, 100, 100, 1, (value) => {
            filterAgentsByTime(value);
        });
        controlsContainer.appendChild(timeSlider);

        // Type filter
        const filterLabel = document.createElement('label');
        filterLabel.textContent = 'Filter by Type:';
        filterLabel.style.color = 'white';
        filterLabel.style.marginTop = '10px';
        controlsContainer.appendChild(filterLabel);

        const typeSelect = document.createElement('select');
        const allOption = document.createElement('option');
        allOption.value = 'all';
        allOption.textContent = 'All';
        typeSelect.appendChild(allOption);

        const types = ['Controller', 'Worker', 'Analyzer', 'Scheduler', 'Temporary'];
        types.forEach(type => {
            const option = document.createElement('option');
            option.value = type;
            option.textContent = type;
            typeSelect.appendChild(option);
        });

        typeSelect.onchange = () => {
            filterAgentsByType(typeSelect.value);
        };

        controlsContainer.appendChild(typeSelect);

        // Search agent
        const searchContainer = document.createElement('div');
        searchContainer.style.display = 'flex';
        searchContainer.style.flexDirection = 'column';
        searchContainer.style.width = '100%';

        const searchLabel = document.createElement('label');
        searchLabel.textContent = 'Search Agent:';
        searchLabel.style.color = 'white';
        searchContainer.appendChild(searchLabel);

        const searchInput = document.createElement('input');
        searchInput.type = 'text';
        searchInput.style.width = '100%';
        searchInput.onchange = () => {
            focusOnAgent(searchInput.value);
        };
        searchContainer.appendChild(searchInput);

        controlsContainer.appendChild(searchContainer);

        // Save and Load buttons
        const saveButton = document.createElement('button');
        saveButton.textContent = 'Save Layout';
        saveButton.onclick = () => {
            saveLayout();
        };
        controlsContainer.appendChild(saveButton);

        const loadButton = document.createElement('button');
        loadButton.textContent = 'Load Layout';
        loadButton.onclick = () => {
            loadLayout();
        };
        controlsContainer.appendChild(loadButton);

        // Export Image button
        const exportButton = document.createElement('button');
        exportButton.textContent = 'Export Image';
        exportButton.onclick = () => {
            exportImage();
        };
        controlsContainer.appendChild(exportButton);

        // Append controls to the shadow DOM
        document.body.appendChild(controlsContainer);
    }

    const createSlider = (label: string, min: number, max: number, value: number, step: number, onChange: (value: number) => void): HTMLDivElement => {
        const container = document.createElement('div');
        container.style.display = 'flex';
        container.style.flexDirection = 'column';
        container.style.alignItems = 'flex-start';
        container.style.width = '100%';

        const labelElement = document.createElement('label');
        labelElement.textContent = label;
        labelElement.style.color = 'white';
        labelElement.style.marginBottom = '5px';

        const slider = document.createElement('input');
        slider.type = 'range';
        slider.min = min.toString();
        slider.max = max.toString();
        slider.value = value.toString();
        slider.step = step.toString();
        slider.style.width = '100%';

        slider.oninput = () => {
            onChange(parseFloat(slider.value));
        };

        container.appendChild(labelElement);
        container.appendChild(slider);

        return container;
    }

    const animate = () => {
        const time = performance.now();

        const timeInSeconds = time * 0.001; // Convert time to seconds for other uses

        if (!isSimulationPaused) {
            controls.update();

            agents.forEach((agent) => {
                if (agent.material instanceof THREE.ShaderMaterial) {
                    agent.material.uniforms.time.value = timeInSeconds;
                }
            });

            particles.rotation.x = timeInSeconds * 0.05;
            particles.rotation.y = timeInSeconds * 0.03;

            updateConnectionPositions();

            speechBubbles = speechBubbles.filter(bubble => {
                const isAlive = bubble.update();
                if (!isAlive) {
                    bubble.remove(scene);
                }
                return isAlive;
            });

            // Update raycaster for tooltips and selection
            raycaster.setFromCamera(mouse, camera);
            const intersects = raycaster.intersectObjects(Array.from(agents.values()));

            if (intersects.length > 0) {
                const intersectedAgent = intersects[0].object;
                const agentName = intersectedAgent.name;
                tooltip.textContent = `Agent: ${agentName}`;
                tooltip.style.display = 'block';
            } else {
                tooltip.style.display = 'none';
            }
        }

        composer.render();
        labelRenderer.render(scene, camera);
    }

    const updateConnectionPositions = () => {
        connections.forEach((connection) => {
            const startAgent = agents.get(connection.userData.start);
            const endAgent = agents.get(connection.userData.end);
            if (startAgent && endAgent) {
                const positions = connection.geometry.attributes.position.array as Float32Array;
                positions[0] = startAgent.position.x;
                positions[1] = startAgent.position.y;
                positions[2] = startAgent.position.z;
                positions[3] = endAgent.position.x;
                positions[4] = endAgent.position.y;
                positions[5] = endAgent.position.z;
                connection.geometry.attributes.position.needsUpdate = true;
            }
        });
    }

    const addEventListeners = () => {
        window.addEventListener('resize', onWindowResize.bind(this));
        labelRenderer.domElement.addEventListener('mousemove', onMouseMove.bind(this));
        labelRenderer.domElement.addEventListener('click', onClick.bind(this));
    }

    const onWindowResize = () => {
        camera.aspect = window.innerWidth / window.innerHeight;
        camera.updateProjectionMatrix();
        renderer.setSize(window.innerWidth, window.innerHeight);
        labelRenderer.setSize(window.innerWidth, window.innerHeight);
        composer.setSize(window.innerWidth, window.innerHeight);
    }

    const onMouseMove = (event: MouseEvent) => {
        const rect = labelRenderer.domElement.getBoundingClientRect();
        mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
        mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

        // Update tooltip position
        tooltip.style.left = `${event.clientX + 10}px`;
        tooltip.style.top = `${event.clientY + 10}px`;
    }

    const onClick = (event: MouseEvent) => {
        const rect = labelRenderer.domElement.getBoundingClientRect();
        const mouse = new THREE.Vector2(
            ((event.clientX - rect.left) / rect.width) * 2 - 1,
            -((event.clientY - rect.top) / rect.height) * 2 + 1
        );

        raycaster.setFromCamera(mouse, camera);
        const intersects = raycaster.intersectObjects(Array.from(agents.values()));

        if (intersects.length > 0) {
            if (selectedAgent) {
                // Reset previous selection
                selectedAgent.scale.set(1, 1, 1);
            }
            selectedAgent = intersects[0].object as THREE.Mesh;
            selectedAgent.scale.set(1.5, 1.5, 1.5);

            // Highlight connections
            highlightConnections(selectedAgent.name);

            // Show agent control panel
            showAgentControlPanel(selectedAgent.name);
            showAgentResponse(selectedAgent.name);
        } else if (selectedAgent) {
            // Deselect if clicked on empty space
            selectedAgent.scale.set(1, 1, 1);
            selectedAgent = null;
            resetConnectionHighlights();
            hideAgentControlPanel();
        }
    }

    const highlightConnections = (agentId: string) => {
        connections.forEach(connection => {
            const isConnected = connection.userData.start === agentId || connection.userData.end === agentId;
            (connection.material as THREE.LineBasicMaterial).color.set(isConnected ? 0xff0000 : 0x00ffff);
        });
    }

    const resetConnectionHighlights = () => {
        connections.forEach(connection => {
            (connection.material as THREE.LineBasicMaterial).color.set(0x00ffff);
        });
    }

    const filterAgentsByType = (type: string) => {
        agents.forEach((agentMesh, agentId) => {
            const agentData = agentDataList.find(agent => agent.id === agentId);
            if (agentData) {
                const shouldBeVisible = type === 'all' || agentData.type === type;
                agentMesh.visible = shouldBeVisible;
            }
        });

        // Update connections visibility
        connections.forEach(connection => {
            const startAgentVisible = agents.get(connection.userData.start)?.visible ?? false;
            const endAgentVisible = agents.get(connection.userData.end)?.visible ?? false;
            connection.visible = startAgentVisible && endAgentVisible;
        });
    }

    const filterAgentsByTime = (maxTime: number) => {
        agents.forEach((agentMesh, agentId) => {
            const agentData = agentDataList.find(agent => agent.id === agentId);
            if (agentData) {
                const shouldBeVisible = agentData.time <= maxTime;
                agentMesh.visible = shouldBeVisible;
            }
        });

        // Update connections visibility
        connections.forEach(connection => {
            const startAgentVisible = agents.get(connection.userData.start)?.visible ?? false;
            const endAgentVisible = agents.get(connection.userData.end)?.visible ?? false;
            connection.visible = startAgentVisible && endAgentVisible;
        });
    }

    const focusOnAgent = (agentId: string) => {
        const agent = agents.get(agentId);
        if (agent) {
            gsap.to(camera.position, {
                x: agent.position.x + 10,
                y: agent.position.y + 10,
                z: agent.position.z + 10,
                duration: 1,
                onUpdate: () => {
                    camera.lookAt(agent.position);
                    controls.target.copy(agent.position);
                }
            });
        } else {
            alert('Agent not found.');
        }
    }

    const saveLayout = () => {
        const layoutData = {
            agents: agentDataList.map(agentData => ({
                id: agentData.id,
                position: agentData.position,
                type: agentData.type,
                time: agentData.time,
                connections: agentData.connections,
                parentId: agentData.parentId
            }))
        };
        const dataStr = JSON.stringify(layoutData);
        const dataUri = 'data:application/json;charset=utf-8,' + encodeURIComponent(dataStr);

        const exportFileDefaultName = 'layout.json';

        const linkElement = document.createElement('a');
        linkElement.setAttribute('href', dataUri);
        linkElement.setAttribute('download', exportFileDefaultName);
        linkElement.click();
    }

    const loadLayout = async () => {
        const inputElement = document.createElement('input');
        inputElement.type = 'file';
        inputElement.accept = 'application/json';
        inputElement.onchange = async (event: any) => {
            const file = event.target.files[0];
            const text = await file.text();
            const data = JSON.parse(text);
            initializeAgentsAndConnections(data);
        };
        inputElement.click();
    }

    const exportImage = () => {
        renderer.render(scene, camera); // Ensure the latest frame is rendered
        const dataURL = renderer.domElement.toDataURL('image/png');
        const link = document.createElement('a');
        link.download = 'visualization.png';
        link.href = dataURL;
        link.click();
    }

    // Loading data from a source (can be modified to fetch from an API or file)
    const loadData = () => {
        generateTestData(); // For demonstration purposes
        initializeAgentsAndConnections({ agents: agentDataList });
        applyLayout('force'); // Apply initial layout
    }

    const initializeAgentsAndConnections = (data: any) => {
        // Clear existing agents and connections
        agents.forEach(agent => scene.remove(agent));
        connections.forEach(connection => scene.remove(connection));
        agents.clear();
        connections = [];

        // Reset agent data list
        agentDataList = data.agents;

        // Add agents
        agentDataList.forEach(agentData => {
            addAgent(agentData);
        });

        // Add connections
        agentDataList.forEach(agentData => {
            agentData.connections.forEach(connectedId => {
                addConnection(agentData.id, connectedId);
            });
        });
    }

    const generateTestData = () => {
        const agentTypes = ['Controller', 'Worker', 'Analyzer', 'Scheduler'];
        const numAgents = 30;

        for (let i = 0; i < numAgents; i++) {
            const type = agentTypes[Math.floor(Math.random() * agentTypes.length)];
            const name = `${type}${i + 1}`;
            const position = new THREE.Vector3();
            const time = Math.random() * 100;

            const agentData: AgentData = {
                id: name,
                position: position,
                connections: [],
                type: type,
                time: time
            };
            agentDataList.push(agentData);
        }

        // Add some random connections
        const numConnections = 45;
        for (let i = 0; i < numConnections; i++) {
            const agent1 = agentDataList[Math.floor(Math.random() * agentDataList.length)];
            const agent2 = agentDataList[Math.floor(Math.random() * agentDataList.length)];
            if (agent1 !== agent2 && !agent1.connections.includes(agent2.id)) {
                agent1.connections.push(agent2.id);
                agent2.connections.push(agent1.id);
            }
        }
    }

    const generateAgentResponse = (agentId: string) => {
        console.log("Generating agent response for", agentId);
        // This is a placeholder. In a real system, you'd get the actual response from the LLM.
        const responses = [
            "Processing data...",
            "Analyzing patterns...",
            "Generating report...",
            "Collaborating with other agents...",
            "Learning from new information..."
        ];
        return responses[Math.floor(Math.random() * responses.length)];
    }

    const showAgentResponse = (agentId: string) => {
        const agent = agents.get(agentId);
        if (agent) {
            const response = generateAgentResponse(agentId);
            const bubble = new SpeechBubble(response, agent.position, scene);
            speechBubbles.push(bubble);
        }
    }

    const startSimulation = () => {
        // Spawn a new agent every few seconds
        setInterval(() => {
            if (!isSimulationPaused) {
                spawnRandomAgent();
            }
        }, 5000);

        // Remove an agent periodically
        setInterval(() => {
            if (!isSimulationPaused) {
                removeRandomAgent();
            }
        }, 8000);
    }

    const spawnRandomAgent = () => {
        const agentTypes = ['Controller', 'Worker', 'Analyzer', 'Scheduler'];
        const type = agentTypes[Math.floor(Math.random() * agentTypes.length)];
        const id = `${type}${Math.floor(Math.random() * 1000)}`;
        const position = new THREE.Vector3(
            (Math.random() - 0.5) * 100,
            (Math.random() - 0.5) * 100,
            (Math.random() - 0.5) * 100
        );
        const time = Date.now() % 100; // Use timestamp for unique time

        const newAgentData: AgentData = {
            id: id,
            position: position,
            connections: [],
            type: type,
            time: time
        };

        // Add to agent data list and scene
        agentDataList.push(newAgentData);
        addAgent(newAgentData);

        // Randomly connect to an existing agent
        if (agentDataList.length > 1) {
            const existingAgents = agentDataList.filter(a => a.id !== id);
            const randomAgent = existingAgents[Math.floor(Math.random() * existingAgents.length)];
            newAgentData.connections.push(randomAgent.id);
            randomAgent.connections.push(id);
            addConnection(id, randomAgent.id);
        }
        showAgentResponse(newAgentData.id);
        // Reapply layout
        applyLayout(currentLayoutName);
    }

    const removeRandomAgent = () => {
        if (agentDataList.length === 0) return;

        const index = Math.floor(Math.random() * agentDataList.length);
        const agentData = agentDataList[index];
        const agentId = agentData.id;

        // Remove from data list
        agentDataList.splice(index, 1);

        // Remove agent mesh
        const agentMesh = agents.get(agentId);
        if (agentMesh) {
            scene.remove(agentMesh);
            agents.delete(agentId);
        }

        // Remove related connections
        connections = connections.filter(connection => {
            if (connection.userData.start === agentId || connection.userData.end === agentId) {
                scene.remove(connection);
                return false;
            }
            return true;
        });

        // Update other agents' connections
        agentDataList.forEach(agent => {
            agent.connections = agent.connections.filter(id => id !== agentId);
        });

        // Reapply layout
        applyLayout(currentLayoutName);
    }

    const showAgentControlPanel = (agentId: string) => {
        // Create a control panel if it doesn't exist
        if (!agentControlPanel) {
            agentControlPanel = document.createElement('div');
            agentControlPanel.style.position = 'absolute';
            agentControlPanel.style.top = '50px';
            agentControlPanel.style.right = '10px';
            agentControlPanel.style.background = 'rgba(0, 0, 0, 0.7)';
            agentControlPanel.style.color = '#fff';
            agentControlPanel.style.padding = '10px';
            agentControlPanel.style.borderRadius = '5px';
            document.body.appendChild(agentControlPanel);
        }

        agentControlPanel.innerHTML = `<h3>Agent: ${agentId}</h3>`;

        // Add controls (e.g., buttons, sliders)
        const spawnButton = document.createElement('button');
        spawnButton.textContent = 'Spawn Child Agent';
        spawnButton.onclick = () => {
            spawnChildAgent(agentId);
        };
        agentControlPanel.appendChild(spawnButton);

        // Show the control panel
        agentControlPanel.style.display = 'block';
    }

    const hideAgentControlPanel = () => {
        if (agentControlPanel) {
            agentControlPanel.style.display = 'none';
        }
    }

    const spawnChildAgent = (parentAgentId: string) => {
        const parentAgentData = agentDataList.find(agent => agent.id === parentAgentId);
        if (!parentAgentData) return;

        const childId = `${parentAgentId}-child${Math.floor(Math.random() * 1000)}`;
        const position = parentAgentData.position.clone().add(new THREE.Vector3(
            (Math.random() - 0.5) * 20,
            (Math.random() - 0.5) * 20,
            (Math.random() - 0.5) * 20
        ));

        const childAgentData: AgentData = {
            id: childId,
            position: position,
            connections: [parentAgentId],
            type: 'Temporary',
            time: Date.now() % 100,
            parentId: parentAgentId
        };

        agentDataList.push(childAgentData);
        addAgent(childAgentData);
        addConnection(parentAgentId, childId);

        // Reapply layout to accommodate new agent
        applyLayout(currentLayoutName);
    }

    const applyTheme = (theme: string) => {
        switch (theme) {
            case 'dark':
                document.documentElement.style.setProperty('--background-color', '#000000');
                document.documentElement.style.setProperty('--text-color', '#ffffff');
                document.documentElement.style.setProperty('--control-background', 'rgba(0, 0, 0, 0.7)');
                break;
            case 'neon':
                document.documentElement.style.setProperty('--background-color', '#1a1a1a');
                document.documentElement.style.setProperty('--text-color', '#39ff14');
                document.documentElement.style.setProperty('--control-background', 'rgba(0, 0, 0, 0.7)');
                break;
            default:
                document.documentElement.style.setProperty('--background-color', '#000011');
                document.documentElement.style.setProperty('--text-color', '#ffffff');
                document.documentElement.style.setProperty('--control-background', 'rgba(0, 0, 0, 0.7)');
                break;
        }
        scene.background = new THREE.Color(getComputedStyle(document.documentElement).getPropertyValue('--background-color'));
    }

    // Initialize the scene immediately
    initializeScene();

    // Return a container element for the timeline
    return (<div id="timeline-container"></div>);
}
