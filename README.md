# Kcache
一种分布式LRU缓存。
特性：
+ 缓存数据支持生命周期
+ 支持并发访问
+ 节点间通过Protobuf通信
+ 支持SingleFlight减轻节点压力
+ 支持分布式部署和单机多端口部署
## 使用
### 创建缓存group
NewGroup参数依次为：名称，LRU容量，数据生命周期，DB获取数据函数
```go
  func createGroup() *cache.Group {
    return cache.NewGroup("scores", 10, 5*time.Second, cache.GetterFunc(func(key string) ([]byte, error) {
      log.Println("[SlowDB] search key", key)
      if v, ok := db[key]; ok {
        return []byte(v), nil
      }
      return nil, fmt.Errorf("%s not exist", key)
    }))
  }
```

### 启动缓存服务器
```go
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
```

### 启动客户端
部署时，某一节点可选做客户端
```go
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
```

## 部署与测试
本机部署为例（缓存节点部署在8001,8002，8003端口，客户端在9999端口，详情见[main.go](https://github.com/826168707/Kcache/blob/main/kcache/main.go)）
```go
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
```
### 测试
```
curl "http://localhost:9999/api?key=???" 
```
