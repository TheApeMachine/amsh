declare namespace JSX {
    // Define what a JSX.Element is in our system
    type Element = Node | Promise<Node> | JSX.Element | Promise<JSX.Element> | string | Promise<string> | number | Promise<number> | boolean | Promise<boolean> | Promise<Node | Element>;

    // Define the base interface for element attributes
    interface ElementAttributesProperty {
        props: {};
    }

    // Define all valid HTML elements
    interface IntrinsicElements {
        [elemName: string]: {
            children?: Element | Element[] | string | string[] | number | (string | number | Element | Element[])[] | null;
            [key: string]: any;
        };
    }

    // Add support for async components
    interface ElementClass {
        render?: () => Element;
    }

    interface IntrinsicAttributes {
        children?: Element | Element[];
    }

    interface ElementChildrenAttribute {
        children: Element | Element[] | string | string[] | number | (string | number | Element | Element[])[] | null;
    }

    // Remove the FunctionElement type to prevent circular references
    type FunctionElement = () => Element | Element[];
}
