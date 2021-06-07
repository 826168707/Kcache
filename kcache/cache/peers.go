package cache

import "kcache/cache/cachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(in *cachepb.Request, out *cachepb.Response) error
}
