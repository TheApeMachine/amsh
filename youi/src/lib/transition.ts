import { onMount, onUnmount } from './lifecycle';

/*
Transitions is a function that takes a duration and an easing and returns an object with enter and exit functions.
@param duration The duration of the transition.
@param easing The easing of the transition.
@returns An object with enter and exit functions.
*/
import gsap from 'gsap';

export const Transition = (
    element: DocumentFragment,
    { enter, exit }: { enter: (el: HTMLElement) => void, exit: (el: HTMLElement) => void }
): DocumentFragment => {
    const targetElement = element.firstElementChild as HTMLElement;
    console.debug("transition", "targetElement", targetElement)

    // Trigger the enter animation when the element is added to the DOM
    onMount(targetElement, () => {
        console.debug("transition", "onMount", targetElement)
        enter(targetElement);  // Apply the enter animation
    });

    // Trigger the exit animation when the element is removed from the DOM
    onUnmount(targetElement, () => {
        console.debug("transition", "onUnmount", targetElement)
        exit(targetElement);  // Apply the exit animation
    });

    return element;
};

export const sequence = (...animations: Array<(el: HTMLElement) => gsap.core.Timeline | gsap.core.Tween>): ((el: HTMLElement) => gsap.core.Timeline) => {
    console.debug("transition", "sequence", animations)
    return (el: HTMLElement) => {
        const tl = gsap.timeline();
        animations.forEach(anim => tl.add(anim(el)));
        return tl;
    };
};

export const parallel = (...animations: Array<(el: HTMLElement) => gsap.core.Timeline | gsap.core.Tween>): ((el: HTMLElement) => gsap.core.Timeline) => {
    console.debug("transition", "parallel", animations)
    return (el: HTMLElement) => {
        const tl = gsap.timeline();
        animations.forEach(anim => tl.add(anim(el), 0)); // Start all animations at time 0
        return tl;
    };
};

export const fadeIn = (el: HTMLElement) => gsap.from(el, { opacity: 0, duration: 1, ease: "power2.out" });
export const fadeOut = (el: HTMLElement) => gsap.to(el, { opacity: 0, duration: 1, ease: "power2.in" });
export const scaleUp = (el: HTMLElement) => gsap.from(el, { scale: 1, duration: 1, ease: "power2.out" });
export const scaleDown = (el: HTMLElement) => gsap.to(el, { scale: 1, duration: 1, ease: "power2.in" });
export const blurIn = (el: HTMLElement) => gsap.from(el, { filter: "blur(10px)", duration: 1, ease: "power2.out" });
export const blurOut = (el: HTMLElement) => gsap.to(el, { filter: "blur(10px)", duration: 1, ease: "power2.in" });