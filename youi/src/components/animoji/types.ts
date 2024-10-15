export type Animoji =
    "crying" 
    | "cursing" 
    | "devil-frown" 
    | "devil-smile" 
    | "dotted-line" 
    | "eyes" 
    | "glowing-star" 
    | "mind-blown" 
    | "monocle" 
    | "nerd-face" 
    | "robot" 
    | "salute" 
    | "shacking" 
    | "smirking" 
    | "sunglasses" 
    | "thinking" 
    | "tornado" 

export const AnimojiIdle: (index: number) => Animoji = gsap.utils.wrap([
    "eyes",
    "robot"
])

export const AnimojiError: (index: number) => Animoji = gsap.utils.wrap([
    "cursing",
    "devil-frown",
    "devil-smile",
    "glowing-star",
    "mind-blown",
    "tornado"
])

export const AnimojiWorking: (index: number) => Animoji = gsap.utils.wrap([
    "salute",
])

export const AnimojiThinking: (index: number) => Animoji = gsap.utils.wrap([
    "thinking",
    "monocle"
])

export const AnimojiConfident: (index: number) => Animoji = gsap.utils.wrap([
    "sunglasses"
])

export const AnimojiUnsure: (index: number) => Animoji = gsap.utils.wrap([
    "dotted-line"
])

export const AnimojiStates: Record<string, (index: number) => Animoji> = {
    idle: AnimojiIdle,
    working: AnimojiWorking,
    thinking: AnimojiThinking,
    confident: AnimojiConfident,
    unsure: AnimojiUnsure
}