package charlescache

import "charlescache/cachepb"

/**
 * @Author Charles
 * @Date 9:46 PM 10/9/2022
 **/

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(in *cachepb.Request, out *cachepb.Response) error
}
