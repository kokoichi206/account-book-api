CREATE TABLE "users" (
	"id" bigserial PRIMARY KEY,
	"name" varchar NOT NULL,
	"password" varchar NOT NULL,
	"email" varchar UNIQUE NOT NULL,
	"age" int NOT NULL,
	"balance" bigint NOT NULL,
	"password_changed_at" timestamptz NOT NULL DEFAULT (now()),
	"created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions" (
	"id" uuid PRIMARY KEY,
	"user_id" bigint NOT NULL,
	"user_agent" varchar NOT NULL,
	"client_ip" varchar NOT NULL,
	"created_at" timestamptz NOT NULL DEFAULT (now()),
	"expires_at" timestamptz NOT NULL
);

CREATE TABLE "expenses" (
	"id" bigserial PRIMARY KEY,
	"user_id" bigint NOT NULL,
	"category_id" bigint NOT NULL,
	"amount" bigint NOT NULL,
	"food_receipt_id" bigint,
	"comment" varchar,
	"created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "categories" (
	"id" bigserial PRIMARY KEY,
	"name" varchar UNIQUE NOT NULL
);

CREATE TABLE "food_receipts" (
	"id" bigserial PRIMARY KEY,
	"store_name" varchar NOT NULL
);

CREATE TABLE "food_contents" (
	"id" bigserial PRIMARY KEY,
	"name" varchar NOT NULL,
	"calories" float4 NOT NULL,
	"lipid" float4 NOT NULL,
	"carbohydrate" float4 NOT NULL,
	"protein" float4 NOT NULL
);

CREATE TABLE "food_receipt_contents" (
	"id" bigserial PRIMARY KEY,
	"food_receipt_id" bigint NOT NULL,
	"food_content_id" bigint NOT NULL,
	"amount" bigint NOT NULL
);

CREATE TABLE "transfers" (
	"id" bigserial PRIMARY KEY,
	"from_user_id" bigint NOT NULL,
	"to_user_id" bigint NOT NULL,
	"amount" bigint NOT NULL,
	"created_at" timestamptz NOT NULL DEFAULT (now())
);

COMMENT ON COLUMN "expenses"."amount" IS 'can be negative or positive';
COMMENT ON COLUMN "food_receipt_contents"."amount" IS 'must be positive';
COMMENT ON COLUMN "transfers"."amount" IS 'must be positive';

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "expenses" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "expenses" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");
ALTER TABLE "expenses" ADD FOREIGN KEY ("food_receipt_id") REFERENCES "food_receipts" ("id");

ALTER TABLE "food_receipt_contents" ADD FOREIGN KEY ("food_receipt_id") REFERENCES "food_receipts" ("id");
ALTER TABLE "food_receipt_contents" ADD FOREIGN KEY ("food_content_id") REFERENCES "food_contents" ("id");

ALTER TABLE "transfers" ADD FOREIGN KEY ("from_user_id") REFERENCES "users" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("to_user_id") REFERENCES "users" ("id");
