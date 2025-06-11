CREATE UNIQUE INDEX products_sku_key ON public.products USING btree (sku);

alter table "public"."products" add constraint "products_sku_key" UNIQUE using index "products_sku_key";


