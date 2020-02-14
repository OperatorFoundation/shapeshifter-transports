# shadow

Shadowsocks is a fast, free, and open-source encrypted proxy project, used to circumvent Internet censorship by utilizing a simple, but effective encryption and a shared password.

## Using shadow

### Go Version:

shadow is one of the transports available in the [Shapeshifter Transports library](https://github.com/OperatorFoundation/shapeshifter-transports).

#### Create a shadow server

config := NewConfig("password", "aes-128-ctr")

serverListener, listenError := config.Listen("127.0.0.1:1234")
if listenError != nil {
    return
}

serverConn, acceptErr := serverListener.Accept()
if acceptErr != nil {
    return
}

#### Create a shadow client

config := NewConfig("password", "aes-128-ctr")

clientConn, dialErr := config.Dial("127.0.0.1:1234")
if dialErr != nil {
    t.Fail()
    return
}
