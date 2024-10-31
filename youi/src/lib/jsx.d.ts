declare namespace JSX {
    // Define what a JSX.Element is in our system
    interface Element extends Node { }

    // Define the base interface for element attributes
    interface ElementAttributesProperty {
        props: {};
    }

    // Define all valid HTML elements
    interface IntrinsicElements extends HTMLElementTagNameMap {
        [elemName: string]: any;
    }
}

// Tell TypeScript to use our JSX namespace
declare module "*.tsx" {
    const content: any;
    export default content;
}