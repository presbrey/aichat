# Changelog

## [1.1.2] - 2025-02-08

### Added
- New message manipulation methods:
  - `PopMessage()`: Remove and return the last message
  - `PopMessageIfRole(role)`: Remove and return last message if role matches
  - `SetSystemContent(content)`: Set or update system message content
  - `SetSystemMessage(msg)`: Set or update system message
  - `ShiftMessages()`: Remove and return the first message
  - `UnshiftMessages(msg)`: Insert a message at the beginning

### Documentation
- Expanded README with detailed sections on:
  - API usage examples
  - Extended code examples for tool calls
  - S3 interface implementation details
  - Rich content handling
  - Comprehensive feature list

## [v1.1.1] - 2025-02-04

### Added
- 100% test coverage
- OpenRouter integration with comprehensive type definitions and tests
- Added testify assertion library for improved test readability

### Changed
- Refactored test suite to use testify assertions
- Improved test readability and maintainability
- Enhanced error handling in tests

## [v1.1.0] - 2025-02-02

### Added
- Idempotent message handling with new `AddMessage` method
- Added comprehensive examples for message handling in README
- Added new convenience methods for message manipulation

### Changed
- Changed `Messages` type from `[]Message` to `[]*Message` for better memory management and performance
- Updated all message-related methods to work with message pointers
- Improved message handling in Range functions to work with pointers

### Developer Experience
- Enhanced documentation with more usage examples
- Added new test cases for message handling functionality

For more details about the changes, please refer to the [GitHub release](https://github.com/presbrey/aichat/releases/tag/v1.1.0).
