---
tags: [python,decorators,programming,functions]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 0
---

# Python Decorators

## Question

What are decorators in Python and how do you implement a decorator with arguments?

## Answer

Decorators are a design pattern in Python that allows you to modify the behavior of a function or class without directly changing its source code.

### Basic Decorator Structure

```python
def my_decorator(func):
    def wrapper(*args, **kwargs):
        # Do something before the function call
        result = func(*args, **kwargs)
        # Do something after the function call
        return result
    return wrapper

@my_decorator
def my_function():
    pass
```

### Decorator with Arguments

To create a decorator that accepts arguments, you need an additional level of nesting:

```python
def repeat(n=1):
    def decorator(func):
        def wrapper(*args, **kwargs):
            result = None
            for _ in range(n):
                result = func(*args, **kwargs)
            return result
        return wrapper
    return decorator

@repeat(3)
def say_hello(name):
    print(f"Hello, {name}!")

# This will print "Hello, World!" three times
say_hello("World")
```

### Common Use Cases

1. **Timing functions**:

   ```python
   def timing_decorator(func):
       def wrapper(*args, **kwargs):
           import time
           start_time = time.time()
           result = func(*args, **kwargs)
           end_time = time.time()
           print(f"{func.__name__} took {end_time - start_time:.2f} seconds")
           return result
       return wrapper
   ```

2. **Caching/memoization**:

   ```python
   def memoize(func):
       cache = {}
       def wrapper(*args):
           if args not in cache:
               cache[args] = func(*args)
           return cache[args]
       return wrapper
   ```

3. **Authentication and authorization**:

   ```python
   def requires_auth(func):
       def wrapper(*args, **kwargs):
           if not is_authenticated():
               raise Exception("Authentication required")
           return func(*args, **kwargs)
       return wrapper
   ```

4. **Logging**:

   ```python
   def log_function_call(func):
       def wrapper(*args, **kwargs):
           print(f"Calling {func.__name__} with {args}, {kwargs}")
           result = func(*args, **kwargs)
           print(f"{func.__name__} returned {result}")
           return result
       return wrapper
   ```
