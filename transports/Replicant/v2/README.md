# Overview of Replicant for Go

# Operator Foundation

[Operator](https://operatorfoundation.org) makes usable tools to help people around the world with censorship, security, and privacy.

## Shapeshifter

The Shapeshifter project provides network protocol shapeshifting technology (also sometimes referred to as obfuscation). The purpose of this technology is to change the characteristics of network traffic so that it is not identified and subsequently blocked by network filtering devices.

There are two components to Shapeshifter: transports and the dispatcher. Each transport provides a different approach to shapeshifting. These transports are provided as a Go library which can be integrated directly into applications. The dispatcher is a command line tool which provides a proxy that wraps the transport library. It has several different proxy modes and can proxy both TCP and UDP network traffic.

If you are an application developer working in the Go programming language, then you probably want to use the transports library directly in your application. If you are an end user that is trying to circumvent filtering on your network or you are a developer that wants to add pluggable transports to an existing tool that is not written in the Go programming language, then you probably want the dispatcher. Please note that familiarity with executing programs on the command line is necessary to use this tool. You can find Shapeshifter Dispatcher here: <https://github.com/OperatorFoundation/shapeshifter-dispatcher>

If you are looking for a complete, easy-to-use VPN that incorporates shapeshifting technology and has a graphical user interface, consider [Moonbounce](https://github.com/OperatorFoundation/Moonbounce), an application for macOS which incorporates Shapeshifter without the need to write code or use the command line.

### Shapeshifter Transports
The transports implement the Pluggable Transports version 2.1 specification, which is available here: <https://www.pluggabletransports.info/spec/#build> Specifically, they implement the Go Transports API v2.1.

The purpose of the transport library is to provide a set of different transports. Each transport implements a different method of shapeshifting network traffic. The goal is for application traffic to be sent over the network in a shapeshifted form that bypasses network filtering, allowing the application to work on networks where it would otherwise be blocked or heavily throttled.

#### Replicant
Replicant is Operator's Pluggable Transport that can be tuned for each adversary. It is designed to be more effective and efficient than older transports. It can be quickly reconfigured as filtering conditions change by updating just the configuration file.

There are two main advantages to using Replicant. First, it can be more effective than other transports. Simple transports such as shadowsocks work well against some adversaries, but other adversaries with more advanced filtering techniques can easily block them. In situations such as this, Replicant can work where other transports fail. Second, Replicant can be more efficient than other transports. Some transports that are very effective at circumventing the filter are also very inefficient, using a lot of bandwidth in order to implement their approach to shapeshifting. This can make it very expensive to run proxy servers using these transports. Replicant is designed to use the minimum amount of additional bandwidth in order to provide shapeshifting, and can therefore save on the cost of providing proxy servers. Less additional bandwidth used also means a faster connection and more reliable performance on slower Internet connections.

#### Installation

The Shapeshifter transports are written in the Go programming language. To compile it you need to install Go:

<https://golang.org/doc/install>

If you already have Go installed, make sure it is a compatible version:

   go version

The version should be 1.14 or higher.

If you get the error "go: command not found", then try exiting your terminal and starting a new one.

In order to use a transport in your project, you must have Go modules enabled in your project. How to do this is beyond the scope of this document. You can find more information about Go modules here: <https://blog.golang.org/using-go-modules>

To use in your project, simply import the specific transport that you want to use, for example:

   import "github.com/OperatorFoundation/shapeshifter-transports/transports/Replicant/v2"
  
Your Go build tools should automatically add this module to your go.mod and go.sum files. Otherwise, you can add it to the go.mod file directly. See the official [Go modules guide](https://blog.golang.org/using-go-modules) for more information on this.   

Please note that the import path includes "/v2" to indicate that you want to use the version of the module compatible with the PT v2.0 specification and above. This is required by the Go modules guide. The modules are actually compatible with the PT2.1 specification, but Go modules only let you specify the major version number, v2.

When you build your project, it should automatically fetch the correct version of the transport module.

#### Using the Library

Like all Shapeshifter transports, Replicant is provided as a Go library. In order to use Replicant, you will need to integrate it into your Go application. For details on the Go API for Pluggable Transports, refer to the [Pluggable Transports v2.1 specification](https://www.pluggabletransports.info/spec/#build).

### Shapeshifter Dispatcher
The purpose of the dispatcher is to provide different proxy interfaces to using transports. Through the use of these proxies, application traffic can be sent over the network in a shapeshifted form that bypasses network filtering, allowing the application to work on networks where it would otherwise be blocked or heavily throttled.

The dispatcher currently supports several modes and transports, not all of which are covered here. For more information on the various modes and transports available, refer to the [Shapeshifter Dispatcher documentation](https://github.com/OperatorFoundation/shapeshifter-dispatcher/edit/master/README.md)

#### Installation

The dispatcher is written in the Go programming language. To compile it you need to install Go 1.14 or higher:

<https://golang.org/doc/install>

If you already have Go installed, make sure it is a compatible version:

   go version

The version should be 1.14 or higher.

If you get the error "go: command not found", then try exiting your terminal and starting a new one.

Get the git repository for shapeshifter-disptacher:

   git clone https://github.com/OperatorFoundation/shapeshifter-dispatcher.git

Go into that directory and build the command line executable:

   cd shapeshifter-dispatcher
   go build

This command will do all the work. It will fetch the source code for shapeshifter-dispatcher and all the dependencies, then it will compile everything, and then it will produce an executable named shapeshifter-dispatcher.

#### Running

Run without argument to get usage information:

   ./shapeshifter-dispatcher

#### Running with Replicant

Here are example command lines to run the dispatcher with the Replicant transport:

##### Server

For this example to work, you need an application server running. You can use netcat to run a simple server on port 3333:

   nc -l 3333

Now launch the transport server, telling it where to find the application server:

   ./shapeshifter-dispatcher -transparent -server -state state -orport 127.0.0.1:3333 -transports Replicant -bindaddr Replicant-127.0.0.1:2222 -logLevel DEBUG -enableLogging -optionsFile ReplicantServerConfig1.json

This runs the server in transparent TCP proxy mode. The directory "state" is used to hold transport state. The destination that the server will proxy to is 127.0.0.1, port 3333. The Replicant transport is enabled and bound to the address 127.0.0.1 and the port 2222. Logging is enabled and set to DEBUG level. To access the log for debugging purposes, look at state/dispatcher.log.

To use Replicant, a configuration file is needed. A sample configuration file, ReplicantServerConfig1.json, is provided in the git repository purely for educational purposes and should not be used except for testing. Email contact@operatorfoundation.org to request a configuration file to use for real deployment.

##### Client

   ./shapeshifter-dispatcher -transparent -client -state state -target 127.0.0.1:2222  -transports Replicant -proxylistenaddr 127.0.0.1:1443 -optionsFile ReplicantClientConfig1.json -logLevel DEBUG -enableLogging

This runs the client in transparent TCP proxy mode. The arguments are similar to the ones for the server. Please note that the client uses a different configuration file than the server, and that compatible configuration files must be used on both the client and the server in order for Replicant to work properly.

To use Replicant, a configuration file is needed. A sample configuration file, ReplicantClientConfig1.json, is provided purely for educational purposes and should not be used except for testing. Email contact@operatorfoundation.org to request a configuration file to use for real deployment.

Once the client is running, you can connect to the client address, which in this case is 127.0.0.1, port 1443. For instance, you can telnet to this address:

   telnet 127.0.0.1 1443

Any bytes sent over this connection will be forwarded through the transport server to the application server, which in the case of this demo is a netcat server. You can also type bytes into the netcat server and they will appear on the telnet client, once again being routed over the transport.

#### Frequently Asked Questions

##### How do I get a Replicant configuration file?

Replicant requires a configuration file. Each configuration file is tuned to circumvent the filtering of a specific adversary. Therefore, significant time and expertise is needed to create a Replicant configuration file and the process of creating one from scratch is outside of the scope of this document. Email contact@operatorfoundation.org for a tuned configuration file for the adversary of interest.

##### My application is not written in Go. Can I still use the transports?

Yes, the Go API is only one way to integrate transports into your application. There is also an interprocess communication (IPC) protocol that allows you to control a separate process (called the dispatcher) which provides access to the transports through a proxy interface. When using this method, your application can be written in any language. You would need to implement the IPC protocol in order to communicate with the dispatcher. The IPC protocol is specified in the [Pluggable Transports 2.1 specification](https://www.pluggabletransports.info/spec/#build) section 3.3 and an implementation of the dispatcher is available which you can bundle with your application here: <https://github.com/OperatorFoundation/shapeshifter-dispatcher>

##### My question was not answered here, where can I find more information?

The [Pluggable Transports](https://www.pluggabletransports.info/) website includes a great deal of information on how to use all Pluggable Transports.

Here are some articles you might want to read from the Pluggable Transports website:
[What Pluggable Transports do](https://www.pluggabletransports.info/how-transports/)
[Censorship Techniques](https://www.pluggabletransports.info/how/)
[OpenVPN Walkthrough](https://www.pluggabletransports.info/implement/openvpn/)
[Installing a mobile VPN with support for Pluggable Transports](https://www.pluggabletransports.info/implement/mobilevpn)
[How to Use Pluggable Transports in Your Go Application](https://www.pluggabletransports.info/implement/go/)
