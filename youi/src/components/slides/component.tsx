// @ts-ignore: JSX factory import is used by the transpiler
import { jsx, type HTMLAttributes } from "@/lib/template";
import Reveal from "reveal.js";

interface Props extends HTMLAttributes {
    children: Node | Node[];
}

export const SlidesComponent = ({ children }: Props): JSX.Element => {
    // Create and initialize reveal after the component is mounted
    setTimeout(() => {
        const revealElement = document.querySelector(".reveal");
        if (revealElement instanceof HTMLElement) {
            const revealInstance = new Reveal(revealElement, {});
            revealInstance.initialize({
                controls: true,
                controlsTutorial: true,
                controlsLayout: "bottom-right",
                controlsBackArrows: "faded",
                progress: true,
                slideNumber: false,
                showSlideNumber: "all",
                hashOneBasedIndex: false,
                hash: true,
                respondToHashChanges: true,
                history: true,
                keyboard: true,
                keyboardCondition: null,
                disableLayout: true,
                overview: true,
                center: true,
                touch: true,
                loop: false,
                rtl: false,
                navigationMode: "default",
                shuffle: false,
                fragments: true,
                fragmentInURL: true,
                embedded: true,
                help: true,
                pause: true,
                showNotes: false,
                autoPlayMedia: null,
                preloadIframes: null,
                autoAnimate: true,
                autoAnimateMatcher: null,
                autoAnimateEasing: "ease",
                autoAnimateDuration: 1.0,
                autoAnimateUnmatched: true,
                autoAnimateStyles: [
                    "opacity",
                    "color",
                    "background-color",
                    "padding",
                    "font-size",
                    "line-height",
                    "letter-spacing",
                    "border-width",
                    "border-color",
                    "border-radius",
                    "outline",
                    "outline-offset"
                ],
                autoSlide: 0,
                autoSlideStoppable: true,
                autoSlideMethod: null,
                defaultTiming: null,
                mouseWheel: false,
                previewLinks: true,
                postMessage: true,
                postMessageEvents: false,
                focusBodyOnPageVisibilityChange: true,
                transition: "convex",
                transitionSpeed: "default",
                backgroundTransition: "fade",
                pdfMaxPagesPerSlide: Number.POSITIVE_INFINITY,
                pdfSeparateFragments: true,
                pdfPageHeightOffset: -1,
                viewDistance: 3,
                mobileViewDistance: 2,
                display: "block",
                hideInactiveCursor: true,
                hideCursorTime: 5000
            });
        }
    }, 0);

    // Create a div to hold slides
    const slidesDiv = document.createElement("div");
    slidesDiv.className = "slides";

    // Handle children properly
    if (Array.isArray(children)) {
        children.forEach((child) => {
            if (child instanceof Node) {
                slidesDiv.appendChild(child);
            }
        });
    } else if (children instanceof Node) {
        slidesDiv.appendChild(children);
    }

    const revealDiv = document.createElement("div");
    revealDiv.className = "reveal";
    revealDiv.appendChild(slidesDiv);

    return revealDiv as unknown as JSX.Element;
};

export default SlidesComponent;
