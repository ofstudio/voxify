/*
Package cancelgroup provides a synchronization primitive similar to sync.WaitGroup,
but with built-in support for cancellation via context.Context.

It is useful for managing a set of goroutines that need to be canceled and
waited for as a group.

# Example

	package main

	import (
		"fmt"
		"time"

		"github.com/your/module/cancelgroup"
	)

	func main() {
		g := cancelgroup.New()

		// Start a worker
		g.Go(func() {
			for {
				select {
				case <-g.Context().Done():
					fmt.Println("worker stopped")
					return
				default:
					fmt.Println("worker running...")
					time.Sleep(200 * time.Millisecond)
				}
			}
		})

		time.Sleep(1 * time.Second)

		// Cancel all workers
		g.Cancel()

		// Wait for them to stop
		g.Wait()
		fmt.Println("all workers stopped")
	}
*/
package cancelgroup
