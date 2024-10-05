interface Route {
    path: string;
    view: () => Promise<DocumentFragment>;
}

const routes: Route[] = [{
    path: "/",
    view: async () => {
        const module = await import("@/routes/home");
        return await module.render();
    }
}, {
    path: "/work",
    view: async () => {
        const module = await import("@/routes/work");
        return await module.render();
    }
}, {
    path: "/learn",
    view: async () => {
        const module = await import("@/routes/learn");
        return await module.render();
    }
}, {
    path: "/connect",
    view: async () => {
        const module = await import("@/routes/connect");
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
            targetElement.innerHTML = `<error-boundary><pre>${error.message}</pre></error-boundary>`;
        }
    }
};

window.addEventListener("popstate", () => {
    const main = document.querySelector('layout-component')?.shadowRoot?.querySelector('main');
    if (main) {
        router(main);
    }
});

export const navigateTo = async (url: string, targetElement: HTMLElement) => {
    if (url !== currentPath) {
        history.pushState(null, "", url);
        await router(targetElement);
    }
};