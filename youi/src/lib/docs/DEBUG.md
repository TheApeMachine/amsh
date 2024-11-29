# YouI Debug System

A powerful, modular debugging system with real-time visualization, time-travel capabilities, and advanced monitoring features.

## 🚀 Getting Started

Initialize the debug system in your application:

```ts
import { createDebugOverlay } from "@/lib/debug";

createDebugOverlay();
```

## 🎯 Core Features

### Main Debug Overlay

The debug overlay provides a comprehensive view of your application's state and behavior, divided into several sections:

1. Console Logs
2. Network Activity
3. Time Travel
4. Code Smells
5. Digital Twin
6. Chaos Testing

### Layout and Controls

The overlay is fully draggable and resizable, with a minimizable interface:

```html
<div class="debug-overlay">
    <div class="debug-header">
        <span class="debug-header-title">Debug Tools</span>
        <div class="debug-header-controls">
            <button id="clear-logs">Clear</button>
            <button id="expand-all">Expand All</button>
            <button id="minimize-debug">−</button>
        </div>
    </div>
    <div class="debug-content-grid">
        <!-- Debug sections -->
    </div>
</div>
```

## 🔍 Feature Modules

### Network Visualization

3D interactive graph showing all network requests and their relationships:

```ts
const networkView = setup3DNetworkView({ addLog, overlay: networkSection });
```

### Time Travel Debugging

Record and replay application states:

```ts
const timeTravel = setupTimeTravel({ addLog, overlay: timeTravelSection });
```

Features:

-   State recording
-   Snapshot management
-   Timeline visualization
-   State restoration

### Code Smell Detection

Automatic detection of potential issues:

```ts
const codeSmells = setupCodeSmellDetector({
    addLog,
    overlay: codeSmellSection
});
```

Detects:

-   Memory leaks
-   Performance bottlenecks
-   Event listener accumulation
-   DOM mutation patterns

### Digital Twin

Real-time simulation of application behavior:

```ts
const digitalTwin = setupDigitalTwin({ addLog, overlay: digitalTwinSection });
```

Features:

-   User behavior prediction
-   State prediction
-   Performance simulation

### Chaos Testing

Controlled chaos engineering:

```ts
const chaosMonkey = setupChaosMonkey({ addLog, overlay: chaosMonkeySection });
```

Features:

-   Network latency simulation
-   Error injection
-   Resource limitation testing

## 🎨 Visual Styling

The debug overlay uses a dark theme with customizable CSS variables:

```css
.debug-overlay {
    --debug-bg-main: rgba(20, 20, 20, 0.95);
    --debug-text: #ffffff;
    --debug-border: rgba(255, 255, 255, 0.1);
    --debug-spacing-sm: 8px;
    --debug-font-size-base: 12px;
}
```

## 🔄 Log Management

### Adding Logs

```ts
addLog({
    id: crypto.randomUUID(),
    timestamp: new Date().toISOString(),
    type: "info",
    category: "debug.general",
    summary: "Log message",
    details: {
        /* Additional data */
    },
    stack: new Error().stack
});
```

### Log Types

-   Console output
-   Network requests
-   DOM mutations
-   State changes
-   Performance metrics
-   Error tracking

## 🛠️ Best Practices

1. **Performance**: The debug overlay is designed to have minimal impact on application performance
2. **Memory Management**: Automatic cleanup of old logs and snapshots
3. **Modularity**: Each feature can be enabled/disabled independently
4. **Error Handling**: All debug features are isolated from application code
5. **Customization**: Extensible architecture for adding custom debug features

## 🌟 Advanced Features

### Custom Visualizations

Add custom visualizations using the choreographer:

```ts
const debugChoreographer = setupDebugChoreographer({
    addLog,
    overlay: choreographSection
});
```

### State Tracking

Monitor global and component-level state changes:

```ts
const stateTracker = createTrackingProxy(target, "root");
```

## 🔒 Security Considerations

1. Debug overlay is only enabled in development
2. Sensitive data is automatically redacted
3. Network requests are monitored but not modified
4. State snapshots exclude secure information

## 🌟 Conclusion

The YouI Debug System provides a comprehensive suite of debugging tools while maintaining performance and security. It's designed to be both powerful for developers and safe for production use when needed.

For more complex scenarios or custom implementations, you can extend the system using the provided APIs or create custom debug modules as needed.
