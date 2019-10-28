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

1. First you will need to initialize the transports you would like Optimizer to use:
    `dialer := proxy.Direct
	obfs4Transport := obfs4.Transport{
		CertString: "InsertCertStringHere",
		IatMode:    0 or 1,
		Address:    "InsertAddressHere",
		Dialer:     dialer,
	}`
	`shadowTransport := shadow.Transport{"InsertPasswordHere", "InsertCipherNameHere", "InsertAddressHere"}`
	
2. Create an array with these transports:
    `transports := []Transport{obfs4Transport, shadowTransport}`
    
3. Initialize the strategy of your choice using the array of transports you created:
    `strategy := NewMinimizeDialDuration(transports)`
    
4. Create an instance of OptimizerConnectionFactory using your new Strategy instance:
    `optimizerTransport := NewOptimizerClient(transports, strategy)`
    
5. Call Dial on optimizerTransport:
    `_, err := optimizerTransport.Dial()`
