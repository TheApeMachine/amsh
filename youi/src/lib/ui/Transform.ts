import { TemplateObject } from "./types";
import { Elements } from './Transform/Elements';
import { NestedDocument } from './Transform/NestedDocument';

export interface Transform {
    path?: string;
    template: string | TemplateObject;
}

const { createFieldElement, createDynamicIsland, createBasicElement, parseTemplate } = await Elements();
const { handleNestedData } = await NestedDocument();

const isTemplateObject = (template: Transform['template']): template is TemplateObject =>
    typeof template !== 'string';

const addChildren = (element: HTMLElement, template: TemplateObject, data: any): void => {
    if (!template.children) return;

    // Handle single child or array of children
    const children = Array.isArray(template.children) ? template.children : [template.children];

    children.forEach(childTemplate => {
        const childElement = createBasicElement(childTemplate, data);
        if (!childElement) return;

        // Set content if it exists
        if (childTemplate.content) {
            childElement.textContent = parseTemplate(childTemplate.content, data);
        }

        // Set props if they exist
        if (childTemplate.props) {
            Object.entries(childTemplate.props).forEach(([key, value]) => {
                childElement.setAttribute(key, parseTemplate(value as string, data));
            });
        }

        // Recursively handle any level of nested children
        if (childTemplate.children) {
            addChildren(childElement, { children: childTemplate.children }, data);
        }

        element.appendChild(childElement);
    });
};

const handleBasicElement = (container: HTMLElement, config: TemplateObject, item: any): void => {
    // Make sure we have a tag to create the element
    if (!config.tag) {
        config.tag = 'div'; // Default to div if no tag specified
    }

    const element = createBasicElement(config, item);
    if (element) {
        if (config.children && Array.isArray(item?.collections)) {
            item.collections.forEach((collection: any) => {
                addChildren(element, config, collection);
            });
        } else if (config.text) {
            element.textContent = config.text;
        } else if (config.content) {
            element.textContent = parseTemplate(config.content, item);
        }

        container.appendChild(element);
    }
};

// Main transformation functions
const handleArrayData = async (
    container: HTMLElement,
    config: TemplateObject,
    data: any[]
): Promise<HTMLElement> => {
    if (!config) return container;

    for (const item of data) {
        if (config.component === 'DynamicIsland') {
            const element = await createDynamicIsland(config, item);
            container.appendChild(element);
        } else {
            handleBasicElement(container, config, item);
        }
    }

    return container;
};

const isBinaryId = (value: any): boolean => {
    return value?.sub_type === 3 && value?.buffer instanceof Uint8Array;
};

// Update the handleDocumentData function
const handleDocumentData = async (
    container: HTMLElement,
    config: TemplateObject,
    documentData: any,
    prefix: string = ''
): Promise<HTMLElement> => {
    for (const [key, value] of Object.entries(documentData)) {
        if (key.startsWith('_')) continue;

        const fieldKey = prefix ? `${prefix}.${key}` : key;

        if (value !== null && typeof value === 'object' && !isBinaryId(value)) {
            // Handle nested object
            await handleNestedData(fieldKey, config, value, container);
            // Don't create a form field for this nested object
            continue;
        }

        // Create form field for non-nested values
        const element = await createFieldElement(config, fieldKey, value);
        container.appendChild(element);
    }
    return container;
};

const handleObjectData = async (
    container: HTMLElement,
    config: TemplateObject,
    data: any,
    prefix: string = ''
): Promise<HTMLElement> => {
    if (!config) return container;

    if (data.document) {
        return handleDocumentData(container, config, data.document, prefix);
    }

    if (config.component === 'DynamicIsland') {
        const element = await createDynamicIsland(config, data);
        container.appendChild(element);
    } else {
        handleBasicElement(container, config, data);
    }

    return container;
};

const handleArrayOfTemplates = async (container: HTMLElement, config: any, data: any) => {
    // First, create all base containers
    for (const templateItem of config.template) {
        const element = createBasicElement(templateItem, data);
        if (element) {
            container.appendChild(element);
        }
    }

    // Then, after containers exist, handle the document data
    for (const templateItem of config.template) {
        if (templateItem.component === 'DynamicIsland') {
            const element = await createDynamicIsland(templateItem, templateItem.data || data);
            container.appendChild(element);
        } else if (templateItem.tag === 'form' && data.document) {
            const formElement = container.querySelector('form');
            if (formElement) {
                await handleObjectData(formElement, templateItem.transforms?.data?.template, data);
            }
        }
    }

    return container;
};

export const transformData = async (
    container: HTMLElement,
    config: any,
    data: any,
    path: string = '*'
): Promise<HTMLElement> => {
    if (!config) return container;

    if (config.template && Array.isArray(config.template)) {
        return handleArrayOfTemplates(container, config, data);
    }

    if (config.template) {
        const element = createBasicElement(config.template, data);
        if (element) container.appendChild(element);
        return container;
    }

    if (Array.isArray(data)) {
        return handleArrayData(container, config.transforms?.data?.template, data);
    }

    if (typeof data === 'object' && data !== null || data?.document) {
        return handleObjectData(container, config.transforms?.data?.template, data);
    }

    const targetData = path === '*'
        ? data
        : path.split('.').reduce((obj: any, key: string) => obj?.[key], data);

    if (config.transforms?.data?.template) {
        return handleArrayData(container, config.transforms.data.template, [targetData]);
    }

    if (isTemplateObject(config)) {
        const element = createBasicElement(config, targetData);
        if (element) container.appendChild(element);
        return container;
    }

    return container;
};
