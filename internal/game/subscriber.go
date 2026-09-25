package game

import "sync"

type Subscriber struct {
	Out          <-chan Event
	In           chan<- Event
	Done         <-chan struct{}
	queue        []Event
	close        chan struct{}
	closeOnce    sync.Once
	shutDownOnce sync.Once
}

func NewSubscriber() *Subscriber {
	out := make(chan Event, 128)
	in := make(chan Event, 128)
	done := make(chan struct{})

	s := &Subscriber{
		Out:   out,
		In:    in,
		Done:  done,
		queue: []Event{},
		close: make(chan struct{}),
	}
	go s.pump(out, in, done)

	return s
}

func (s *Subscriber) CloseImmediately() {
	s.closeOnce.Do(func() {
		close(s.close)
	})
}

func (s *Subscriber) CloseGracefully() {
	s.shutDownOnce.Do(func() {
		close(s.In)
	})
}

func (s *Subscriber) pump(out chan Event, in chan Event, done chan struct{}) {
	defer func() {
		close(out)
		close(done)
		s.CloseGracefully()
	}()

	inOrNil := in
	for {
		var outOrNil chan Event
		var nextEvent Event

		if inOrNil == nil && len(s.queue) == 0 {
			return
		}

		if len(s.queue) > 0 {
			outOrNil = out
			nextEvent = s.queue[0]
		}

		select {
		case event, ok := <-inOrNil:
			if !ok {
				inOrNil = nil
				break
			}
			s.queue = append(s.queue, event)
		case outOrNil <- nextEvent:
			s.queue[0] = nil
			s.queue = s.queue[1:]
			if len(s.queue) == 0 || cap(s.queue) > 256 {
				s.queue = nil
			}
		case <-s.close:
			return
		}
	}
}
