package cache

type Manager struct {
	// 记录节点本身的地址,包括主机名/IP 和 端口。  e.g."https://127.0.0.1:8000"
	self string
	// 节点间通信地址的前缀，默认是/_kcache/
	basePath string
	// 所有节点地址
	addrs []string
}

func NewManager(self string) *Manager {
	return &Manager{
		self:     self,
		basePath: defaultBasePath,
		addrs:    make([]string, 0),
	}
}
