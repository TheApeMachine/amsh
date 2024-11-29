import { jsx } from "@/lib/template";
import gsap from "gsap";
import { Flip } from "gsap/Flip";
import { Component, Ref } from "@/lib/ui/Component";
import { transformData } from "@/lib/ui/Transform";
import { eventBus } from "@/lib/event";
import { eventAnimations } from "./animations";
import { onUnmount } from "@/lib/lifecycle";

gsap.registerPlugin(Flip);

// Add this type alias before the Props interface
type ContentType = string | JSX.Element | JSX.Element[];
interface Props {
    variant: string;
    data?: Record<string, any>;
    header?: ContentType;
    aside?: ContentType;
    main?: ContentType;
    article?: ContentType;
    footer?: ContentType;
}

type DynamicIslandRefs = {
    dynamicIsland: Ref;
    header: Ref;
    aside: Ref;
    main: Ref;
    article: Ref;
    footer: Ref;
};

// Create a global timeline map outside component to persist across renders
const globalTimelineMap = new WeakMap<Element, gsap.core.Timeline>();

export const getAnimation = (
    effect: keyof typeof eventAnimations,
    variant: keyof (typeof eventAnimations)[keyof typeof eventAnimations]
) => {
    return eventAnimations[effect][variant];
};

const getTimeline = (element: Element) => {
    return (
        globalTimelineMap.get(element) ??
        (() => {
            const tl = eventAnimations.toggle.vertical.tl(
                element as HTMLElement
            );
            globalTimelineMap.set(element, tl);
            return tl;
        })()
    );
};

const applyConfig = (
    data: Record<string, any> | undefined,
    refs: DynamicIslandRefs,
    config: any
) => {
    if (data) {
        Object.entries(refs).forEach(([key, ref]) => {
            if (config[key]) {
                if (config[key].styles) {
                    gsap.set(ref.current, config[key].styles);
                }

                transformData(ref.current as HTMLElement, config[key], data);
            }
        });
    }
};

let state: ReturnType<typeof Flip.getState> | null = null;

export const DynamicIsland = Component.create(
    async (refs: DynamicIslandRefs, { variant, data }: Props) => {
        const config = (await import(`./config/${variant}.json`)).default;

        // Add this check to prevent duplicate subscriptions
        if (
            !refs.dynamicIsland.current?.hasAttribute("data-events-initialized")
        ) {
            // Store unsubscribe functions
            const cleanup: Array<() => void> = [];

            // Subscribe to morph events with cleanup
            cleanup.push(
                eventBus.subscribe(
                    "morph",
                    async (payload: {
                        variant: string;
                        originalEvent: Event;
                    }) => {
                        const currentElement = refs.dynamicIsland.current;
                        if (!currentElement) return;

                        // Only process if the click came from within this DynamicIsland
                        const clickTarget = payload.originalEvent
                            ?.target as HTMLElement;
                        if (!currentElement.contains(clickTarget)) return;

                        state = Flip.getState(currentElement);

                        // Load the new config
                        const morphCfg = (
                            await import(`./config/${payload.variant}.json`)
                        ).default;

                        // Transform with new config using the same pattern as initial setup
                        if (morphCfg) {
                            applyConfig(data?.content, refs, morphCfg);
                        }

                        if (state) {
                            Flip.from(state, {
                                duration: 0.5,
                                ease: "power2.inOut",
                                nested: true
                            });
                        }
                    }
                )
            );

            // Subscribe to toggle events with cleanup
            cleanup.push(
                eventBus.subscribe(
                    "toggle",
                    (payload: {
                        target: string;
                        originalEvent: Event;
                        variant: string;
                    }) => {
                        // Only process events for this specific instance
                        if (
                            !refs.dynamicIsland.current?.contains(
                                payload.originalEvent.target as Node
                            )
                        )
                            return;

                        const targetElement = (
                            payload.originalEvent.target as HTMLElement
                        )
                            ?.closest(".dynamic-island")
                            ?.querySelector(payload.target ?? "");

                        if (!targetElement || !payload.variant) return;

                        const animation = getAnimation(
                            "toggle",
                            payload.variant as "vertical"
                        );
                        animation.set(targetElement as HTMLElement);
                        const timeline = getTimeline(targetElement);

                        if (!timeline.paused()) return;

                        const isActive =
                            targetElement.classList.contains("active");

                        if (isActive) {
                            targetElement.classList.remove("active");
                            timeline.reverse();
                        } else {
                            targetElement.classList.add("active");
                            timeline.play();
                        }
                    }
                )
            );

            // Subscribe to submit events with cleanup
            cleanup.push(
                eventBus.subscribe(
                    "submit",
                    async (payload: { originalEvent: Event }) => {
                        console.log(payload);
                        const form = refs.main.current?.querySelector("form");
                        if (!form) return;

                        // Prevent default form submission
                        payload.originalEvent.preventDefault();

                        // Gather form data
                        const formData = new FormData(form as HTMLFormElement);
                        const data = Object.fromEntries(formData.entries());

                        // Emit save event with form data
                        eventBus.emit("save", { data });
                    }
                )
            );

            // Subscribe to input events with cleanup
            cleanup.push(
                eventBus.subscribe(
                    "input",
                    (payload: { originalEvent: Event }) => {
                        const input = payload.originalEvent
                            .target as HTMLInputElement;
                        if (!input) return;

                        // Mark form as dirty
                        const form = input.closest("form");
                        if (form) {
                            form.dataset.dirty = "true";
                        }
                    }
                )
            );

            // Mark this instance as initialized
            refs.dynamicIsland.current?.setAttribute(
                "data-events-initialized",
                "true"
            );

            // Set up cleanup on unmount
            if (refs.dynamicIsland?.current) {
                const element = refs.dynamicIsland.current;
                const cleanupFn = () => {
                    cleanup.forEach((unsubscribe) => unsubscribe());
                    element.removeAttribute("data-events-initialized");
                    // Clean up timelines
                    const timeline = globalTimelineMap.get(element);
                    if (timeline) {
                        timeline.kill();
                        globalTimelineMap.delete(element);
                    }
                };

                onUnmount(element, cleanupFn);
            }
        }

        applyConfig(data, refs, config);
    },
    ({
        variant,
        data,
        refs,
        ...sections
    }: Props & { refs: DynamicIslandRefs }): JSX.Element => {
        return (
            <div
                ref={(el: HTMLElement | null) => {
                    refs.dynamicIsland.current = el;
                }}
                className="dynamic-island"
                data-variant={variant}
            >
                <header
                    ref={(el: HTMLElement | null) => {
                        refs.header.current = el;
                    }}
                >
                    {sections.header || ""}
                </header>
                <aside
                    ref={(el: HTMLElement | null) => {
                        refs.aside.current = el;
                    }}
                >
                    {sections.aside || ""}
                </aside>
                <main
                    ref={(el: HTMLElement | null) => {
                        refs.main.current = el;
                    }}
                >
                    {sections.main || ""}
                </main>
                <article
                    ref={(el: HTMLElement | null) => {
                        refs.article.current = el;
                    }}
                >
                    {sections.article || ""}
                </article>
                <footer
                    ref={(el: HTMLElement | null) => {
                        refs.footer.current = el;
                    }}
                >
                    {sections.footer || ""}
                </footer>
            </div>
        );
    }
);
