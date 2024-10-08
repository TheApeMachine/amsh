import Reveal from 'reveal.js';
import '../lookingglass/lens';
import '../agentviz/conversation';
import '../timeline/3dtimeline';
class SlidesComponent extends HTMLElement {
    private revealInstance: Reveal.Api | undefined;

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });

        const template = document.createElement('template');
        template.innerHTML = `
            <style>
                :host {
                    display: flex;
                    width: 100%;
                    height: 100%;
                }
                :root {
                    --background-color: #000011;
                    --text-color: #ffffff;
                    --control-background: rgba(0, 0, 0, 0.7);
                }

                .theme-light {
                    --background-color: #ffffff;
                    --text-color: #000000;
                    --control-background: rgba(255, 255, 255, 0.7);
                }

                button {
                    background-color: var(--control-background);
                    color: var(--text-color);
                    border: none;
                    padding: 8px;
                    cursor: pointer;
                    border-radius: 4px;
                    margin-top: 5px;
                }

                button:hover {
                    background-color: rgba(255, 255, 255, 0.1);
                }

                input[type="range"], input[type="text"], select {
                    width: 100%;
                    margin-top: 5px;
                }

                label {
                    color: var(--text-color);
                }

                .reveal {
                    width: 100%;
                    height: 100%;
                    transform-style: preserve-3d;
                }
                .slides {
                    position: relative;
                    width: 100%;
                    height: 100%;
                    transform-style: preserve-3d;
                }
                section {
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    width: 100%;
                    height: 100%;
                    background-image: linear-gradient(to right, #434343 0%, black 100%);
                }
            </style>
            <div class="reveal">
                <div class="slides">
                    <section>
                        <nodegraph-editor></nodegraph-editor>
                    </section>
                    <section>
                        <conversation-visualizer></conversation-visualizer>
                    </section>
                    <section>
                        <looking-glass-lens></looking-glass-lens>
                    </section>
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
                keyboard: false,
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
                embedded: true,
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
