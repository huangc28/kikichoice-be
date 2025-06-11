import { describe, it } from "vitest";
import { upsertProducts } from "./upsert-products.ts";

describe("updateProducts", () => {
  it.skip("should update products", async () => {
    const products = await upsertProducts([
      {
        sku: "123",
        name: "123",
        ready_for_sale: true,
        stock_adjust_count: 1,
        price: 1,
        short_desc: "123",
      },
    ]);

    console.log("upsert result", products);
  });
});
