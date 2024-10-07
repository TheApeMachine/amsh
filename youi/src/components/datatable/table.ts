import * as echarts from 'echarts';
import * as aq from 'arquero';

const sanitizeHeader = (header: string): string => {
    return header.replace(/\s+/g, '-').replace(/[^a-zA-Z0-9-_]/g, '').toLowerCase();
};

class DataTable extends HTMLElement {
    private arqueroTable: aq.Table | null = null;
    private table: HTMLTableElement;
    private thead: HTMLTableSectionElement;
    private tbody: HTMLTableSectionElement;
    private originalHeaders: string[] = [];
    private sanitizedHeaders: string[] = [];
    private template: HTMLTemplateElement;

    constructor() {
        super();
        this.template = document.createElement("template")
        this.template.innerHTML = `
        <style>
            :host {
                width: 100%;
                background: var(--white);
                border-radius: var(--border-radius);
                border-spacing: 0;
                text-align: left;
                overflow-y: auto;
            }
            table {
                width: 100%;
                border-collapse: collapse;
            }
            thead {
                position: sticky;
                top: 0;
                z-index: 1;
                background: var(--subtle);
            }
            th, td {
                padding: 0.5rem;
                border-bottom: 1px solid hsl(0, 0%, 90%);
                font-size: 0.75rem;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
                max-width: 0;
            }
            th {
                background: hsl(0, 0%, 90%);
                text-transform: uppercase;
            }
            th h4 {
                padding: 0.25rem;
                background: hsl(0, 50%, 90%);
            }
            th div {
                margin-top: 0.5rem;
            }
            td {
                padding: 0.5rem;
                border-bottom: 1px solid hsl(0, 0%, 90%);
            }
            dl {
                display: flex;
                flex-flow: row wrap;
                align-items: center;
            }
            dt {
                flex-basis: 20%;
                padding: 0.5rem;
                text-align: left;
            }
            dd {
                flex-basis: 70%;
                flex-grow: 1;
                margin: 0;
                padding: 0.25rem;
                text-align: right;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }
        </style>
        <table>
            <thead>
            </thead>
            <tbody>
            </tbody>
            <tfoot>
            </tfoot>
        </table>
        `
        this.attachShadow({ mode: 'open' });
        console.debug('DataTable initialized', 'DataTable');
    }

    connectedCallback() {
        this.shadowRoot!.appendChild(this.template.content.cloneNode(true));
        this.table = this.shadowRoot!.querySelector('table')!;
        this.thead = this.table.querySelector('thead')!;
        this.tbody = this.table.querySelector('tbody')!;
        const observeKey = this.getObserveKey();
        console.log('DataTable.connectedCallback.observeKey', observeKey);

        if (observeKey) {
            window.stateManager.subscribe(observeKey, this.handleStateChange);
            const initialData = window.stateManager.getState(observeKey);
            if (initialData) {
                this.setData(initialData);
            }
        }
    }

    disconnectedCallback() {
        const observeKey = this.getObserveKey();
        console.log('DataTable.disconnectedCallback.observeKey', observeKey);
    }

    private getObserveKey(): string | null {
        return this.getAttribute('observe');
    }

    private handleStateChange = (data: any) => {
        console.log('DataTable.handleStateChange', data);
        this.setData(data.users);
    };

    public setData(data: any[]) {
        console.log('DataTable.setData', data);

        if (!data.length) {
            this.renderPlaceholder('No data available');
            return;
        }

        this.arqueroTable = aq.from(data);
        this.originalHeaders = Object.keys(data[0]);
        this.sanitizedHeaders = this.originalHeaders.map(header => sanitizeHeader(header));
        this.renderTableHeader();
        this.renderTableBody();
        this.initializeCharts();
    }

    private renderPlaceholder(message: string) {
        this.tbody.innerHTML = `<tr><td colspan="${this.originalHeaders.length}">${message}</td></tr>`;
    }

    private renderTableHeader() {
        this.thead.innerHTML = `
            <tr>
                ${this.originalHeaders.map((header, index) => {
                    const sanitizedHeader = this.sanitizedHeaders[index];
                    const columnData = this.arqueroTable!.array(header) as any[];
                    const columnType = this.determineColumnType(columnData);
                    const summaryStats = this.calculateSummaryStatsForColumn(columnData, columnType);

                    let summaryDisplay;
                    switch (columnType) {
                        case 'number':
                            summaryDisplay = `
                                <dt part="dt">Min</dt>
                                <dd part="dd">${summaryStats.min}</dd>
                                <dt part="dt">Max</dt>
                                <dd part="dd">${summaryStats.max}</dd>
                                <dt part="dt">Mean</dt>
                                <dd part="dd">${summaryStats.mean.toFixed(2)}</dd>
                            `;
                            break;
                        case 'string':
                            summaryDisplay = `
                                <dt part="dt">Mode</dt>
                                <dd part="dd">${summaryStats.mode}</dd>
                                <dt part="dt">Unique</dt>
                                <dd part="dd">${summaryStats.uniqueCount}</dd>
                                <dt part="dt">Total</dt>
                                <dd part="dd">${summaryStats.totalCount}</dd>
                            `;
                            break;
                        case 'boolean':
                            summaryDisplay = `
                                <dt part="dt">True</dt>
                                <dd part="dd">${summaryStats.trueCount}</dd>
                                <dt part="dt">False</dt>
                                <dd part="dd">${summaryStats.falseCount}</dd>
                                <dt part="dt">Total</dt>
                                <dd part="dd">${summaryStats.totalCount}</dd>
                            `;
                            break;
                        case 'datetime':
                            summaryDisplay = `
                                <dt part="dt">Earliest</dt>
                                <dd part="dd">${summaryStats.earliest}</dd>
                                <dt part="dt">Latest</dt>
                                <dd part="dd">${summaryStats.latest}</dd>
                                <dt part="dt">Range</dt>
                                <dd part="dd">${summaryStats.range} days</dd>
                            `;
                            break;
                        default:
                            summaryDisplay = `<dt part="dt">-</dt><dd part="dd">-</dd>`;
                    }

                    return `
                        <th id="header-${sanitizedHeader}">
                            <h4>${header}</h4>
                            <dl>
                                <dt part="dt">Type</dt>
                                <dd part="dd">
                                    <select part="select">
                                        <option value="number" ${columnType === 'number' ? 'selected' : ''}>Numeric</option>
                                        <option value="string" ${columnType === 'string' ? 'selected' : ''}>String</option>
                                        <option value="boolean" ${columnType === 'boolean' ? 'selected' : ''}>Boolean</option>
                                        <option value="datetime" ${columnType === 'datetime' ? 'selected' : ''}>Datetime</option>
                                    </select>
                                </dd>
                                ${summaryDisplay}
                            </dl>
                            <div id="chart-${sanitizedHeader}" style="width: 100%; height: 50px;"></div>
                        </th>
                    `;
                }).join('')}
            </tr>
        `;
        this.addEventListenersForSorting();
        this.addEventListenersForFiltering();
    }

    private renderTableBody() {
        if (!this.arqueroTable) return;

        const validRows = this.arqueroTable.objects().filter(row =>
            Object.values(row).some(value => value !== null && value !== undefined)
        );

        this.tbody.innerHTML = validRows.map((row: any) => `
            <tr>
                ${this.originalHeaders.map(header => `
                    <td>
                        ${row[header] !== undefined ? row[header] : ''}
                    </td>
                `).join('')}
            </tr>
        `).join('');
    }

    private initializeCharts() {
        requestAnimationFrame(() => {
            this.sanitizedHeaders.forEach(sanitizedHeader => {
                this.createMiniChart(sanitizedHeader);
            });
        });
    }

    private createMiniChart(sanitizedHeader: string) {
        if (!this.arqueroTable) return;

        const columnData = this.arqueroTable.array(sanitizedHeader) as any[];
        const columnType = this.determineColumnType(columnData);

        let chartData = [];

        switch (columnType) {
            case 'number':
                chartData = this.binNumericData(Array.from(columnData).filter((v): v is number => typeof v === 'number'));
                break;
            case 'datetime':
                chartData = this.binDatetimeData(Array.from(columnData).map(value => new Date(value as string)));
                break;
            case 'string':
                chartData = this.binStringData(Array.from(columnData).filter((v): v is string => typeof v === 'string'));
                break;
            default:
                console.warn(`Unsupported column type: ${columnType} for header "${sanitizedHeader}"`);
                return;
        }

        const chartDom = this.shadowRoot!.getElementById(`chart-${sanitizedHeader}`);
        if (!chartDom) return;

        try {
            const myChart = echarts.init(chartDom, 'light');

            const option = {
                tooltip: {
                    confine: true,
                    textStyle: {
                        fontSize: 10,
                    },
                },
                xAxis: {
                    type: 'category',
                    data: chartData.map(row => row.label),
                },
                yAxis: {
                    type: 'value',
                    show: false,
                },
                series: [{
                    type: 'bar',
                    data: chartData.map(row => row.value),
                }],
            };

            myChart.setOption(option);
            window.addEventListener('resize', () => {
                myChart.resize();
            });
        } catch (error: any) {
            console.error(`Failed to initialize chart for header "${sanitizedHeader}": ${error.message}`, 'DataTable');
        }
    }

    private binNumericData(data: number[]) {
        const min = Math.min(...data);
        const max = Math.max(...data);
        const binCount = 10;
        const binSize = (max - min) / binCount;

        const bins = Array.from({ length: binCount }, (_, i) => ({
            range: `${(min + i * binSize).toFixed(2)} - ${(min + (i + 1) * binSize).toFixed(2)}`,
            count: 0,
        }));

        data.forEach(value => {
            const binIndex = Math.min(Math.floor((value - min) / binSize), binCount - 1);
            bins[binIndex].count++;
        });

        return bins.map(bin => ({
            label: bin.range,
            value: bin.count,
        }));
    }

    private binDatetimeData(data: Date[]) {
        const binnedData = data.reduce((acc, date) => {
            const year = date.getFullYear();
            if (!acc[year]) acc[year] = 0;
            acc[year]++;
            return acc;
        }, {} as Record<string, number>);

        return Object.keys(binnedData).map(year => ({
            label: year,
            value: binnedData[year],
        }));
    }

    private binStringData(data: string[]) {
        const frequency = data.reduce((acc, value) => {
            if (!acc[value]) acc[value] = 0;
            acc[value]++;
            return acc;
        }, {} as Record<string, number>);

        const sortedKeys = Object.keys(frequency).sort((a, b) => frequency[b] - frequency[a]);

        return sortedKeys.map(key => ({
            label: key,
            value: frequency[key],
        }));
    }

    private determineColumnType(columnData: any[]): string {
        const nonNullData = columnData.filter(value => value !== null && value !== undefined);
        if (nonNullData.every(value => typeof value === 'number')) return 'number';
        if (nonNullData.every(value => typeof value === 'boolean')) return 'boolean';
        if (nonNullData.every(value => typeof value === 'string' && this.isValidDate(value))) return 'datetime';
        return 'string';
    }

    private isValidDate(dateString: string): boolean {
        const date = new Date(dateString);
        return !isNaN(date.getTime());
    }

    calculateSummaryStatsForColumn(columnData: any[], columnType: string) {
        switch (columnType) {
            case 'number':
                const numericData = columnData.filter((value) => typeof value === 'number' && !isNaN(value));
                if (numericData.length === 0) return { min: 0, max: 0, mean: 0 };
    
                const min = Math.min(...numericData);
                const max = Math.max(...numericData);
                const mean = numericData.reduce((sum, value) => sum + value, 0) / numericData.length;
    
                return { min, max, mean };
    
            case 'string':
                const uniqueValues = Array.from(new Set(columnData));
                const mostCommonValue = uniqueValues.sort((a, b) =>
                    columnData.filter(v => v === b).length - columnData.filter(v => v === a).length
                )[0];
    
                return { mode: mostCommonValue, uniqueCount: uniqueValues.length, totalCount: columnData.length };
    
            case 'boolean':
                const trueCount = columnData.filter(value => value === true).length;
                const falseCount = columnData.filter(value => value === false).length;
    
                return { trueCount, falseCount, totalCount: columnData.length };
    
            case 'datetime':
                const dateData = columnData
                    .map(value => new Date(value))
                    .filter(date => !isNaN(date.getTime()));
    
                if (dateData.length === 0) return { earliest: '-', latest: '-', range: '-' };
    
                const earliest = new Date(Math.min(...dateData));
                const latest = new Date(Math.max(...dateData));
                const range = latest.getTime() - earliest.getTime();
    
                return { earliest: earliest.toISOString().split('T')[0], latest: latest.toISOString().split('T')[0], range };
    
            default:
                return {};
        }
    }

    private addEventListenersForSorting() {
        this.thead.querySelectorAll('.sort-asc').forEach(button => {
            button.addEventListener('click', () => {
                const sanitizedHeader = button.closest('th')!.id.replace('header-', '');
                this.sortTable(sanitizedHeader, 'asc');
            });
        });

        this.thead.querySelectorAll('.sort-desc').forEach(button => {
            button.addEventListener('click', () => {
                const sanitizedHeader = button.closest('th')!.id.replace('header-', '');
                this.sortTable(sanitizedHeader, 'desc');
            });
        });
    }

    private addEventListenersForFiltering() {
        this.thead.querySelectorAll('.filter-input').forEach(input => {
            const debouncedFilter = this.debounce(() => {
                const sanitizedHeader = input.closest('th')!.id.replace('header-', '');
                const query = (input as HTMLInputElement).value.toLowerCase();
                this.filterTable(sanitizedHeader, query);
            }, 300);
            input.addEventListener('input', debouncedFilter);
        });
    }

    private debounce(func: Function, delay: number) {
        let timer: NodeJS.Timeout;
        return (...args: any[]) => {
            clearTimeout(timer);
            timer = setTimeout(() => func.apply(this, args), delay);
        };
    }

    private sortTable(sanitizedHeader: string, direction: 'asc' | 'desc') {
        this.arqueroTable = this.arqueroTable!.orderby(
            direction === 'asc' ? aq.asc(sanitizedHeader) : aq.desc(sanitizedHeader)
        );
        this.renderTableBody();
    }

    private filterTable(sanitizedHeader: string, query: string) {
        this.arqueroTable = this.arqueroTable!.filter(
            (d: any) => String(d[sanitizedHeader]).toLowerCase().includes(query)
        );
        this.renderTableBody();
    }
}

customElements.define('data-table', DataTable);