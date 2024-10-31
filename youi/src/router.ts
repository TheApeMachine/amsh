import { Transition } from '@/lib/transition';
import gsap from 'gsap';

interface Route {
    path: string;
    view: () => Promise<Node>;
    effect: () => Promise<void | null>;
}

const routes: Route[] = [{
    path: "/404",
    view: async () => {
        const module = await import("@/routes/notfound");
        return await module.render();
    },
    effect: async () => {
        return null;
    }
}, {
    path: "/",
    view: async () => {
        const module = await import("@/routes/home");
        return await module.render();
    },
    effect: async () => {
        const module = await import("@/routes/home");
        return await module.effect();
    }
}, {
    path: "/chat",
    view: async () => {
        const module = await import("@/routes/chat");
        return await module.render();
    },
    effect: async () => {
        return null;
    }
}, {
    path: "/product",
    view: async () => {
        const module = await import("@/routes/product");
        return await module.render();
    },
    effect: async () => {
        return null;
    }
}, {
    path: "/learn",
    view: async () => {
        const module = await import("@/routes/learn");
        return await module.render();
    },
    effect: async () => {
        return null;
    }
}];

let currentPath = '';

export const router = async (targetElement: HTMLElement) => {
    const path = window.location.pathname;

    // Always render if it's the initial load or the path has changed
    if (path !== currentPath) {
        const previousContent = targetElement.firstChild as HTMLElement;
        currentPath = path;

        const route = routes.find(route => route.path === path) || routes[0]; // Default to first route if no match

        try {
            // Apply exit transition to the current route content
            if (previousContent) {
                Transition(previousContent, {
                    exit: (el: HTMLElement) => gsap.to(el, { opacity: 0, duration: 0.5, ease: "power2.in" })
                });
                await new Promise(resolve => setTimeout(resolve, 500)); // Wait for exit animation to finish
                targetElement.removeChild(previousContent);
            }

            const content = await route.view();
            targetElement.appendChild(content);

            // Apply enter transition for the new route content
            const newContent = targetElement.firstChild as HTMLElement;
            Transition(newContent, {
                enter: (el: HTMLElement) => gsap.from(el, { opacity: 0, duration: 0.5, ease: "power2.out" })
            });

            await route.effect();
        } catch (error: any) {
            console.error("Routing error:", error);
            const errorElement = document.createElement("error-boundary");
            errorElement.textContent = error.message;
            targetElement.innerHTML = '';
            targetElement.appendChild(errorElement);
        }
    }
};

window.addEventListener("popstate", () => {
    const main = document.body;
    if (main) {
        router(main as HTMLElement);
    }
});

export const navigateTo = async (url: string, targetElement: HTMLElement) => {
    if (url !== currentPath) {
        history.pushState(null, "", url);
        await router(targetElement);
    }
};