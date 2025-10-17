# Contributing to OpenID Connect Server

Thank you for your interest in contributing to this project! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Node.js 20 or later
- Git

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/openid-golang.git
   cd openid-golang
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/prasenjit-net/openid-golang.git
   ```

4. Run the setup script:
   ```bash
   ./setup.sh
   ```

5. Install UI dependencies:
   ```bash
   cd ui/admin
   npm install
   cd ../..
   ```

## Development Workflow

### Creating a Branch

Always create a new branch for your work:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-description
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test improvements

### Making Changes

1. Make your changes in your branch
2. Write or update tests as needed
3. Update documentation as needed
4. Ensure all tests pass
5. Commit your changes with clear commit messages

### Running Tests

**Backend tests:**
```bash
go test -v ./...
```

**Frontend build:**
```bash
cd ui/admin
npm run build
cd ../..
```

**Run the server:**
```bash
./test.sh
```

### Keeping Your Fork Updated

```bash
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
```

## Submitting Changes

### Pull Request Process

1. Update your branch with the latest changes from main:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Push your changes to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

3. Create a Pull Request on GitHub with:
   - Clear title describing the change
   - Detailed description of what changed and why
   - Reference to any related issues
   - Screenshots (if UI changes)

4. Wait for review and address any feedback

### Commit Message Guidelines

Write clear, descriptive commit messages:

```
type(scope): brief description

Longer description if needed, explaining what and why.

Fixes #123
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(admin): add user bulk delete functionality
fix(auth): resolve token expiration race condition
docs(api): update OAuth 2.0 flow documentation
```

## Coding Standards

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Run `gofmt` before committing
- Run `go vet` to catch common mistakes
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small

**Example:**
```go
// AuthenticateUser verifies user credentials and returns a user object.
// Returns an error if authentication fails.
func AuthenticateUser(username, password string) (*User, error) {
    // Implementation
}
```

### React/TypeScript Code

- Use TypeScript for type safety
- Follow React hooks best practices
- Use functional components
- Add PropTypes or TypeScript interfaces
- Keep components small and focused
- Use meaningful component and variable names

**Example:**
```typescript
interface UserListProps {
  users: User[];
  onDelete: (id: string) => void;
}

const UserList: React.FC<UserListProps> = ({ users, onDelete }) => {
  // Implementation
};
```

### Code Review Checklist

Before submitting, ensure:
- [ ] Code follows project style guidelines
- [ ] All tests pass
- [ ] New code has tests
- [ ] Documentation is updated
- [ ] No commented-out code or debug logs
- [ ] Error handling is appropriate
- [ ] Code is DRY (Don't Repeat Yourself)

## Testing

### Writing Tests

**Go tests:**
```go
func TestUserAuthentication(t *testing.T) {
    // Arrange
    user := &User{Username: "test", Password: "hashed"}
    
    // Act
    result, err := AuthenticateUser(user.Username, "password")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Test Coverage

Aim for good test coverage:
```bash
go test -cover ./...
```

## Documentation

### Code Documentation

- Add godoc comments for exported types and functions
- Include usage examples in comments
- Document edge cases and error conditions

### User Documentation

Update relevant files in the `docs/` directory when:
- Adding new features
- Changing APIs
- Updating configuration options
- Modifying deployment procedures

## Questions?

If you have questions or need help:
- Open an issue for discussion
- Join our community chat (if available)
- Email the maintainers

Thank you for contributing! 🎉
