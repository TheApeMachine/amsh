interface JsonViewerOptions {
    initialExpandDepth?: number;
    maxStringLength?: number;
    showTypes?: boolean;
    theme?: 'dark' | 'light';
}

export const createJsonViewer = (json: any, options: JsonViewerOptions = {}) => {
    const {
        initialExpandDepth = 1,
        maxStringLength = 100,
        showTypes = true,
        theme = 'dark'
    } = options;

    const colors = theme === 'dark' ? {
        key: '#79b6f2',        // Light blue
        string: '#c3e88d',     // Light green
        number: '#f78c6c',     // Orange
        boolean: '#c792ea',    // Purple
        null: '#ff5370',       // Red
        type: '#546e7a',       // Gray
        bracket: '#89ddff',    // Cyan
        background: '#263238', // Dark blue-gray
        hover: '#37474f'       // Lighter blue-gray
    } : {
        key: '#2196f3',        // Blue
        string: '#4caf50',     // Green
        number: '#ff9800',     // Orange
        boolean: '#9c27b0',    // Purple
        null: '#f44336',       // Red
        type: '#90a4ae',       // Gray
        bracket: '#00bcd4',    // Cyan
        background: '#ffffff', // White
        hover: '#f5f5f5'       // Light gray
    };

    // Generate unique IDs for each node
    const generateId = (() => {
        let counter = 0;
        return () => `json-node-${counter++}`;
    })();

    const getType = (value: any): string => {
        if (value === null) return 'null';
        if (Array.isArray(value)) return 'array';
        return typeof value;
    };

    const getPreview = (value: any, type: string): string => {
        switch (type) {
            case 'array':
                return `Array(${value.length})`;
            case 'object':
                return `Object(${Object.keys(value).length})`;
            case 'string':
                if (value.length > maxStringLength) {
                    return `"${value.substring(0, maxStringLength)}..."`;
                }
                return `"${value}"`;
            case 'null':
                return 'null';
            default:
                return String(value);
        }
    };

    const getCollapsedPreview = (value: any, type: string, colors: any): string => {
        return `
            <span class="json-preview-content">${getPreview(value, type)}</span>
            <span style="color: ${colors.bracket}">${type === 'array' ? ']' : '}'}</span>
        `;
    };

    const formatValue = (value: any, type: string): string => {
        switch (type) {
            case 'string':
                return `<span style="color: ${colors.string}">"${escapeHtml(value)}"</span>`;
            case 'number':
                return `<span style="color: ${colors.number}">${value}</span>`;
            case 'boolean':
                return `<span style="color: ${colors.boolean}">${value}</span>`;
            case 'null':
                return `<span style="color: ${colors.null}">null</span>`;
            default:
                return escapeHtml(String(value));
        }
    };

    const escapeHtml = (str: string): string => {
        return str
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    };

    const getBracketOpen = (type: string): string => {
        return type === 'array' ? '[' : '{';
    };

    const getExpandedPreviewContent = (): string => {
        return '';
    };

    const getCollapsedPreviewContent = (value: any, type: string, colors: any): string => {
        return getCollapsedPreview(value, type, colors);
    };

    const getPreviewContent = (isExpanded: boolean, value: any, type: string, colors: any): string => {
        return isExpanded ?
            getExpandedPreviewContent() :
            getCollapsedPreviewContent(value, type, colors);
    };

    const renderJsonNode = (key: string | number | null, value: any, depth: number): string => {
        const type = getType(value);
        const id = generateId();
        const isExpanded = depth < initialExpandDepth;
        const indentation = '  '.repeat(depth);
        const hasChildren = type === 'object' || type === 'array';

        const typeLabel = showTypes ?
            `<span class="json-type" style="color: ${colors.type}">${type}</span>` : '';

        let toggleHtml = '<span class="json-toggle-placeholder" style="width: 12px; display: inline-block"></span>';
        if (hasChildren) {
            toggleHtml = `
                <span class="json-toggle" style="color: ${colors.bracket}">
                    ${isExpanded ? '▼' : '▶'}
                </span>
            `;
        }

        let html = `
            <div class="json-node" id="${id}" style="position: relative;">
                <div class="json-line" style="padding-left: ${depth * 16}px; cursor: ${hasChildren ? 'pointer' : 'default'};">
                    ${toggleHtml}
                    ${key !== null ? `
                        <span class="json-key" style="color: ${colors.key}">${key}</span>
                        <span style="color: ${colors.bracket}">: </span>
                    ` : ''}
                    ${typeLabel}
                    ${hasChildren ? `
                        <span class="json-preview">
                            <span style="color: ${colors.bracket}">
                                ${getBracketOpen(type)}
                            </span>
                            ${getPreviewContent(isExpanded, value, type, colors)}
                        </span>
                    ` : formatValue(value, type)}
                </div>`;

        if (hasChildren && isExpanded) {
            const entries = type === 'array' ?
                value.map((v: any, i: number) => [i, v]) :
                Object.entries(value);

            html += '<div class="json-children">';
            entries.forEach(([k, v]: [string | number, any]) => {
                html += renderJsonNode(k, v, depth + 1);
            });
            html += `${indentation}${type === 'array' ? ']' : '}'}</div>`;
        }

        html += '</div>';
        return html;
    };

    const createStyles = () => `
        .json-viewer {
            font-family: ui-monospace, SFMono-Regular, Monaco, monospace;
            font-size: 13px;
            line-height: 1.4;
            background: ${colors.background};
            border-radius: 4px;
            padding: 8px;
            overflow: auto;
        }
        .json-line {
            white-space: nowrap;
            transition: background-color 0.2s;
            border-radius: 2px;
            padding: 2px 4px;
        }
        .json-line:hover {
            background: ${colors.hover};
        }
        .json-toggle {
            display: inline-block;
            width: 12px;
            margin-right: 4px;
            cursor: pointer;
            transition: transform 0.2s;
        }
        .json-key {
            margin-right: 4px;
        }
        .json-type {
            font-size: 11px;
            margin-right: 4px;
            opacity: 0.7;
        }
        .json-preview-content {
            opacity: 0.7;
            margin: 0 4px;
        }
        .json-children {
            position: relative;
        }
        .json-children:before {
            content: '';
            position: absolute;
            left: 9px;
            top: 0;
            bottom: 0;
            width: 1px;
            background: ${colors.type};
            opacity: 0.3;
        }
    `;

    const initializeViewer = (container: HTMLElement) => {
        // Add styles
        const style = document.createElement('style');
        style.textContent = createStyles();
        document.head.appendChild(style);

        // Create viewer
        const viewer = document.createElement('div');
        viewer.className = 'json-viewer';
        viewer.innerHTML = renderJsonNode(null, json, 0);
        container.appendChild(viewer);

        // Add click handlers
        viewer.addEventListener('click', (e) => {
            const target = e.target as HTMLElement;
            const line = target.closest('.json-line');
            if (!line) return;

            const node = line.closest('.json-node');
            if (!node) return;

            const toggle = line.querySelector('.json-toggle');
            if (!toggle) return;

            const children = node.querySelector('.json-children') as HTMLElement;
            const preview = line.querySelector('.json-preview');
            if (!children || !preview) return;

            const isExpanded = toggle.textContent === '▼';
            toggle.textContent = isExpanded ? '▶' : '▼';

            if (isExpanded) {
                children.style.display = 'none';
                const nodeType = preview.parentElement?.querySelector('.json-type')?.textContent?.toLowerCase() ?? 'object';
                preview.innerHTML += `
                    <span class="json-preview-content">${getPreview(json, nodeType)}</span>
                    <span style="color: ${colors.bracket}">${nodeType === 'array' ? ']' : '}'}</span>
                `;
            } else {
                children.style.display = 'block';
                const nodeType = preview.parentElement?.querySelector('.json-type')?.textContent?.toLowerCase() ?? 'object';
                preview.innerHTML = `<span style="color: ${colors.bracket}">${nodeType === 'array' ? '[' : '{'}</span>`;
            }
        });

        return {
            destroy: () => {
                style.remove();
                viewer.remove();
            }
        };
    };

    return initializeViewer;
};