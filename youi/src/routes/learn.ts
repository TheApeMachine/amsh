import { loader } from "@/lib/loader"
import { html } from "@/lib/template"
import { match } from "@/lib/match"
import { blurIn, blurOut, sequence, Transition } from "@/lib/transition"
import "@/components/agentviz/conversation"

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
                        <conversation-visualizer></conversation-visualizer>
                    </section>
                </slides-component>
            `
        }
    }), {
        enter: sequence(blurIn),
        exit: sequence(blurOut)
    })
    
}