package hospital

func (h *Hospital) Subscribe(buffer int) (events <-chan EpisodeRecord, cancel func()) {
	if buffer < 1 {
		buffer = 1 // un canal sin buffer descartaría todo
	}
	ch := make(chan EpisodeRecord, buffer)

	h.subMu.Lock()
	id := h.nextSubID
	h.nextSubID++
	h.subs[id] = ch
	h.subMu.Unlock()

	cancel = func() {
		h.subMu.Lock()
		defer h.subMu.Unlock()
		if c, ok := h.subs[id]; ok {
			delete(h.subs, id)
			close(c)
		}
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
