import { loader } from "@/lib/loader"
import { sequence, Transition, blurIn, blurOut } from "@/lib/transition"
import { html } from "@/lib/template"
import { match } from "@/lib/match"
import { switchLayer } from "@/lib/layer"
import { gsap } from "gsap"
import * as vega from 'vega';
import * as SandDance from '@msrvida/sanddance';
import "@/components/layers/manager"
import "@/components/slides/presentation"
import "@/components/slides/zlide"
import "@/components/animoji/assistant"
import "@/components/island/dynamic"
import "@/components/ui/progress"
import "@/components/ui/popover"
import "@/components/ui/scrollmask"
import "@/components/sanddance/component"

export const effect = async () => {
    let layers: Record<string, HTMLElement> = {};
    let positions: string[] = ["product"];
    let distance: number = 1000;
    let currentLayer: number | undefined = 1;

    layers = {
        product: document.querySelector("slides-component") as HTMLElement,

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
                currentLayer, parseInt(evt.key, 10), positions, layers, distance
            );
        }
    });

    window.addEventListener("keydown", (evt: KeyboardEvent) => {
        if (["4"].includes(evt.key)) {
            const el = document.querySelector('toast-container') as ToastContainer;
            console.log("el", el);
            el.addToast('Success!', 'success');
        }
    });

    const data = [
        { myX: 0, myY: 0, myZ: 0 },
        { myX: 1, myY: 1, myZ: 1 },
        { myX: 2, myY: 2, myZ: 2 },
    ];

    const viewer = new SandDance.Viewer(document.querySelector('#vis') as HTMLElement);
    const insight = {
        columns: {
            x: 'myX',
            y: 'myY',
            z: 'myZ'
        },
        size: {
            height: 700,
            width: 700
        },
        chart: SandDance.Chart.Scatterplo
    };

    viewer.render({ insight: insight }, data);
}

export const render = async () => {

    return Transition(match(await loader({}), {
        loading: () => {
            return html`<div>Loading...</div>`
        },
        error: (error: any) => {
            return html`<div>Error: ${error}</div>`
        },
        success: (_: any) => {
            return html`
                <slides-component>
                    <section>
                        <dynamic-island></dynamic-island>
                    </section>
                    <section>
                        <resizable></resizable>
                    </section>
                </slides-component>
                <div id="vis"></div>
            `
        }
    }), {
        enter: sequence(blurIn),
        exit: sequence(blurOut)
    })
    
}