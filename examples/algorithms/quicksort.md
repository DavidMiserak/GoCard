---
tags: [algorithms,sorting,divide-and-conquer,complexity]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 0
---

# Quicksort Algorithm

## Question

Explain the quicksort algorithm, including its time complexity, space complexity, and when it performs well or poorly.

## Answer

Quicksort is a divide-and-conquer sorting algorithm that works by selecting a 'pivot' element and partitioning the array around it.

### Algorithm Steps

1. Choose a pivot element from the array
2. Partition the array around the pivot (elements < pivot go left, elements > pivot go right)
3. Recursively apply the above steps to the sub-arrays

### Implementation

```python
def quicksort(arr, low=0, high=None):
    if high is None:
        high = len(arr) - 1

    if low < high:
        # Partition the array and get pivot position
        pivot_index = partition(arr, low, high)

        # Recursively sort the sub-arrays
        quicksort(arr, low, pivot_index - 1)
        quicksort(arr, pivot_index + 1, high)

    return arr

def partition(arr, low, high):
    # Choose rightmost element as pivot
    pivot = arr[high]

    # Index of smaller element
    i = low - 1

    for j in range(low, high):
        # If current element is smaller than the pivot
        if arr[j] <= pivot:
            # Increment index of smaller element
            i += 1
            arr[i], arr[j] = arr[j], arr[i]

    # Place pivot in its correct position
    arr[i + 1], arr[high] = arr[high], arr[i + 1]
    return i + 1
```

### Time Complexity

- **Best case**: O(n log n) - When the pivot divides the array into roughly equal halves
- **Average case**: O(n log n)
- **Worst case**: O(n²) - When the smallest or largest element is always chosen as pivot (e.g., already sorted array)

### Space Complexity

- O(log n) for the recursion stack in average case
- O(n) in worst case

### Performance Characteristics

#### When Quicksort Performs Well

- Random or evenly distributed data
- Cache efficiency (good locality of reference)
- When implemented with optimizations like:

  - Random pivot selection
  - Median-of-three pivot selection
  - Switching to insertion sort for small subarrays

#### When Quicksort Performs Poorly

- Already sorted or nearly sorted arrays
- Arrays with many duplicate elements
- When the pivot selection consistently produces unbalanced partitions

### Key Advantages

- In-place sorting (requires small additional space)
- Cache-friendly
- Typically faster than other O(n log n) algorithms like mergesort in practice
- Easy to implement and optimize

### Key Disadvantages

- Not stable (equal elements may change relative order)
- Worst-case performance is O(n²)
- Recursive, so can cause stack overflow for very large arrays
