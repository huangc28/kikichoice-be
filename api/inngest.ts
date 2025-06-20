import express from "express";
import { Inngest } from "inngest";
import { serve } from "inngest/express";

import { env } from "#shared/env.js";
import { syncProducts } from "./_inngest/sync-products/main.js";
import { syncProductVariants } from "./_inngest/sync-product-variants/main.js";

const app = express();
const inngest = new Inngest({ id: env.INNGEST_APP_ID });

app.use(
  "/api/inngest",
  express.json(),
  serve({
    client: inngest,
    functions: [
      syncProducts(inngest),
      syncProductVariants(inngest),
    ],
  }),
);

export default app;
