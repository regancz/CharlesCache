package charlescache

import (
	"charlescache/cachepb"
	"charlescache/singleflight"
	"fmt"
	"log"
	"sync"
)

/**
 * @Author Charles
 * @Date 6:16 PM 10/9/2022
 **/

type Group struct {
	name      string
	getter    Getter
	mainCache Cache
	peers     PeerPicker
	// use singleflight.Group to make sure that  each key is only fetched once, support concurrent
	loader *singleflight.Group
}

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: Cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	mutex.Lock()
	groups[name] = g
	mutex.Unlock()
	return g
}

func GetGroup(name string) *Group {
	mutex.RLock()
	g := groups[name]
	mutex.RUnlock()
	return g
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.Get(key); ok {
		log.Println("[CharlesCache] hit")
		return v, nil
	}
	return g.Load(key)
}

func (g *Group) Load(key string) (value ByteView, err error) {
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.GetFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.GetLocally(key)
	})
	if err != nil {
		return
	}
	return view.(ByteView), nil
}

func (g *Group) GetLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.PopulateCache(key, value)
	return value, nil
}

func (g *Group) GetFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &cachepb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &cachepb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

func (g *Group) PopulateCache(key string, value ByteView) {
	g.mainCache.Add(key, value)
}

type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc is a functional implements
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}
