import { loader } from "@/lib/loader"
import { sequence, Transition, blurIn, blurOut } from "@/lib/transition"
import { html } from "@/lib/template"
import { match } from "@/lib/match"
import "../components/dashboard/editor"

export const render = async () => {

    return Transition(match(await loader({
        users: {url: "https://fakestoreapi.com/products", method: "GET", params: {active: true}}
    }), {
        loading: () => {
            return html`<div>Loading...</div>`
        },
        error: (error: any) => {
            return html`<div>Error: ${error}</div>`
        },
        success: (data: any) => {
            window.stateManager.setState("users", data);    
            return html`<dashboard-editor></dashboard-editor>`
        }
    }), {
        enter: sequence(blurIn),
        exit: sequence(blurOut)
    })
    
}