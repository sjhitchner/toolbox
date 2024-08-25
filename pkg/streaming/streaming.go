package streaming

import (
	"log"
	"sync"
	"time"
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
	return MergeBuffer[T](1, in...)
}

func MergeBuffer[T any](buffer int, in ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T, buffer)

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

func MergeError[T any](inCh []<-chan T, errCh []<-chan error) (<-chan T, <-chan error) {

	var wg sync.WaitGroup

	outCh := make(chan T)
	outErrCh := make(chan error)

	fnOut := func(ch <-chan T) {
		for v := range ch {
			outCh <- v
		}
		wg.Done()
	}

	fnErr := func(ch <-chan error) {
		for v := range ch {
			outErrCh <- v
		}
		wg.Done()
	}

	// Start a Goroutine for each input channel to forward values to the output channel.
	for _, ch := range inCh {
		if ch == nil {
			continue
		}

		wg.Add(1)
		go fnOut(ch)
	}

	for _, err := range errCh {
		if err == nil {
			continue
		}

		wg.Add(1)
		go fnErr(err)
	}

	// Start a Goroutine to close the output channel when all input channels are closed.
	go func() {
		wg.Wait()
		close(outCh)
		close(outErrCh)
	}()

	return outCh, outErrCh
}

func Demultiplex[T any](done <-chan struct{}, in <-chan T, ch ...chan<- T) {

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

func FanOut[T any](in <-chan T, num int) []chan T {

	streams := make([]chan T, num)
	for i := range streams {
		streams[i] = make(chan T)
	}

	go func() {
		defer func() {
			for i := range streams {
				close(streams[i])
			}
		}()

		for val := range in {
			for _, ch := range streams {
				ch <- val
			}
		}
	}()

	return streams
}

func Consume[T any](in <-chan T) {
	for _ = range in {
	}
}

func Gather[T any](in <-chan T) []T {
	arr := make([]T, 0, 10)
	for val := range in {
		arr = append(arr, val)
	}
	return arr
}

func Error(in <-chan error) error {
	for err := range in {
		if err != nil {
			return err
		}
	}
	return nil
}

func ErrorLog(in <-chan error) {
	go func() {
		for err := range in {
			if err != nil {
				log.Println(err)
			}
		}
	}()
}

func Copy[T any](out chan<- T, in <-chan T) {
	go func() {
		for v := range in {
			out <- v
		}
	}()
}

func CopyWG[T any](wg *sync.WaitGroup, out chan<- T, in <-chan T) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range in {
			out <- v
		}
	}()
}

func Loop[T any](done <-chan struct{}, data []T) (<-chan T, error) {
	out := make(chan T)
	go func() {
		defer close(out)

		var index int
		for {
			select {
			case <-done:
				return
			default:
				out <- data[index]

				index = (index + 1) % len(data)
			}
		}

	}()

	return out, nil
}

func Batch[T any](doneCh <-chan struct{}, inCh <-chan T, batchSize int, timeout time.Duration) <-chan []T {
	outCh := make(chan []T)

	go func() {
		defer close(outCh)
		batch := make([]T, 0, batchSize)
		timer := time.NewTimer(timeout)

		for {
			select {
			case <-doneCh:
				if len(batch) > 0 {
					outCh <- batch
				}
				return

			case item, ok := <-inCh:
				if !ok {
					if len(batch) > 0 {
						outCh <- batch
					}
					return
				}

				batch = append(batch, item)
				if len(batch) == batchSize {
					outCh <- batch
					batch = batch[:0] // make([]T, 0, batchSize)
					timer.Reset(timeout)
				}
			case <-timer.C:
				// Timeout, send the current batch even if it's not full
				if len(batch) > 0 {
					outCh <- batch
					batch = batch[:0] // make([]T, 0, batchSize)
				}
				timer.Reset(timeout)
			}
		}
	}()
	return outCh
}
