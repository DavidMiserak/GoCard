# Contributing to GoCard

First off, thank you for considering contributing to GoCard! It's people like you that make GoCard such a great tool for learning and knowledge management.

## Code of Conduct

This project and everyone participating in it are governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

1. **Ensure the bug is not already reported** by searching existing GitHub issues.
2. If you can't find an existing issue, [open a new one](https://github.com/DavidMiserak/GoCard/issues/new?template=bug_report.md).
3. Include a clear title and description, as much relevant information as possible, and a code sample or an executable test case demonstrating the expected behavior that is not occurring.

### Suggesting Enhancements

1. Check the [issues list](https://github.com/DavidMiserak/GoCard/issues) to see if your suggestion is already there.
2. If not, [open a new feature request](https://github.com/DavidMiserak/GoCard/issues/new?template=feature_request.md).
3. Provide a clear and detailed explanation of the feature you want to see.

### Your First Code Contribution

Unsure where to begin contributing? You can start by looking through these `good-first-issue` and `help-wanted` issues:

- [Good First Issues](https://github.com/DavidMiserak/GoCard/labels/good%20first%20issue) - issues that should only require a few lines of code
- [Help Wanted](https://github.com/DavidMiserak/GoCard/labels/help%20wanted) - issues that are more involved but not necessarily difficult

### Pull Requests

1. Fork the repository and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue the pull request!

#### Pull Request Process

1. Update the README.md or documentation with details of changes if applicable.
2. Increase version numbers in any examples files and the README.md to the new version that this Pull Request would represent.
3. You may merge the Pull Request once it has been reviewed and approved by a maintainer.

## Development Setup

### Prerequisites

- Go 1.23 or later
- Git
- Pre-commit (optional but recommended)

### Setup Steps

1. Clone the repository

```bash
git clone https://github.com/DavidMiserak/GoCard.git
cd GoCard
```

2. Install dependencies

```bash
go mod download
```

3. Install pre-commit hooks (recommended)

```bash
make pre-commit-setup
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
make test-cover

# Run linters
make lint
```

## Coding Conventions

### Go Formatting

- Use `gofmt` for formatting
- Follow Go best practices and idioms
- Keep functions small and focused
- Add comments to explain complex logic

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) format:

- `feat:` A new feature
- `fix:` A bug fix
- `docs:` Documentation changes
- `style:` Formatting, missing semicolons, etc.
- `refactor:` Code refactoring
- `test:` Adding or modifying tests
- `chore:` Maintenance tasks

Example:

```markdown
feat: add import/export functionality for cards

- Implement Anki package (.apkg) import
- Add export to markdown feature
- Update documentation with new features
```

### Git Workflow

1. Create a feature branch: `git checkout -b feat/new-feature`
2. Make your changes
3. Run tests and linters
4. Commit with a conventional commit message
5. Push your branch
6. Open a pull request

## Reporting Security Issues

Please do not create public GitHub issues for security vulnerabilities.
Instead, send a detailed description to [david.miserak@gmail.com](mailto: david.miserak@gmail.com).

## Questions?

If you have any questions, feel free to open an issue or reach out to the maintainers.

**Happy Contributing!** 🚀📚
