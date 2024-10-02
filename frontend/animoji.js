class AnimojiLoader extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.currentState = 'intro';
        this.currentIndex = 0;
        this.currentPlayer = null;
        this.transitioning = false;
        this.lastSwitchTime = 0;
        this.players = {
            intro: [],
            thinking: [],
            high: [],
            medium: [],
            low: []
        }
        this.animojis = {
            intro: [
                '/public/animoji/salute/lottie.json'
            ],
            thinking: [
                '/public/animoji/thinking/lottie.json',
                '/public/animoji/robot/lottie.json',
                '/public/animoji/monocle/lottie.json'
            ],
            high: [
                '/public/animoji/sunglasses/lottie.json'
            ],
            medium: [
                '/public/animoji/nerd-face/lottie.json'
            ],
            low: [
                '/public/animoji/mind-blown/lottie.json'
            ]
        };
    }

    connectedCallback() {
        this.render();
    }

    render() {
        this.shadowRoot.innerHTML = `
            <style>
                :host {
                    display: block;
                    position: absolute;
                }
                dotlottie-player {
                    position: absolute;
                    width: 100%;
                    height: 100%;
                    opacity: 0;
                }
                dotlottie-player.active {
                    opacity: 1;
                }
                #animoji-container {
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    width: 64px;
                    height: 64px;
                    position: absolute;
                    top: 0;
                    left: 0;
                    overflow: hidden;
                    filter: drop-shadow(0px 0px 4px rgba(0, 0, 0, 0.5));
                }
            </style>
            <div id="animoji-container">
            </div>
        `;

        // Create the animoji players.
        for (const state in this.animojis) {
            for (const animoji of this.animojis[state]) {
                const player = this.createAnimoji(animoji);
                this.players[state].push(player);
                this.shadowRoot.getElementById('animoji-container').appendChild(player);
                console.log(this.players);
            }
        }
        this.startIntroSequence();
    }

    createAnimoji(src, isActive = false) {
        const player = document.createElement('dotlottie-player');
        player.classList.add('animoji');
        player.setAttribute('src', src);
        player.setAttribute('background', 'transparent');
        player.setAttribute('speed', '1');
        player.setAttribute('loop', 'true');
        player.setAttribute('autoplay', 'true');
        if (isActive) player.classList.add('active');
        this.shadowRoot.appendChild(player);
        return player;
    }

    startIntroSequence() {
        gsap.set(this.shadowRoot.querySelector('.animoji'), { opacity: 0 });
        this.playNextAnimoji();
    }

    playNextAnimoji() {
        const currentTime = Date.now();
        if (currentTime - this.lastSwitchTime < 3000) {
            return;
        }

        this.transitioning = true;
        const index = this.currentIndex % this.animojis[this.currentState].length;
        // Crossfade the current animoji out and the next one in.
        const tl = gsap.timeline({
            onComplete: () => {
                this.currentPlayer = this.players[this.currentState][index];
                this.transitioning = false;
                this.lastSwitchTime = Date.now();
            }
        });
        tl.to(this.currentPlayer, { opacity: 0, duration: 0.25, ease: 'power2.inOut' });
        tl.to(this.players[this.currentState][index], { opacity: 1, duration: 0.25, ease: 'power2.inOut' }, "<");
        tl.play();
    }

    setPosition(x, y) {
        gsap.to(this.shadowRoot.getElementById('animoji-container'), {
            x: x-128,
            y: y-64,
            duration: 1,
            ease: "power2.out"
        })
    }

    setState(newState) {
        const currentTime = Date.now();
        if (this.currentState !== newState && !this.transitioning && currentTime - this.lastSwitchTime >= 3000) {
            this.currentState = newState;
            this.currentIndex = 0;
            this.playNextAnimoji();
        }
    }
}

customElements.define('animoji-loader', AnimojiLoader);