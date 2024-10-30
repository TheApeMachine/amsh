# YouI

## 🚀 Getting Started

First things first, let's initialize our event management duo:

With the EventManager up and running, your application is now ready to listen and respond to a variety of events. But wait, there's more!

## 🎯 Subscribing to Events

```ts
eventBus.subscribe("stateChange", (payload) => {
    console.log("State changed:", payload);
});
```

```ts
eventBus.subscribe(
    "stateChange",
    (payload) => {
        console.log("State changed:", payload);
    },
    (payload) => payload.key === "isLoggedIn"
); // Only trigger when key is "isLoggedIn"
```

## 📢 Publishing Events

```ts
eventBus.publish("stateChange", { key: "isLoggedIn", value: true });
```

## 🖱️ Handling DOM Events

Need to handle user interactions like clicks, drags, or wheel movements? The EventManager has you covered:

You can dynamically add or remove event listeners as your component's lifecycle evolves. It's like having a personal assistant for your events! 🤖

## 🏗️ Managing Component Lifecycle

```ts
const myElement = document.createElement("div");
myElement.setAttribute("data-event", "myCustomEvent");

eventManager.manageComponentLifecycle(
    myElement,
    () => console.log("Element was added to the DOM:", myElement),
    () => console.log("Element was removed from the DOM:", myElement)
);

// Append the element to trigger the lifecycle
document.body.appendChild(myElement);
// Remove the element to trigger the unmount lifecycle
document.body.removeChild(myElement);
```

## 🌟 Conclusion
