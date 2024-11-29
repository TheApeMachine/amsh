# YouI Template System

A lightweight, performant JSX-compatible templating system with built-in sanitization and transition support.

## 🚀 Getting Started

Import the template system in your application:

```tsx
// @ts-ignore: JSX factory import is used by the transpiler
import { jsx, Fragment } from "@/lib/template";
```

## 🎯 Core Features

### JSX Support

Write components using familiar JSX syntax:

```tsx
const MyComponent = ({ title, children }) => (
    <div className="container">
        <h1>{title}</h1>
        {children}
    </div>
);
```

### Fragment Support

Group elements without adding extra nodes:

```tsx
const List = () => (
    <>
        <li>Item 1</li>
        <li>Item 2</li>
    </>
);
```

## 🔧 Template Functions

### HTML Template Literal

Use the `html` tagged template literal for string-based templating:

```ts
const template = html`
    <div class="card">
        <h2>${title}</h2>
        <p>${content}</p>
    </div>
`;
```

### Sanitization

Automatic HTML sanitization is enabled by default:

```ts
// Safe by default
const content = "<script>alert('xss')</script>";
html`<div>${content}</div>`; // Sanitizes potentially harmful content

// Opt out when needed
const options = {
    sanitize: false
};
```

## 🎨 Props and Attributes

### Standard Props

```tsx
// Class names
<div className="my-class" />

// Styles
<div style={{ color: 'red', marginTop: '10px' }} />

// Data attributes
<div data-test="value" />

// ARIA attributes
<div role="button" aria-label="Close" />
```

### Event Handlers

```tsx
// Click events
<button onClick={(e) => handleClick(e)}>Click me</button>

// Other events
<input
    onFocus={handleFocus}
    onBlur={handleBlur}
    onChange={handleChange}
/>
```

### Refs

Access DOM elements directly:

```tsx
const MyComponent = () => {
    const handleRef = (element: HTMLElement) => {
        // Do something with the element
        element.focus();
    };

    return <input ref={handleRef} />;
};
```

## 🔄 Transitions

Built-in transition support:

```tsx
const FadeComponent = () => (
    <div
        transitionEnter={(el) => gsap.from(el, { opacity: 0 })}
        transitionExit={(el) => gsap.to(el, { opacity: 0 })}
    >
        Fade me
    </div>
);
```

## 🏗️ Component Patterns

### Function Components

```tsx
// Basic component
const Button = ({ label, onClick }) => (
    <button onClick={onClick}>{label}</button>
);

// With children
const Card = ({ title, children }) => (
    <div className="card">
        <h2>{title}</h2>
        {children}
    </div>
);
```

### List Rendering

```tsx
const List = ({ items }) => (
    <ul>
        {items.map((item) => (
            <li key={item.id}>{item.text}</li>
        ))}
    </ul>
);
```

## 🛠️ Best Practices

1. **Keys**: Always use unique keys when rendering lists
2. **Sanitization**: Keep sanitization enabled unless explicitly needed
3. **Fragments**: Use fragments to avoid unnecessary wrapper elements
4. **Type Safety**: Leverage TypeScript for better type checking
5. **Performance**: Use memoization for expensive computations

## 🔍 Type System

### Component Props

```tsx
interface ButtonProps {
    label: string;
    onClick?: () => void;
    variant?: "primary" | "secondary";
    disabled?: boolean;
}

const Button = ({
    label,
    onClick,
    variant = "primary",
    disabled
}: ButtonProps) => (
    <button
        onClick={onClick}
        className={`btn btn-${variant}`}
        disabled={disabled}
    >
        {label}
    </button>
);
```

### Event Types

```tsx
type EventHandlers = {
    [K in keyof HTMLElementEventMap]?: (event: HTMLElementEventMap[K]) => void;
};

type EventHandlerProps = {
    [K in keyof EventHandlers as `on${Capitalize<K>}`]?: EventHandlers[K];
};
```

## 🌟 Advanced Features

### Conditional Rendering

```tsx
const ConditionalComponent = ({ isVisible, content }) => (
    <div>
        {isVisible && <div>{content}</div>}
        {isVisible ? <span>Visible</span> : <span>Hidden</span>}
    </div>
);
```

### Dynamic Attributes

```tsx
const DynamicComponent = ({ attributes }) => (
    <div {...attributes}>Dynamic Props</div>
);
```

## 🌟 Conclusion

The YouI Template System provides a modern, type-safe approach to building user interfaces. It combines the familiarity of JSX with the performance of native DOM operations, while maintaining security through built-in sanitization.

For more complex scenarios, you can extend the system using custom components, transitions, or event handlers as needed.
