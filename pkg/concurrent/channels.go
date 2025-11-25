package concurrent

import (
	"fmt"
)

func TestChannels() {
	input := make(chan string)
	result := make(chan string)

	// producer
	go func() {
		defer close(input)
		for _, s := range []string{"hi", "yo", "hey"} {
			input <- s
		}
	}()

	// transformer
	go func() {
		defer close(result)
		for msg := range input {
			result <- msg + "LOL"
		}
	}()

	// consumer
	for out := range result {
		fmt.Println(out)
	}
}

// func TestChannels() {

// 	input := make(chan string)
// 	result := make(chan string)
// 	var wg sync.WaitGroup
// 	wg.Add(2)

// 	go func() {
// 		defer close(input)
// 		defer wg.Done()
// 		input <- "hallo"
// 	}()

// 	go func() {
// 		defer wg.Done()
// 		defer close(result)
// 		messageFromChannel := <-input
// 		result <- messageFromChannel + "LOL"
// 	}()

// 	fmt.Println(<-result)
// 	wg.Wait()
// }
