# amsh

The Ape Machine Shell is the personal terminal workspace of Daniel Owen van Dommelen, a.k.a. The Ape Machine.

It is primarily based on his personal Neovim configuration, but adds various features, the most important of which 
is a deep integration with Large Language Models to improve the programming experience using A.I.

It is also able to attempt self-improvement by leveraging the power of A.I. to rewrite itself. To facilitate this, it uses
an integration with the Docker API to build itself into a Docker image and run it in a container. It can then call the
OpenAI API to instantiate various A.I. agents within the container to perform tasks such as writing more of itself,
fixing bugs, and self-reviewing its code to improve its own design.

This README therefore also serves as a guide for how to improve the amsh project itself, and should be used as a reference
for the A.I. to understand the overall design and architecture of the amsh project.

## Quick Reference / Guidelines

- **Visual Components**: Should be split into model.go, update.go, and view.go files, and implement the BubbleTea Model interface.
  - model.go: Contains the state of the component.
  - update.go: Contains the update logic for the component.
  - view.go: Contains the view logic for the component.

- **Purely Functional Components**: Need to implement the io.ReadWriteCloser interface for passing data around.

- **Code Comments**: Should be written in a way that is easy to understand, and should be used to explain the "why" behind the code, as well as any
  "tricky" parts. The what and how can be understood by reading the code, but the why is what is important, and what is often the result of trial
  and error, or a design decision that was made. It also provides a guide for the A.I. to understand the code and make changes to it.
  - Top-level comments, above functions, methods, or types should use the following format:
    ```
    /*
    NameOfThing ... (then use guidelines described above)
    */
    ```
    All other comments nested one or deeper should use the `//` format.

## Goals

- **Unified Workspace**: amsh should provide a unified workspace for development, and related tasks and act as both a terminal and editor.
  - Switching between editor and terminal can be done using an Alt-Screen toggle.
    Example: https://github.com/charmbracelet/bubbletea/tree/master/examples/altscreen-toggle
- **Familiarity**: amsh should take the bulk of its inpiration from Neovimodel, especially in terms of its interface and general mode-based operation.
  - It should implement Neovim's modes, Normal, Insert, and Visual.
  - It should have built-in features for auto-complete.
    Example: https://github.com/charmbracelet/bubbletea/blob/master/examples/autocomplete/main.go
- **Enjoyable**: amsh should be enjoyable to use, and combine functionality with a top-shelf visual experience encoded in the TUI.
  - Using the BubbleTea framework, and various pre-built components will get us very far along the way.
    Examples:
      https://github.com/charmbracelet/bubbletea/tree/master/examples
      https://github.com/charmbracelet/bubbles/tree/master/examples
- **Modularity**: amsh should be highly modular, and each component should be designed to be able to be replaced or improved independently of the others.
  - Every component needs to do one thing, and do that one thing well, and decoupled from any other components. No component should be directly referencing,
    or importing another component. Communication should only happen using the BubbleTea framework's messaging and command systemodel, and each component should
    implement the Update method according to the BubbleTea Model interface, and handle any messages it should somehow respond to.
  - Each component is responsible for its own local state, behavior, and sending messages or commands.
  - The buffer is where everything comes together to render the final, composed TUI corresponding to the current state of the systemodel. The buffer acts as a
    hub for messages and commands, although it can only pass through messages, and render a view. The buffer does not have any specific implementation regarding
    any component, and even the view it renders is merely a concatenation of messages it receives containing view updates from other components. Components
    are able to register themselves with the buffer, opening a two-way message channel.
- **Self-Similar**: amsh should, as much as possible, be composed of components that are all following a similar implementation and design.
  - Each component should have a model.go, update.go, and view.go file, separating the three parts of the BubbleTea Model interface.
  - If a component does not implement a visual representation, it should implement io.ReadWriteCloser for passing data around. For this, the buffer can implement
    a secondary message queue based on goroutines and channels, including a conversion path between messages received on this channel to BubbleTea messages.
- **Self-Improvement**: amsh should be able to improve itself by leveraging the power of A.I. to rewrite itself and improve its own design.
  - The application should leverage its implementation of the Docker API to build and run itself as a Docker container, using an alternative startup command,
    which runs the self-improvement cycle.
  - It should be able to refer to this document to understand the goals and guidelines of this project.
  - It should be able to call the OpenAI API and start sending it code that needs reviewing, and receive back tasks that need to be executed.
  - It should be able to call the OpenAI API and start sending it code that needs modification.
  - It should be able to run itself inside the container, and monitor its output to detect any bugs or other issues.
  - It should be able to self-review its current strategies and long-term horizon, using this document as a source of truth.
  - It should be able to connect with the user that is currently using a running version of the programodel, via a chat window, so progress and plans of attack
    can be verified with the product owner.
  - It should be able to commit its changes to a branch, so they can be verified before merging into the main branch.