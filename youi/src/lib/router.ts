import { Transition } from '@/lib/transition';
import gsap from 'gsap';

interface RouteModule {
    render: (params: Record<string, string>) => Promise<Node>;
}

export interface Route {
    path: string;
    view: (params: Record<string, string>) => Promise<Node>;
}

// Function to load and discover all route files dynamically
async function discoverRoutes(): Promise<Route[]> {
    const routeModules = import.meta.glob('@/routes/*.{ts,tsx}');
    const routes: Route[] = [];

    for (const [path, importFn] of Object.entries(routeModules)) {
        const moduleName = path.replace(/.*\/(.*)\.ts[x]?$/, '$1');
        const module = await importFn() as RouteModule;
        if (module.render) {
            if (moduleName === 'collection') {
                routes.push({
                    path: '/collection/:id',
                    view: async (params) => module.render(params)
                });
            } else {
                routes.push({
                    path: moduleName === 'home' ? '/' : `/${moduleName}`,
                    view: async (params) => module.render(params)
                });
            }
        }
    }

    return routes;
}

let currentPath = '';

export const createRouter = async () => {
    const routes = await discoverRoutes();

    const router = async (targetElement: HTMLElement) => {
        const path = window.location.pathname;

        if (path !== currentPath) {
            const previousContent = targetElement.firstChild as HTMLElement;
            currentPath = path;

            const { route, params } = matchRoute(path, routes);

            try {
                if (previousContent) {
                    Transition(previousContent, {
                        exit: (el: HTMLElement) => gsap.to(el, { opacity: 0, duration: 0.5, ease: "power2.in" })
                    });
                    await new Promise(resolve => setTimeout(resolve, 500));
                    targetElement.removeChild(previousContent);
                }

                const content = await route.view(params);
                targetElement.appendChild(content);

                const newContent = targetElement.firstChild as HTMLElement;
                Transition(newContent, {
                    enter: (el: HTMLElement) => gsap.from(el, { opacity: 0, duration: 0.5, ease: "power2.out" })
                });
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

    return { router, navigateTo: createNavigateTo(router) };
};

const createNavigateTo = (router: (targetElement: HTMLElement) => Promise<void>) => {
    return async (url: string, targetElement: HTMLElement) => {
        if (url !== currentPath) {
            history.pushState(null, "", url);
            await router(targetElement);
        }
    };
};

// Function to match the current path against the available routes
const matchRoute = (path: string, routes: Route[]) => {
    for (const route of routes) {
        const paramNames: string[] = [];
        const regexPath = route.path
            .replace(/\/:([^/]+)/g, (_, paramName) => {
                paramNames.push(paramName);
                return '/([^/]+)';
            })
            .replace(/\//g, '\\/');

        const pathRegex = new RegExp(`^${regexPath}$`);
        const match = path.match(pathRegex);

        if (match) {
            const params: Record<string, string> = {};
            paramNames.forEach((name, index) => {
                params[name] = match[index + 1];
            });
            return { route, params };
        }
    }

    // Add fallback if no route matches and no 404 page exists
    const notFoundRoute = routes.find(route => route.path === "/404") ?? {
        path: "/404",
        view: async () => {
            const container = document.createElement('div');
            container.style.cssText = `
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                width: 100%;
                height: 100vh;
                font-family: system-ui, -apple-system, sans-serif;
                color: #374151;
                text-align: center;
                padding: 0 1rem;
            `;

            const heading = document.createElement('h1');
            heading.textContent = '404';
            heading.style.cssText = `
                font-size: 6rem;
                font-weight: bold;
                margin: 0;
                color: #1F2937;
            `;

            const message = document.createElement('p');
            message.textContent = 'Page Not Found';
            message.style.cssText = `
                font-size: 1.5rem;
                margin: 1rem 0;
            `;

            const subMessage = document.createElement('p');
            subMessage.textContent = `The page you are looking for doesn't exist or has been moved.`;
            subMessage.style.cssText = `
                color: #6B7280;
                margin-top: 0.5rem;
            `;

            container.appendChild(heading);
            container.appendChild(message);
            container.appendChild(subMessage);

            return container;
        }
    };

    return { route: notFoundRoute, params: {} };
};
