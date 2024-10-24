import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

class ScrollMask extends HTMLElement {
    private CONFIG: {
        [key: string]: number | boolean;
    } = {
        size: 120,
        trigger: false,
        bar: true,
        range: 120,
        light: false,
    };
    private scroller: HTMLElement | null = null;
    private sig: HTMLElement | null = null;
    private template: HTMLTemplateElement;

    constructor() {
        super();
        this.template = document.createElement('template');
        this.template.innerHTML = `
        <style>
            :host {
                display: block;
            }

            @import url('https://unpkg.com/normalize.css') layer(normalize);

@layer base {
	* {
		box-sizing: border-box;
	}

	:root {
		color-scheme: light only;
		--accent: hsl(10 90% 50%);
	}

	[data-light='false'] {
		color-scheme: dark only;
	}

	body {
		display: grid;
		place-items: center;
		min-height: 100vh;
		accent-color: hsl(250 90% 70%);
		background: var(--bg);
		font-family: 'SF Pro Text', 'SF Pro Icons', 'AOS Icons', 'Helvetica Neue';
		line-height: 1.6;
		color: var(--color);
	}

	body::before {
		--line: color-mix(in lch, canvasText, transparent 65%);
		--size: 40px;
		content: '';
		height: 100vh;
		width: 100vw;
		position: fixed;
		background: linear-gradient(
					90deg,
					var(--line) 1px,
					transparent 1px var(--size)
				)
				0 -5vmin / var(--size) var(--size),
			linear-gradient(var(--line) 1px, transparent 1px var(--size)) 0 -5vmin / var(
					--size
				) var(--size);
		mask: linear-gradient(-15deg, transparent 80%, white);
		top: 0;
		z-index: -1;
	}
	p::selection {
		background: var(--accent);
	}
	p:first-of-type {
		font-weight: 600;
		text-wrap: balance;
		line-height: 1.2;
	}
	p:not(:first-of-type) {
		opacity: .875;
		font-weight: 300;
	}
	p {
		padding: 0;
		margin: 0;
	}
}

h1 {
	font-size: 1.5rem;
	margin: 0;
	position: relative;
	display: inline-block;
	align-self: start;
}

h1 span {
	position: absolute;
	left: calc(100% + 0.25ch);
	bottom: 50%;
	font-size: 0.5em;
	color: var(--accent);
}

@layer scroller {
	.resizer {
		--bg: color-mix(in lch, canvas, canvasText 2%);
		width: 36ch;
		aspect-ratio: 3 / 4;
		resize: both;
		overflow: hidden;
		min-width: 260px;
		min-height: 320px;
    max-height: calc(100vh - 4rem);
    max-width: calc(100vw - 4rem);
		background:
			linear-gradient(var(--bg), var(--bg)) padding-box,
			linear-gradient(
				color-mix(in lch, canvasText, transparent 70%),
				color-mix(in lch, canvasText, transparent 90%)
			) border-box;
		border-radius: 12px;
		border: 1px solid transparent;
		padding: 0.25rem;
	}
	.scroller {
		width: 100%;
		height: 100%;		
		overflow: auto;
		position: relative;
		overflow-x: hidden;
		padding: 0.75rem;
		scrollbar-width: thin;
		scrollbar-color: var(--accent) transparent;
	}
	article {
		display: flex;
		gap: 1rem;
		flex-direction: column;
	}
}

@layer mask {
	.scroller {
		--mask-width: 100%;
		mask: linear-gradient(#000, #0000)
				0 0 / var(--mask-width) calc((var(--scroll-progress-top, 0) / 100) * (var(--mask-size) * 1px)) no-repeat,
			linear-gradient(#000, #000) 0 50% / 100% 100% no-repeat,
			linear-gradient(#0000, #000)
				0 100% / var(--mask-width) calc(((100 - var(--scroll-progress-bottom, 100)) / 100) * (var(--mask-size) * 1px)) no-repeat;
		mask-composite: exclude;
	}
	[data-mask-bar="false"] .scroller {
		--mask-width: calc(100% - 10px);
	}
	[data-trigger="true"] .scroller {
		transition: mask-size 0.25s linear(
	    0 0%, 0.0027 3.64%, 0.0106 7.29%,
	    0.0425 14.58%, 0.0957 21.87%,
	    0.1701 29.16%, 0.2477 35.19%,
	    0.3401 41.23%, 0.5982 55.18%,
	    0.7044 61.56%, 0.7987 68.28%, 0.875 75%,
	    0.9297 81.25%, 0.9687 87.5%,
	    0.9922 93.75%, 1 100%
	  );
	}
	@supports (animation-timeline: scroll()) {
		.scroller {
			scroll-timeline: --scroller;
			mask: linear-gradient(white, transparent) 0 0 / var(--mask-width) 0 no-repeat,
				linear-gradient(white, white) 0 50% / 100% 100% no-repeat,
				linear-gradient(transparent, white) 0 100% / var(--mask-width) calc(var(--mask-size) * 1px) no-repeat;
			mask-composite: exclude;
			mask-repeat: none;
			animation: mask-up both linear, mask-down both linear;
			animation-timeline: scroll(self);
			animation-range: 0 calc(var(--mask-range) * 1px), calc(100% - (var(--mask-range) * 1px)) 100%;
		}
		@keyframes mask-up {
			100% {
				mask-size: var(--mask-width) calc(var(--mask-size) * 1px), 100% 100%, var(--mask-width) calc(var(--mask-size) * 1px);
			}
		}
		@keyframes mask-down {
			100% {
				mask-size: var(--mask-width) calc(var(--mask-size) * 1px), 100% 100%, var(--mask-width) 0;
			}
		}

		@property --trigger-up {
			initial-value: 0;
			inherits: true;
			syntax: '<integer>';
		}
		@property --trigger-down {
			initial-value: 1;
			inherits: true;
			syntax: '<integer>';
		}

		[data-trigger=true] {
			timeline-scope: --scroller;
		}

		[data-trigger=true] .resizer {
			animation: trigger-up both linear, trigger-down both linear;
			animation-timeline: --scroller;
			animation-range: 0 calc(var(--mask-range) * 1px), calc(100% - (var(--mask-range) * 1px)) 100%;
		}

		[data-trigger=true] .scroller {
			animation: none;
			mask: linear-gradient(white, transparent) 0 0 / var(--mask-width) calc(var(--trigger-up, 0) * (var(--mask-size) * 1px)) no-repeat,
					linear-gradient(white, white) 0 50% / 100% 100% no-repeat,
					linear-gradient(transparent, white) 0 100% / var(--mask-width) calc(var(--trigger-down, 0) * (var(--mask-size) * 1px)) no-repeat;
			mask-composite: exclude;
			mask-repeat: none;
		}

		@keyframes trigger-up {
			100% {
				--trigger-up: 1;
			}
		}
		@keyframes trigger-down {
			100% {
				--trigger-down: 0;
			}
		}

	}
}

@layer sign-off {
	.sig path {
		--draw: 1.025;
		--end: 1.025;
		stroke-dasharray: var(--end);
		stroke-dashoffset: var(--draw);
	}

	.sig :is(.eye, .nose) {
		fill: transparent;
	}

	.sig {
		width: max(100px, 20vmin);
		max-width: 35%;
		color: color-mix(in lch, canvasText, canvas 25%);
		z-index: 10;
		rotate: -15deg;
		translate: -5% 0;
		pointer-events: none;
		align-self: flex-end;
		margin: 1rem 0 2rem 0;
	}
	@supports (animation-timeline: scroll()) {
		@property --sign {
			initial-value: 0;
			inherits: true;
			syntax: '<integer>';
		}
		.sig {
			animation: sign both steps(1, end);
			animation-timeline: view();
			animation-range: entry;
		}
		.sig path {
			animation: none;
			stroke-dashoffset: calc(1.025 - (var(--sign) * 1.025));
			transition: stroke-dashoffset 1s 0.25s;
		}
		.sig :is(.eye, .nose) {
			animation: none;
		}
		@keyframes sign {
			to { --sign: 1; }
		}
	}
}

@layer socials {
	.bear-link {
		color: currentColor;
		position: fixed;
		top: 1rem;
		left: 1rem;
		width: 48px;
		aspect-ratio: 1;
		display: grid;
		place-items: center;
		opacity: 0.8;
	}

	:where(.x-link, .bear-link):is(:hover, :focus-visible) {
		opacity: 1;
	}
	.bear-link svg {
		width: 75%;
	}
}



@layer transition {
	@property --column-one {
		inherits: true;
		initial-value: 0;
		syntax: '<number>';
	}
	@property --column-two {
		inherits: true;
		initial-value: 0;
		syntax: '<number>';
	}
	@property --column-three {
		inherits: true;
		initial-value: 0;
		syntax: '<number>';
	}
	@property --column-four {
		inherits: true;
		initial-value: 0;
		syntax: '<number>';
	}
	@property --column-five {
		inherits: true;
		initial-value: 0;
		syntax: '<number>';
	}

	@keyframes one   { from { --column-one:   100; }}
	@keyframes two   { from { --column-two:   100; }}
	@keyframes three { from { --column-three: 100; }}
	@keyframes four  { from { --column-four:  100; }}
	@keyframes five  { from { --column-five:  100; }}
	
	::view-transition-new(root) {
		clip-path: polygon(
			/*	1st column */
			0 100%, 0 calc(var(--column-one) * 1%), 20% calc(var(--column-one) * 1%), 20% 100%,
			/*	2nd column */
			20% 100%, 20% calc(var(--column-two) * 1%), 40% calc(var(--column-two) * 1%), 40% 100%,
			/*	3rd column */
			40% 100%, 40% calc(var(--column-three) * 1%), 60% calc(var(--column-three) * 1%), 60% 100%,
			/*	4th column */
			60% 100%, 60% calc(var(--column-four) * 1%), 80% calc(var(--column-four) * 1%), 80% 100%,
			/*	5th column */
			80% 100%, 80% calc(var(--column-five) * 1%), 100% calc(var(--column-five) * 1%), 100% 100%
		);
	}
	::view-transition-new(root) {
		--speed: 0.625;
		--columns: 5;
		animation:
			one calc(var(--speed) * 1s) calc(sin((0 / 5) * 45deg) * var(--speed) * 1s),
			two calc(var(--speed) * 1s) calc(sin((1 / 5) * 45deg) * var(--speed) * 1s),
			three calc(var(--speed) * 1s) calc(sin((2 / 5) * 45deg) * var(--speed) * 1s),
			four calc(var(--speed) * 1s) calc(sin((3 / 5) * 45deg) * var(--speed) * 1s),
			five calc(var(--speed) * 1s) calc(sin((4 / 5) * 45deg) * var(--speed) * 1s);
		animation-fill-mode: both;
		animation-timing-function: linear(
	    0 0%, 0.0027 3.64%, 0.0106 7.29%,
	    0.0425 14.58%, 0.0957 21.87%,
	    0.1701 29.16%, 0.2477 35.19%,
	    0.3401 41.23%, 0.5982 55.18%,
	    0.7044 61.56%, 0.7987 68.28%, 0.875 75%,
	    0.9297 81.25%, 0.9687 87.5%,
	    0.9922 93.75%, 1 100%
	  );
		z-index: 2;
	}
	::view-transition-old(root) {
		animation: none;
	}
}
        </style>
        <div class="resizer">
            <section class="scroller">
                <article>
                    <h1>Scroll Masking<span>PRO</span></h1>
                    <p>As you scroll a container, animate its <code>mask-size</code> composited with <code>exclude</code>.
                    <p>Lorem ipsum dolor sit, amet consectetur adipisicing, elit. Non, earum omnis nihil, atque labore fugit provident minus voluptas a laboriosam amet quidem minima molestias ipsum eos ea dicta quod corrupti.</p>
                    <p>Aliquam dolores ipsum sit cum delectus voluptas, repellat obcaecati deleniti omnis alias sed sapiente debitis repudiandae sunt, autem quisquam. Corporis itaque, beatae dolorem repellendus! Alias nobis ipsum rem odit dicta.</p>
                    <p>Exercitationem voluptatem maxime quas facilis sequi repellendus laudantium, mollitia eaque cum cumque saepe magnam enim explicabo quia autem totam, placeat vitae voluptatibus ipsa accusantium distinctio, quibusdam eos iste molestiae! Quaerat!</p>
                </article>
            </section>
        </div>
        `;

        this.attachShadow({ mode: "open" });
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
        this.scroller = this.shadowRoot?.querySelector('.scroller') as HTMLElement;
        this.sig = this.shadowRoot?.querySelector('.sig') as HTMLElement;
        this.setupControls();
        this.update();
        this.setupScrollTrigger();
    }

    private setupControls() {
        this.updateConfig = this.updateConfig.bind(this);
    }

    private updateConfig(key: string, value: any) {
        (this.CONFIG as any)[key] = value;
        this.update();
    }

    private update() {
        if (!CSS.supports('animation-timeline: scroll()')) {
            ScrollTrigger.refresh();
        }
        this.setAttribute('data-light', this.CONFIG.light.toString());
        this.setAttribute('data-trigger', this.CONFIG.trigger.toString());
        this.setAttribute('data-mask-bar', this.CONFIG.bar.toString());
        this.style.setProperty('--mask-size', this.CONFIG.size.toString());
        this.style.setProperty('--mask-range', this.CONFIG.range.toString());
    }

    private setupScrollTrigger() {
        if (!CSS.supports('animation-timeline: scroll()')) {
            gsap.registerPlugin(ScrollTrigger);

            ScrollTrigger.create({
                scroller: this.scroller,
                scrub: true,
                start: 0,
                end: () => this.CONFIG.range as number,
                trigger: 'article',
                onUpdate: (self) => {
                    if (this.scroller) {
                        this.scroller.style.setProperty('--scroll-progress-top', 
                            ((this.CONFIG.trigger ? Math.floor(self.progress) : self.progress) * 100).toString());
                    }
                }
            });

            ScrollTrigger.create({
                scroller: this.scroller,
                trigger: 'article',
                scrub: true,
                start: () => ScrollTrigger.maxScroll(this.scroller!) - ((this.CONFIG.range as number) * 1),
                end: () => ScrollTrigger.maxScroll(this.scroller!),
                onUpdate: (self) => {
                    if (this.scroller) {
                        this.scroller.style.setProperty('--scroll-progress-bottom', 
                            ((this.CONFIG.trigger ? Math.ceil(self.progress) : self.progress) * 100).toString());
                    }
                }
            });

            if (this.sig) {
                gsap.fromTo(this.sig.querySelectorAll('path'), {
                    '--draw': 1.025,
                }, {
                    '--draw': 0,
                    scrollTrigger: {
                        trigger: this.sig,
                        scroller: this.scroller,
                        toggleActions: 'play reset play reset',
                        start: `top bottom-=${this.sig.getBoundingClientRect().height * 0.5}`,
                    }
                });
            }

            const refreshScrollTrigger = () => {
                ScrollTrigger.refresh();
            };
            const obs = new ResizeObserver(refreshScrollTrigger);
            if (this.scroller) obs.observe(this.scroller);
        }
    }
}

customElements.define('scroll-mask', ScrollMask);
