/*
html is a function that takes a TemplateStringsArray and any number of values and returns a DocumentFragment.
@param strings The TemplateStringsArray to render.
@param values The values to render.
@returns A DocumentFragment containing the rendered TemplateStringsArray and values.
*/
export const html = (strings: TemplateStringsArray, ...values: any[]): DocumentFragment => {
    const template = document.createElement("template");

    const htmlString = strings.reduce((result, string, i) => {
        let value = values[i];
        
        // Handle arrays
        if (Array.isArray(value)) {
            return `${result}${string}${value.join("")}`;
        }

        // Handle DocumentFragments
        if (value instanceof DocumentFragment) {
            return `${result}${string}`;
        }

        // Handle strings and other values
        return `${result}${string}${options.sanitize ? sanitizeHTML(value) : value}`;
    }, "");

    template.innerHTML = htmlString.trim();
    const fragment = document.importNode(template.content, true);

    return fragment;
};

/* Define the options object with necessary properties */
const options = {
    eventPrefix: 'event-', // Prefix for event attributes
    sanitize: true // Enable HTML sanitization
};

/*
sanitizeHTML is a function that takes a string and returns a sanitized string.
@param str The string to sanitize.
@returns A sanitized string.
*/
export const sanitizeHTML = (str: string): string => {
    const temp = document.createElement('div');
    temp.textContent = str;
    return temp.innerHTML;
}

export function jsx(tag: string | Function, props: any, ...children: any[]) {
    if (typeof tag === 'function') {
        return tag({ ...props, children });
    }

    const element = document.createElement(tag);

    for (const [name, value] of Object.entries(props || {})) {
        if (name.startsWith('on') && typeof value === 'function') {
            element.addEventListener(name.slice(2).toLowerCase(), value);
        } else {
            element.setAttribute(name, value);
        }
    }

    for (const child of children) {
        element.appendChild(
            child instanceof Node ? child : document.createTextNode(child)
        );
    }

    return element;
}