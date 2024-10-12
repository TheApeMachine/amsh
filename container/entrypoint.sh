#!/bin/bash

# Use a default username if none is provided
USERNAME=${USERNAME:-devuser}

# Create a new user if it doesn't exist.
# This allows for dynamic user creation at runtime, enhancing container flexibility.
if ! id "$USERNAME" &>/dev/null; then
    useradd -m -s /bin/bash "$USERNAME"
    # Grant sudo access without password for convenience in development environments.
    # Note: This should be used cautiously in production settings.
    echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME
    chmod 0440 /etc/sudoers.d/$USERNAME
fi

# Store the custom message in a file for persistence across sessions
CUSTOM_MESSAGE_FILE="/etc/motd_custom"
echo "$CUSTOM_MESSAGE" > "$CUSTOM_MESSAGE_FILE"

# Modify .bashrc to display the custom message on login.
# This ensures the message is shown each time a new shell session starts,
# which is particularly useful for providing context to language models or users.
echo "cat $CUSTOM_MESSAGE_FILE" >> /home/$USERNAME/.bashrc

# Switch to the specified user and execute the provided command.
# This maintains the principle of least privilege by running as a non-root user.
exec sudo -u "$USERNAME" "$@"