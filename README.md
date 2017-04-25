# amsh
amsh (APE MACHINE SHELL) is a command-line shell enhanced with an A.I. chatbot

This uses the API.AI service to pipe all command-line input through.

SCENARIO 1: It finds an appropiate response from API.AI, and echoes that back to the terminal, giving it not only the ability to perform "small talk," but also build executable commands in a more organic, NLP, or user friendly way.

SCENARIO 2: When there is no appropriate response from the chatbot, I have removed all the default fallback content, that just randomizes various strings saying that the bot did not understand, to just one fallback line: escalate-command.
This then gets picked up again from the code, and ran through a method_missing in the code, which tries to include a file from the "bin" directory with that method as a filename, and instantiate the class and run it.
This allows for new commands to be added, and in fact you can add them dynamically without ever restarting your shell.

IMRPOVEMENTS: Before running the method_missing call, I would first like to pass the command off to bash, to see if it can process the command, because then we can get that functionality out of the box as well.
