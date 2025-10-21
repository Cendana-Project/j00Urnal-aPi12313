# Migration

The `migration` directory contains database migration scripts that manage the evolution of your database schema over time.

## Subdirectories

- **db/**: Contains SQL migration files for database schema changes.

## Files

- **20240615140438_create_extension_uuid.sql**: Migration to create the UUID extension in the database.
- **20240615155416_create_table_users.sql**: Migration to create the users table.

## Purpose

The migration layer is responsible for:

1. Defining the database schema in a version-controlled manner
2. Providing a history of schema changes over time
3. Allowing for database upgrades and downgrades between versions
4. Ensuring database schema consistency across different environments

This layer uses migration files that are executed in order based on their timestamps. Each migration represents a discrete change to the database schema, allowing for controlled evolution of the database structure as the application grows and changes.