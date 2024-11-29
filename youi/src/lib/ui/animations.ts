import gsap from "gsap";

export const eventAnimations = {
    toggle: {
        horizontal: {
            set: (target: HTMLElement) => {
                gsap.set(target, {
                    width: 0,
                    overflow: 'hidden'
                });
            },
            tl: (target: HTMLElement) => {
                const tl = gsap.timeline({
                    paused: true,
                });

                tl.to(target, {
                    width: "auto",
                    duration: 0.3,
                    ease: 'power4.inOut'
                });

                return tl;
            }
        },
        vertical: {
            set: (target: HTMLElement) => {
                gsap.set(target, {
                    height: 0,
                    overflow: 'hidden'
                });
            },
            tl: (target: HTMLElement) => {
                const tl = gsap.timeline({
                    paused: true,
                });

                // Animate container
                tl.to(target, {
                    height: 'auto',
                    duration: 0.3,
                    ease: 'power4.inOut'
                }).to(target.parentElement, {
                    background: 'var(--white)',
                    borderBottom: '1px solid var(--lighter)',
                    zIndex: 2,
                    duration: 0.25,
                    ease: 'power4.inOut'
                }, "<").from(target.querySelector('li'), {
                    opacity: 0,
                    transform: 'translate3d(0, 10px, 0)',
                    duration: 0.25,
                    ease: 'power4.inOut'
                }, "<");

                tl.eventCallback("onComplete", () => { tl.pause(); });
                tl.eventCallback("onReverseComplete", () => { tl.pause(); });

                return tl;
            }
        }
    }
};
