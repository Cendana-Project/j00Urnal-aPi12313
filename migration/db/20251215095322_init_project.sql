-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "users" (
  "id" uuid DEFAULT gen_random_uuid(),
  "username" varchar NOT NULL,
  "first_name" varchar,
  "last_name" varchar,
  "email" varchar NOT NULL,
  "password_hash" varchar NOT NULL,
  "affiliation" varchar(300),
  "phone" varchar,
  "last_login" timestamp,
  "status" varchar(16) NOT NULL DEFAULT 'pending',
  "verified_at" timestamptz,
  "updated_at" timestamp,
  "deleted_at" timestamp,
  "created_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS "idx_users_deleted_at" ON "users" ("deleted_at");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_username" ON "users" ("username");

CREATE TABLE IF NOT EXISTS "roles" (
  "id" uuid DEFAULT gen_random_uuid(),
  "name" varchar(64) NOT NULL,
  "slug" varchar(64) NOT NULL,
  "description" text,
  "active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz,
  "deleted_at" timestamptz,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_roles_name" UNIQUE ("name")
);

CREATE INDEX IF NOT EXISTS "idx_roles_deleted_at" ON "roles" ("deleted_at");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_roles_slug" ON "roles" ("slug");

CREATE TABLE IF NOT EXISTS "permissions" (
  "id" uuid DEFAULT gen_random_uuid(),
  "name" varchar(255) NOT NULL,
  "slug" varchar(255) NOT NULL,
  "description" text,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_permissions_slug" ON "permissions" ("slug");

CREATE TABLE IF NOT EXISTS "role_permissions" (
  "id" uuid DEFAULT gen_random_uuid(),
  "role_id" uuid NOT NULL,
  "permission_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS "idx_role_permissions_permission_id" ON "role_permissions" ("permission_id");
CREATE INDEX IF NOT EXISTS "idx_role_permissions_role_id" ON "role_permissions" ("role_id");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_rp_unique" ON "role_permissions" ("role_id", "permission_id");

CREATE TABLE IF NOT EXISTS "user_roles" (
  "user_id" uuid,
  "role_id" uuid,
  "assigned_at" timestamptz NOT NULL DEFAULT now(),
  "created_by" uuid,
  PRIMARY KEY ("user_id", "role_id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "user_roles";
DROP TABLE IF EXISTS "role_permissions";
DROP TABLE IF EXISTS "permissions";
DROP TABLE IF EXISTS "roles";
DROP TABLE IF EXISTS "users";
-- +goose StatementEnd
