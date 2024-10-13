import Reveal from 'reveal.js';

class SlidesComponent extends HTMLElement {
    private revealInstance: Reveal.Api | undefined;

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });

        const template = document.createElement('template');
        template.innerHTML = `
            <link rel="stylesheet" href="node_modules/reveal.js/dist/reveal.css" />
            <link rel="stylesheet" href="node_modules/reveal.js/dist/theme/white.css" />
            <div class="reveal">
                <div class="slides">
                    <section>
                        <button 
                            data-event="click"
                            data-effect="switch-layer" 
                            data-target="logic"
                        >
                            Go to Logic Layer
                        </button>
                    </section>
                    <section>Slide 2</section>
                </div>
            </div>
        `;

        this.shadowRoot!.appendChild(template.content.cloneNode(true));
    }

    connectedCallback() {
        const revealElement = this.shadowRoot!.querySelector('.reveal');
        if (revealElement instanceof HTMLElement) {
            this.revealInstance = new Reveal(revealElement, {});
            this.revealInstance.initialize({
                controls: true,
                controlsTutorial: true,
                controlsLayout: 'bottom-right',
                controlsBackArrows: 'faded',
                progress: true,
                slideNumber: false,
                showSlideNumber: 'all',
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
                navigationMode: 'default',
                shuffle: false,
                fragments: true,
                fragmentInURL: true,
                embedded: false,
                help: true,
                pause: true,
                showNotes: false,
                autoPlayMedia: null,
                preloadIframes: null,
                autoAnimate: true,
                autoAnimateMatcher: null,
                autoAnimateEasing: 'ease',
                autoAnimateDuration: 1.0,
                autoAnimateUnmatched: true,
                autoAnimateStyles: [
                    'opacity',
                    'color',
                    'background-color',
                    'padding',
                    'font-size',
                    'line-height',
                    'letter-spacing',
                    'border-width',
                    'border-color',
                    'border-radius',
                    'outline',
                    'outline-offset',
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
                transition: 'convex',
                transitionSpeed: 'default',
                backgroundTransition: 'fade',
                pdfMaxPagesPerSlide: Number.POSITIVE_INFINITY,
                pdfSeparateFragments: true,
                pdfPageHeightOffset: -1,
                viewDistance: 3,
                mobileViewDistance: 2,
                display: 'block',
                hideInactiveCursor: true,
                hideCursorTime: 5000,
            });
        }
    }
}

customElements.define("slides-component", SlidesComponent);
