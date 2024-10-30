/*
Matches the input value against the patterns and returns the result of the render function if a match is found.
If no match is found, an empty DocumentFragment is returned.
@param value The input value to match against.
@param cases An array of Matcher tuples.
@returns A DocumentFragment containing the matched result
*/
export const match = (stateObj: { state: "loading" | "error" | "success", results: any }, handlers: Record<string, Function>): DocumentFragment => {
    const { state, results } = stateObj;
    const result = handlers[state](results);

    // Create a DocumentFragment to hold the result
    const fragment = document.createDocumentFragment();

    if (result instanceof HTMLElement) {
        fragment.appendChild(result);
    } else {
        // Handle primitive values
        const wrapper = document.createElement('div');
        wrapper.appendChild(document.createTextNode(String(result)));
        fragment.appendChild(wrapper);
    }

    return fragment;
};
