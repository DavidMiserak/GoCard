---
tags: [go,concurrency,channels,programming]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 0
---

# Go Concurrency Patterns

## Question

What is the fan-out fan-in concurrency pattern in Go and when should you use it?

## Answer

The fan-out fan-in pattern is a concurrency pattern in Go where:

1. **Fan-out**: Start multiple goroutines to handle input from a single source (distributing work)
2. **Fan-in**: Combine multiple results from those goroutines into a single channel

### Implementation

```go
func fanOut(input <-chan int, n int) []<-chan int {
    // Create n output channels
    outputs := make([]<-chan int, n)

    for i := 0; i < n; i++ {
        outputs[i] = worker(input)
    }

    return outputs
}

func worker(input <-chan int) <-chan int {
    output := make(chan int)

    go func() {
        defer close(output)
        for n := range input {
            // Do some work with n
            result := process(n)
            output <- result
        }
    }()

    return output
}

func fanIn(inputs []<-chan int) <-chan int {
    output := make(chan int)
    var wg sync.WaitGroup

    // Start a goroutine for each input channel
    for _, ch := range inputs {
        wg.Add(1)
        go func(ch <-chan int) {
            defer wg.Done()
            for n := range ch {
                output <- n
            }
        }(ch)
    }

    // Close output once all input channels are drained
    go func() {
        wg.Wait()
        close(output)
    }()

    return output
}
```

### When to use it

- CPU-intensive operations that can be parallelized
- Operations that have independent work units
- When you need to process many items but control the level of concurrency
- Example use cases: image processing, data transformation pipelines, web scraping
