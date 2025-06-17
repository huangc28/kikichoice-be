import { syncProductVariantsSheet } from "./sync-product-variants-sheet.js";
import type { ProcessedProductVariant } from "./upsert-product-variants.js";

// Simple test to verify the sheet sync functionality
const testSheetSync = async () => {
  console.log("🧪 Testing syncProductVariantsSheet function...");

  // Sample test data
  const testData: ProcessedProductVariant[] = [
    {
      sku: "TEST-001",
      stock_count: 50,
      price: 19.99,
    },
    {
      sku: "TEST-002",
      stock_count: 25,
      price: 29.99,
    },
  ];

  try {
    await syncProductVariantsSheet(testData);
    console.log("✅ Sheet sync test completed successfully!");
  } catch (error) {
    console.error("❌ Sheet sync test failed:", error);
  }
};

// Run the test if this file is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  testSheetSync();
}
