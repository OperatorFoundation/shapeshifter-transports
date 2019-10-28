# shadow

Shadowsocks is a fast, free, and open-source encrypted proxy project, used to circumvent Internet censorship by utilizing a simple, but effective  encryption and a shared password

## Using shadow


### Go Version:

shadow is one of the transports available in the [Shapeshifter-Transports library](https://github.com/OperatorFoundation/Shapeshifter-Transports).

1. Create an instance of an shadow server
   `shadowTransport := shadow.Transport{"banana", "aes-192-ctr", "127.0.0.1:1234"`

2. Call Dial on shadowTransport:
   `_, err := shadowTransport.Dial()`
