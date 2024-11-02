// @ts-ignore: JSX factory import is used by the transpiler
import { Transition } from "@/lib/transition";

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

// Define base HTML attributes similar to React's approach
export interface HTMLAttributes extends Record<string, any> {
    className?: string;
    id?: string;
    style?: Partial<CSSStyleDeclaration> | string;
    role?: string;
    tabIndex?: number;
    ref?: ((el: HTMLElement) => void) | { current: HTMLElement | null }; // Added `ref` prop
}

// Update event handler types to be more specific
type EventHandlers = {
    [K in keyof HTMLElementEventMap]?: (event: HTMLElementEventMap[K]) => void;
};

// Combine HTML attributes with event handlers for our final Props type
type Props = HTMLAttributes & {
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
        children.flat().forEach(child => {
            fragment.appendChild(
                child instanceof Node ? child : document.createTextNode(String(child))
            );
        });
        return fragment;
    }

    // Handle function components
    if (typeof tag === 'function') {
        const componentProps = props ? { ...props } : {};
        if (children.length > 0) {
            componentProps.children = children.length === 1 ? children[0] : children;
        }
        return (tag as (props: HTMLAttributes & { children?: any }) => Node)(componentProps);
    }

    // Create the element
    const element = document.createElement(tag);

    if (props) {
        // Handle prop spreading for HTML elements
        Object.entries(props).forEach(([name, value]) => {
            if (name === 'className') {
                element.setAttribute('class', String(value));
            } else if (name.startsWith('on') && typeof value === 'function') {
                element.addEventListener(name.slice(2).toLowerCase(), value as EventListener);
            } else if (name === 'ref' && (typeof value === 'function' || typeof value === 'object')) {
                // Handle `ref`
                if (typeof value === 'function') {
                    value(element);
                } else if (typeof value === 'object' && value !== null) {
                    value.current = element;
                }
            } else if (typeof value === 'boolean') {
                if (value) {
                    element.setAttribute(name, '');
                } else {
                    element.removeAttribute(name);
                }
            } else if (name === 'transitionEnter' || name === 'transitionExit') {
                element.dataset[name] = JSON.stringify(value);
            } else {
                element.setAttribute(name, String(value));
            }
        });
    }

    // Flatten and append children
    children.flat().forEach(child => {
        if (Array.isArray(child)) {
            // Recursively flatten arrays
            child.forEach(subChild => {
                if (subChild instanceof Node) {
                    element.appendChild(subChild);
                } else {
                    element.appendChild(document.createTextNode(String(subChild)));
                }
            });
        } else if (child instanceof Node) {
            element.appendChild(child);
        } else {
            element.appendChild(document.createTextNode(String(child)));
        }
    });

    // Handle transitions
    if (props?.transitionEnter || props?.transitionExit) {
        Transition(
            element,
            {
                enter: props.transitionEnter || (() => { }),
                exit: props.transitionExit || (() => { })
            }
        );
    }

    return element;
}
