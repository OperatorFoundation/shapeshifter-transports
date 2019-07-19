# The Operator Foundation

[Operator](https://operatorfoundation.org) makes useable tools to help people around the world with censorship, security, and privacy.

## Shapeshifter

The Shapeshifter project provides network protocol shapeshifting technology
(also sometimes referred to as obfuscation). The purpose of this technology is
to change the characteristics of network traffic so that it is not identified
and subsequently blocked by network filtering devices.

There are two components to Shapeshifter: transports and the dispatcher. Each
transport provide different approach to shapeshifting. These transports are
provided as a Go library which can be integrated directly into applications.
The dispatcher is a command line tool which provides a proxy that wraps the
transport library. It has several different proxy modes and can proxy both
TCP and UDP traffic.

If you are a tool developer working in the Go programming language, then you
probably want to use the transports library directly in your application.
<https://github.com/OperatorFoundation/shapeshifter-transports>

If you want a end user that is trying to circumvent filtering on your network or
you are a developer that wants to add pluggable transports to an existing tool
that is not written in the Go programming language, then you probably want the
dispatcher. Please note that familiarity with executing programs on the command
line is necessary to use this tool.
<https://github.com/OperatorFoundation/shapeshifter-dispatcher>

If you are looking for a complete, easy-to-use VPN that incorporates
shapeshifting technology and has a graphical user interface, consider
[Moonbounce](https://github.com/OperatorFoundation/Moonbounce), an application for macOS which incorporates shapeshifting without
the need to write code or use the command line.

### Shapeshifter Transports

This is the repository for the shapeshifter transports library for the Go
programming language. If you are looking for a tool which you can install and
use from the command line, take a look at the dispatcher instead:
<https://github.com/OperatorFoundation/shapeshifter-transports>

The transports implement the Pluggable Transports 2.1 draft 1 specification available here:
<https://github.com/Pluggable-Transports/Pluggable-Transports-spec/blob/master/releases/PTSpecV2.1Draft1/Pluggable%20Transport%20Specification%20v2.1%20-%20Go%20Transport%20API%20v2.1%2C%20Draft%201.pdf> Specifically,
they implement the Go Transports API v2.1 draft 1.

The purpose of the transport library is to provide a set of different
transports. Each transport implements a different method of shapeshifting
network traffic. The goal is for application traffic to be sent over the network
in a shapeshifted form that bypasses network filtering, allowing
the application to work on networks where it would otherwise be blocked or
heavily throttled.

The following transports are currently provided by the library:

#### obfs4

This is a look-like nothing obfuscation protocol that incorporates ideas and concepts from Philipp Winter's ScrambleSuit protocol. The obfs naming was chosen primarily because it was shorter, in terms of protocol ancestery obfs4 is much closer to ScrambleSuit than obfs2/obfs3.

The notable differences between ScrambleSuit and obfs4:

* The handshake always does a full key exchange (no such thing as a Session Ticket Handshake).
* The handshake uses the Tor Project's ntor handshake with public keys obfuscated via the Elligator 2 mapping.
* The link layer encryption uses NaCl secret boxes (Poly1305/XSalsa20).
* As an added bonus, obfs4proxy also supports acting as an obfs2/3 client and bridge to ease the transition to the new protocol.

#### shadow (Shadowsocks)

Shadowsocks is a fast, free, and open-source encrypted proxy project, used to circumvent Internet censorship by utilizing a simple, but effective  encryption and a shared password

#### Optimizer

Optimizer is a pluggable transport that finds the optimal pluggable 
transport, pulled from multiple servers from the transports above and assigns you the 
transport that is most effective for each individual user.

#### Installation

The dispatcher is written in the Go programming language. To compile it you need
to install Go:

<https://golang.org/doc/install>

If you just installed Go for the first time, you will need to create a directory
to keep all of your Go source code:

    mkdir ~/go

If you already have Go installed, make sure it is a compatible version:

    go version

The version should be 1.10.2 or higher.

If you get the error "go: command not found", then trying exiting your terminal
and starting a new one.

If you have a compatible Go installed, you should go to the directory where you
keep all of your Go source code and set your GOPATH:

    cd ~/go
    export GOPATH=~/go

Software written in Go is installed using the `go get` command. You need to
install the transports that you want to use in your application. For instance,
for obfs4:

    go get github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4

This will fetch the source code for the obfs4 transport, and all the
dependencies, and compile everything.

If you want to install all provided transports:

    go get github.com/OperatorFoundation/shapeshifter-transports/transports

This will fetch the source code for all of the transports, and all the
dependencies, and compile everything.

#### Using the Library

The Shapeshifter Transports are just a Go library. In order to use the
transports, you will need to integrate them into your application, which is
presumably also written in Go. For details on the Go API for Pluggable
Transports, refer to the [Pluggable Transports 2.0 draft 1 specification](http://www.pluggabletransports.info/assets/PTSpecV2Draft1.pdf).

#### Frequently Asked Questions

##### What transport should I use in my application?

You should use a transport that is likely to work on the network where your
application is currently being blocked. This changes over time as the network
administrators alter their blocking rules. obfs4 is a popular transport, so
that might be a good starting point if you don't know how blocking is
implementing on your network. If you are an application developer using
Pluggable Transports, feel free to reach out to the Operator Foundation for
help in determining which transport might work best for your application. Email
brandon@operatorfoundation.org.

##### My application is not written in Go. Can I still use the transports?

Yes, the Go API is only one way to integrate transports into your application.
There is also an interprocess communication (IPC) protocol that allows you to
control a separate process (called the dispatcher) which provides access to the
transports through a proxy interface. When using this method, your application
can be written in any language. You just need to implement the IPC protocol so
that you can communicate with the dispatcher. The IPC protocol is specified in
the [Pluggable Transports 2.0 specification](https://github.com/Pluggable-Transports/Pluggable-Transports-spec/blob/master/releases/pt-2_0.pdf) section 3.3 and an implementation of the dispatcher is available which you can bundle with your
application here: <https://github.com/OperatorFoundation/shapeshifter-dispatcher>

### Credits

shapeshifter-transports is based on the transport implementations included with
the Tor project's "obfs4proxy" tool.

 * Yawning Angel for obfs4proxy and the obfs4 protocol
 * David Fifield for goptlib
 * Adam Langley for the Go Elligator implementation.
 * Philipp Winter for the ScrambleSuit protocol.
