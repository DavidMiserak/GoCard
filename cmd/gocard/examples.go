// cmd/gocard/examples.go - Example content generation
package main

import (
	"fmt"
	"time"

	"github.com/DavidMiserak/GoCard/internal/storage"
)

// createExampleContent creates comprehensive example content
//
//nolint:unused // Will be used in the future
func createExampleContent(store storage.CardStoreInterface) error {
	// Create main category decks
	programmingDeck, err := store.CreateDeck("Programming", nil)
	if err != nil {
		return fmt.Errorf("failed to create programming deck: %w", err)
	}

	conceptsDeck, err := store.CreateDeck("Concepts", nil)
	if err != nil {
		return fmt.Errorf("failed to create concepts deck: %w", err)
	}

	// Create sub-decks for programming
	goDeck, err := store.CreateDeck("Go", programmingDeck)
	if err != nil {
		return fmt.Errorf("failed to create go deck: %w", err)
	}

	pythonDeck, err := store.CreateDeck("Python", programmingDeck)
	if err != nil {
		return fmt.Errorf("failed to create python deck: %w", err)
	}

	// Create example programming cards (Go)

	// Define constants for code blocks to handle nested backticks
	const goStart = "```go"
	const pythonStart = "```python"
	const jsStart = "```javascript"
	const javaStart = "```java"
	const codeEnd = "```"

	// Go Channels card
	concurrencyContent := `# Go Concurrency Patterns

## 1. Goroutines
Lightweight threads managed by the Go runtime:

` + goStart + `
go func() {
    // code to run concurrently
}()
` + codeEnd + `

## 2. Channels
Type-safe pipes for communication between goroutines:

` + goStart + `
// Unbuffered channel
ch := make(chan int)

// Buffered channel with capacity 10
buffered := make(chan string, 10)

// Send to channel (blocks if unbuffered or buffer full)
ch <- 42

// Receive from channel (blocks if empty)
value := <-ch

// Close a channel (no more sends allowed)
close(ch)

// Range over channel until closed
for v := range ch {
    fmt.Println(v)
}
` + codeEnd + `

## 3. Select Statement
Waits on multiple channel operations:

` + goStart + `
select {
case v1 := <-ch1:
    // use v1
case ch2 <- v2:
    // sent v2
case <-time.After(1 * time.Second):
    // timeout after 1 second
default:
    // non-blocking case
}
` + codeEnd + `

## 4. Wait Groups
Synchronization for multiple goroutines:

` + goStart + `
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // do work
    }(i)
}

// Wait for all goroutines to finish
wg.Wait()
` + codeEnd + `

## 5. Mutex
Locks for mutual exclusion:

` + goStart + `
var mu sync.Mutex
var count int

func increment() {
    mu.Lock()
    defer mu.Unlock()
    count++
}
` + codeEnd

	_, err = store.CreateCardInDeck(
		"Go Concurrency Patterns",
		"What are the main concurrency patterns in Go?",
		concurrencyContent,
		[]string{"go", "concurrency", "goroutines", "channels"},
		goDeck,
	)
	if err != nil {
		return fmt.Errorf("failed to create go concurrency card: %w", err)
	}

	// Go Slices vs Arrays
	slicesContent := `# Go Slices vs Arrays

## Arrays
- Fixed length, determined at declaration
- Value types, copying an array creates a new copy
- Passed by value to functions
- Type includes the size: ` + "`[5]int`" + ` is different from ` + "`[10]int`" + `

` + goStart + `
// Array declaration
var arr [5]int
arr := [5]int{1, 2, 3, 4, 5}
arr := [...]int{1, 2, 3, 4, 5} // Compiler counts elements
` + codeEnd + `

## Slices
- Dynamic length, can grow and shrink
- Reference types, pointing to an underlying array
- Passed by reference (technically by value, but the value is a reference)
- Type does not include size: just ` + "`[]int`" + `

` + goStart + `
// Slice declaration
var s []int
s := []int{1, 2, 3, 4, 5}
s := make([]int, 5)      // len=5, cap=5
s := make([]int, 3, 10)  // len=3, cap=10

// Creating slices from arrays or other slices
s := arr[1:4]
` + codeEnd + `

## Key operations
` + goStart + `
// Append to slice (may reallocate underlying array)
s = append(s, 6, 7, 8)

// Length and capacity
len(s) // Number of elements
cap(s) // Size of underlying array

// Slicing
s[1:4]  // Elements from index 1 to 3
s[:3]   // Elements from start to index 2
s[2:]   // Elements from index 2 to end
` + codeEnd

	slicesCard, err := store.CreateCardInDeck(
		"Go Slices vs Arrays",
		"What's the difference between slices and arrays in Go?",
		slicesContent,
		[]string{"go", "arrays", "slices", "data-structures"},
		goDeck,
	)
	if err != nil {
		return fmt.Errorf("failed to create go slices card: %w", err)
	}

	// Modify the card to simulate a review history
	slicesCard.LastReviewed = time.Now().AddDate(0, 0, -2)
	slicesCard.ReviewInterval = 4
	slicesCard.Difficulty = 3
	if err := store.SaveCard(slicesCard); err != nil {
		return fmt.Errorf("failed to update go slices card: %w", err)
	}

	// Create example programming cards (Python)

	// Python Decorators
	decoratorsContent := `# Python Decorators

Decorators are a design pattern in Python that allows you to modify the behavior of a function or class without directly changing its source code.

## Basic Syntax

` + pythonStart + `
@decorator_function
def my_function():
    pass
` + codeEnd + `

This is equivalent to:

` + pythonStart + `
def my_function():
    pass
my_function = decorator_function(my_function)
` + codeEnd + `

## Example: Simple Function Decorator

` + pythonStart + `
def timing_decorator(func):
    def wrapper(*args, **kwargs):
        import time
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()
        print(f"{func.__name__} executed in {end_time - start_time:.4f} seconds")
        return result
    return wrapper

@timing_decorator
def slow_function():
    import time
    time.sleep(1)
    print("Function executed")

# When called:
slow_function()
# Output:
# Function executed
# slow_function executed in 1.0012 seconds
` + codeEnd + `

## Decorators with Arguments

` + pythonStart + `
def repeat(n=1):
    def decorator(func):
        def wrapper(*args, **kwargs):
            for _ in range(n):
                result = func(*args, **kwargs)
            return result
        return wrapper
    return decorator

@repeat(3)
def say_hello():
    print("Hello!")

# When called:
say_hello()
# Output:
# Hello!
# Hello!
# Hello!
` + codeEnd + `

## Class Decorators

` + pythonStart + `
def add_greeting(cls):
    cls.greet = lambda self: f"Hello, I'm {self.name}"
    return cls

@add_greeting
class Person:
    def __init__(self, name):
        self.name = name

# Usage:
p = Person("Alice")
print(p.greet())  # Output: Hello, I'm Alice
` + codeEnd + `

## Built-in Decorators
- ` + "`@property`" + `: Convert a method to a read-only attribute
- ` + "`@classmethod`" + `: Define a method that operates on the class, not instance
- ` + "`@staticmethod`" + `: Define a method that doesn't need class or instance
- ` + "`@contextlib.contextmanager`" + `: Define a context manager using a generator
- ` + "`@functools.lru_cache`" + `: Cache function results for performance`

	_, err = store.CreateCardInDeck(
		"Python Decorators",
		"What are decorators in Python and how do they work?",
		decoratorsContent,
		[]string{"python", "decorators", "functions", "metaprogramming"},
		pythonDeck,
	)
	if err != nil {
		return fmt.Errorf("failed to create python decorators card: %w", err)
	}

	// Python List Comprehensions
	listCompContent := `# Python List Comprehensions

List comprehensions provide a concise way to create lists based on existing sequences.

## Basic Syntax

` + pythonStart + `
[expression for item in iterable]
` + codeEnd + `

## Examples

Simple list generation:
` + pythonStart + `
# Create a list of squares from 0 to 9
squares = [x**2 for x in range(10)]
# Result: [0, 1, 4, 9, 16, 25, 36, 49, 64, 81]
` + codeEnd + `

With filtering condition:
` + pythonStart + `
# Create a list of even squares
even_squares = [x**2 for x in range(10) if x % 2 == 0]
# Result: [0, 4, 16, 36, 64]
` + codeEnd + `

Nested loops:
` + pythonStart + `
# Create coordinates for a 3x3 grid
coordinates = [(x, y) for x in range(3) for y in range(3)]
# Result: [(0,0), (0,1), (0,2), (1,0), (1,1), (1,2), (2,0), (2,1), (2,2)]
` + codeEnd + `

With conditional expression:
` + pythonStart + `
# Classify numbers as 'even' or 'odd'
classifications = ['even' if x % 2 == 0 else 'odd' for x in range(5)]
# Result: ['even', 'odd', 'even', 'odd', 'even']
` + codeEnd + `

## Dictionary Comprehensions

` + pythonStart + `
# Create a dictionary mapping numbers to their squares
square_dict = {x: x**2 for x in range(5)}
# Result: {0: 0, 1: 1, 2: 4, 3: 9, 4: 16}
` + codeEnd + `

## Set Comprehensions

` + pythonStart + `
# Create a set of all vowels in a string
vowels = {char for char in "hello world" if char in 'aeiou'}
# Result: {'e', 'o'}
` + codeEnd + `

## Generator Expressions
Similar to list comprehensions but use parentheses and generate values on demand:

` + pythonStart + `
# Create a generator of squares
square_gen = (x**2 for x in range(10))
# Usage:
for square in square_gen:
    print(square)
` + codeEnd + `

## Performance and Usage Tips
- List comprehensions are often faster than equivalent for loops
- Avoid using them for complex operations where readability suffers
- Don't nest too deeply (2+ levels) as readability decreases
- Can replace many map() and filter() functions more elegantly`

	listCompCard, err := store.CreateCardInDeck(
		"Python List Comprehensions",
		"What are list comprehensions in Python and how do they work?",
		listCompContent,
		[]string{"python", "list-comprehensions", "syntax"},
		pythonDeck,
	)
	if err != nil {
		return fmt.Errorf("failed to create python list comprehensions card: %w", err)
	}

	// Modify the card to simulate a review history
	listCompCard.LastReviewed = time.Now().AddDate(0, 0, -5)
	listCompCard.ReviewInterval = 10
	listCompCard.Difficulty = 4
	if err := store.SaveCard(listCompCard); err != nil {
		return fmt.Errorf("failed to update python list comprehensions card: %w", err)
	}

	// Create example computer science concept cards

	// Big O Notation
	bigOContent := `# Big O Notation

Big O notation describes the performance or complexity of an algorithm by characterizing how its runtime or space requirements grow relative to input size.

## Definition
O(g(n)) = { f(n) : there exist positive constants c and n₀ such that 0 ≤ f(n) ≤ c*g(n) for all n ≥ n₀ }

## Common Big O Complexities (from fastest to slowest)

| Notation    | Name          | Example                                      |
|-------------|---------------|----------------------------------------------|
| O(1)        | Constant      | Array access, hash table lookup              |
| O(log n)    | Logarithmic   | Binary search, balanced tree operations      |
| O(n)        | Linear        | Simple loops, linear search                  |
| O(n log n)  | Linearithmic  | Efficient sorting (merge sort, quicksort)    |
| O(n²)       | Quadratic     | Nested loops, bubble sort                    |
| O(n³)       | Cubic         | Triple nested loops, some matrix operations  |
| O(2ⁿ)       | Exponential   | Recursive Fibonacci, power set               |
| O(n!)       | Factorial     | Permutations, travelling salesman (brute)    |

## Why Important?
- Predicts how algorithms scale with larger inputs
- Language-independent way to compare algorithms
- Focuses on the dominant factors affecting performance
- Helps identify bottlenecks in large-scale systems
- Key consideration in algorithm design and selection

## Big O Rules of Thumb
1. **Drop constants**: O(2n) simplifies to O(n)
2. **Drop lower-order terms**: O(n² + n) simplifies to O(n²)
3. **Consider worst-case scenario** unless specified otherwise
4. **Different inputs get different variables**: O(a + b) for two separate inputs
5. **Multiplication for nested operations**: O(n*m) for nested loops over different inputs

## Example Analysis

**Finding an element in an array**
` + jsStart + `
function findElement(array, element) {
    for (let i = 0; i < array.length; i++) {
        if (array[i] === element) return i;
    }
    return -1;
}
` + codeEnd + `
Time Complexity: O(n) - linear time (worst case checks every element)`

	bigOCard, err := store.CreateCardInDeck(
		"Big O Notation",
		"What is Big O Notation and why is it important?",
		bigOContent,
		[]string{"algorithm", "complexity", "big-o", "computer-science"},
		conceptsDeck,
	)
	if err != nil {
		return fmt.Errorf("failed to create big o card: %w", err)
	}

	// Modify the card to simulate a review history
	bigOCard.LastReviewed = time.Now().AddDate(0, 0, -1)
	bigOCard.ReviewInterval = 3
	bigOCard.Difficulty = 2
	if err := store.SaveCard(bigOCard); err != nil {
		return fmt.Errorf("failed to update big o card: %w", err)
	}

	// Design Patterns
	designPatternsContent := `# Factory vs. Builder Design Patterns

## Factory Pattern

The Factory pattern provides an interface for creating objects without specifying their concrete classes.

### Factory Method Example (Java)

` + javaStart + `
// Product interface
interface Vehicle {
    void drive();
}

// Concrete products
class Car implements Vehicle {
    public void drive() {
        System.out.println("Driving a car...");
    }
}

class Motorcycle implements Vehicle {
    public void drive() {
        System.out.println("Riding a motorcycle...");
    }
}

// Factory
class VehicleFactory {
    public Vehicle createVehicle(String type) {
        if ("car".equals(type)) {
            return new Car();
        } else if ("motorcycle".equals(type)) {
            return new Motorcycle();
        }
        throw new IllegalArgumentException("Unknown vehicle type");
    }
}

// Usage
VehicleFactory factory = new VehicleFactory();
Vehicle car = factory.createVehicle("car");
car.drive(); // Output: Driving a car...
` + codeEnd + `

## Builder Pattern

The Builder pattern separates the construction of a complex object from its representation, allowing the same construction process to create different representations.

### Builder Example (Java)

` + javaStart + `
// Product
class Pizza {
    private String dough;
    private String sauce;
    private String topping;

    public void setDough(String dough) { this.dough = dough; }
    public void setSauce(String sauce) { this.sauce = sauce; }
    public void setTopping(String topping) { this.topping = topping; }

    public void describe() {
        System.out.println("Pizza with " + dough + " dough, " + sauce + " sauce, and " + topping + " topping");
    }
}

// Builder interface
interface PizzaBuilder {
    void buildDough();
    void buildSauce();
    void buildTopping();
    Pizza getPizza();
}

// Concrete builder
class HawaiianPizzaBuilder implements PizzaBuilder {
    private Pizza pizza = new Pizza();

    public void buildDough() { pizza.setDough("thin"); }
    public void buildSauce() { pizza.setSauce("mild"); }
    public void buildTopping() { pizza.setTopping("ham and pineapple"); }
    public Pizza getPizza() { return pizza; }
}

// Director
class Cook {
    private PizzaBuilder pizzaBuilder;

    public void setPizzaBuilder(PizzaBuilder pb) { pizzaBuilder = pb; }

    public Pizza getPizza() { return pizzaBuilder.getPizza(); }

    public void constructPizza() {
        pizzaBuilder.buildDough();
        pizzaBuilder.buildSauce();
        pizzaBuilder.buildTopping();
    }
}

// Usage
Cook cook = new Cook();
PizzaBuilder hawaiianPizzaBuilder = new HawaiianPizzaBuilder();
cook.setPizzaBuilder(hawaiianPizzaBuilder);
cook.constructPizza();
Pizza hawaiianPizza = cook.getPizza();
hawaiianPizza.describe(); // Output: Pizza with thin dough, mild sauce, and ham and pineapple topping
` + codeEnd + `

## Key Differences:

1. **Purpose**:
   - Factory: Creates objects without exposing instantiation logic
   - Builder: Constructs complex objects step by step

2. **Flexibility**:
   - Factory: Returns different types based on input
   - Builder: Constructs the same type with different configurations

3. **Construction**:
   - Factory: Creates objects in a single step
   - Builder: Creates objects in multiple steps

4. **Use Case**:
   - Factory: When the exact type of object isn't known until runtime
   - Builder: When an object has many optional parameters or complex initialization`

	_, err = store.CreateCardInDeck(
		"Design Patterns: Factory vs. Builder",
		"What are the Factory and Builder design patterns, and how do they differ?",
		designPatternsContent,
		[]string{"design-patterns", "factory-pattern", "builder-pattern", "java"},
		conceptsDeck,
	)
	if err != nil {
		return fmt.Errorf("failed to create design patterns card: %w", err)
	}

	return nil
}

// runCLIMode runs the original example code from the previous main function
func runCLIMode(store storage.CardStoreInterface) {
	fmt.Println("Running in CLI mode (for backwards compatibility).")
	fmt.Println("This mode is deprecated and will be removed in a future version.")
	fmt.Println("We recommend using the terminal UI mode (default).")
	fmt.Println()

	// Example: Create a new flashcard
	exampleCard, err := store.CreateCard(
		"Two-Pointer Technique",
		"What is the two-pointer technique in algorithms and when should it be used?",
		`The two-pointer technique uses two pointers to iterate through a data structure simultaneously.

It's particularly useful for:
- Sorted array operations
- Finding pairs with certain conditions
- String manipulation (palindromes)
- Linked list cycle detection

Example (Two Sum in sorted array):
`+"```python\ndef two_sum(nums, target):\n    left, right = 0, len(nums) - 1\n    while left < right:\n        current_sum = nums[left] + nums[right]\n        if current_sum == target:\n            return [left, right]\n        elif current_sum < target:\n            left += 1\n        else:\n            right -= 1\n    return [-1, -1]  # No solution\n```",
		[]string{"algorithms", "techniques", "arrays"},
	)
	if err != nil {
		fmt.Printf("Failed to create example card: %v\n", err)
		return
	}

	fmt.Printf("Created new card: %s at %s\n", exampleCard.Title, exampleCard.FilePath)

	// Rest of the CLI mode code...
	// (This is just a placeholder - in a real implementation, you'd add the rest of the CLI mode code)
	fmt.Println("CLI mode example complete. Run with -tui flag (or no arguments) for full functionality.")
}
