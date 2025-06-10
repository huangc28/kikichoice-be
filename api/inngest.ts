import express from "express";
import { Inngest } from "inngest";
import { serve } from "inngest/express";
import { env } from "#shared/env.js";
import { syncInventory } from "./_inngest/sync-inventory/main";

const app = express();
const inngest = new Inngest({ id: env.INNGEST_APP_ID });

app.use(
  "/api/inngest",
  express.json(),
  serve({
    client: inngest,
    functions: [
      syncInventory(inngest),
    ],
  }),
);

export default app;
