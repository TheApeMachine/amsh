import { jsx as h } from "@/lib/template";
import { DynamicIsland, getAnimation } from "../DynamicIsland";
import { TemplateObject } from "../types";
import gsap from "gsap";
export const Elements = async () => {
    // Template parsing utils
    const parseTemplate = (text: string, data: any): string =>
        text.replace(/\{\{([^}]+)\}\}/g, (_, key) => {
            const value = key.split('.').reduce((obj: any, k: string) => obj?.[k], data);
            return value?.toString() || '';
        });

    // Element creation utils
    const applyDataAttributes = (element: HTMLElement, data: Record<string, string>, context: any): void => {
        Object.entries(data).forEach(([key, value]) => {
            element.dataset[key] = parseTemplate(value, context);
        });
    };

    const createFieldElement = async (
        config: TemplateObject,
        fieldKey: string,
        value: any
    ): Promise<HTMLElement> => {
        return await createDynamicIsland(config, {
            key: fieldKey,
            value: value?.toString() ?? ''
        });
    };

    const createDynamicIsland = async (template: TemplateObject, item: any): Promise<HTMLElement> => {
        const dynamicIslandProps: any = {
            variant: template.props?.variant,
            data: item
        };

        const element = await h(DynamicIsland, dynamicIslandProps);
        if (!(element instanceof HTMLElement)) {
            throw new Error('DynamicIsland did not return an HTMLElement');
        }
        return element;
    };

    const createBasicElement = (template: TemplateObject, data: any): HTMLElement | null => {
        if (!template.tag) return null;

        const element = document.createElement(template.tag);

        // Add class if it exists
        if (template.class) {
            element.className = template.class;
        }

        // Handle both content and text properties
        if (template.content) {
            element.textContent = parseTemplate(template.content, data);
        } else if (template.text) {
            element.textContent = template.text;
        }

        if (template.styles) {
            gsap.set(element, template.styles);
        }

        if (template.data) {
            applyDataAttributes(element, template.data, data);
        }

        // Add props if they exist
        if (template.props) {
            Object.entries(template.props).forEach(([key, value]) => {
                if (typeof value === 'string') {
                    element.setAttribute(key, parseTemplate(value, data));
                }
            });
        }

        // Add event attributes if they exist
        if (template.events) {
            element.dataset.event = template.events.effect;
            element.dataset.trigger = template.events.trigger;
            element.dataset.target = template.events.target;
            element.dataset.variant = parseTemplate(template.events.variant, data);

            if (template.events.target) {
                // Capture events data to avoid undefined checks later
                const { target, effect, variant } = template.events;

                // Set initial styles after a microtask to ensure DOM is ready
                queueMicrotask(() => {
                    const targetElement = element.closest('.dynamic-island')?.querySelector(target);
                    if (targetElement instanceof HTMLElement && effect && variant) {
                        const animation = getAnimation(effect, variant);
                        animation?.set(targetElement);
                    }
                });
            }
        }

        return element;
    };

    return {
        createFieldElement,
        createDynamicIsland,
        createBasicElement,
        parseTemplate,
        applyDataAttributes
    }
}