# The Operator Foundation

[Operator](https://operatorfoundation.org) makes usable tools to help people around the world with censorship, security, and privacy.

## Shapeshifter

The Shapeshifter project provides network protocol shapeshifting technology
(also sometimes referred to as obfuscation). The purpose of this technology is
to change the characteristics of network traffic so that it is not identified
and subsequently blocked by network filtering devices.

There are two components to Shapeshifter: transports and the dispatcher. 

If you are an end user that is trying to circumvent filtering on your network, or a developer that wants to add pluggable transports to an existing tool that is not written in the Go programming language, then you probably want [shapeshifter-dispatcher](https://github.com/OperatorFoundation/shapeshifter-dispatcher). Please note that familiarity with executing programs on the command line is necessary to use this tool.

If you are looking for a complete, easy-to-use VPN that incorporates shapeshifting technology and has a graphical user interface, consider
[Moonbounce](https://github.com/OperatorFoundation/Moonbounce), an application for macOS which incorporates shapeshifting without the need to write code or use the command line.

### Shapeshifter Transports

The purpose of the transport suite is to provide a variety of transports to choose from. Each transport implements a different method of shapeshifting network traffic. The goal is for application traffic to be sent over the network in a shapeshifted form that bypasses network filtering, allowing the application to work on networks where it would otherwise be blocked or heavily throttled.

Each transport provides a different approach to shapeshifting. These transports are provided as a Go library which can be integrated directly into applications. The dispatcher is a command line tool which provides a proxy that wraps the transport library. It has several different proxy modes and can proxy both TCP and UDP traffic.

These transports implement the [Pluggable Transports 3.0](https://github.com/Pluggable-Transports/Pluggable-Transports-spec/tree/main/releases/PTSpecV3.0) specification. Specifically, they implement the Go Transports API v3.

If you are a tool developer working in the Go programming language, then you
probably want to use one or more transport libraries directly in your application.


The following transports are currently implemented in Go:

#### Replicant

[Replicant](https://github.com/OperatorFoundation/Replicant-go) is Operator's flagship transport which can be tuned for each adversary. It is designed to be more effective and efficient that older transports.
It can be quickly reconfigured as filtering conditions change by updating just the configuration file.

A [Swift implementation](https://github.com/OperatorFoundation/ReplicantSwift.git) is also available.

#### Starbridge

[Starbridge](https://github.com/OperatorFoundation/Starbridge-go.git) is a Pluggable Transport that requires only minimal configuration information from the user. Under the hood, it uses the Replicant Pluggable Transport technology for network protocol obfuscation. [Replicant](https://github.com/OperatorFoundation/Replicant-go) is more complex to configure, so Starbridge is a good starting point for those wanting to use the technology to circumvent Internet cenorship, but wanting a minimal amount of setup.

A [Swift implementation](https://github.com/OperatorFoundation/Starbridge.git) is also available.

#### Shadow (Shadowsocks)

Shadowsocks is a simple, but effective and popular network traffic obfuscation tool that uses basic encryption with a shared password.
[Shadow](https://github.com/OperatorFoundation/Shadow-go) is a wrapper for Shadowsocks that makes it available as a Pluggable Transport.

A [Swift implementation](https://github.com/OperatorFoundation/ShadowSwift.git) is also available.

#### Optimizer

[Optimizer](https://github.com/OperatorFoundation/Optimizer-go) is a pluggable transport that works with your other transports to find the best option. It has multiple configurable strategies to find
the optimal choice among the available transports. It can be used for numerous optimization tasks, such as round
robin load spreading among multiple transport servers or minimizing latency given multiple transport configurations.


#### Installation

For individual installation instructions, see the README's for the individual transports:

- [Replicant README](https://github.com/OperatorFoundation/Replicant-go/blob/main/README.md)

- [Starbridge README](https://github.com/OperatorFoundation/Starbridge-go/blob/main/README.md)

- [Shadow README](https://github.com/OperatorFoundation/Shadow-go/blob/main/README.md)

- [Optimizer README](https://github.com/OperatorFoundation/Optimizer-go/blob/main/README.md)


#### Frequently Asked Questions

##### What transport should I use in my application?

Try Replicant, Operator's flagship transport which can be tuned for each adversary. Email contact@operatorfoundation.org for a sample config file for the adversary of interest.
shadow is also a good choice as it works on many networks and is easy to configure.

If you are an application developer using Pluggable Transports, feel free to reach out to the Operator Foundation for
help in determining which transport might work best for your application. Email contact@operatorfoundation.org.

##### My application is not written in Go. Can I still use the transports?

Yes, the Go API is only one way to integrate transports into your application.
There is also an interprocess communication (IPC) protocol that allows you to
control a separate process (called the dispatcher) which provides access to the
transports through a proxy interface. When using this method, your application
can be written in any language. You just need to implement the IPC protocol so
that you can communicate with the dispatcher. The IPC protocol is specified in
the [Pluggable Transports 2.1 specification](https://www.pluggabletransports.info/spec/#build) section 3.3 and an implementation of the [dispatcher](https://github.com/OperatorFoundation/shapeshifter-dispatcher) is available which you can bundle with your
application.

In addition, we have native Swift implementations available for those developers looking to integrate transports directly into their iOS, macOS, or Linux applications:
- [Replicant](https://github.com/OperatorFoundation/ReplicantSwift.git)
- [Starbridge](https://github.com/OperatorFoundation/Starbridge.git)
- [Shadow](https://github.com/OperatorFoundation/ShadowSwift.git)

### Credits

 * Replicant, Starbridge, Shadow, and Optimizer were developed by [Operator Foundation](https://operatorfoundation.org)
 * [Shadowsocks](https://shadowsocks.org/guide/what-is-shadowsocks.html) was developed by the Shadowsocks team.
