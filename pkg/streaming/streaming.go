package streaming

import (
	"sync"
)

func Apply[T any](done <-chan struct{}, ch <-chan T, fn func(T) T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for v := range ch {
			select {
			case out <- fn(v):
			case <-done:
				return
			}
		}
	}()

	return out
}

func Morph[I any, O any](done <-chan struct{}, ch <-chan I, fn func(I) (O, error)) <-chan O {
	out := make(chan O)

	go func() {
		defer close(out)

		for v := range ch {

			o, err := fn(v)
			if err != nil {
				continue
			}

			select {
			case out <- o:
			case <-done:
				return
			}
		}
	}()

	return out
}

func Merge[T any](in ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T)

	fn := func(ch <-chan T) {
		for v := range ch {
			out <- v
		}
		wg.Done()
	}

	// Start a Goroutine for each input channel to forward values to the output channel.
	for _, ch := range in {
		if ch == nil {
			continue
		}

		wg.Add(1)
		go fn(ch)
	}

	// Start a Goroutine to close the output channel when all input channels are closed.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func MergeDone[T any](done <-chan struct{}, ch ...<-chan T) <-chan T {
	out := make(chan T)

	var wg sync.WaitGroup

	merge := func(ch <-chan T) {
		defer wg.Done()

		for v := range ch {
			select {
			case out <- v:
			case <-done:
				return
			}
		}
	}

	wg.Add(len(ch))
	for _, c := range ch {
		go merge(c)
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out
}

func Multiplex[T any](done <-chan struct{}, in <-chan T, ch ...chan<- T) {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for v := range in {
			for _, c := range ch {
				select {
				case c <- v:
				case <-done:
					return
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		for _, c := range ch {
			close(c)
		}
	}()

	return
}

func Generate[T any](done <-chan struct{}, v ...T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, n := range v {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}()
	return out
}

// Done combine done channels
func Done(in ...<-chan struct{}) <-chan struct{} {

	switch len(in) {
	case 0:
		panic("need at least one channel")
	case 1:
		if in[0] == nil {
			panic("only one nil channel")
		}
		return in[0]
	}

	orDone := make(chan struct{})
	go func() {
		defer close(orDone)

		switch len(in) {
		case 2:
			select {
			case <-in[0]:
			case <-in[1]:
			}
		default:
			select {
			case <-in[0]:
			case <-in[1]:
			case <-in[2]:
			case <-Done(append(in[3:], orDone)...):
			}
		}
	}()

	return orDone
}

func AllDone(in ...<-chan struct{}) <-chan struct{} {
	var wg sync.WaitGroup

	// Create a channel to signal when all input channels are closed
	done := make(chan struct{})

	// Function to wait for a single channel to be closed
	fn := func(ch <-chan struct{}) {
		defer wg.Done()
		for range ch {
			// Do nothing, just consume values until the channel is closed
		}
	}

	// Start a goroutine for each input channel
	for _, ch := range in {
		wg.Add(1)
		go fn(ch)
	}

	// Start a goroutine to close the 'done' channel when all input channels are closed
	go func() {
		wg.Wait()
		close(done)
	}()

	return done
}
