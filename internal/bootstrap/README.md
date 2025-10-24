# Bootstrap

This directory contains initialization code for the application. It serves as the entry point for all application functionalities.

## Files

- **common.go**: Common initialization utilities
- **migrate.bootstrap.go**: Database migration setup
- **server.bootstrap.go**: HTTP server initialization

## Purpose

The bootstrap layer:

1. Initializes components in the correct order
2. Loads configuration
3. Connects to external dependencies (database, cache)
4. Wires up application dependencies
5. Starts services (HTTP server, workers)