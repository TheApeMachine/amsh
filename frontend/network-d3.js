< !--File: network - graph - visualization.js-- >

    class NetworkGraphVisualization extends HTMLElement {
        constructor() {
            super();
            this.attachShadow({ mode: 'open' });
            this.nodes = [];
            this.links = [];
        }

        connectedCallback() {
            this.render();
            this.initializeGraph();
        }

        render() {
            this.shadowRoot.innerHTML = `
        <style>
            :host {
                display: block;
                width: 100%;
                height: 100%;
            }
            svg {
                width: 100%;
                height: 100%;
            }
                /* In the component's <style> */
                .nodes circle {
                    stroke: #fff;
                    stroke-width: 1.5px;
                }

                .links line {
                    stroke: #999;
                    stroke-opacity: 0.6;
                }

                .tooltip {
                    position: absolute;
                    text-align: center;
                    width: 120px;
                    padding: 4px;
                    font: 12px sans-serif;
                    background: lightsteelblue;
                    border: 0px;
                    border-radius: 8px;
                    pointer-events: none;
                }

        </style>
        <svg></svg>
        `;
        }

        initializeGraph() {
            const svg = d3.select(this.shadowRoot).select('svg');
            const width = this.clientWidth;
            const height = this.clientHeight;

            const simulation = d3.forceSimulation(this.nodes)
                .force('link', d3.forceLink(this.links).id(d => d.id).distance(50))
                .force('charge', d3.forceManyBody().strength(-300))
                .force('center', d3.forceCenter(width / 2, height / 2));

            const link = svg.append('g')
                .attr('class', 'links')
                .selectAll('line')
                .data(this.links)
                .enter().append('line')
                .attr('stroke-width', d => Math.sqrt(d.value));

            const node = svg.append('g')
                .attr('class', 'nodes')
                .selectAll('circle')
                .data(this.nodes)
                .enter().append('circle')
                .attr('r', 5)
                .attr('fill', d => d.color)
                .call(drag(simulation));

            node.append('title')
                .text(d => d.id);

            simulation.on('tick', () => {
                link
                    .attr('x1', d => d.source.x)
                    .attr('y1', d => d.source.y)
                    .attr('x2', d => d.target.x)
                    .attr('y2', d => d.target.y);

                node
                    .attr('cx', d => d.x)
                    .attr('cy', d => d.y);
            });

            function drag(simulation) {
                function dragstarted(event) {
                    if (!event.active) simulation.alphaTarget(0.3).restart();
                    event.subject.fx = event.subject.x;
                    event.subject.fy = event.subject.y;
                }

                function dragged(event) {
                    event.subject.fx = event.x;
                    event.subject.fy = event.y;
                }

                function dragended(event) {
                    if (!event.active) simulation.alphaTarget(0);
                    event.subject.fx = null;
                    event.subject.fy = null;
                }

                return d3.drag()
                    .on('start', dragstarted)
                    .on('drag', dragged)
                    .on('end', dragended);
            }
        }

        addThought(thought) {
            // Add node
            this.nodes.push({
                id: thought.id,
                group: thought.agentName,
                content: thought.content,
                color: this.getColorForAgent(thought.agentName),
            });

            // Add links (you may need to define how links are created)
            // For example, link to previous thought
            if (this.nodes.length > 1) {
                this.links.push({
                    source: this.nodes[this.nodes.length - 2].id,
                    target: thought.id,
                    value: 1,
                });
            }

            // Restart the simulation
            this.initializeGraph();
        }

        getColorForAgent(agentName) {
            // Assign colors to agents
            const agentColors = {
                'reasoner': '#6E95F7',
                'verifier': '#F7746D',
                'learning': '#F7B96D',
                'metacognition': '#06C26F',
                'prompt_engineer': '#FFB347',
                'context_manager': '#F76D95',
            };
            return agentColors[agentName] || '#000000';
        }

        handleEvents(node) {
            node.on('mouseover', (event, d) => {
                // Show tooltip
                const tooltip = d3.select(this.shadowRoot)
                    .append('div')
                    .attr('class', 'tooltip')
                    .style('left', `${event.pageX}px`)
                    .style('top', `${event.pageY}px`)
                    .html(`<strong>${d.id}</strong><br>${d.content}`);
            })
                .on('mouseout', () => {
                    // Remove tooltip
                    d3.select(this.shadowRoot).select('.tooltip').remove();
                });

            const zoom = d3.zoom()
                .scaleExtent([0.1, 10])
                .on('zoom', (event) => {
                    svg.attr('transform', event.transform);
                });

            svg.call(zoom);

            simulation.force('cluster', clustering());

            function clustering() {
                // Implement clustering logic
            }

        }

    }

customElements.define('network-graph-visualization', NetworkGraphVisualization);
