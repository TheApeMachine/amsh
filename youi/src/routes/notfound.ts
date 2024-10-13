import { Transition } from "@/lib/transition"
import { html } from "@/lib/template"
import { match } from "@/lib/match"
import { loader } from "@/lib/loader"
import { sequence, blurIn, blurOut } from "@/lib/transition"

export const render = async () => {
    return Transition(match(await loader({}), {
        loading: () => {
            return html`<div>Loading...</div>`
        },
        error: (error: any) => {
            return html`<div>Error: ${error}</div>`
        },
        success: (data: any) => {
            return html`<div>404 - Not Found</div>`
        }
    }), {
        enter: sequence(blurIn),
        exit: sequence(blurOut)
    })
}