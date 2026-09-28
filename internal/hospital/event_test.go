package hospital

import (
	"testing"
	"time"
)

func registerOne(t *testing.T, h *Hospital, p *Patient) EpisodeRecord {
	t.Helper()
	e := NewEpisodeRecord(p, fakeAttender{"S-1", "Dr. Karen"}, "cafeteria", nil)
	if err := h.RegisterEpisode(e); err != nil {
		t.Fatalf("RegisterEpisode: %v", err)
	}
	return e
}

func receive(t *testing.T, ch <-chan EpisodeRecord) EpisodeRecord {
	t.Helper()
	select {
	case e, ok := <-ch:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		return e
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
		return EpisodeRecord{}
	}
}

func TestSubscribe_ReceivesEpisodes(t *testing.T) {
	h := newTestHospital(t)
	p := admit(t, h, "A", NarcolepsyLevelMild)
	events, cancel := h.Subscribe(4)
	defer cancel()

	sent := registerOne(t, h, p)

	if got := receive(t, events); got.ID() != sent.ID() {
		t.Errorf("got %v, want %v", got, sent)
	}
}

func TestSubscribe_FanOutToEverySubscriber(t *testing.T) {
	h := newTestHospital(t)
	p := admit(t, h, "A", NarcolepsyLevelMild)
	logger, cancelLogger := h.Subscribe(4)
	stream, cancelStream := h.Subscribe(4)
	defer cancelLogger()
	defer cancelStream()

	sent := registerOne(t, h, p)

	if receive(t, logger).ID() != sent.ID() || receive(t, stream).ID() != sent.ID() {
		t.Error("both subscribers must receive the same event")
	}
}

func TestSubscribe_SlowSubscriberNeverBlocksTheHospital(t *testing.T) {
	h := newTestHospital(t)
	p := admit(t, h, "A", NarcolepsyLevelMild)
	events, cancel := h.Subscribe(1)
	defer cancel()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 5; i++ { // nadie lee: el buffer se llena tras el primero
			registerOne(t, h, p)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RegisterEpisode blocked on a slow subscriber")
	}

	if len(h.History()) != 5 {
		t.Errorf("history is the source of truth: got %d, want 5", len(h.History()))
	}
	if len(events) != 1 {
		t.Errorf("slow subscriber should hold only its buffer, got %d", len(events))
	}
}

func TestSubscribe_CancelClosesChannelAndIsIdempotent(t *testing.T) {
	h := newTestHospital(t)
	p := admit(t, h, "A", NarcolepsyLevelMild)
	events, cancel := h.Subscribe(1)

	cancel()
	cancel() // no debe hacer panic

	if _, ok := <-events; ok {
		t.Error("channel must be closed after cancel")
	}
	registerOne(t, h, p) // publicar sin suscriptores tampoco debe fallar
}

func TestSubscribe_ZeroBufferIsRaisedToOne(t *testing.T) {
	h := newTestHospital(t)
	p := admit(t, h, "A", NarcolepsyLevelMild)
	events, cancel := h.Subscribe(0)
	defer cancel()

	registerOne(t, h, p)
	receive(t, events)
}
