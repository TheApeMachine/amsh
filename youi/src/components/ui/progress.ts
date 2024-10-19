import gsap from "gsap";

class YouiProgress extends HTMLElement {
    template: HTMLTemplateElement;
    tl: gsap.core.Timeline = gsap.timeline();

    constructor() {
        super();
        this.template = document.createElement("template");
        this.template.innerHTML = `
            <style>
                progress[value] {
                    /* Reset the default appearance */
                    -webkit-appearance: none;
                    appearance: none;

                    width: 20px;
                    height: 20px;
                }

                progress[value]::-webkit-progress-bar {
                    background-color: #eee;
                    border-radius: 2px;
                    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.25) inset;
                    border-radius: 20px;
                }

                progress[value]::-webkit-progress-value {
                    -webkit-linear-gradient(-45deg, transparent 33%, rgba(0, 0, 0, .1) 33%, rgba(0,0, 0, .1) 66%, transparent 66%),
                    -webkit-linear-gradient(top, rgba(255, 255, 255, .25), rgba(0, 0, 0, .25)),
                    -webkit-linear-gradient(left, #09c, #f44);
                    border-radius: 2px; 
                    background-size: 35px 20px, 100% 100%, 100% 100%;
                    border-radius: 20px;
                }
            </style>
            <progress value="50" max="100"></progress>
            <progress value="50" max="100"></progress>
            <progress value="50" max="100"></progress>
        `;
        this.attachShadow({ mode: "open" });
    }

    connectedCallback() {
        this.shadowRoot?.appendChild(this.template.content.cloneNode(true));
        this.animate();
    }

    animate() {
        const progress = this.shadowRoot?.querySelectorAll("progress") as NodeListOf<HTMLProgressElement>;
        if (!progress) return;
        
        this.tl.set(progress, {
            width: 20,
            ease: "back.inOut(1.2)",
        }).to(progress, {
            width: 100,
            duration: 1,
            ease: "back.inOut(1.2)",
        }).to(progress, {
            value: 100,
            duration: 3,
            ease: "back.inOut(1.2)",
        }).to(progress, {
            width: 20,
            duration: 1,
            ease: "back.inOut(1.2)",
        });
    }
}

customElements.define("youi-progress", YouiProgress);
