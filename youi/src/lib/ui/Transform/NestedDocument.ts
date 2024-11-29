import { TemplateObject } from "../types";
import { Elements } from "./Elements";
import gsap from "gsap";

const isBinaryId = (value: any): boolean => {
    return value?.sub_type === 3 && value?.buffer instanceof Uint8Array;
};

export const NestedDocument = async () => {
    const { createFieldElement } = await Elements();

    const createArrayItem = async (value: any[], i: number, fieldKey: string, config: TemplateObject, arrayContent: HTMLDivElement) => {
        const arrayItem = value[i];
        if (typeof arrayItem === 'object' && arrayItem !== null) {
            const itemWrapper = document.createElement('div');
            gsap.set(itemWrapper, {
                display: 'flex',
                flexDirection: 'column',
                boxShadow: 'var(--ring)',
                borderRadius: 'var(--xs)',
                margin: 'var(--md)'
            });

            // Create a header for each item
            const itemHeader = document.createElement('div');
            gsap.set(itemHeader, {
                display: 'flex',
                alignItems: 'center',
                cursor: 'pointer'
            });
            itemHeader.dataset.event = 'toggle';
            itemHeader.dataset.trigger = 'click';
            itemHeader.dataset.target = `#nested-${fieldKey.replace(/\./g, '-')}-${i}-content`;
            itemHeader.dataset.variant = 'vertical';

            const itemLabel = document.createElement('span');
            itemLabel.textContent = `Item ${i + 1}`;
            gsap.set(itemLabel, {
                fontWeight: '700',
                color: 'var(--gray-700)',
                padding: 'var(--md)',
            });
            itemHeader.appendChild(itemLabel);

            // Create content container for the item
            const itemContent = document.createElement('div');
            itemContent.id = `nested-${fieldKey.replace(/\./g, '-')}-${i}-content`;
            gsap.set(itemContent, {
                display: 'flex',
                flexDirection: 'column',
                gap: 'var(--md)',
                height: '0',
                overflow: 'hidden'
            });

            // Add the header and content container to the wrapper
            itemWrapper.appendChild(itemHeader);
            itemWrapper.appendChild(itemContent);

            // Move the item content creation into the content container
            for (const [itemKey, itemValue] of Object.entries(arrayItem)) {
                if (itemKey.startsWith('_')) continue;

                if (itemValue !== null && typeof itemValue === 'object' && !isBinaryId(itemValue)) {
                    await handleNestedData(`${fieldKey}.${i}.${itemKey}`, config, itemValue, itemContent);
                } else {
                    const element = await createFieldElement(config, itemKey, itemValue);
                    itemContent.appendChild(element);
                }
            }
            arrayContent.appendChild(itemWrapper);
        }
    }

    const handleNestedData = async (
        fieldKey: string,
        config: TemplateObject,
        value: any,
        container: HTMLElement
    ) => {
        // Find or create the appropriate parent container
        const parentKey = fieldKey.split('.').slice(0, -1).join('-');
        const parentContainer = parentKey
            ? document.querySelector(`#nested-${parentKey}`)
            : document.querySelector('.nested-documents');

        if (!parentContainer) {
            console.warn(`No container found for ${fieldKey}`);
            return;
        }

        // For array items, create individual nested documents
        if (Array.isArray(value)) {
            // Create array wrapper
            const arrayWrapper = document.createElement('div');
            arrayWrapper.id = 'nested-' + fieldKey.replace(/\./g, '-');
            gsap.set(arrayWrapper, {
                display: 'flex',
                flexDirection: 'column',
                gap: 'var(--sm)',
            });

            // Create array header
            const arrayHeader = document.createElement('div');
            gsap.set(arrayHeader, {
                backgroundColor: 'var(--muted)',
                borderRadius: 'var(--xs)',
                padding: 'var(--md)',
                cursor: 'pointer'
            });

            arrayHeader.dataset.event = 'toggle';
            arrayHeader.dataset.trigger = 'click';
            arrayHeader.dataset.target = '#nested-' + fieldKey.replace(/\./g, '-') + '-content';
            arrayHeader.dataset.variant = 'vertical';

            const arrayLabel = document.createElement('span');
            arrayLabel.textContent = fieldKey.split('.').pop() ?? fieldKey;
            gsap.set(arrayLabel, {
                fontWeight: '700',
                color: 'var(--gray-700)'
            });
            arrayHeader.appendChild(arrayLabel);

            const arrayContent = document.createElement('div');
            arrayContent.id = 'nested-' + fieldKey.replace(/\./g, '-') + '-content';
            gsap.set(arrayContent, {
                display: 'flex',
                flexDirection: 'column',
                height: '0',
                overflow: 'hidden',
            });

            // Create individual items
            for (let i = 0; i < value.length; i++) {
                await createArrayItem(value, i, fieldKey, config, arrayContent);
            }

            arrayWrapper.appendChild(arrayHeader);
            arrayWrapper.appendChild(arrayContent);
            parentContainer.appendChild(arrayWrapper);
            return;
        }

        // Handle non-array objects (existing code)
        const wrapper = document.createElement('div');
        gsap.set(wrapper, {
            display: 'flex',
            padding: 'var(--md)',
            flexDirection: 'column'
        });

        const header = document.createElement('div');
        gsap.set(header, {
            display: 'flex',
            alignItems: 'center',
            cursor: 'pointer'
        });
        header.dataset.event = 'toggle';
        header.dataset.trigger = 'click';
        header.dataset.target = '#nested-' + fieldKey.replace(/\./g, '-');
        header.dataset.variant = 'vertical';

        const icon = document.createElement('span');
        icon.textContent = '▼';
        icon.style.transition = 'transform 0.2s ease';
        header.appendChild(icon);

        const label = document.createElement('span');
        label.textContent = fieldKey.split('.').pop() ?? fieldKey;
        gsap.set(label, {
            fontWeight: '700',
            color: 'var(--gray-700)'
        });
        header.appendChild(label);

        const nestedContent = document.createElement('div');
        nestedContent.id = 'nested-' + fieldKey.replace(/\./g, '-');
        gsap.set(nestedContent, {
            height: '0',
            overflow: 'hidden'
        });

        // Handle regular objects
        for (const [nestedKey, nestedValue] of Object.entries(value)) {
            if (nestedKey.startsWith('_')) continue;

            const nestedFieldKey = `${fieldKey}.${nestedKey}`;
            if (nestedValue !== null && typeof nestedValue === 'object' && !isBinaryId(nestedValue)) {
                await handleNestedData(nestedFieldKey, config, nestedValue, container);
            } else {
                const element = await createFieldElement(config, nestedKey, nestedValue);
                nestedContent.appendChild(element);
            }
        }

        wrapper.appendChild(header);
        wrapper.appendChild(nestedContent);
        parentContainer.appendChild(wrapper);
    };

    return {
        createArrayItem,
        handleNestedData
    };
}