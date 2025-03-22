---
tags: [math,calculus,limits,basic-concepts]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 0
---

# Limits in Calculus

## Question

What is a limit in calculus? Explain the formal (ε-δ) definition and provide examples of evaluating limits, including cases where direct substitution doesn't work.

## Answer

A limit describes the value a function approaches as the input approaches a particular value.

### Informal Definition

For a function f(x), the limit of f(x) as x approaches a value c is written as:

$$\lim_{x \to c} f(x) = L$$

This means that as x gets arbitrarily close to c (but not necessarily equal to c), the function value f(x) gets arbitrarily close to L.

### Formal (ε-δ) Definition

$$\lim_{x \to c} f(x) = L \text{ if and only if for every } \varepsilon > 0 \text{ there exists a } \delta > 0 \text{ such that if } 0 < |x - c| < \delta \text{ then } |f(x) - L| < \varepsilon$$

In plain language: We can make f(x) as close as we want to L by making x sufficiently close to c.

### Evaluating Limits

#### Method 1: Direct Substitution

If f(x) is continuous at x = c, then:

$$\lim_{x \to c} f(x) = f(c)$$

**Example**:

$$\lim_{x \to 2} (x^2 + 3x) = 2^2 + 3(2) = 4 + 6 = 10$$

#### Method 2: Algebraic Manipulation

When direct substitution gives an indeterminate form (like 0/0), algebraic manipulation can help.

**Example**:

$$\lim_{x \to 3} \frac{x^2 - 9}{x - 3}$$

Direct substitution gives $\frac{0}{0}$ (indeterminate), so we factor:
$$\lim_{x \to 3} \frac{(x - 3)(x + 3)}{x - 3} = \lim_{x \to 3} (x + 3) = 6$$

#### Method 3: L'Hôpital's Rule

For indeterminate forms like $\frac{0}{0}$ or $\frac{\infty}{\infty}$:
$$\lim_{x \to c} \frac{f(x)}{g(x)} = \lim_{x \to c} \frac{f'(x)}{g'(x)}$$

**Example**:

$$\lim_{x \to 0} \frac{\sin(x)}{x} = \lim_{x \to 0} \frac{\cos(x)}{1} = 1$$

#### Method 4: Special Limit Results

**Example**: Squeeze Theorem

If g(x) ≤ f(x) ≤ h(x) near x = c, and $\lim_{x \to c} g(x) = \lim_{x \to c} h(x) = L$, then $\lim_{x \to c} f(x) = L$

### Common Indeterminate Forms

- $\frac{0}{0}$ - Use factoring or L'Hôpital's rule
- $\frac{\infty}{\infty}$ - Use algebraic manipulation or L'Hôpital's rule
- $0 \cdot \infty$ - Rewrite as $\frac{0}{1/\infty}$ or $\frac{\infty}{1/0}$
- $\infty - \infty$ - Find a common denominator or use algebraic manipulation
- $0^0$, $1^{\infty}$, $\infty^0$ - Use logarithms or exponential properties

### One-sided Limits

- Left-hand limit: $\lim_{x \to c^-} f(x)$ (x approaches c from values less than c)
- Right-hand limit: $\lim_{x \to c^+} f(x)$ (x approaches c from values greater than c)
- A limit exists if and only if both one-sided limits exist and are equal
