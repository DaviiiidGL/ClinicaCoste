package hospital

import (
	"sync"
)

func (h *Hospital) Subscribe(bufSize int) (<-chan EpisodeRecord, func()) {
	if bufSize <= 0 {
		bufSize = 1
	}

	h.subMu.Lock()
	defer h.subMu.Unlock()

	id := h.nextSubID
	h.nextSubID++

	ch := make(chan EpisodeRecord, bufSize)
	h.subs[id] = ch

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.subMu.Lock()
			defer h.subMu.Unlock()
			if c, ok := h.subs[id]; ok {
				close(c)
				delete(h.subs, id)
			}
		})
	}

	return ch, cancel
}

func (h *Hospital) publish(e EpisodeRecord) {
	h.subMu.Lock()
	defer h.subMu.Unlock()

	for _, ch := range h.subs {
		select {
		case ch <- e:
		default:
		}
	}
}
