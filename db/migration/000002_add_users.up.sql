CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");
-- CREATE UNIQUE INDEX ON "transfers" ("from_account_id", "to_account_id"); same as below
ALTER TABLE "accounts" ADD CONSTRAINT "owner_currency_key" UNIQUE ("owner","currency")
