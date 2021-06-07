package cache

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// ConsistentHash 通过一致性哈希结构寻找节点
type ConsistentHash struct {
	hash     Hash
	replicas int            // 虚拟节点副本数
	keys     []int          // 切片模拟一致性哈希环
	hashMap  map[int]string // 虚拟节点-真实节点 映射
}

func NewConsistentHash(replicas int, fn Hash) *ConsistentHash {
	m := &ConsistentHash{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *ConsistentHash) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *ConsistentHash) Delete(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))

			if _, ok := m.hashMap[hash]; !ok {
				break
			}

			// todo 感觉这里的遍历逻辑以后可以优化
			var idx int
			for i, v := range m.keys {
				if v == hash {
					idx = i
					break
				}
			}
			m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
			delete(m.hashMap, hash)
		}
	}
}

func (m *ConsistentHash) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
