package main

import (
    "fmt"
)

func main() {
    // Create a channel to pass data between goroutines
    chan1 := make(chan int)
    chan2 := make(chan int)

    // Start goroutine 1: generates numbers and sends them to chan1
    go func() {
        for i := 0; i < 10; i++ {
            chan1 <- i
        }
        close(chan1)
    }()

    // Start goroutine 2: receives numbers from chan1, squares them, and sends result to chan2
    go func() {
        for num := range chan1 {
            result := num * num
            chan2 <- result
        }
        close(chan2)
    }()

    // Print results from chan2
    for result := range chan2 {
        fmt.Println(result)
    }
}