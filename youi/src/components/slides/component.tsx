import { jsx } from "@/lib/template";
import Reveal from "reveal.js";

interface Props {
    children: any;
}

export const SlidesComponent = ({ children }: Props) => {
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
            keyboard: false,
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

    return <div className="reveal">{children}</div>;
};

export default SlidesComponent;
