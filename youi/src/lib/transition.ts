import { onMount, onUnmount } from './lifecycle';
import gsap from 'gsap';

/*
Transition function to manage both enter and exit animations for elements.
Allows components or routes to declare animations, making the entire system more declarative.
*/
export const Transition = (
    element: Node,
    { enter, exit }: { enter?: (el: HTMLElement) => void, exit?: (el: HTMLElement) => void }
) => {
    const targetElement = element instanceof HTMLElement ? element : (element as Element).firstElementChild as HTMLElement;

    if (targetElement) {
        // Apply enter animation when the element is added to the DOM
        if (enter) {
            onMount(targetElement, () => {
                console.debug("transition", "onMount", targetElement);
                enter(targetElement);
            });
        }

        // Apply exit animation when the element is removed from the DOM
        if (exit) {
            onUnmount(targetElement, () => {
                console.debug("transition", "onUnmount", targetElement);
                exit(targetElement);
            });
        }
    } else {
        console.warn("Transition: No element found in DocumentFragment");
    }

    return element;
};

// Helper functions for creating transitions
export const sequence = (...animations: Array<(el: HTMLElement) => gsap.core.Timeline | gsap.core.Tween>): ((el: HTMLElement) => gsap.core.Timeline) => {
    return (el: HTMLElement) => {
        const tl = gsap.timeline();
        animations.forEach(anim => tl.add(anim(el)));
        return tl;
    };
};

export const parallel = (...animations: Array<(el: HTMLElement) => gsap.core.Timeline | gsap.core.Tween>): ((el: HTMLElement) => gsap.core.Timeline) => {
    return (el: HTMLElement) => {
        const tl = gsap.timeline();
        animations.forEach(anim => tl.add(anim(el), 0)); // Start all animations at time 0
        return tl;
    };
};

// Predefined animations for convenience
export const fadeIn = (el: HTMLElement) => gsap.from(el, { opacity: 0, duration: 1, ease: "power2.out" });
export const fadeOut = (el: HTMLElement) => gsap.to(el, { opacity: 0, duration: 1, ease: "power2.in" });
export const scaleUp = (el: HTMLElement) => gsap.from(el, { scale: 0.9, duration: 1, ease: "power2.out" });
export const scaleDown = (el: HTMLElement) => gsap.to(el, { scale: 0.9, duration: 1, ease: "power2.in" });
export const blurIn = (el: HTMLElement) => gsap.from(el, { filter: "blur(10px)", duration: 1, ease: "power2.out" });
export const blurOut = (el: HTMLElement) => gsap.to(el, { filter: "blur(10px)", duration: 1, ease: "power2.in" });
