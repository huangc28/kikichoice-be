

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


COMMENT ON SCHEMA "public" IS 'standard public schema';



CREATE EXTENSION IF NOT EXISTS "pg_graphql" WITH SCHEMA "graphql";






CREATE EXTENSION IF NOT EXISTS "pg_stat_statements" WITH SCHEMA "extensions";






CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA "extensions";






CREATE EXTENSION IF NOT EXISTS "supabase_vault" WITH SCHEMA "vault";






CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA "extensions";






CREATE TYPE "public"."address_kind" AS ENUM (
    'shipping',
    'billing'
);


ALTER TYPE "public"."address_kind" OWNER TO "postgres";


CREATE TYPE "public"."auth_provider" AS ENUM (
    'clerk'
);


ALTER TYPE "public"."auth_provider" OWNER TO "postgres";


CREATE TYPE "public"."entity_type" AS ENUM (
    'product',
    'product_variant'
);


ALTER TYPE "public"."entity_type" OWNER TO "postgres";


CREATE TYPE "public"."order_status" AS ENUM (
    'pending_payment',
    'paid',
    'processing',
    'shipped',
    'delivered',
    'canceled',
    'refunded'
);


ALTER TYPE "public"."order_status" OWNER TO "postgres";


CREATE TYPE "public"."payment_status" AS ENUM (
    'initiated',
    'authorized',
    'captured',
    'failed',
    'refunded'
);


ALTER TYPE "public"."payment_status" OWNER TO "postgres";


CREATE TYPE "public"."shipping_status" AS ENUM (
    'pending_label',
    'label_purchased',
    'in_transit',
    'out_for_delivery',
    'delivered',
    'delivery_failed',
    'returned'
);


ALTER TYPE "public"."shipping_status" OWNER TO "postgres";


CREATE TYPE "public"."status_actor" AS ENUM (
    'system',
    'customer',
    'staff',
    'webhook'
);


ALTER TYPE "public"."status_actor" OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."update_user_sessions_updated_at"() RETURNS "trigger"
    LANGUAGE "plpgsql"
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION "public"."update_user_sessions_updated_at"() OWNER TO "postgres";

SET default_tablespace = '';

SET default_table_access_method = "heap";


CREATE TABLE IF NOT EXISTS "public"."addresses" (
    "id" bigint NOT NULL,
    "kind" "public"."address_kind" NOT NULL,
    "name" "text" NOT NULL,
    "phone" "text",
    "address_line_1" "text" NOT NULL,
    "address_line_2" "text",
    "city" "text" NOT NULL,
    "state_province" "text",
    "postal_code" "text" NOT NULL,
    "country" "text" DEFAULT 'TW'::"text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"()
);


ALTER TABLE "public"."addresses" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."addresses_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."addresses_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."addresses_id_seq" OWNED BY "public"."addresses"."id";



CREATE TABLE IF NOT EXISTS "public"."image_entities" (
    "id" bigint NOT NULL,
    "entity_id" bigint NOT NULL,
    "alt_text" character varying(255),
    "is_primary" boolean DEFAULT false NOT NULL,
    "sort_order" integer DEFAULT 0,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "image_id" bigint NOT NULL,
    "entity_type" "public"."entity_type" DEFAULT 'product'::"public"."entity_type" NOT NULL
);


ALTER TABLE "public"."image_entities" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."image_entities_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."image_entities_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."image_entities_id_seq" OWNED BY "public"."image_entities"."id";



CREATE TABLE IF NOT EXISTS "public"."images" (
    "id" bigint NOT NULL,
    "url" "text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."images" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."images_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."images_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."images_id_seq" OWNED BY "public"."images"."id";



CREATE TABLE IF NOT EXISTS "public"."order_items" (
    "id" bigint NOT NULL,
    "order_id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "variant_id" bigint,
    "name" "text" NOT NULL,
    "image_url" "text",
    "unit_price" numeric(10,2) NOT NULL,
    "quantity" integer NOT NULL,
    "line_total" numeric(10,2) GENERATED ALWAYS AS (("unit_price" * ("quantity")::numeric)) STORED,
    "metadata" "jsonb",
    CONSTRAINT "order_items_quantity_check" CHECK (("quantity" > 0))
);


ALTER TABLE "public"."order_items" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."order_items_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."order_items_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."order_items_id_seq" OWNED BY "public"."order_items"."id";



CREATE TABLE IF NOT EXISTS "public"."orders" (
    "id" bigint NOT NULL,
    "order_number" bigint NOT NULL,
    "status" "public"."order_status" DEFAULT 'pending_payment'::"public"."order_status" NOT NULL,
    "currency" "text" DEFAULT 'TWD'::"text" NOT NULL,
    "subtotal" numeric(12,2) NOT NULL,
    "discount_total" numeric(12,2) DEFAULT 0 NOT NULL,
    "shipping_total" numeric(12,2) DEFAULT 0 NOT NULL,
    "tax_total" numeric(12,2) DEFAULT 0 NOT NULL,
    "grand_total" numeric(12,2) GENERATED ALWAYS AS (((("subtotal" - "discount_total") + "shipping_total") + "tax_total")) STORED,
    "email" "text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"(),
    CONSTRAINT "orders_email_check" CHECK (("email" ~* '^[^@]+@[^@]+\.[^@]+$'::"text"))
);


ALTER TABLE "public"."orders" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."orders_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."orders_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."orders_id_seq" OWNED BY "public"."orders"."id";



CREATE SEQUENCE IF NOT EXISTS "public"."orders_order_number_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."orders_order_number_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."orders_order_number_seq" OWNED BY "public"."orders"."order_number";



CREATE TABLE IF NOT EXISTS "public"."payments" (
    "id" bigint NOT NULL,
    "order_id" bigint NOT NULL,
    "provider" "text" NOT NULL,
    "provider_txn_id" "text",
    "status" "public"."payment_status" DEFAULT 'initiated'::"public"."payment_status" NOT NULL,
    "currency" "text" NOT NULL,
    "amount_authorized" numeric(12,2),
    "amount_captured" numeric(12,2),
    "captured_at" timestamp with time zone,
    "raw_response" "jsonb",
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"()
);


ALTER TABLE "public"."payments" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."payments_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."payments_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."payments_id_seq" OWNED BY "public"."payments"."id";



CREATE TABLE IF NOT EXISTS "public"."product_specs" (
    "id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "spec_name" character varying(100) NOT NULL,
    "spec_value" character varying(255) NOT NULL,
    "sort_order" integer DEFAULT 0,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."product_specs" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."product_specs_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."product_specs_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."product_specs_id_seq" OWNED BY "public"."product_specs"."id";



CREATE TABLE IF NOT EXISTS "public"."product_variants" (
    "id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "name" character varying(255) NOT NULL,
    "stock_count" integer DEFAULT 0 NOT NULL,
    "reserved_count" integer DEFAULT 0 NOT NULL,
    "sku" character varying(100) NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "price" numeric(10,2) DEFAULT 0 NOT NULL,
    "uuid" "text",
    CONSTRAINT "product_variant_reserved_count_check" CHECK (("reserved_count" >= 0)),
    CONSTRAINT "product_variant_stock_count_check" CHECK (("stock_count" >= 0)),
    CONSTRAINT "product_variants_price_check" CHECK (("price" >= (0)::numeric))
);


ALTER TABLE "public"."product_variants" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."product_variant_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."product_variant_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."product_variant_id_seq" OWNED BY "public"."product_variants"."id";



CREATE TABLE IF NOT EXISTS "public"."products" (
    "id" bigint NOT NULL,
    "uuid" character varying DEFAULT "gen_random_uuid"() NOT NULL,
    "sku" character varying(100) NOT NULL,
    "name" character varying(255) NOT NULL,
    "price" numeric(10,2) NOT NULL,
    "original_price" numeric(10,2),
    "category" character varying(100),
    "stock_count" integer DEFAULT 0 NOT NULL,
    "specs" "jsonb" DEFAULT '[]'::"jsonb",
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "ready_for_sale" boolean DEFAULT false NOT NULL,
    "full_desc" "text",
    "reserved_count" integer DEFAULT 0 NOT NULL,
    "short_desc" "text",
    "slug" "text",
    CONSTRAINT "products_original_price_check" CHECK (("original_price" >= (0)::numeric)),
    CONSTRAINT "products_price_check" CHECK (("price" >= (0)::numeric)),
    CONSTRAINT "products_reserved_count_check" CHECK (("reserved_count" >= 0)),
    CONSTRAINT "products_stock_count_check" CHECK (("stock_count" >= 0))
);


ALTER TABLE "public"."products" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."products_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."products_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."products_id_seq" OWNED BY "public"."products"."id";



CREATE TABLE IF NOT EXISTS "public"."shipments" (
    "id" bigint NOT NULL,
    "order_id" bigint NOT NULL,
    "address_id" bigint NOT NULL,
    "status" "public"."shipping_status" DEFAULT 'pending_label'::"public"."shipping_status" NOT NULL,
    "carrier" "text" NOT NULL,
    "service_level" "text" NOT NULL,
    "tracking_number" "text",
    "tracking_url" "text",
    "shipping_cost" numeric(12,2) NOT NULL,
    "insurance_cost" numeric(12,2) DEFAULT 0,
    "weight_kg" numeric(8,3) DEFAULT 0,
    "dimensions_cm" "jsonb",
    "eta" timestamp with time zone,
    "shipped_at" timestamp with time zone,
    "delivered_at" timestamp with time zone,
    "carrier_payload" "jsonb",
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"()
);


ALTER TABLE "public"."shipments" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."shipments_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."shipments_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."shipments_id_seq" OWNED BY "public"."shipments"."id";



CREATE TABLE IF NOT EXISTS "public"."user_sessions" (
    "id" bigint NOT NULL,
    "chat_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "session_type" character varying(50) DEFAULT 'add_product'::character varying NOT NULL,
    "state" "jsonb" DEFAULT '{}'::"jsonb" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "expires_at" timestamp with time zone DEFAULT ("now"() + '24:00:00'::interval) NOT NULL,
    "expected_reply_message_id" bigint
);


ALTER TABLE "public"."user_sessions" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."user_sessions_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "public"."user_sessions_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."user_sessions_id_seq" OWNED BY "public"."user_sessions"."id";



CREATE TABLE IF NOT EXISTS "public"."users" (
    "id" bigint NOT NULL,
    "name" "text" NOT NULL,
    "email" "text",
    "created_at" timestamp with time zone DEFAULT "now"(),
    "updated_at" timestamp with time zone DEFAULT "now"(),
    "deleted_at" timestamp with time zone,
    "auth_provider" "public"."auth_provider" DEFAULT 'clerk'::"public"."auth_provider",
    "auth_provider_id" "text"
);


ALTER TABLE "public"."users" OWNER TO "postgres";


ALTER TABLE "public"."users" ALTER COLUMN "id" ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME "public"."users_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);



ALTER TABLE ONLY "public"."addresses" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."addresses_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."image_entities" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."image_entities_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."images" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."images_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."order_items" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."order_items_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."orders" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."orders_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."orders" ALTER COLUMN "order_number" SET DEFAULT "nextval"('"public"."orders_order_number_seq"'::"regclass");



ALTER TABLE ONLY "public"."payments" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."payments_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."product_specs" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."product_specs_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."product_variants" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."product_variant_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."products" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."products_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."shipments" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."shipments_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."user_sessions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."user_sessions_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."addresses"
    ADD CONSTRAINT "addresses_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."images"
    ADD CONSTRAINT "images_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."order_items"
    ADD CONSTRAINT "order_items_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."orders"
    ADD CONSTRAINT "orders_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."payments"
    ADD CONSTRAINT "payments_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."image_entities"
    ADD CONSTRAINT "product_images_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."product_specs"
    ADD CONSTRAINT "product_specs_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."product_variants"
    ADD CONSTRAINT "product_variant_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."product_variants"
    ADD CONSTRAINT "product_variant_sku_key" UNIQUE ("sku");



ALTER TABLE ONLY "public"."products"
    ADD CONSTRAINT "products_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."products"
    ADD CONSTRAINT "products_sku_key" UNIQUE ("sku");



ALTER TABLE ONLY "public"."products"
    ADD CONSTRAINT "products_uuid_key" UNIQUE ("uuid");



ALTER TABLE ONLY "public"."shipments"
    ADD CONSTRAINT "shipments_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."shipments"
    ADD CONSTRAINT "shipments_tracking_number_key" UNIQUE ("tracking_number");



ALTER TABLE ONLY "public"."product_specs"
    ADD CONSTRAINT "unique_product_spec" UNIQUE ("product_id", "spec_name");



ALTER TABLE ONLY "public"."user_sessions"
    ADD CONSTRAINT "user_sessions_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."user_sessions"
    ADD CONSTRAINT "user_sessions_user_id_session_type_key" UNIQUE ("user_id", "session_type");



ALTER TABLE ONLY "public"."users"
    ADD CONSTRAINT "users_email_key" UNIQUE ("email");



ALTER TABLE ONLY "public"."users"
    ADD CONSTRAINT "users_pkey" PRIMARY KEY ("id");



CREATE INDEX "addresses_kind_idx" ON "public"."addresses" USING "btree" ("kind");



CREATE INDEX "idx_image_entities_entity_id" ON "public"."image_entities" USING "btree" ("entity_id");



CREATE INDEX "idx_image_entities_entity_type" ON "public"."image_entities" USING "btree" ("entity_type");



CREATE INDEX "idx_image_entities_image_id" ON "public"."image_entities" USING "btree" ("image_id");



CREATE INDEX "idx_image_entities_is_primary" ON "public"."image_entities" USING "btree" ("is_primary");



CREATE INDEX "idx_images_created_at" ON "public"."images" USING "btree" ("created_at");



CREATE INDEX "idx_images_url" ON "public"."images" USING "btree" ("url");



CREATE INDEX "idx_product_specs_name" ON "public"."product_specs" USING "btree" ("spec_name");



CREATE INDEX "idx_product_specs_product_id" ON "public"."product_specs" USING "btree" ("product_id");



CREATE INDEX "idx_product_variants_name" ON "public"."product_variants" USING "btree" ("name");



CREATE INDEX "idx_product_variants_price" ON "public"."product_variants" USING "btree" ("price");



CREATE INDEX "idx_product_variants_product_id" ON "public"."product_variants" USING "btree" ("product_id");



CREATE INDEX "idx_product_variants_sku" ON "public"."product_variants" USING "btree" ("sku");



CREATE INDEX "idx_products_category" ON "public"."products" USING "btree" ("category");



CREATE INDEX "idx_products_created_at" ON "public"."products" USING "btree" ("created_at");



CREATE INDEX "idx_products_sku" ON "public"."products" USING "btree" ("sku");



CREATE INDEX "idx_products_uuid" ON "public"."products" USING "btree" ("uuid");



CREATE INDEX "idx_user_sessions_expires_at" ON "public"."user_sessions" USING "btree" ("expires_at");



CREATE INDEX "idx_user_sessions_user_id" ON "public"."user_sessions" USING "btree" ("user_id");



CREATE INDEX "order_items_order_idx" ON "public"."order_items" USING "btree" ("order_id");



CREATE INDEX "order_items_product_idx" ON "public"."order_items" USING "btree" ("product_id");



CREATE INDEX "order_items_variant_idx" ON "public"."order_items" USING "btree" ("variant_id");



CREATE INDEX "orders_created_idx" ON "public"."orders" USING "btree" ("created_at" DESC);



CREATE INDEX "payments_order_idx" ON "public"."payments" USING "btree" ("order_id");



CREATE INDEX "shipments_order_idx" ON "public"."shipments" USING "btree" ("order_id");



CREATE OR REPLACE TRIGGER "update_user_sessions_updated_at" BEFORE UPDATE ON "public"."user_sessions" FOR EACH ROW EXECUTE FUNCTION "public"."update_user_sessions_updated_at"();



ALTER TABLE ONLY "public"."image_entities"
    ADD CONSTRAINT "image_entities_image_id_fkey" FOREIGN KEY ("image_id") REFERENCES "public"."images"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."order_items"
    ADD CONSTRAINT "order_items_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "public"."orders"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."order_items"
    ADD CONSTRAINT "order_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products"("id");



ALTER TABLE ONLY "public"."order_items"
    ADD CONSTRAINT "order_items_variant_id_fkey" FOREIGN KEY ("variant_id") REFERENCES "public"."product_variants"("id");



ALTER TABLE ONLY "public"."payments"
    ADD CONSTRAINT "payments_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "public"."orders"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."product_specs"
    ADD CONSTRAINT "product_specs_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."product_variants"
    ADD CONSTRAINT "product_variant_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."shipments"
    ADD CONSTRAINT "shipments_address_id_fkey" FOREIGN KEY ("address_id") REFERENCES "public"."addresses"("id");



ALTER TABLE ONLY "public"."shipments"
    ADD CONSTRAINT "shipments_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "public"."orders"("id") ON DELETE CASCADE;





ALTER PUBLICATION "supabase_realtime" OWNER TO "postgres";


GRANT USAGE ON SCHEMA "public" TO "postgres";
GRANT USAGE ON SCHEMA "public" TO "anon";
GRANT USAGE ON SCHEMA "public" TO "authenticated";
GRANT USAGE ON SCHEMA "public" TO "service_role";

























































































































































GRANT ALL ON FUNCTION "public"."update_user_sessions_updated_at"() TO "anon";
GRANT ALL ON FUNCTION "public"."update_user_sessions_updated_at"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."update_user_sessions_updated_at"() TO "service_role";


















GRANT ALL ON TABLE "public"."addresses" TO "anon";
GRANT ALL ON TABLE "public"."addresses" TO "authenticated";
GRANT ALL ON TABLE "public"."addresses" TO "service_role";



GRANT ALL ON SEQUENCE "public"."addresses_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."addresses_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."addresses_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."image_entities" TO "anon";
GRANT ALL ON TABLE "public"."image_entities" TO "authenticated";
GRANT ALL ON TABLE "public"."image_entities" TO "service_role";



GRANT ALL ON SEQUENCE "public"."image_entities_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."image_entities_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."image_entities_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."images" TO "anon";
GRANT ALL ON TABLE "public"."images" TO "authenticated";
GRANT ALL ON TABLE "public"."images" TO "service_role";



GRANT ALL ON SEQUENCE "public"."images_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."images_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."images_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."order_items" TO "anon";
GRANT ALL ON TABLE "public"."order_items" TO "authenticated";
GRANT ALL ON TABLE "public"."order_items" TO "service_role";



GRANT ALL ON SEQUENCE "public"."order_items_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."order_items_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."order_items_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."orders" TO "anon";
GRANT ALL ON TABLE "public"."orders" TO "authenticated";
GRANT ALL ON TABLE "public"."orders" TO "service_role";



GRANT ALL ON SEQUENCE "public"."orders_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."orders_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."orders_id_seq" TO "service_role";



GRANT ALL ON SEQUENCE "public"."orders_order_number_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."orders_order_number_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."orders_order_number_seq" TO "service_role";



GRANT ALL ON TABLE "public"."payments" TO "anon";
GRANT ALL ON TABLE "public"."payments" TO "authenticated";
GRANT ALL ON TABLE "public"."payments" TO "service_role";



GRANT ALL ON SEQUENCE "public"."payments_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."payments_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."payments_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."product_specs" TO "anon";
GRANT ALL ON TABLE "public"."product_specs" TO "authenticated";
GRANT ALL ON TABLE "public"."product_specs" TO "service_role";



GRANT ALL ON SEQUENCE "public"."product_specs_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."product_specs_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."product_specs_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."product_variants" TO "anon";
GRANT ALL ON TABLE "public"."product_variants" TO "authenticated";
GRANT ALL ON TABLE "public"."product_variants" TO "service_role";



GRANT ALL ON SEQUENCE "public"."product_variant_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."product_variant_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."product_variant_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."products" TO "anon";
GRANT ALL ON TABLE "public"."products" TO "authenticated";
GRANT ALL ON TABLE "public"."products" TO "service_role";



GRANT ALL ON SEQUENCE "public"."products_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."products_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."products_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."shipments" TO "anon";
GRANT ALL ON TABLE "public"."shipments" TO "authenticated";
GRANT ALL ON TABLE "public"."shipments" TO "service_role";



GRANT ALL ON SEQUENCE "public"."shipments_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."shipments_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."shipments_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."user_sessions" TO "anon";
GRANT ALL ON TABLE "public"."user_sessions" TO "authenticated";
GRANT ALL ON TABLE "public"."user_sessions" TO "service_role";



GRANT ALL ON SEQUENCE "public"."user_sessions_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."user_sessions_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."user_sessions_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."users" TO "anon";
GRANT ALL ON TABLE "public"."users" TO "authenticated";
GRANT ALL ON TABLE "public"."users" TO "service_role";



GRANT ALL ON SEQUENCE "public"."users_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."users_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."users_id_seq" TO "service_role";









ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES  TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES  TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES  TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES  TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS  TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS  TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS  TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS  TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES  TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES  TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES  TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES  TO "service_role";






























RESET ALL;
