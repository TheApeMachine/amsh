import gsap from 'gsap';

export const Transitions = (duration: number, easy: number) => {
    return {
        switchLayer: (
            currentLayer: number | undefined,
            targetLayer: number,
            positions: string[],
            layers: any,
            distance: number,
        ) => {
            if (!currentLayer || currentLayer === targetLayer) return;

            const steps = targetLayer - currentLayer;
            const tl = gsap.timeline();
            const animationOrder = steps > 0 ? positions : [...positions].reverse();
            const reverse = steps < 0;

            animationOrder.forEach((position: string) => {
                const zPosition = (positions.indexOf(position) - (targetLayer - 1)) * -distance;

                tl.to(layers[position], {
                    z: zPosition,
                    duration: duration,
                    ease: `back.inOut(${easy})`,
                }, 0);

                tl.to(layers[position], {
                    opacity: 0.5,
                    filter: "blur(10px)",
                    rotationX: reverse ? 20 : -20,
                    duration: duration / 2,
                    ease: `back.inOut(${easy})`,
                }, duration / 4);

                tl.to(layers[position], {
                    opacity: 1,
                    filter: "blur(0px)",
                    rotationX: 0,
                    duration: duration / 4,
                    ease: `back.inOut(${easy})`,
                }, duration / 2);
            });

            tl.play();
            return targetLayer;
        }
    }

}