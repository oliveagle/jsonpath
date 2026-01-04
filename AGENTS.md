# AGENTS.md - JSONPath Project

This file provides guidance to AI agents working on the JSONPath project.

## Project Overview

JSONPath is a Go library for querying JSON data using JSONPath expressions. It provides:

- **Language**: Go (Golang) 1.25+
- **Type**: Library/package
- **Purpose**: JSON data querying and manipulation
- **Features**: JSONPath expression parsing and evaluation

## Key Configuration Files

- `go.mod` - Go module definition
- `jsonpath.go` - Main library implementation
- `jsonpath_test.go` - Comprehensive test suite
- `.travis.yml` - CI/CD configuration

## Build and Test Commands

### Installation
```bash
# Install the package
go get github.com/your-repo/jsonpath

# Install dependencies
go mod tidy
```

### Development
```bash
# Build the project
go build

# Run the library
go run .
```

### Testing
```bash
# Run all tests
go test -v

# Run tests with coverage
go test -cover

# Run specific test functions
go test -run TestFunctionName
```

### Code Quality
```bash
# Format code
gofmt -w .

# Check for formatting issues
gofmt -d .
```

## Project Structure

```
jsonpath.go        # Main library implementation
jsonpath_test.go   # Test suite
go.mod             # Go module definition
readme.md          # Project documentation
```

## Code Style Guidelines

- **Go Standards**: Follow official Go code review comments
- **Formatting**: Use `gofmt` for consistent formatting
- **Naming**: Use camelCase for variables, PascalCase for exported types
- **Error Handling**: Explicit error handling (no panic for expected errors)
- **Documentation**: Add godoc comments for exported functions/types
- **Testing**: Comprehensive test coverage for all functionality

## Testing Instructions

- **Test Files**: `jsonpath_test.go` contains comprehensive tests
- **Test Patterns**: Table-driven tests for JSONPath expressions
- **Coverage**: Aim for high test coverage of all code paths
- **Edge Cases**: Test malformed JSON, invalid paths, boundary conditions

## JSONPath Implementation Details

- **Expression Parsing**: Custom JSONPath parser implementation
- **Query Evaluation**: Efficient JSON traversal algorithms
- **Result Handling**: Support for various return types
- **Error Handling**: Clear error messages for invalid expressions

## Security Considerations

- **Input Validation**: Validate JSONPath expressions
- **Memory Safety**: Handle large JSON documents efficiently
- **Error Messages**: Don't expose sensitive information
- **Dependency Management**: Minimal dependencies for security

## Performance Considerations

- **Parsing Optimization**: Efficient JSONPath expression parsing
- **Traversal Algorithms**: Optimized JSON document traversal
- **Memory Usage**: Minimize allocations during query execution
- **Benchmarking**: Consider adding performance benchmarks

## Usage Examples

```go
// Basic usage example
result, err := jsonpath.Get(pathExpression, jsonData)
if err != nil {
    // Handle error
}
// Use result...
```

## Git Conventions

- **Commit Messages**: Clear, descriptive commit messages
- **Branching**: Use feature branches for new development
- **Pull Requests**: Required for merging to main branch
- **Tags**: Use semantic versioning for releases

## CI/CD

- **Travis CI**: Configured in `.travis.yml`
- **Automated Testing**: Runs on every push/PR
- **Build Verification**: Ensures project builds successfully
- **Test Coverage**: Reports test coverage metrics

## Documentation

- **readme.md**: Contains usage examples and API documentation
- **Godoc**: Use godoc comments for inline documentation
- **Examples**: Include practical usage examples

## Dependency Management

- **Minimal Dependencies**: Only standard library dependencies
- **Go Modules**: Uses Go modules for dependency management
- **Updates**: Regularly update Go version in `go.mod`

## Future Enhancements

- **Additional Features**: Consider adding more JSONPath features
- **Performance**: Optimize for large JSON documents
- **Compatibility**: Ensure compatibility with JSONPath standards
- **Documentation**: Expand usage examples and tutorials

## Task Implementation
1. **Analyze Requirements**: Refer to `README.md` for detailed feature specifications and system design.
2. **Implementation**: Modify source code in the respective directories (e.g., `src/`, `internal/`).
3. **Verification**: Run provided build and test commands (see above) to ensure correctness.
4. **Push Changes**:
   - Commit changes: `git commit -m "feat: implement <feature>"`
   - Push to remote: `git push origin <branch-name>`
