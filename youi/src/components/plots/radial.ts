import * as echarts from 'echarts';

class RadialPlot extends HTMLElement {
    private template: HTMLTemplateElement;
    private chart: echarts.ECharts | null = null;
    private resizeObserver: ResizeObserver | null = null;

    constructor() {
        super();
        this.template = document.createElement("template");
        this.template.innerHTML = `
            <style>
                :host {
                    display: block;
                    width: 100%;
                    height: 100%;
                }
            </style>
            <div id="radialContainer" style="width: 100%; height: 100%;"></div>
        `;
        this.attachShadow({ mode: "open" });
    }

    connectedCallback() {
        this.shadowRoot!.appendChild(this.template.content.cloneNode(true));

        const container = this.shadowRoot!.getElementById('radialContainer')!;
        this.chart = echarts.init(container);

        this.renderChart();

        // Set up the ResizeObserver
        this.resizeObserver = new ResizeObserver(() => {
            this.chart?.resize();
        });
        this.resizeObserver.observe(container);
    }

    disconnectedCallback() {
        // Clean up the ResizeObserver when the element is removed from the DOM
        this.resizeObserver?.disconnect();
        this.chart?.dispose();
    }

    private renderChart() {
        type EChartsOption = echarts.EChartsOption;

        var data = [
        {
            name: 'Grandpa',
            children: [
            {
                name: 'Uncle Leo',
                value: 15,
                children: [
                {
                    name: 'Cousin Jack',
                    value: 2
                },
                {
                    name: 'Cousin Mary',
                    value: 5,
                    children: [
                    {
                        name: 'Jackson',
                        value: 2
                    }
                    ]
                },
                {
                    name: 'Cousin Ben',
                    value: 4
                }
                ]
            },
            {
                name: 'Aunt Jane',
                children: [
                {
                    name: 'Cousin Kate',
                    value: 4
                }
                ]
            },
            {
                name: 'Father',
                value: 10,
                children: [
                {
                    name: 'Me',
                    value: 5,
                    itemStyle: {
                    color: 'red'
                    }
                },
                {
                    name: 'Brother Peter',
                    value: 1
                }
                ]
            }
            ]
        },
        {
            name: 'Mike',
            children: [
            {
                name: 'Uncle Dan',
                children: [
                {
                    name: 'Cousin Lucy',
                    value: 3
                },
                {
                    name: 'Cousin Luck',
                    value: 4,
                    children: [
                    {
                        name: 'Nephew',
                        value: 2
                    }
                    ]
                }
                ]
            }
            ]
        },
        {
            name: 'Nancy',
            children: [
            {
                name: 'Uncle Nike',
                children: [
                {
                    name: 'Cousin Betty',
                    value: 1
                },
                {
                    name: 'Cousin Jenny',
                    value: 2
                }
                ]
            }
            ]
        }
        ];

        var option: EChartsOption;

        option = {
            visualMap: {
                type: 'continuous',
                min: 0,
                max: 10,
                inRange: {
                color: ['#2F93C8', '#AEC48F', '#FFDB5C', '#F98862']
                }
            },
            series: {
                type: 'sunburst',
                data: data,
                radius: [0, '90%'],
                label: {
                rotate: 'radial'
                }
            }
        };

        this.chart?.setOption(option);
    }
}

customElements.define("radial-plot", RadialPlot);