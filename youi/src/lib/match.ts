/*
Matches the input value against the patterns and returns the result of the render function if a match is found.
If no match is found, an empty DocumentFragment is returned.
@param value The input value to match against.
@param cases An array of Matcher tuples.
@returns The result of the render function if a match is found, otherwise an empty DocumentFragment.
*/
export const match = (stateObj: { state: "loading" | "error" | "success", results: any }, handlers: Record<string, Function>) => {
    const { state, results } = stateObj;
    console.debug("matcher", "match", stateObj, handlers)
    return handlers[state](results);
};
