import { navigateTo } from '@/router';

class LayoutComponent extends HTMLElement {
    private template: HTMLTemplateElement;
    private overflowClass: string = "no-scroll";
    private openClass: string = "nav--open";
    private activeClass: string = "nav__item--active";

    constructor() {
        super();
        this.template = document.createElement('template');
        this.template.innerHTML = `
        <style>
        :host {
            --hue: 223;
            --bg: hsl(var(--hue),10%,90%);
            --fg: hsl(var(--hue),10%,10%);
            --primary: hsl(var(--hue),90%,55%);
            --trans-dur: 0.3s;
            font-size: calc(16px + (20 - 16) * (100vw - 320px) / (1280 - 320));
            display: flex;
            flex-direction: column;
            height: 100%;
            width: 100%;
            transform-style: preserve-3d;
            perspective: 500px;
        }
        main {
            display: flex;
            flex-direction: column;
            height: 100%;
            width: 100%;
            margin: auto;
            padding: 4em 0 0 0;
            transform-style: preserve-3d;
            perspective: 500px;
        }
        .no-scroll {
            overflow: hidden;
        }
        nav {
            position: fixed;
            top: 0;
            text-align: center;
	        text-transform: uppercase;
	        width: 100vw;
        }
        .nav__arrow,
        .nav__items {
            z-index: 0;
        }
        .nav__arrow,
        .nav__item {
            color: hsl(0,0%,0%,0.7);
        }
        .nav__arrow {
            display: block;
            pointer-events: none;
            position: absolute;
            top: 3em;
            left: calc(50% - 0.375em);
            width: 0.75em;
            height: 0.375em;
            transition:
                opacity 0.15s 0.15s ease-in-out,
                transform 0.15s 0.15s ease-in-out;
        }
        .nav__items {
            list-style: none;
            position: relative;
            width: 100%;
            margin: 0;
            padding: 0;
        }
        .nav__item {
            background-color: hsl(var(--hue),90%,70%);
            box-shadow: 0 0 0 hsla(0,0%,0%,0.3);
            font-size: 0.75em;
            font-weight: 600;
            letter-spacing: 0.25em;
            position: absolute;
            display: flex;
            align-items: center;
            justify-content: center;
            width: 100%;
            height: 25vh;
            min-height: 8rem;
            transition:
                box-shadow var(--trans-dur) ease-in-out,
                transform var(--trans-dur) ease-in-out,
                visibility var(--trans-dur) steps(1);
            transform: translateY(calc(-100% + 4rem));
            visibility: hidden;
            z-index: 100;
        }
        .nav__item:nth-of-type(2) {
            background-color: hsl(3,90%,70%);
            z-index: 99;
        }
        .nav__item:nth-of-type(3) {
            background-color: hsl(33,90%,70%);
            z-index: 98;
        }
        .nav__item:nth-of-type(4) {
            background-color: hsl(153,90%,40%);
            z-index: 97;
        }
        .nav__item-link {
            background-color: hsla(0,0%,100%,0);
            color: inherit;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            padding: 1.5rem;
            text-decoration: none;
            transition: background-color 0.15s ease-in-out;
            width: 100%;
            height: 100%;
        }
        .nav__item-link:focus {
            background-color: hsla(0,0%,100%,0.2);
            outline: transparent;
        }
        .nav__item-icon {
            display: block;
            margin: 0 auto 1.5em;
            opacity: 1;
            pointer-events: none;
            width: 2em;
            height: 2em;
            transition:
                opacity var(--trans-dur) ease-in-out,
                transform var(--trans-dur) ease-in-out;
            transform: scale(1);
        }
        .nav__item-text {
            transition: transform var(--trans-dur) ease-in-out;
        }
        .nav--open {
            overflow-x: hidden;
            overflow-y: hidden;
            height: 100%;
            z-index: 100;
        }
        .nav--open .nav__arrow {
            opacity: 0;
            transform: scale(0);
            transition-delay: 0s;
        }
        .nav--open .nav__item-icon {
            opacity: 1;
            transform: scale(1);
            transition-delay: 0.05s;
            transition-timing-function: cubic-bezier(0.42,0,0.58,1.5);
        }
        .nav--open .nav__item {
            box-shadow: 0 0.5em 0.5em hsla(0,0%,0%,0.3);
            transform: translateY(0);
            transition-duration: var(--trans-dur), var(--trans-dur), 0s;
            visibility: visible;
        }
        .nav--open .nav__item:nth-of-type(2) {
            transform: translateY(100%);
        }
        .nav--open .nav__item:nth-of-type(3) {
            transform: translateY(200%);
        }
        .nav--open .nav__item:nth-of-type(4) {
            transform: translateY(300%);
        }
        .nav:not(.nav--open) .nav__item--active {
            box-shadow: 0 0.5em 0.5em hsla(0,0%,0%,0.3);
            visibility: visible;
            z-index: 1;
        }
        .nav:not(.nav--open) .nav__item--active .nav__item-link {
            justify-content: center;
        }
        .nav:not(.nav--open) .nav__item--active .nav__item-icon {
            opacity: 0;
            transform: scale(0.5) translateY(-20px);
            transition-delay: 0s;
        }
        .nav:not(.nav--open) .nav__item--active .nav__item-text {
            transform: translateY(0);
        }
        .nav--open .nav__item-icon {
            transition-delay: 0.1s;
        }
        .nav--open .nav__item-text {
            transition-delay: 0.2s;
        }
        .nav--open .nav__item:nth-of-type(2) {
            transform: translateY(100%);
        }
        .nav--open .nav__item:nth-of-type(2) .nav__item-icon {
            transition-delay: 0.1s;
        }
        .nav--open .nav__item:nth-of-type(3) {
            transform: translateY(200%);
        }
        .nav--open .nav__item:nth-of-type(3) .nav__item-icon {
            transition-delay: 0.15s;
        }
        .nav--open .nav__item:nth-of-type(4) {
            transform: translateY(300%);
        }
        .nav--open .nav__item:nth-of-type(4) .nav__item-icon {
            transition-delay: 0.2s;
        }
        .nav:not(.nav--open) .nav__item--active {
            box-shadow: 0 0.5em 0.5em hsla(0,0%,0%,0.3);
            visibility: visible;
            z-index: 1;
        }

        .nav--open .nav__item--active {
            pointer-events: none;
        }
            
        @supports selector(:focus-visible) {
            .nav__item-link:focus {
                background-color: hsla(0,0%,100%,0);
            }
            .nav__item-link:focus-visible {
                background-color: hsla(0,0%,100%,0.2);
            }
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --bg: hsl(var(--hue),10%,20%);
                --fg: hsl(var(--hue),10%,90%);
            }
        }
        </style>
        <svg display="none">
            <symbol id="home" viewBox="0 0 32 32">
                <g fill="currentColor">
                    <polygon points="16 0,0 10,0 32,10 32,10 16,22 16,22 32,32 32,32 10"/>
                </g>
            </symbol>
            <symbol id="work" viewBox="0 0 32 32">
                <g fill="currentColor">
                    <path d="M30,8h-6v-2c0-1.1-.9-2-2-2H10c-1.1,0-2,.9-2,2v2H2c-1.1,0-2,.9-2,2V26c0,1.1,.9,2,2,2H30c1.1,0,2-.9,2-2V10c0-1.1-.9-2-2-2Zm-20-1c0-.55,.45-1,1-1h10c.55,0,1,.45,1,1v1H10v-1Z"/>
                </g>
            </symbol>
            <symbol id="learn" viewBox="0 0 32 32">
                <g fill="currentColor">
                    <path d="M16,0C7.163,0,0,7.163,0,16s7.163,16,16,16,16-7.163,16-16S24.837,0,16,0Zm2,22c0,1.1-.9,2-2,2s-2-.9-2-2v-6c0-1.1,.9-2,2-2s2,.9,2,2v6Zm-2-10c-1.105,0-2-.895-2-2s.895-2,2-2,2,.895,2,2-.895,2-2,2Z"/>
                </g>
            </symbol>
            <symbol id="connect" viewBox="0 0 32 32">
                <g fill="currentColor">
                    <path d="M30,4H2c-1.1,0-2,.9-2,2v14c0,1.1,.9,2,2,2h3.169l5.417,5.417c.778,.778,2.051,.778,2.828,0l5.417-5.417h11.169c1.1,0,2-.9,2-2V6c0-1.1-.9-2-2-2ZM5,8h6c.55,0,1,.45,1,1s-.45,1-1,1H5c-.55,0-1-.45-1-1s.45-1,1-1Zm0,4h14c.55,0,1,.45,1,1s-.45,1-1,1H5c-.55,0-1-.45-1-1s.45-1,1-1Zm22,6H5c-.55,0-1-.45-1-1s.45-1,1-1H27c.55,0,1,.45,1,1s-.45,1-1,1Z"/>
                </g>
            </symbol>
        </svg>

        <nav class="nav">
            <ul class="nav__items">
                <li class="nav__item nav__item--active">
                    <a class="nav__item-link" href="/">
                        <svg class="nav__item-icon" width="32px" height="32px" aria-hidden="true">
                            <use xlink:href="#home" />
                        </svg>
                        <span class="nav__item-text">Home</span>
                    </a>
                </li>
                <li class="nav__item">
                    <a class="nav__item-link" href="/work">
                        <svg class="nav__item-icon" width="32px" height="32px" aria-hidden="true">
                            <use xlink:href="#work" />
                        </svg>
                        <span class="nav__item-text">Work</span>
                    </a>
                </li>
                <li class="nav__item">
                    <a class="nav__item-link" href="/learn">
                        <svg class="nav__item-icon" width="32px" height="32px" aria-hidden="true">
                            <use xlink:href="#learn" />
                        </svg>
                        <span class="nav__item-text">Learn</span>
                    </a>
                </li>
                <li class="nav__item">
                    <a class="nav__item-link" href="/connect">
                        <svg class="nav__item-icon" width="32px" height="32px" aria-hidden="true">
                            <use xlink:href="#connect" />
                        </svg>
                        <span class="nav__item-text">Connect</span>
                    </a>
                </li>
            </ul>
            <svg class="nav__arrow" width="12px" height="6px" aria-hidden="true">
                <polyline fill="none" stroke="currentColor" stroke-width="2" points="1,1 6,5 11,1" />
            </svg>
        </nav>
        <main>
            <slot></slot>
        </main>
        `;

        this.attachShadow({ mode: "open" });
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
    }

    connectedCallback() {
        this.shadowRoot?.querySelector('.nav')?.addEventListener('click', this.toggle.bind(this));
        window.addEventListener('popstate', this.handlePopState.bind(this));
        this.loadContent(location.pathname);
    }

    toggle(e: Event) {
        e.preventDefault();
        const target = e.target as HTMLElement;
        if (target.closest('.nav__item-link')) {
            const body = document.body;
            const nav = this.shadowRoot?.querySelector('.nav');
            if (!nav) return;
            
            nav.classList.toggle(this.openClass);
            if (nav.classList.contains(this.openClass)) {
                body.classList.add(this.overflowClass);
                this.resetNavItemStyles();
            } else {
                body.classList.remove(this.overflowClass);
                const active = nav.querySelector(`.${this.activeClass}`);
                if (active) {
                    active.classList.remove(this.activeClass);
                }
                const newActive = target.closest('.nav__item');
                if (newActive) {
                    newActive.classList.add(this.activeClass);
                    this.animateNavItemIcon(newActive as HTMLElement);
                    const route = newActive.querySelector('.nav__item-link')?.getAttribute('href');
                    if (route) {
                        history.pushState(null, '', route);
                        this.loadContent(route);
                    }
                }
            }
        }
    }

    handlePopState() {
        this.loadContent(location.pathname);
    }

    async loadContent(route: string) {
        const main = this.shadowRoot?.querySelector('main');
        if (!main) return;

        // Use the router to load the content
        await navigateTo(route, main);
    }

    resetNavItemStyles() {
        const navItems = this.shadowRoot?.querySelectorAll('.nav__item');
        navItems?.forEach(item => {
            const icon = item.querySelector('.nav__item-icon') as HTMLElement;
            const text = item.querySelector('.nav__item-text') as HTMLElement;
            if (icon && text) {
                icon.style.opacity = '1';
                icon.style.transform = 'scale(1) translateY(0)';
                text.style.transform = 'translateY(0)';
            }
        });
    }

    animateNavItemIcon(item: HTMLElement) {
        const icon = item.querySelector('.nav__item-icon') as HTMLElement;
        const text = item.querySelector('.nav__item-text') as HTMLElement;
        if (icon && text) {
            // Trigger reflow to ensure the animation runs
            void icon.offsetWidth;
            icon.style.opacity = '0';
            icon.style.transform = 'scale(0.5) translateY(-20px)';
            text.style.marginTop = '1.5em'; // Adjust text position to center it vertically
        }
    }
}

customElements.define('layout-component', LayoutComponent);