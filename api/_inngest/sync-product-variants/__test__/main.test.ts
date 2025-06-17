import { describe, expect, it } from "vitest";
import { syncProductVariantsFunc } from "../main.js";

describe("syncProductVariantsFunc", () => {
  it.only("should sync product variants", async () => {
    // Mock the step object that Inngest provides
    const mockStep = {
      run: async (stepName: string, fn: any) => {
        console.log(`🧪 Running step: ${stepName}`);

        // Execute the function directly for testing
        if (typeof fn === "function") {
          return await fn();
        }
        return fn;
      },
    };
    // Call the function with the mocked step parameter
    const result = await syncProductVariantsFunc({ step: mockStep as any });

    console.log("🧪 Test result:", result);

    // Basic assertions
    expect(result).toBeDefined();

    if (result) {
      expect(result).toHaveProperty("inserted");
      expect(result).toHaveProperty("updated");
      expect(result).toHaveProperty("total");
      expect(result).toHaveProperty("skipped");

      expect(typeof result.inserted).toBe("number");
      expect(typeof result.updated).toBe("number");
      expect(typeof result.total).toBe("number");
      expect(typeof result.skipped).toBe("number");

      console.log(
        `✅ Sync completed: ${result.inserted} inserted, ${result.updated} updated, ${result.total} total, ${result.skipped} skipped`,
      );
    } else {
      console.log("📭 No product variants to sync");
    }
  });
});
