import * as vg from "@uwdata/vgplot"

class BarPlot extends HTMLElement {
    private template: HTMLTemplateElement;
    private chart: any;
    constructor() {
        super();
        this.template = document.createElement("template");
        this.template.innerHTML = `
            <div class="bar-chart"></div>
        `;
        this.attachShadow({ mode: "open" });
    }

    async connectedCallback() {
        this.shadowRoot!.appendChild(this.template.content.cloneNode(true));
        await this.render();
    }

    private async render() {
        const coordinator = vg.coordinator();
        await coordinator.databaseConnector(vg.wasmConnector());

        // Load the CSV data
        const stocksData = await vg.loadCSV("https://raw.githubusercontent.com/uwdata/mosaic/refs/heads/main/data/stocks.csv");

        // Create the plot
        this.chart = vg.plot(
            vg.areaY(stocksData, {x: "Date", y: "Price"})
        );

        // Append the chart to the shadow DOM
        const chartContainer = this.shadowRoot!.querySelector('.bar-chart');
        chartContainer!.appendChild(this.chart);
    }
}

customElements.define("bar-plot", BarPlot);