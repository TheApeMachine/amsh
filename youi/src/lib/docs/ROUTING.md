# YouI Router

A dynamic, client-side router with built-in transitions and automatic route discovery.

## 🚀 Getting Started

Initialize the router in your application:

```ts
import { createRouter } from "@/lib/router";

const { router, navigateTo } = await createRouter();
const app = document.getElementById("app");
if (app) {
    router(app);
}
```

## 📁 Route Structure

Routes are automatically discovered from the `@/routes` directory. Each route file should export a `render` function:

```tsx
// @/routes/home.tsx
export const render = async (params: Record<string, string>) => {
    return <div>Home Page</div>;
};

// @/routes/collection.tsx - Dynamic route
export const render = async (params: Record<string, string>) => {
    return <div>Collection {params.id}</div>;
};
```

### Supported Route Patterns

```plaintext
/                   -> home.tsx
/about              -> about.tsx
/collection/:id     -> collection.tsx
/collection/:collection/document/:document -> document.tsx
```

## 🔄 Navigation

### Programmatic Navigation

```ts
// Using the global navigation function
window.navigateTo("/about");

// Using the router instance
navigateTo("/about", targetElement);
```

### Link Elements

Add the following data attribute to your links:

```html
<a href="/about" onclick="navigateTo('/about'); return false;">About</a>
```

## 🎨 Transitions

The router includes built-in page transitions using GSAP:

```ts
// Default transition behavior
exit: (el) => gsap.to(el, { opacity: 0, duration: 0.5, ease: "power2.in" });
enter: (el) => gsap.from(el, { opacity: 0, duration: 0.5, ease: "power2.out" });
```

## 🏗️ Layout System

The router automatically wraps your routes in a layout component:

```tsx
// Custom layout structure
<Layout>
    <div class="dynamic-island">
        <main>{/* Route content is rendered here */}</main>
    </div>
</Layout>
```

## ⚠️ Error Handling

### Built-in 404 Page

The router includes a default 404 page when no matching route is found:

```tsx
// Custom 404 page can be added at @/routes/404.tsx
export const render = async () => {
    return <div>Custom 404 Page</div>;
};
```

### Error Boundary

Route rendering errors are caught and displayed:

```html
<error-boundary>Error Message</error-boundary>
```

## 🛠️ Best Practices

1. **Route Organization**: Keep route files in the `@/routes` directory with clear, descriptive names
2. **Dynamic Routes**: Use parameter notation (`:paramName`) for dynamic segments
3. **Transitions**: Keep transitions subtle and quick for better user experience
4. **Error Handling**: Always provide fallback UI for error states
5. **Type Safety**: The router includes TypeScript support for better type checking

## 🔍 Route Matching

The router uses a sophisticated matching system:

```ts
// Examples of how routes are matched
/about              -> matches /about
/collection/:id     -> matches /collection/123
/collection/:collection/document/:document
                    -> matches /collection/blog/document/post-1
```

## 🌟 Conclusion

This routing system provides a seamless, transition-enabled navigation experience with automatic route discovery and error handling. It's designed to work with the YouI framework's performance and developer ergonomics goals.

For more complex scenarios, you can extend the routing system by modifying the route discovery process or adding custom transition effects.
