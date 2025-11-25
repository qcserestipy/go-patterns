package concurrent

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// WorkItem represents an input item to process.
type WorkItem struct {
	ID int
}

// Result represents a processed output.
type Result struct {
	ID      int
	Outcome string
	Err     error
}

const maxInFlight = 200

var inflight = make(chan struct{}, maxInFlight)

func FanOut() {
	// Create a cancelable context so we can stop long-running pipelines gracefully.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Example input batch (finite). For streaming, see generator below.
	items := make([]WorkItem, 0, 20)
	for i := range 20 {
		items = append(items, WorkItem{ID: i})
	}

	// Build the pipeline: produce -> fan-out workers -> fan-in results -> consume
	input := make(chan WorkItem)      // unbuffered gives strong backpressure (safe default)
	results := make(chan Result, 128) // small buffer helps absorb worker bursts

	const workers = 4

	var wg sync.WaitGroup
	wg.Add(workers)

	// Start workers (fan-out): each worker reads from the same input channel.
	for w := range workers {
		go func(workerID int) {
			defer wg.Done()
			worker(ctx, workerID, input, results)
		}(w)
	}

	// Close results after *all* workers exit (fan-in closer).
	go func() {
		wg.Wait()
		close(results)
	}()

	// Start producer in its own goroutine to feed input and then close it.
	go func() {
		defer close(input) // sender closes the channel when done
		// produce(ctx, items, input)
		generateBatches(ctx, input)
		// For streaming, use: generate(ctx, input) to send until canceled
	}()

	// Consume results until results channel is closed by the closer goroutine.
	consume(results)
}

// produce sends a finite batch to the input channel.
// Closing is handled by the caller (the goroutine that invoked this function).
func produce(ctx context.Context, items []WorkItem, input chan<- WorkItem) {
	for _, it := range items {
		select {
		case <-ctx.Done():
			return // Stop early on cancel
		case input <- it: // Block until a worker is ready (unbuffered) or buffer available (buffered)
		}
	}
}

// generate demonstrates an infinite/long-running producer with cancellation.
// Call this instead of produce() if you want a stream.
// NOTE: the caller must close(input) after this returns (on cancel).
func generate(ctx context.Context, input chan<- WorkItem) {
	id := 0
	ticker := time.NewTicker(20 * time.Millisecond) // rate limit the producer
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case <-ctx.Done():
				return
			case input <- WorkItem{ID: id}:
				id++
			}
		}
	}
}

func generateBatches(ctx context.Context, input chan<- WorkItem) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	const batchSize = 100
	batchID := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Println("generator stopped")
			return

		case <-ticker.C:
			fmt.Printf("starting batch %d\n", batchID)

			for i := range batchSize {
				// 1) acquire a slot (blocks if cap reached)
				select {
				case <-ctx.Done():
					return
				case inflight <- struct{}{}:
				}

				// 2) enqueue work (may block if input buffer is full)
				select {
				case <-ctx.Done():
					// give slot back if we’re shutting down while blocked
					<-inflight
					return
				case input <- WorkItem{ID: batchID*batchSize + i}:
				}
			}

			fmt.Printf("batch %d done\n", batchID)
			batchID++
		}
	}
}

// worker simulates work by sleeping for a random duration.
// It reads from input until the input channel is closed or context is canceled.
func worker(ctx context.Context, workerID int, input <-chan WorkItem, results chan<- Result) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

	for {
		select {
		case <-ctx.Done():
			return
		case it, ok := <-input:
			if !ok {
				// Input closed: no more work; exit gracefully.
				return
			}
			// Simulate variable work cost.
			time.Sleep(time.Duration(10+rng.Intn(50)) * time.Millisecond)

			// Send result (non-blocking semantics still apply: may block if results buffer is full).
			select {
			case <-ctx.Done():
				return
			case results <- Result{
				ID:      it.ID,
				Outcome: fmt.Sprintf("worker %d processed %d", workerID, it.ID),
				Err:     nil,
			}:
			}

			//Release slot
			<-inflight
		}
	}
}

// consume drains the results channel until it is closed.
func consume(results <-chan Result) {
	count := 0
	for r := range results {
		if r.Err != nil {
			fmt.Printf("ERR: id=%d err=%v\n", r.ID, r.Err)
			continue
		}
		fmt.Printf("OK:  id=%02d msg=%s\n", r.ID, r.Outcome)
		count++
	}
	fmt.Printf("done. processed=%d\n", count)
}

/*
CHEATSHEET

- Fan-out: start N workers reading from the same input channel.
- Fan-in: have all workers send to a single results channel, then close results
  *after* all workers finish (using a WaitGroup + a closer goroutine).

- Backpressure:
  Unbuffered input  -> producer blocks until a worker is ready (fair & safe).
  Buffered input    -> producer can get ahead up to cap; tune for throughput.

- Closing rules:
  * Only the SENDER closes a channel.
  * Close(input) when no more items will be sent.
  * Close(results) after all workers are done sending.

- Cancellation:
  Use context to stop infinite/long tasks. Always select on ctx.Done() in producers and workers.

- Ordering:
  If output ordering must match input, include an index and reorder at the consumer,
  or use a per-item result channel / semaphore pattern.
*/
