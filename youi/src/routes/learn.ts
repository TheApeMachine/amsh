import { loader } from "@/lib/loader"
import { html } from "@/lib/template"
import { match } from "@/lib/match"
import { blurIn, blurOut, sequence, Transition } from "@/lib/transition"
import "@/components/slides/zlide"

export const render = async () => {

    return Transition(match(await loader({}), {
        loading: () => {
            return html`<div>Loading...</div>`
        },
        error: (error: any) => {
            return html`<div>Error: ${error}</div>`
        },
        success: (_: any) => {
            return html`<youi-zlide></youi-zlide>`
        }
    }), {
        enter: sequence(blurIn),
        exit: sequence(blurOut)
    })
    
}