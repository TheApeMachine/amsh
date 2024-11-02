// @ts-ignore: JSX factory import is used by the transpiler
import { jsx, Fragment } from "@/lib/template";
import { gsap } from "gsap";
import Flip from "gsap/Flip";
import { onMount } from "@/lib/lifecycle";
import { eventBus } from "@/lib/event";

gsap.registerPlugin(Flip);

interface Props {
    header?: JSX.Element;
    aside?: JSX.Element | JSX.Element[];
    main?: JSX.Element;
    article?: JSX.Element;
    footer?: JSX.Element;
    styles?: Record<string, any>;
    variant?: string;
    data?: Record<string, unknown>;
}

export const DynamicIsland = ({ ...props }: Props): JSX.Element => {
    // Define a ref to reference the main container element
    const elementRef = { current: null as HTMLElement | null };
    const { header, aside, main, article, footer, variant, styles } = props;

    // Load the configuration asynchronously on mount
    onMount(elementRef.current, async () => {
        console.log("props", props);

        if (variant) {
            try {
                const config = await import(
                    `@/components/ui/configs/${variant}.json`
                );
                console.log("Loaded config:", config);

                // Apply animations once loaded
                if (elementRef.current) {
                    const state = Flip.getState(elementRef.current);

                    if (styles) applyStyles(config.styles);

                    Flip.from(state, {
                        duration: 0.3,
                        ease: "power2.inOut"
                    });
                }

                // Subscribe to a specific click event using the centralized event bus
                eventBus.subscribe("click", (e: Event) => {
                    console.log("Click event triggered:", e);
                    // Update component state or style based on this event
                    if (elementRef.current) {
                        gsap.to(elementRef.current, {
                            scale: 1.1,
                            duration: 0.5,
                            ease: "power2.out"
                        });
                    }
                });
            } catch (err) {
                console.error("Error loading variant configuration:", err);
            }
        }
    });

    // Function to apply styles to the elements
    const applyStyles = (styles: Record<string, any>) => {
        if (elementRef.current) {
            Object.entries(styles).forEach(([selector, value]) => {
                const el = elementRef.current?.querySelector(
                    selector
                ) as HTMLElement;
                if (el) {
                    el.style.cssText = value;
                }
            });
        }
    };

    // Render the content
    return (
        <div ref={(el) => (elementRef.current = el)}>
            <header>{header}</header>
            <aside>{Array.isArray(aside) ? aside : aside}</aside>
            <main>{main}</main>
            <article>{article}</article>
            <footer>{footer}</footer>
        </div>
    );
};
