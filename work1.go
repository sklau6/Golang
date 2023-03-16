package main

import (
    "fmt"
    "sync"
)

func main() {
    // Create a WaitGroup to wait for goroutines to finish
    wg := sync.WaitGroup{}

    // Create a channel to receive results
    results := make(chan int)

    // Launch several goroutines to perform some computation
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()

            // Perform some computation
            result := 0
            for j := 0; j < n; j++ {
                result += j
            }

            // Send the result back on the channel
            results <- result
        }(i)
    }

    // Wait for all goroutines to finish
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect the results and print them
    total := 0
    for result := range results {
        total += result
    }
    fmt.Println(total)
}
