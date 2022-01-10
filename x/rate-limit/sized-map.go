package rate_limit

import "sync"

type SizedMap struct {
	mutex *sync.Mutex

	chunks     []*chunkSizedMap
	freshIndex int
	oldIndex   int

	creator func() interface{}
}

func NewSizedMap(chunkNum int, limitPerChunk int, creator func() interface{}) *SizedMap {
	chunks := make([]*chunkSizedMap, 0, chunkNum)
	for i := 0; i < chunkNum; i++ {
		chunks = append(chunks, newChunkSizedMap(limitPerChunk))
	}

	return &SizedMap{
		mutex:      new(sync.Mutex),
		chunks:     chunks,
		freshIndex: 0,
		oldIndex:   0,
		creator:    creator,
	}
}

func (m *SizedMap) Get(key string) interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, chunk := range m.chunks {
		if chunk.Has(key) {
			return chunk.Get(key)
		}
	}

	r := m.creator()

	chunk := m.chunks[m.freshIndex]
	if !chunk.IsMax() {
		chunk.Set(key, r)
		return r
	} else {
		m.freshIndex = (m.freshIndex + 1) % len(m.chunks)

		if m.freshIndex == m.oldIndex {
			m.chunks[m.oldIndex].Clear()
			m.oldIndex = (m.oldIndex + 1) % len(m.chunks)
		}

		m.chunks[m.freshIndex].Set(key, r)
		return r
	}
}

func (m *SizedMap) Size() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	r := 0
	for _, chunk := range m.chunks {
		r += chunk.Size()
	}
	return r
}

type chunkSizedMap struct {
	limit int
	m     map[string]interface{}
}

func newChunkSizedMap(limit int) *chunkSizedMap {
	return &chunkSizedMap{limit: limit, m: make(map[string]interface{})}
}

func (m *chunkSizedMap) Get(key string) interface{} {
	r, ok := m.m[key]
	if !ok {
		return nil
	}
	return r
}

func (m *chunkSizedMap) Set(key string, value interface{}) {
	m.m[key] = value
}

func (m *chunkSizedMap) Has(key string) bool {
	_, ok := m.m[key]
	return ok
}

func (m *chunkSizedMap) IsMax() bool {
	return len(m.m) >= m.limit
}

func (m *chunkSizedMap) Clear() {
	m.m = make(map[string]interface{})
}

func (m *chunkSizedMap) Size() int {
	return len(m.m)
}
