# YouI Dynamic Island

A powerful, configuration-driven component system that can morph between different states based on JSON configurations and dynamic data sources.

## 🚀 Getting Started

Initialize a Dynamic Island in your application:

```tsx
import { DynamicIsland } from "@/components/ui/DynamicIsland";

const MyPage = () => (
    <DynamicIsland
        variant="page"
        sources={{
            navigation: {
                import: "@/data/navigation.json"
            }
        }}
    />
);
```

## 🏗️ Architecture

The Dynamic Island consists of three main parts:

1. **Core Component**: Manages the overall structure and lifecycle
2. **Source Loader**: Handles data fetching and imports
3. **Content Builder**: Constructs the DOM based on configuration

### Core Structure

```plaintext
<div class="dynamic-island">
    <header>...</header>
    <aside>...</aside>
    <main>...</main>
    <article>...</article>
    <footer>...</footer>
</div>
```

## 📝 Configuration

### Basic Configuration Structure

```json
{
    "aside": {
        "styles": {
            "display": "flex",
            "flexDirection": "column"
        },
        "children": [
            {
                "tag": "nav",
                "source": [
                    // Source configuration
                ]
            }
        ]
    }
}
```

### Element Configuration

```json
{
    "tag": "button",
    "text": "{{category}}",
    "styles": {
        "display": "flex",
        "padding": "var(--sm)"
    },
    "events": {
        "trigger": "click",
        "effects": [
            {
                "target": "ul",
                "styles": {
                    "height": "auto"
                }
            }
        ]
    }
}
```

## 🔄 Data Sources

### Source Configuration

```tsx
interface SourceConfig {
    import?: string; // Dynamic imports
    fetch?: string; // API endpoints
    static?: any; // Static data
}

const sources = {
    navigation: {
        import: "@/data/navigation.json"
    },
    users: {
        fetch: "https://api.example.com/users"
    }
};
```

### Template Variables

Use double curly braces for data binding:

```json
{
    "text": "{{category}}",
    "children": [
        {
            "text": "{{collections.label}}"
        }
    ]
}
```

## 🎨 Animations

### GSAP Integration

The Dynamic Island uses GSAP for smooth transitions:

```json
"events": {
    "trigger": "click",
    "effects": [
        {
            "target": "ul > li",
            "styles": {
                "opacity": "1",
                "transform": "translateX(0)"
            }
        }
    ]
}
```

### Timeline Management

Animations are managed using GSAP timelines:

```ts
const timeline = gsap.timeline({ paused: true });
timeline.to(targets, {
    ...effect.styles,
    duration: 0.3,
    ease: "power4.inOut",
    stagger: 0.05
});
```

## 🔧 Event Handling

### Event Configuration

```json
"events": {
    "trigger": "click",
    "effects": [
        {
            "target": "selector",
            "styles": {
                "property": "value"
            }
        }
    ]
}
```

### Event Bus Integration

```ts
eventBus.subscribe("toggleAccordion", (payload) => {
    // Handle accordion toggle
});
```

## 🛠️ Best Practices

1. **Configuration Organization**

    - Keep configurations in separate files
    - Use semantic naming for variants
    - Leverage CSS variables for consistent styling

2. **Performance**

    - Use dynamic imports for large data sources
    - Cache GSAP timelines
    - Minimize DOM operations

3. **Maintainability**

    - Keep configurations flat when possible
    - Use template variables consistently
    - Document complex animations

4. **Accessibility**
    - Include ARIA attributes in configurations
    - Ensure keyboard navigation
    - Maintain proper focus management

## 🔍 Advanced Usage

### Nested Sources

```json
{
    "source": [
        {
            "children": [
                {
                    "source.collections": [
                        {
                            "tag": "li",
                            "text": "{{collections.label}}"
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Custom Effects

```json
"effects": [
    {
        "target": "ul",
        "styles": {
            "height": "auto",
            "overflow": "visible",
            "marginTop": "var(--xs)"
        }
    }
]
```

## 🌟 Conclusion

The Dynamic Island system provides a flexible, configuration-driven approach to building complex UI components. Its combination of declarative configuration, dynamic data loading, and smooth animations makes it ideal for creating morphing interfaces.

For more complex scenarios, you can extend the system by adding new source types, animation effects, or event handlers as needed.
