import { describe, expect, it } from "vitest";
import { syncFunc } from "./main.js";

describe("syncFunc - Integration Test", () => {
  it(
    "Should sync products from excel sheet. This is a integration test",
    async () => {
      console.log("🚀 Starting real integration test...");
      console.log("📋 This test will:");
      console.log("  - Fetch real data from Google Sheets");
      console.log("  - Update real database with upsert operations");
      console.log("  - No mocking involved");

      // Create a real step implementation that actually executes the functions
      const realStep = {
        run: async (stepName: string, fn: () => Promise<any>) => {
          console.log(`🔄 Executing step: ${stepName}`);
          const startTime = Date.now();

          try {
            const result = await fn();
            const duration = Date.now() - startTime;
            console.log(`✅ Step "${stepName}" completed in ${duration}ms`);
            return result;
          } catch (error) {
            console.error(`❌ Step "${stepName}" failed:`, error);
            throw error;
          }
        },
      };

      // Execute the real sync function
      try {
        const result = await syncFunc({ step: realStep });

        // Log the results
        console.log("🎉 Integration test completed successfully!");
        console.log("📊 Final result:", result);

        // If the result has statistics, log them
        if (result && typeof result === "object") {
          if (
            "inserted" in result && "updated" in result && "total" in result
          ) {
            console.log(`📈 Database operations:`);
            console.log(`  - Inserted: ${result.inserted} products`);
            console.log(`  - Updated: ${result.updated} products`);
            console.log(`  - Total: ${result.total} products`);

            // Assert that we processed some products
            expect(result.total).toBeGreaterThan(0);
          }
        }
      } catch (error) {
        console.error("💥 Integration test failed:");
        console.error("Error details:", error);
        console.error("Stack trace:", error.stack);

        // Re-throw to fail the test
        throw error;
      }
    },
    30000,
  ); // 30 second timeout for the integration test
});
