import { jsx } from "@/lib/template";
import { Component } from "@/lib/ui/Component";
import Reveal from "reveal.js";
interface LayoutProps {
    children?: JSX.Element;
}

/** Layout Component - A wrapper component for page content */
export const Layout = Component<LayoutProps>({
    effect: () => {
        Reveal.initialize();
    },
    render: async ({ children }) => {
        return (
            <div className="reveal">
                <div className="slides">
                    <section>Slide 1</section>
                    <section>Slide 2</section>
                </div>
            </div>
        );
    }
});
