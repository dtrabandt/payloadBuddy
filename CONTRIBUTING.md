# Contributing to PayloadBuddy

We welcome contributions to PayloadBuddy! This project follows git-flow and Test-Driven Development (TDD) practices.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Quality Standards](#code-quality-standards)
- [Testing Guidelines](#testing-guidelines)
- [CI/CD Pipeline](#cicd-pipeline)
- [Project Structure](#project-structure)
- [Areas for Improvement](#areas-for-improvement)

## Getting Started

### Prerequisites

- Go 1.21 or newer
- Git with git-flow extension (optional but recommended)
- golangci-lint for linting
- gosec for security scanning

### Development Setup

1. **Fork the repository** on GitHub
2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/payloadBuddy.git
   cd payloadBuddy
   ```
3. **Set up git-flow** (optional):
   ```bash
   git flow init -d  # Use default branch names
   ```
4. **Install development tools**:
   ```bash
   # Install linter
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   
   # Install security scanner
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   ```

## Development Workflow

### 1. Create a Feature Branch

Using git-flow:
```bash
git flow feature start your-feature-name
```

Or manually:
```bash
git checkout develop
git checkout -b feature/your-feature-name
```

### 2. Follow TDD Approach

This project uses **Test-Driven Development**. See [CLAUDE.md](CLAUDE.md) for detailed TDD guidelines.

**Red-Green-Refactor Cycle**:
1. **Write a failing test** first
2. **Write minimal code** to make it pass
3. **Refactor** while keeping tests green

### 3. Code Quality Checklist

Before committing, ensure your code meets these standards:

```bash
# Format code (required)
gofmt -s -w .

# Run tests with coverage
go test -v -cover ./...

# Run linting
golangci-lint run

# Run security scanning
gosec ./...

# Build verification
go build -o payloadBuddy
```

### 4. Commit and Push

```bash
git add .
git commit -m "feat: add your feature description"
git push origin feature/your-feature-name
```

### 5. Create Pull Request

1. **Target the `develop` branch** (not `main`)
2. **Fill out the PR template** with:
   - Description of changes
   - Test coverage information
   - Breaking changes (if any)
3. **Wait for CI checks** to pass
4. **Address review feedback**

### 6. Finish Feature

Using git-flow:
```bash
git flow feature finish your-feature-name
```

## Code Quality Standards

### Formatting

- **Required**: All code must be formatted with `gofmt -s`
- **Enforced**: CI will fail if code is not properly formatted
- **Standard**: Follow Go community formatting conventions

### Linting

We use `golangci-lint` with strict settings:
- **Static analysis**: Catches common bugs and issues
- **Code style**: Enforces consistent patterns
- **Performance**: Identifies potential optimizations

### Security

- **Security scanning**: All code is scanned with `gosec`
- **Credentials**: Never commit secrets or keys
- **Dependencies**: Keep dependencies updated and secure

### Testing

- **Minimum coverage**: 75% (currently at 76.9%)
- **TDD approach**: Write tests before implementation
- **Comprehensive**: Cover happy paths, edge cases, and error conditions
- **Fast execution**: Test suite should complete in under 10 seconds

## Testing Guidelines

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {
            name:     "descriptive test case name",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)
            
            if tt.wantErr && err == nil {
                t.Error("Expected error but got none")
            }
            if !tt.wantErr && err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
            if result != tt.expected {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### Test Types

1. **Unit Tests**: Test individual functions in isolation
2. **Integration Tests**: Test HTTP handlers end-to-end
3. **Plugin Tests**: Test PayloadPlugin interface compliance
4. **Authentication Tests**: Test auth middleware scenarios

### Coverage Requirements

- **Minimum**: 75% statement coverage
- **Target**: 80%+ for new features
- **Exceptions**: `main()` function and some error paths are acceptable to exclude

## CI/CD Pipeline

### Automated Testing

Every PR triggers:
- **Go Testing**: Full test suite with race detection
- **Code Coverage**: Must meet 75% minimum
- **Linting**: `golangci-lint` with strict rules
- **Security**: `gosec` vulnerability scanning
- **Formatting**: `gofmt` compliance check
- **Build**: Cross-platform compilation test

### Quality Gates

All checks must pass before merging:
- ✅ All tests pass
- ✅ Coverage ≥ 75%
- ✅ No linting violations
- ✅ No security issues
- ✅ Code properly formatted
- ✅ Build successful

### Release Process

1. **Development**: Work happens on `develop` branch
2. **Release**: Create release branch from `develop`
3. **Main**: Merge to `main` triggers automated release
4. **Tagging**: Git tags (e.g., `v1.0.0`) create GitHub releases
5. **Binaries**: Cross-platform builds for all supported architectures

## Project Structure

```
├── main.go                          # Server setup and plugin registration
├── auth.go                          # HTTP Basic Authentication middleware
├── openapi.go                       # OpenAPI 3.1.1 data structures
├── rest_payload_handler.go          # Large single-response endpoint
├── streaming_payload_handler.go     # Advanced streaming endpoint
├── documentation_handler.go         # OpenAPI spec and Swagger UI
├── *_test.go                        # Comprehensive test suite
├── .github/workflows/               # CI/CD pipeline
│   ├── test.yml                     # PR testing workflow
│   └── release.yml                  # Release automation
├── README.md                        # User documentation
├── CONTRIBUTING.md                  # This file
├── CLAUDE.md                        # Development guidance for AI
└── DEPLOYMENT.md                    # Deployment strategies
```

### Plugin Architecture

PayloadBuddy uses a plugin system where endpoints implement the `PayloadPlugin` interface:

```go
type PayloadPlugin interface {
    Path() string
    Handler() http.HandlerFunc
    OpenAPISpec() OpenAPIPathSpec
}
```

### Adding New Endpoints

1. **Create handler file**: `your_handler.go`
2. **Implement interface**: Create plugin struct implementing `PayloadPlugin`
3. **Register plugin**: Add `registerPlugin(YourPlugin{})` in `init()`
4. **Add tests**: Create `your_handler_test.go`
5. **Update documentation**: Plugin automatically appears in OpenAPI spec

## Areas for Improvement

We welcome contributions in these areas:

### Features
- Additional ServiceNow scenarios and delay patterns
- Performance optimizations and monitoring
- Additional output formats (XML, CSV, etc.)
- New authentication methods (JWT, OAuth)
- Docker containerization and Kubernetes deployments

### Testing
- Increase test coverage above 80%
- Add benchmark tests for performance scenarios
- Integration tests with real ServiceNow instances
- Load testing scenarios

### Documentation
- API usage examples for different languages
- Performance tuning guides
- Troubleshooting documentation
- Video tutorials for ServiceNow consultants

### DevOps
- Docker images for easy deployment
- Kubernetes manifests
- Monitoring and observability
- Automated security updates

## Getting Help

- **Issues**: Report bugs and request features via [GitHub Issues](https://github.com/dtrabandt/payloadBuddy/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/dtrabandt/payloadBuddy/discussions)
- **Documentation**: Check [CLAUDE.md](CLAUDE.md) for detailed development guidance

## Code of Conduct

Please be respectful and collaborative. This project is designed to help ServiceNow consultants and developers, so constructive feedback and patience with different skill levels is appreciated.

## License

By contributing to PayloadBuddy, you agree that your contributions will be licensed under the same [MIT License](LICENSE) that covers the project.