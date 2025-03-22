---
tags: [gocard,features,markdown,syntax-highlighting,tables]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 0
---

# GoCard Markdown Features Demo

## Question

What Markdown features does GoCard support and how can they be used effectively in flashcards?

## Answer

GoCard supports a wide range of Markdown features, making it powerful for creating rich flashcards. Here's a comprehensive demonstration:

### Text Formatting

**Bold text** is created with `**double asterisks**`

*Italic text* is created with `*single asterisks*`

***Bold and italic*** text is created with `***triple asterisks***`

~~Strikethrough~~ is created with `~~double tildes~~`

### Lists

Unordered lists:

- Item 1
- Item 2
  - Nested item
  - Another nested item
- Item 3

Ordered lists:

1. First item
2. Second item
3. Third item
   1. Nested ordered item
   2. Another nested item

### Code Blocks with Syntax Highlighting

JavaScript example:

```javascript
function calculateFactorial(n) {
  if (n === 0 || n === 1) {
    return 1;
  }
  return n * calculateFactorial(n - 1);
}

// Calculate and display factorial of 5
console.log(calculateFactorial(5)); // Output: 120
```

SQL example:

```sql
SELECT
  users.name,
  COUNT(orders.id) AS order_count,
  SUM(orders.amount) AS total_spent
FROM users
JOIN orders ON users.id = orders.user_id
WHERE orders.created_at > DATE_SUB(NOW(), INTERVAL 1 YEAR)
GROUP BY users.id
HAVING total_spent > 1000
ORDER BY total_spent DESC
LIMIT 10;
```

Inline code: `const x = 42;`

### Tables

| Feature | Markdown Syntax | Notes |
|---------|-----------------|-------|
| Headers | `# H1`, `## H2` | Up to 6 levels |
| Bold | `**text**` | For emphasis |
| Tables | `\| col1 \| col2 \|` | Requires header row |
| Code | \`\`\`language | Specify language for highlighting |

### Blockquotes

> This is a blockquote.
>
> It can span multiple lines.
>
> > And it can be nested.

### Links and Images

[Link to GoCard Repository](https://github.com/DavidMiserak/GoCard)

![GoCard Logo](assets/gocard-logo.webp)

### Task Lists

- [x] Implemented core features
- [x] Added markdown support
- [ ] Complete all example cards
- [ ] Release v1.0

### Horizontal Rules

---

### Mathematical Notation

Inline math: $E = mc^2$

Block math:
$$
\frac{d}{dx}(e^x) = e^x
$$

$$
\int_{a}^{b} f(x) \, dx = F(b) - F(a)
$$

### Tips for Effective Flashcards

1. **Keep questions focused** on one concept
2. **Use formatting** to emphasize key points
3. **Include code examples** where relevant
4. **Use tables** to organize related information
5. **Add visual elements** when they help understanding
6. **Structure answers** with clear headings and sections
