import { describe, expect, it } from "vitest";
import { fetchSheetData } from "./fetch-sheet-data.ts";

// Simple test to call real function and see output
describe("fetchSheetData - Real Function Call", () => {
  it("should call real function and output result", async () => {
    console.log("🚀 Calling fetchSheetData with real Google Sheets API...");

    try {
      const result = await fetchSheetData();

      console.log("✅ SUCCESS! Function executed without errors");
      console.log("📊 RESULT:");
      console.log(`- Total products fetched: ${result.length}`);
      console.log(
        `- Result type: ${Array.isArray(result) ? "Array" : typeof result}`,
      );

      if (result.length > 0) {
        console.log("- First 3 products:");
        result.slice(0, 3).forEach((product, index) => {
          console.log(`  ${index + 1}. ${JSON.stringify(product)}`);
        });

        console.log("- All product names:");
        const names = result.map((p) => p.name).slice(0, 10);
        console.log(`  ${names.join(", ")}${result.length > 10 ? "..." : ""}`);
      } else {
        console.log("- No products found (empty result)");
      }

      // Basic assertions
      expect(Array.isArray(result)).toBe(true);
    } catch (error) {
      console.log("❌ ERROR occurred:");
      console.log(`- Error type: ${error.constructor.name}`);
      console.log(`- Error message: ${error.message}`);
      console.log(`- Full error:`, error);

      // Re-throw to fail the test
      throw error;
    }
  });
});
