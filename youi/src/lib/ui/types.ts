import { eventAnimations } from "./animations";

export type DynamicIslandVariant = 'page' | 'accordion' | 'menu' | 'menuItem' | 'button' | 'formfield';
export type AnimationVariant = keyof typeof eventAnimations['toggle'];
export type AnimationEffect = keyof typeof eventAnimations;

export interface TemplateObject {
    component?: string;
    tag?: string;
    class?: string;
    content?: string;
    styles?: Record<string, string>;
    props?: {
        variant: DynamicIslandVariant;
        header?: string | TemplateObject;
        [key: string]: any;
    };
    children?: ChildTemplate;
    data?: Record<string, string>;
    text?: string;
    events?: {
        effect: AnimationEffect;
        trigger: string;
        target: string;
        variant: AnimationVariant;
    };
}

export interface ChildTemplate {
    tag: string;
    content: string;
    data?: Record<string, string>;
    children?: ChildTemplate;
}