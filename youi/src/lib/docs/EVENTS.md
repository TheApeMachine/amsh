# YouI Event System

A powerful and flexible event management system that combines a global event bus with DOM event handling.

## 🚀 Getting Started

First, initialize the event management system in your application:

```ts
import { eventBus, EventManager } from "@/lib/event";

const eventManager = EventManager();
eventManager.init();
```

## 🎯 Event Bus: Subscribe & Publish

### Basic Subscription

```ts
eventBus.subscribe("stateChange", (payload) => {
    console.log("State changed:", payload);
});
```

### Conditional Subscription

```ts
// Only trigger when specific conditions are met
eventBus.subscribe(
    "stateChange",
    (payload) => {
        console.log("Login state changed:", payload);
    },
    (payload) => payload.key === "isLoggedIn"
);

// Multiple subscribers for the same event
eventBus.subscribe(
    "userAction",
    (payload) => {
        console.log("User performed action:", payload);
    },
    (payload) => payload.type === "click"
);
```

### Publishing Events

```ts
// Basic event publishing
eventBus.publish("stateChange", { key: "isLoggedIn", value: true });

// Publishing with complex payload
eventBus.publish("userAction", {
    type: "click",
    target: "submitButton",
    timestamp: Date.now(),
    data: { formId: "login-form" }
});
```

### Event Queuing

The event system includes automatic event queuing to handle race conditions where events might be published before subscribers are ready:

```ts
// Event published before any subscribers
eventBus.publish("earlyEvent", { data: "important" });

// Later subscription will still receive the queued event
eventBus.subscribe("earlyEvent", (payload) => {
    console.log("Received queued event:", payload);
});
```

Events are:

- Automatically queued if published before subscribers exist
- Processed in order when subscribers are added
- Cleared from queue after processing
- Logged for debugging purposes

## 🖱️ DOM Event Handling

### Basic DOM Event Attributes

Add data attributes to your HTML elements to enable automatic event handling:

```html
<!-- Basic click event -->
<button data-event="buttonClicked" data-trigger="click">Click Me</button>

<!-- Event with variant -->
<div data-event="hover" data-trigger="mouseenter" data-variant="primary">
    Hover Me
</div>

<!-- Directional event -->
<div data-event="scroll" data-trigger="wheel" data-direction="vertical">
    Scroll Me
</div>
```

### Custom Event Handlers

```ts
// Add custom event handler
eventManager.addEvent("mousemove", (event) => {
    console.log("Mouse position:", event.clientX, event.clientY);
});

// Scoped event listener
const container = document.querySelector(".container");
if (container instanceof HTMLElement) {
    eventManager.addScopedEventListener(container, "scroll", (event) => {
        console.log("Container scrolled");
    });
}
```

## 🏗️ Component Lifecycle Management

Track when elements are added to or removed from the DOM:

```ts
const myElement = document.createElement("div");
myElement.setAttribute("data-event", "myCustomEvent");

eventManager.manageComponentLifecycle(
    myElement,
    () => {
        console.log("Element mounted");
        // Setup component-specific listeners
    },
    () => {
        console.log("Element unmounted");
        // Cleanup listeners
    }
);

// Example usage
document.body.appendChild(myElement); // Triggers mount
document.body.removeChild(myElement); // Triggers unmount
```

## 🔄 State Management Integration

The event system automatically integrates with the state manager:

```ts
// State changes through events
eventBus.publish("stateChange", {
    key: "theme",
    value: "dark"
});

// Listen for specific state changes
eventBus.subscribe(
    "stateChange",
    (payload) => {
        console.log("Theme changed to:", payload.value);
    },
    (payload) => payload.key === "theme"
);
```

## 🛠️ Best Practices

1. **Event Naming**: Use descriptive, action-based names (e.g., `userLoggedIn`, `dataLoaded`, `modalClosed`)
2. **Error Handling**: Event listeners are automatically wrapped in try-catch blocks
3. **Conditional Subscriptions**: Use conditions to filter events and reduce unnecessary callback executions
4. **Cleanup**: Always remove event listeners when components are unmounted
5. **Type Safety**: The system includes TypeScript support for better type checking
6. **Event Queuing**: Let the system handle race conditions by utilizing the built-in event queue

## 🌟 Conclusion

This event system provides a robust foundation for handling both application-level events and DOM interactions. It combines the flexibility of a publish/subscribe pattern with the convenience of declarative DOM event handling.

For more complex scenarios or custom implementations, you can extend the system using the provided methods or create custom event handlers as needed.
