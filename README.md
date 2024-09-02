# amsh

The Ape Machine Shell is the personal terminal workspace of Daniel Owen van Dommelen, a.k.a. The Ape Machine.

It is primarily based on his personal Neovim configuration, but adds various features, the most important of which 
is a deep integration with Large Language Models to improve the programming experience using A.I.

## Features

- **File Browser**: A file browser that allows you to navigate your filesystem.
- **Buffer**: A buffer that allows you to edit files.
- **Chat**: A chat that allows you to chat with the A.I.
- **Terminal**: A terminal that allows you to run commands.
- **Tasks**: A task manager that allows you to manage your tasks.
- **Notes**: A note manager that allows you to manage your notes.
- **Bookmarks**: A bookmark manager that allows you to manage your bookmarks.
- **Settings**: A settings manager that allows you to manage your settings.

## Architecture

amsh is built using the Bubble Tea framework, which is a popular choice for building terminal applications in Go.

The application is organized into several packages:

- **buffer**: Acts as the main container for the applications and is meant to multiplex components of the application. It is relatively unaware of the
specifics of the components, and merely receives messages which contain everything from view updates to commands. It does not implement any behavior
regarding commands, except for rendering the main composed view. Components can register with the buffer to send and receive messages.
- **filebrowser**: A fully featured file browser that allows you to navigate your filesystem, create, delete, and edit files and directories.
- **editor**: A component that provides a fully featured editor with line numbers, syntax highlighting, code folding, autocompletion, and more.
- **statusbar**: A component that provides a status bar that displays information about the current buffer and mode.
- **layout**: A component that can be used to automatically layout components within the buffer.

Each component is a separate entity that can be used to build the application. They are all designed to be modular and can be used independently of each other.