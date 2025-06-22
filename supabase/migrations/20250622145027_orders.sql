-- Order lifecycle
create type order_status as enum
  ('pending_payment','paid','processing','shipped','delivered','canceled','refunded');

-- Payment lifecycle
create type payment_status as enum
  ('initiated','authorized','captured','failed','refunded');

-- Shipment lifecycle
create type shipping_status as enum
  ('pending_label','label_purchased','in_transit',
   'out_for_delivery','delivered','delivery_failed','returned');

-- Address role for a given order
create type address_kind as enum ('shipping','billing');

-- What changed an order status (for audit trail)
create type status_actor as enum ('system','customer','staff','webhook');

-- Add addresses table (missing dependency)
create table addresses (
  id                bigserial primary key,
  kind              address_kind not null,
  name              text not null,
  phone             text,
  address_line_1    text not null,
  address_line_2    text,
  city              text not null,
  state_province    text,
  postal_code       text not null,
  country           text not null default 'TW',
  created_at        timestamptz default now(),
  updated_at        timestamptz default now()
);

create table orders (
  id                  bigserial primary key,
  order_number        bigserial,                       -- short, customer-visible
  status              order_status not null default 'pending_payment',
  currency            text        not null default 'TWD',
  subtotal            numeric(12,2) not null,
  discount_total      numeric(12,2) not null default 0,
  shipping_total      numeric(12,2) not null default 0,
  tax_total           numeric(12,2) not null default 0,
  grand_total         numeric(12,2) generated always as
                       (subtotal - discount_total + shipping_total + tax_total) stored,
  email               text not null check (email ~* '^[^@]+@[^@]+\.[^@]+$'),
  created_at          timestamptz default now(),
  updated_at          timestamptz default now()
);

-- Fast listing for dashboards
create index orders_created_idx on orders (created_at desc);

create table order_items (
  id             bigserial primary key,
  order_id       bigint not null references orders(id) on delete cascade,
  product_id     bigint not null references products(id),
  variant_id     bigint references product_variants(id),
  name           text not null,                        -- Restore name field for order history
  image_url      text,                                 -- Restore image_url for order history
  unit_price     numeric(10,2) not null,              -- Match product price precision
  quantity       int not null check (quantity > 0),
  line_total     numeric(10,2) generated always as (unit_price * quantity) stored,
  metadata       jsonb                                -- size, color, etc.
);

create index order_items_order_idx on order_items(order_id);
create index order_items_product_idx on order_items(product_id);
create index order_items_variant_idx on order_items(variant_id);

create table payments (
  id                bigserial primary key,
  order_id          bigint not null references orders(id) on delete cascade,
  provider          text not null,              -- 'unipay','stripe',…
  provider_txn_id   text,                       -- returned by gateway
  status            payment_status not null default 'initiated',
  currency          text not null,
  amount_authorized numeric(12,2),
  amount_captured   numeric(12,2),
  captured_at       timestamptz,
  raw_response      jsonb,                      -- full webhook payload
  created_at        timestamptz default now(),
  updated_at        timestamptz default now()
);

create index payments_order_idx on payments(order_id);

create table shipments (
  id                   bigserial primary key,
  order_id             bigint not null references orders(id) on delete cascade,
  address_id           bigint not null references addresses(id),
  status               shipping_status not null default 'pending_label',
  carrier              text not null,                -- 'BlackCat','郵局',…
  service_level        text not null,                -- '宅配','超取','ECONOMY_INTL'
  tracking_number      text unique,
  tracking_url         text,
  shipping_cost        numeric(12,2) not null,
  insurance_cost       numeric(12,2) default 0,
  weight_kg            numeric(8,3) default 0,
  dimensions_cm        jsonb,                        -- {"l":30,"w":20,"h":15}
  eta                  timestamptz,
  shipped_at           timestamptz,
  delivered_at         timestamptz,
  carrier_payload      jsonb,
  created_at           timestamptz default now(),
  updated_at           timestamptz default now()
);

create index shipments_order_idx on shipments(order_id);
create index addresses_kind_idx on addresses(kind);
