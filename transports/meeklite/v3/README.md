# meeklite

meeklite tunnels your network traffic through HTTPS, and uses domain fronting to conceal where your traffic is coming from.

## Using meeklite

1. Parse the Url from a string to a Url:
`unparsedUrl := "https://TheUrl.Here/"
 	Url, _ := url.Parse(unparsedUrl)`
 	
2. Create an instance of a meeklite server:
`meekliteTransport := meeklite.Transport{Url, "Front.Here", "Insert your Address here"`

3. Call dial on meekliteTransport:
`_, err := meekliteTransport.Dial()`

### Go Version:

meeklite is one of the transports available in the [Shapeshifter-Transports library](https://github.com/OperatorFoundation/Shapeshifter-Transports).

