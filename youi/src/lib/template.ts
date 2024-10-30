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

// Add Fragment type
export const Fragment = Symbol('Fragment');

type JSXElementType = string | Function | typeof Fragment;

// Update event handler types to be more specific
type EventHandlers = {
    [K in keyof HTMLElementEventMap]?: (event: HTMLElementEventMap[K]) => void;
};

type Props = {
    [key: string]: any;
} & {
    [K in keyof EventHandlers as `on${Capitalize<K>}`]?: EventHandlers[K];
};

/*
jsx is a function that takes a tag, props, and children and returns a Node.
@param tag The tag to render.
@param props The props to render.
@param children The children to render.
@returns A Node containing the rendered tag, props, and children.
*/
export function jsx(
    tag: JSXElementType,
    props: Props | null,
    ...children: (Node | string | Array<Node | string>)[]
) {
    // Handle Fragments
    if (tag === Fragment) {
        const fragment = document.createDocumentFragment();
        children.forEach(child => {
            fragment.appendChild(
                child instanceof Node ? child : document.createTextNode(String(child))
            );
        });
        return fragment;
    }

    // Handle function components
    if (typeof tag === 'function') {
        // Handle prop spreading
        const componentProps = props ? { ...props } : {};

        // If there are children, add them to props
        if (children.length > 0) {
            componentProps.children = children.length === 1 ? children[0] : children;
        }

        return tag(componentProps);
    }

    const element = document.createElement(tag);

    if (props) {
        // Handle prop spreading for HTML elements
        Object.entries(props).forEach(([name, value]) => {
            if (name === 'className') {
                element.setAttribute('class', String(value));
            } else if (name.startsWith('on') && typeof value === 'function') {
                element.addEventListener(
                    name.slice(2).toLowerCase(),
                    value as EventListener
                );
            } else if (typeof value === 'boolean') {
                if (value) {
                    element.setAttribute(name, '');
                } else {
                    element.removeAttribute(name);
                }
            } else {
                element.setAttribute(name, String(value));
            }
        });
    }

    children.forEach(child => {
        if (child instanceof Node) {
            element.appendChild(child);
        } else {
            element.appendChild(document.createTextNode(String(child)));
        }
    });

    return element;
}