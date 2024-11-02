import { jsx, Fragment } from "@/lib/template"; // Suppressing warning as this import is required for JSX processing
import { loader } from "@/lib/loader";
import { sequence, Transition, blurIn, blurOut } from "@/lib/transition";
import { match } from "@/lib/match";
import { switchLayer } from "@/lib/layer";
import { gsap } from "gsap";
import SlidesComponent from "@/components/slides/component";

export const effect = async () => {
    let layers: Record<string, HTMLElement> = {};
    let positions: string[] = [".product"];
    let distance: number = 1000;
    let currentLayer: number | undefined = 1;

    layers = {
        product: document.querySelector("slides-component") as HTMLElement
    };

    positions.forEach((position: string, index: number) => {
        gsap.set(layers[position], {
            position: "absolute",
            width: "100%",
            height: "100%",
            transformStyle: "preserve-3d",
            backfaceVisibility: "hidden",
            z: index * -distance,
            zIndex: positions.length - index
        });
    });

    window.addEventListener("keydown", (evt: KeyboardEvent) => {
        if (["1", "2", "3"].includes(evt.key)) {
            currentLayer = switchLayer(
                currentLayer,
                parseInt(evt.key, 10),
                positions,
                layers,
                distance
            );
        }
    });

    window.addEventListener("keydown", (evt: KeyboardEvent) => {
        if (["4"].includes(evt.key)) {
            const el = document.querySelector(
                "toast-container"
            ) as ToastContainer;
            console.log("el", el);
            el.addToast("Success!", "success");
        }
    });
};

export const render = async () => {
    return Transition(
        match(await loader({}), {
            loading: () => {
                console.log("Rendering loading state");
                return <div>Loading...</div>;
            },
            error: (error: any) => {
                console.log("Rendering error state:", error);
                return <div>Error: {error}</div>;
            },
            success: (_: any) => (
                <SlidesComponent className="product">
                    <section>
                        <h1>Hello</h1>
                    </section>
                </SlidesComponent>
            )
        }),
        {
            enter: sequence(blurIn),
            exit: sequence(blurOut)
        }
    );
};
