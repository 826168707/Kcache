package cache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"

	"kcache/cache/cachepb"
)

const (
	defaultBasePath = "/_kcache/"
	defaultReplicas = 50
)

// HTTPPool 提供被其他节点访问和访问其他节点的能力
type HTTPPool struct {
	self        string // 记录节点本身的地址,包括主机名/IP 和 端口。  e.g."https://127.0.0.1:8000"
	basePath    string // 节点间通信地址的前缀，默认是/_kcache/
	mu          sync.Mutex
	peers       *ConsistentHash
	httpGetters map[string]*httpGetter // 每一个远程节点对应一个httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// 正确Path /<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&cachepb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// 重置远程节点信息
func (p *HTTPPool) ResetPeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = NewConsistentHash(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// 添加远程节点信息
func (p *HTTPPool) AddPeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.peers == nil {
		p.peers = NewConsistentHash(defaultReplicas, nil)
	}
	p.peers.Add(peers...)

	if p.httpGetters == nil {
		p.httpGetters = make(map[string]*httpGetter, len(peers))
	}
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// 删除远程节点信息
func (p *HTTPPool) RemovePeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.peers == nil || p.httpGetters == nil {
		return
	}

	p.peers.Delete(peers...)

	for _, peer := range peers {
		delete(p.httpGetters, peer)
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

type httpGetter struct {
	// baseURL表示将要访问对远程节点地址 e.g https://example.com/_kcache/
	baseURL string
}

//	当前节点当作客户端对其他节点发送请求
func (h *httpGetter) Get(in *cachepb.Request, out *cachepb.Response) error {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding reponse body: %v", err)
	}

	return nil
}

var _ PeerGetter = (*httpGetter)(nil)
