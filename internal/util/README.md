# Util

The `util` directory contains utility functions and helpers that are used throughout the application. These utilities provide common functionality that doesn't fit into any specific layer.

## Files

- **common.go**: Contains general utility functions used across the application.
- **error.go**: Provides error handling utilities and custom error types.
- **log.go**: Contains logging utilities and helpers.
- **string.go**: Provides string manipulation and processing utilities.
- **transaction.go**: Contains utilities for database transaction management.
- **validation.go**: Provides data validation utilities.

## Purpose

The util layer is responsible for:

1. Providing reusable helper functions that are used across multiple layers
2. Implementing cross-cutting concerns like logging and error handling
3. Offering utility functions for common operations (string manipulation, validation, etc.)
4. Abstracting complex operations into simple, reusable functions

This layer helps reduce code duplication by centralizing common functionality. It contains stateless helper functions that can be used by any other layer in the application.