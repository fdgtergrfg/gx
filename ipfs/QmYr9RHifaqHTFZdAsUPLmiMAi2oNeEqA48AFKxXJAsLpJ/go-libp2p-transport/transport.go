package transport

import (
	"context"
	"net"
	"time"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	logging "gx/ipfs/QmRREK2CAZ5Re2Bd9zZFG6FeYDppUWt5cMgsoUEp3ktgSr/go-log"
	inet "gx/ipfs/QmX5J1q63BrrDTbpcHifrFbxH3cMZsvaNajy6u3zCpzBXs/go-libp2p-net"
	smux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
)

// DialTimeout is the maximum duration a Dial is allowed to take.
// This includes the time between dialing the raw network connection,
// protocol selection as well the handshake, if applicable.
var DialTimeout = 60 * time.Second

// AcceptTimeout is the maximum duration an Accept is allowed to take.
// This includes the time between accepting the raw network connection,
// protocol selection as well as the handshake, if applicable.
var AcceptTimeout = 60 * time.Second

var log = logging.Logger("transport")

// Conn is an extension of the net.Conn interface that provides multiaddr
// information, and an accessor for the transport used to create the conn
type Conn interface {
	smux.Conn
	inet.ConnSecurity
	inet.ConnMultiaddrs

	// Transport returns the transport to which this connection belongs.
	Transport() Transport
}

// Transport represents any device by which you can connect to and accept
// connections from other peers. The built-in transports provided are TCP and UTP
// but many more can be implemented, sctp, audio signals, sneakernet, UDT, a
// network of drones carrying usb flash drives, and so on.
type Transport interface {
	// Dial dials a remote peer. It should try to reuse local listener
	// addresses if possible but it may choose not to.
	Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (Conn, error)

	// CanDial returns true if this transport knows how to dial the given
	// multiaddr.
	//
	// Returning true does not guarantee that dialing this multiaddr will
	// succeed. This function should *only* be used to preemptively filter
	// out addresses that we can't dial.
	CanDial(addr ma.Multiaddr) bool

	// Listen listens on the passed multiaddr.
	Listen(laddr ma.Multiaddr) (Listener, error)

	// Protocol returns the set of protocols handled by this transport.
	//
	// See the Network interface for an explanation of how this is used.
	Protocols() []int

	// Proxy returns true if this is a proxy transport.
	//
	// See the Network interface for an explanation of how this is used.
	// TODO: Make this a part of the go-multiaddr protocol instead?
	Proxy() bool
}

// Listener is an interface closely resembling the net.Listener interface.  The
// only real difference is that Accept() returns Conn's of the type in this
// package, and also exposes a Multiaddr method as opposed to a regular Addr
// method
type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
	Multiaddr() ma.Multiaddr
}

// Network is an inet.Network with methods for managing transports.
type Network interface {
	inet.Network

	// AddTransport adds a transport to this Network.
	//
	// When dialing, this Network will iterate over the protocols in the
	// remote multiaddr and pick the first protocol registered with a proxy
	// transport, if any. Otherwise, it'll pick the transport registered to
	// handle the last protocol in the multiaddr.
	//
	// When listening, this Network will iterate over the protocols in the
	// local multiaddr and pick the *last* protocol registered with a proxy
	// transport, if any. Otherwise, it'll pick the transport registered to
	// handle the last protocol in the multiaddr.
	AddTransport(t Transport) error
}
