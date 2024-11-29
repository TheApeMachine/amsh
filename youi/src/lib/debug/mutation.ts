export const setupMutationObserver = ({ addLog }: { addLog: (entry: any) => void }) => {
    let domTimeout: ReturnType<typeof setTimeout>;  // Use ReturnType instead of NodeJS.Timeout
    const observer = new MutationObserver((mutations) => {
        clearTimeout(domTimeout);
        domTimeout = setTimeout(() => {
            for (const mutation of mutations) {
                if (mutation.type === 'attributes') {
                    const element = mutation.target as Element;
                    const value = element.getAttribute(mutation.attributeName ?? '');

                    addLog({
                        id: crypto.randomUUID(),
                        timestamp: new Date().toISOString(),
                        type: 'dom-change',
                        category: 'dom.attribute',
                        summary: `${element.tagName.toLowerCase()}.${mutation.attributeName} = ${value?.slice(0, 50)}${value && value.length > 50 ? '...' : ''}`,
                        details: {
                            element: element.outerHTML,
                            attribute: mutation.attributeName,
                            value: value
                        }
                    });
                }
            }
        }, 100);
    });

    observer.observe(document.body, {
        attributes: true,
        childList: true,
        subtree: true,
        characterData: true
    });
};