# amsh
amsh (APE MACHINE SHELL) is a command-line shell enhanced with an A.I. chatbot

For the A.I. chatbot we use the api.ai service, which handles all the small talk, as well as attempting to encode
natural language into processable commands.
This shell tries to combine the functionality of many tools on the system, so there are 4 steps in which any given input
is processed, detailed below.

PLEASE NOTE: This project is more of a joke than anything else. I was bored...

### STEP 1
Send the user input off to api.ai and try to get a valid response. If this is categorized as small talk, just echo back
the response, and run it through text-to-speech for a futuristic effect.
If the chatbot recognizes this as an actual command, and all the required parameters are in place, send back an encoded
string response that will perform the command inside the shell, using any of the following methods.
If there are required parameters missing, use the api.ai chatbot to ask the user to define them in a natural language way.

### STEP 2
If the chatbot ends up in its fallback state, echo back a singular response "escalate-command" which is caught by the
interpreter, and results in its own fallback mechanism to kick in.
In this step, we hand the command off to the host shell, and see what exit status and output receive.
If the exit status is 0, we know that the command ran successfully, and we are returned back to the amsh prompt, with
the output from the host shell echoed to us.
If the exit status is anything else than 0, we fall back to the following step.

### STEP 3
We now hand the command off to the Ruby interpreter, and see what the result of this is.
If we do not receive an error, we know that the command (line of code), was interpreted, and we can either return to the amsh prompt, or in the case of a block, continue inputing lines of code.
If there is an error, we continue to step 4.
(currently not implemented)

### STEP 4
Finally, after the previous steps all failed to produce correct results, we hand the command off to the internal command mechanism.
This first tries to include a file from the ./bin directory, with the filename equal to the command method, without parameters.
It then converts the method name into a Class, and tries to instantiate that, before running the .run method on the newly instantiated class.
The way this is implemented means, obviously, that additional commands are basically hot-swappable, and there is no need to restart amsh when making changes to, or adding new commands in the ./bin directory.

### INSTALLATION

* Make sure to install all the gems required at the top of amsh.rb
* Install mpg123 (for text-to-speech)

```
sudo apt-get install mpg123 #for debain based
brew install mpg123 #mac
```

### IMPROVEMENTS

* After trying to pass the command to the shell, and the exit status being anything else than 0, try to pass the command
to Ruby itself, and see if we can run it that way. Then, and only then, try to handle it as an internal command.
