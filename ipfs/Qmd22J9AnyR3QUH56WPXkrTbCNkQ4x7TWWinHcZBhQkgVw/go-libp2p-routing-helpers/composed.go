package routinghelpers

import (
	"context"

	ci "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	routing "gx/ipfs/QmS4niovD1U6pRjUBXivr1zvvLBqiTKbERjFo994JU7oQS/go-libp2p-routing"
	ropts "gx/ipfs/QmS4niovD1U6pRjUBXivr1zvvLBqiTKbERjFo994JU7oQS/go-libp2p-routing/options"
	cid "gx/ipfs/QmZFbDTY9jfSBms2MchvYM9oYRbAF19K7Pby47yDBfpPrb/go-cid"
	pstore "gx/ipfs/QmeKD8YT7887Xu6Z86iZmpYNxrLogJexqxEugSmaf14k64/go-libp2p-peerstore"
	multierror "gx/ipfs/QmfGQp6VVqdPCDyzEM6EGwMY74YPabTSEoQWHUxZuCSWj3/go-multierror"
)

// Compose composes the components into a single router. Not specifying a
// component (leaving it nil) is equivalent to specifying the Null router.
//
// It also implements Bootstrap. All *distinct* components implementing
// Bootstrap will be bootstrapped in parallel. Identical components will not be
// bootstrapped twice.
type Compose struct {
	ValueStore     routing.ValueStore
	PeerRouting    routing.PeerRouting
	ContentRouting routing.ContentRouting
}

// note: we implement these methods explicitly to avoid having to manually
// specify the Null router everywhere we don't want to implement some
// functionality.

// PutValue adds value corresponding to given Key.
func (cr *Compose) PutValue(ctx context.Context, key string, value []byte, opts ...ropts.Option) error {
	if cr.ValueStore == nil {
		return routing.ErrNotSupported
	}
	return cr.ValueStore.PutValue(ctx, key, value, opts...)
}

// GetValue searches for the value corresponding to given Key.
func (cr *Compose) GetValue(ctx context.Context, key string, opts ...ropts.Option) ([]byte, error) {
	if cr.ValueStore == nil {
		return nil, routing.ErrNotFound
	}
	return cr.ValueStore.GetValue(ctx, key, opts...)
}

// Provide adds the given cid to the content routing system. If 'true' is
// passed, it also announces it, otherwise it is just kept in the local
// accounting of which objects are being provided.
func (cr *Compose) Provide(ctx context.Context, c *cid.Cid, local bool) error {
	if cr.ContentRouting == nil {
		return routing.ErrNotSupported
	}
	return cr.ContentRouting.Provide(ctx, c, local)
}

// FindProvidersAsync searches for peers who are able to provide a given key
func (cr *Compose) FindProvidersAsync(ctx context.Context, c *cid.Cid, count int) <-chan pstore.PeerInfo {
	if cr.ContentRouting == nil {
		ch := make(chan pstore.PeerInfo)
		close(ch)
		return ch
	}
	return cr.ContentRouting.FindProvidersAsync(ctx, c, count)
}

// FindPeer searches for a peer with given ID, returns a pstore.PeerInfo
// with relevant addresses.
func (cr *Compose) FindPeer(ctx context.Context, p peer.ID) (pstore.PeerInfo, error) {
	if cr.PeerRouting == nil {
		return pstore.PeerInfo{}, routing.ErrNotFound
	}
	return cr.PeerRouting.FindPeer(ctx, p)
}

// GetPublicKey returns the public key for the given peer.
func (cr *Compose) GetPublicKey(ctx context.Context, p peer.ID) (ci.PubKey, error) {
	if cr.ValueStore == nil {
		return nil, routing.ErrNotFound
	}
	return routing.GetPublicKey(cr.ValueStore, ctx, p)
}

// Bootstrap the router.
func (cr *Compose) Bootstrap(ctx context.Context) error {
	// Deduplicate. Technically, calling bootstrap multiple times shouldn't
	// be an issue but using the same router for multiple fields of Compose
	// is common.
	routers := make(map[Bootstrap]struct{}, 3)
	for _, value := range [...]interface{}{
		cr.ValueStore,
		cr.ContentRouting,
		cr.PeerRouting,
	} {
		switch b := value.(type) {
		case nil:
		case Null:
		case Bootstrap:
			routers[b] = struct{}{}
		}
	}

	var me multierror.Error
	for b := range routers {
		if err := b.Bootstrap(ctx); err != nil {
			me.Errors = append(me.Errors, err)
		}
	}
	return me.ErrorOrNil()
}

var _ routing.IpfsRouting = (*Compose)(nil)
var _ routing.PubKeyFetcher = (*Compose)(nil)
