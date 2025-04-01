// File: internal/data/dummy_store.go

package data

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
)

// addDummyData adds sample decks and cards
func GetDummyDecks() []model.Deck {
	decks := []model.Deck{
		getGoDeck(),
		getComputerScienceDeck(),
		getDataStructuresDeck(),
		getAlgoDeck(),
		getBubbleTeaDeck(),
		getPythonDeck(),
		getLongAnswerDeck(),
	}

	return decks
}

// getGoDeck creates a sample deck for Go programming
func getGoDeck() model.Deck {
	goCards := []model.Card{
		{
			ID:           "go-1",
			Question:     "What is the purpose of the \"defer\" keyword in Go?",
			Answer:       "The \"defer\" keyword in Go schedules a function call to be executed just before the function returns. This is often used for cleanup actions, ensuring they will be executed even if the function panics.",
			DeckID:       "go-programming",
			LastReviewed: time.Now(),
			NextReview:   time.Now().Add(24 * time.Hour),
			Ease:         2.5,
			Interval:     1,
			Rating:       4,
		},
		{
			ID:           "go-2",
			Question:     "What are goroutines in Go?",
			Answer:       "Goroutines are lightweight threads managed by the Go runtime. They allow concurrent execution of functions without the overhead of traditional OS threads.",
			DeckID:       "go-programming",
			LastReviewed: time.Now().Add(-12 * time.Hour),
			NextReview:   time.Now().Add(36 * time.Hour),
			Ease:         2.3,
			Interval:     2,
			Rating:       3,
		},
		{
			ID:           "go-3",
			Question:     "How does a slice differ from an array in Go?",
			Answer:       "A slice is a reference to a contiguous segment of an array. Unlike arrays, slices are dynamic in size and don't carry their length as part of their type.",
			DeckID:       "go-programming",
			LastReviewed: time.Now().Add(-24 * time.Hour),
			NextReview:   time.Now().Add(48 * time.Hour),
			Ease:         2.7,
			Interval:     3,
			Rating:       5,
		},
	}

	goDeck := model.Deck{
		ID:          "go-programming",
		Name:        "Go Programming",
		Description: "Basic Go programming concepts",
		Cards:       goCards,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		LastStudied: time.Now(),
	}

	return goDeck
}

// getComputerScienceDeck creates a sample deck for Computer Science
func getComputerScienceDeck() model.Deck {
	csCards := []model.Card{
		{
			ID:           "cs-1",
			Question:     "What is a compiler?",
			Answer:       "A compiler is a program that translates source code written in a high-level programming language into machine code or another lower-level form.",
			DeckID:       "computer-science",
			LastReviewed: time.Now().Add(-24 * time.Hour),
			NextReview:   time.Now().Add(48 * time.Hour),
			Ease:         2.3,
			Interval:     2,
			Rating:       3,
		},
		{
			ID:           "cs-2",
			Question:     "What is the difference between process and thread?",
			Answer:       "A process is an instance of a program execution that has its own memory space. A thread is the smallest unit of execution within a process, and multiple threads share the memory space of the process.",
			DeckID:       "computer-science",
			LastReviewed: time.Now().Add(-36 * time.Hour),
			NextReview:   time.Now().Add(72 * time.Hour),
			Ease:         2.4,
			Interval:     3,
			Rating:       4,
		},
		{
			ID:           "cs-3",
			Question:     "What is cache memory?",
			Answer:       "Cache memory is a small, fast memory that stores frequently accessed data to reduce the time needed to access it from slower main memory.",
			DeckID:       "computer-science",
			LastReviewed: time.Now().Add(-48 * time.Hour),
			NextReview:   time.Now().Add(96 * time.Hour),
			Ease:         2.2,
			Interval:     4,
			Rating:       3,
		},
	}

	csDeck := model.Deck{
		ID:          "computer-science",
		Name:        "Computer Science",
		Description: "General computer science concepts",
		Cards:       csCards,
		CreatedAt:   time.Now().Add(-45 * 24 * time.Hour),
		LastStudied: time.Now().Add(-24 * time.Hour),
	}

	return csDeck
}

// getDataStructuresDeck creates a sample deck for Data Structures
func getDataStructuresDeck() model.Deck {
	dsCards := []model.Card{
		{
			ID:           "ds-1",
			Question:     "What is a stack data structure?",
			Answer:       "A stack is a linear data structure that follows the Last In First Out (LIFO) principle, where elements are added and removed from the same end, called the top.",
			DeckID:       "data-structures",
			LastReviewed: time.Now().Add(-72 * time.Hour),
			NextReview:   time.Now().Add(15 * 24 * time.Hour),
			Ease:         2.6,
			Interval:     15,
			Rating:       4,
		},
		{
			ID:           "ds-2",
			Question:     "What is a queue data structure?",
			Answer:       "A queue is a linear data structure that follows the First In First Out (FIFO) principle, where elements are added at the rear and removed from the front.",
			DeckID:       "data-structures",
			LastReviewed: time.Now().Add(-84 * time.Hour),
			NextReview:   time.Now().Add(20 * 24 * time.Hour),
			Ease:         2.5,
			Interval:     20,
			Rating:       4,
		},
		{
			ID:           "ds-3",
			Question:     "What is a binary search tree?",
			Answer:       "A binary search tree is a tree data structure where each node has at most two children, and for each node, all elements in the left subtree are less than the node, and all elements in the right subtree are greater.",
			DeckID:       "data-structures",
			LastReviewed: time.Now().Add(-96 * time.Hour),
			NextReview:   time.Now().Add(25 * 24 * time.Hour),
			Ease:         2.7,
			Interval:     25,
			Rating:       5,
		},
	}

	dsDeck := model.Deck{
		ID:          "data-structures",
		Name:        "Data Structures",
		Description: "Common data structures and operations",
		Cards:       dsCards,
		CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		LastStudied: time.Now().Add(-72 * time.Hour),
	}

	return dsDeck
}

// getAlgoDeck creates a sample deck for Algorithms
func getAlgoDeck() model.Deck {
	algoCards := []model.Card{
		{
			ID:           "algo-1",
			Question:     "What is the time complexity of quicksort in the average case?",
			Answer:       "The average time complexity of quicksort is O(n log n), where n is the number of elements to sort.",
			DeckID:       "algorithms",
			LastReviewed: time.Now().Add(-48 * time.Hour),
			NextReview:   time.Now().Add(18 * 24 * time.Hour),
			Ease:         2.4,
			Interval:     18,
			Rating:       3,
		},
		{
			ID:           "algo-2",
			Question:     "What is dynamic programming?",
			Answer:       "Dynamic programming is a method for solving complex problems by breaking them down into simpler subproblems and storing the results of these subproblems to avoid redundant calculations.",
			DeckID:       "algorithms",
			LastReviewed: time.Now().Add(-60 * time.Hour),
			NextReview:   time.Now().Add(22 * 24 * time.Hour),
			Ease:         2.3,
			Interval:     22,
			Rating:       3,
		},
		{
			ID:           "algo-3",
			Question:     "What is breadth-first search?",
			Answer:       "Breadth-first search is a graph traversal algorithm that explores all neighbors at the present depth before moving on to nodes at the next depth level.",
			DeckID:       "algorithms",
			LastReviewed: time.Now().Add(-72 * time.Hour),
			NextReview:   time.Now().Add(26 * 24 * time.Hour),
			Ease:         2.6,
			Interval:     26,
			Rating:       4,
		},
	}

	algoDeck := model.Deck{
		ID:          "algorithms",
		Name:        "Algorithms",
		Description: "Common algorithms and their analysis",
		Cards:       algoCards,
		CreatedAt:   time.Now().Add(-50 * 24 * time.Hour),
		LastStudied: time.Now().Add(-48 * time.Hour),
	}

	return algoDeck
}

// getBubbleTeaDeck creates a sample deck for Bubble Tea UI
func getBubbleTeaDeck() model.Deck {
	btCards := []model.Card{
		{
			ID:           "bt-1",
			Question:     "What is the Elm Architecture used by Bubble Tea?",
			Answer:       "The Elm Architecture is a design pattern consisting of three main components: Model (application state), View (renders the UI based on the state), and Update (handles events and updates the state).",
			DeckID:       "bubble-tea-ui",
			LastReviewed: time.Now().Add(-7 * 24 * time.Hour),
			NextReview:   time.Now().Add(10 * 24 * time.Hour),
			Ease:         2.1,
			Interval:     10,
			Rating:       4,
		},
		{
			ID:           "bt-2",
			Question:     "What is Lipgloss in the context of Bubble Tea?",
			Answer:       "Lipgloss is a styling library for terminal applications, often used with Bubble Tea to create visually appealing terminal UIs with colors, borders, and alignment.",
			DeckID:       "bubble-tea-ui",
			LastReviewed: time.Now().Add(-9 * 24 * time.Hour),
			NextReview:   time.Now().Add(12 * 24 * time.Hour),
			Ease:         2.2,
			Interval:     12,
			Rating:       4,
		},
	}

	btDeck := model.Deck{
		ID:          "bubble-tea-ui",
		Name:        "Bubble Tea UI",
		Description: "Bubble Tea TUI framework concepts",
		Cards:       btCards,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		LastStudied: time.Now().Add(-7 * 24 * time.Hour),
	}

	return btDeck
}

// getPythonDeck creates a sample deck for Python programming
func getPythonDeck() model.Deck {
	pythonCards := []model.Card{
		{
			ID:           "python-1",
			Question:     "What are list comprehensions in Python?",
			Answer:       "List comprehensions are a concise way to create lists in Python.\n\n```python\n# Example\nnumbers = [1, 2, 3, 4, 5]\nsquares = [x**2 for x in numbers]\n# Result: [1, 4, 9, 16, 25]\n```\n\nThey can also include conditions:\n\n```python\neven_squares = [x**2 for x in numbers if x % 2 == 0]\n# Result: [4, 16]\n```",
			DeckID:       "python-programming",
			LastReviewed: time.Now().Add(-24 * time.Hour),
			NextReview:   time.Now().Add(48 * time.Hour),
			Ease:         2.5,
			Interval:     2,
			Rating:       4,
		},
		{
			ID:           "python-2",
			Question:     "What are decorators in Python?",
			Answer:       "Decorators are a way to modify or enhance functions without changing their code directly.\n\n```python\n# Simple decorator example\ndef my_decorator(func):\n    def wrapper():\n        print(\"Something before the function is called.\")\n        func()\n        print(\"Something after the function is called.\")\n    return wrapper\n\n@my_decorator\ndef say_hello():\n    print(\"Hello!\")\n\n# When calling say_hello(), the output will be:\n# Something before the function is called.\n# Hello!\n# Something after the function is called.\n```",
			DeckID:       "python-programming",
			LastReviewed: time.Now().Add(-36 * time.Hour),
			NextReview:   time.Now().Add(72 * time.Hour),
			Ease:         2.3,
			Interval:     3,
			Rating:       3,
		},
		{
			ID:           "python-3",
			Question:     "How do context managers work in Python?",
			Answer:       "Context managers in Python handle setup and teardown operations using the `with` statement.\n\n```python\n# Example using file handling\nwith open('file.txt', 'r') as file:\n    data = file.read()\n# File is automatically closed after the block\n```\n\nYou can create custom context managers using either:\n\n1. A class with `__enter__` and `__exit__` methods\n2. The `@contextmanager` decorator\n\n```python\nfrom contextlib import contextmanager\n\n@contextmanager\ndef my_context():\n    print(\"Setup\")\n    try:\n        yield\n    finally:\n        print(\"Teardown\")\n```",
			DeckID:       "python-programming",
			LastReviewed: time.Now().Add(-48 * time.Hour),
			NextReview:   time.Now().Add(96 * time.Hour),
			Ease:         2.7,
			Interval:     4,
			Rating:       5,
		},
		{
			ID:           "python-4",
			Question:     "What are Python's magic methods?",
			Answer:       "Magic methods (dunder methods) are special methods with double underscores that allow classes to implement operator overloading and other language features.\n\nCommon examples:\n\n* `__init__`: Constructor\n* `__str__`: String representation for users\n* `__repr__`: String representation for developers\n* `__len__`: Length behavior\n* `__add__`: Addition behavior\n\n```python\nclass Vector:\n    def __init__(self, x, y):\n        self.x = x\n        self.y = y\n        \n    def __add__(self, other):\n        return Vector(self.x + other.x, self.y + other.y)\n        \n    def __str__(self):\n        return f\"Vector({self.x}, {self.y})\"\n```",
			DeckID:       "python-programming",
			LastReviewed: time.Now().Add(-60 * time.Hour),
			NextReview:   time.Now().Add(120 * time.Hour),
			Ease:         2.4,
			Interval:     5,
			Rating:       4,
		},
		{
			ID:           "python-5",
			Question:     "What are generators in Python and how do they differ from regular functions?",
			Answer:       "Generators are functions that return an iterator that yields values one at a time, calculated on-demand.\n\nKey differences:\n\n* Use `yield` instead of `return`\n* Maintain state between calls\n* Memory efficient for large sequences\n\n```python\ndef count_up_to(max):\n    count = 1\n    while count <= max:\n        yield count\n        count += 1\n\n# Usage\nfor number in count_up_to(5):\n    print(number)\n# Output: 1 2 3 4 5\n```\n\nGenerator expressions (similar to list comprehensions):\n\n```python\nsquares_gen = (x**2 for x in range(1000000))\n# Doesn't compute all values immediately\n```",
			DeckID:       "python-programming",
			LastReviewed: time.Now().Add(-72 * time.Hour),
			NextReview:   time.Now().Add(144 * time.Hour),
			Ease:         2.6,
			Interval:     6,
			Rating:       4,
		},
	}

	pythonDeck := model.Deck{
		ID:          "python-programming",
		Name:        "Python Programming",
		Description: "Core Python programming concepts and features",
		Cards:       pythonCards,
		CreatedAt:   time.Now().Add(-40 * 24 * time.Hour),
		LastStudied: time.Now().Add(-24 * time.Hour),
	}

	return pythonDeck
}

func getLongAnswerDeck() model.Deck {
	longAnswer := "```go\n" + `
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
` + "```" + `

### When to use it

- CPU-intensive operations that can be parallelized
- Operations that have independent work units
- When you need to process many items but control the level of concurrency
- Example use cases: image processing, data transformation pipelines, web scraping
`
	longAnswerCards := []model.Card{
		{
			ID:           "long-1",
			DeckID:       "long-answer",
			LastReviewed: time.Now().Add(-24 * time.Hour),
			NextReview:   time.Now().Add(48 * time.Hour),
			Ease:         2.5,
			Interval:     2,
			Rating:       4,
			Question:     "What is the fan-out fan-in concurrency pattern in Go and when should you use it?",
			Answer:       longAnswer,
		},
	}

	longDeck := model.Deck{
		ID:          "long-answer",
		Name:        "Long Answer",
		Description: "Long answer questions",
		Cards:       longAnswerCards,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		LastStudied: time.Now().Add(-24 * time.Hour),
	}

	return longDeck
}
