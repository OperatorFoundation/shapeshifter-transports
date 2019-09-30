# Optimizer


Optimizer is a pluggable transport that uses one of several possible “Strategies” to choose between the transports you provide to create a connection. It is not a standalone transport, but is rather a mechanism for choosing between various transports in order to find the one best suited for the user’s needs. For more information about pluggable transports, please refer to [pluggabletransports.info](https://www.pluggabletransports.info/).

Here is a list of the currently available Optimizer strategies:

**Rotate Strategy**: This strategy simply rotates through the list of provided transports and tries the next one in the list each time a connection is needed.

**Choose Random Strategy**: A transport is selected at random from the list for each connection request.

**Track Strategy**: A strategy that  attempts to connect with each of the provided transports. It keeps track of which transports are connecting successfully and favors using those.

**Minimize Dial Strategy**: The transport is chosen based on which has been shown to connect the fastest.

## Using Optimizer

### Go Version:

Optimizer is one of the transports available in the [Shapeshifter-Transports library](https://github.com/OperatorFoundation/Shapeshifter-Transports).