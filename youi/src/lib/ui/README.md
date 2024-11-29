# YouI Dynamic Island

The Dynamic Island is a component that is able to become many different things. In face, in the YouI framework, everything is a Dynamic Island.

## Concept

To make this possible, the Dynamic Island is built from the following parts:

- The container div, to which you must never add or remove any elements, but you can apply styles to it.
  One thing to note is that it has to be rendered with `display: grid` or `display: grid-inline`.
- The `header`, `aside`, `main`, `article`, and `footer` parts, which are the actual parts that you can add or remove elements from.
- A configuration file, which describes the styles of the container and each of the parts.

## Required Styles

The styling rules below is what makes everything work. Because of the way the CSS grid layout work, when there is no content in a part, it will not take up any space.

```css
.dynamic-island {
  display: grid;
    grid-template-columns: auto 1fr auto;
    grid-template-rows: auto 1fr auto;
    grid-template-areas:
      "header header header"
      "aside main article"
      "footer footer footer";
}

header {
  grid-area: header;
}

aside {
  grid-area: aside;
}

main {
  grid-area: main;
}

article {
  grid-area: article;
}

footer {
  grid-area: footer;
}
```

Coupled with the various elements you can add content to, it allows for a lot of flexibility.

For instance, in the case of a full page, you can use the header and footer, and have the main area take up the entire space. The aside can be the sidebar, or it can be empty. And the article could serve as a flyout menu, or it can be empty.

In the case of an accordion, you can use the header to create the toggle, and have the main area house the content.

A form field could use the header to display a label, and the main area to display the input, and the footer to display additional information such as validation messages.

The real beauty comes from the ability to now morph these components into one another at will.

Think of a button that can become a dropdown, or a flyout menu, or a modal. A form field that can become a dialog.

The possibilities are endless.
