package notes

import "sync"

// Broadcaster fans out published messages (rendered HTML fragments) to
// every subscribed SSE client — the pub/sub layer behind the example
// feature's real-time updates (Story #4, AD-4's real-time layer). It's
// deliberately in-process only: a single go-htmx instance is the
// template's default deployment shape, so there's no cross-instance
// fan-out to solve here (a team that outgrows one instance would swap
// this for a Redis/NATS-backed broadcaster behind the same Subscribe/
// Publish shape).
type Broadcaster struct {
	mu   sync.Mutex
	subs map[chan string]struct{}
}

// NewBroadcaster returns a ready-to-use Broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subs: make(map[chan string]struct{})}
}

// Subscribe registers a new subscriber and returns its message channel
// plus an unsubscribe func the caller must call exactly once (typically
// deferred) to release it.
func (b *Broadcaster) Subscribe() (ch chan string, unsubscribe func()) {
	ch = make(chan string, 8)

	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe = func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subs, ch)
			close(ch)
			b.mu.Unlock()
		})
	}
	return ch, unsubscribe
}

// Publish fans msg out to every current subscriber. A subscriber whose
// buffered channel is full is skipped rather than blocking Publish — a
// slow SSE client falls behind rather than stalling every other write.
func (b *Broadcaster) Publish(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- msg:
		default:
		}
	}
}
