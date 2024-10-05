interface Route {
    path: string;
    view: () => Promise<DocumentFragment>;
}

const routes: Route[] = [{
    path: "/404",
    view: async () => {
        const module = await import("@/routes/notfound");
        return await module.render();
    }
}, {
    path: "/",
    view: async () => {
        const module = await import("@/routes/home");
        return await module.render();
    }
}];

let currentPath = '';

export const router = async (targetElement: HTMLElement) => {
    const path = window.location.pathname;

    // Always render if it's the initial load or the path has changed
    if (path !== currentPath) {
        currentPath = path;
        const route = routes.find(route => route.path === path) || routes[0]; // Default to first route if no match

        try {
            const content = await route.view();
            targetElement.innerHTML = '';
            targetElement.appendChild(content);
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
    const main = document.querySelector('layout-component');
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