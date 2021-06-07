package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"kcache/cache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *cache.Group {
	return cache.NewGroup("scores", 10, 5*time.Second, cache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func startCacheServer(addr string, addrs []string, group *cache.Group) {
	// 初始化本机HTTP服务
	peers := cache.NewHTTPPool(addr)
	// 添加其他缓存节点信息
	peers.AddPeers(addrs...)
	// group绑定HTTP服务
	group.RegisterPeers(peers)
	log.Println("kCache is running at", addr)
	// 开启server
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, group *cache.Group) {
	http.Handle("/api", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")
		view, err := group.Get(key)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(view.ByteSlice())
	}))
	log.Println("kCache server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "kCache server port")
	flag.BoolVar(&api, "api", false, "start a api server?")
	flag.Parse()

	apiAddr := "http://127.0.0.1:9999"
	addrMap := map[int]string{
		8001: "http://0.0.0.0:8001",
		8002: "http://0.0.0.0:8002",
		8003: "http://0.0.0.0:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], addrs, group)
}
