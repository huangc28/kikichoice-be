package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/azure"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/logger"

	"github.com/huangc28/kikichoice-be/api/go/_internal/configs"
	appfx "github.com/huangc28/kikichoice-be/api/go/_internal/fx"

	"github.com/jmoiron/sqlx"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Read images from a folder.
// Generate a nanoid for each image
// Upload the image to azure blob storage
// Create records in images table and image_entities table. You can relate the images by directory name since the directory name is the product sku.

type ImageUploader struct {
	cfg         *configs.Config
	db          *sqlx.DB
	azureClient *azure.BlobStorageWrapperClient
	logger      *zap.SugaredLogger
	dryRun      bool
	cleanFirst  bool
}

type ImageUploaderParams struct {
	fx.In

	Cfg         *configs.Config
	DB          *sqlx.DB
	AzureClient *azure.BlobStorageWrapperClient
	Logger      *zap.SugaredLogger
}

type UploadResult struct {
	ProcessedProducts int
	UploadedImages    int
	SkippedImages     int
	Errors            []error
	TotalSizeBytes    int64
	DryRun            bool
}

type ImageFile struct {
	SKU          string
	OriginalPath string
	NewFileName  string
	ContentType  string
	Size         int64
}

func NewImageUploader(p ImageUploaderParams) *ImageUploader {
	return &ImageUploader{
		cfg:         p.Cfg,
		db:          p.DB,
		azureClient: p.AzureClient,
		logger:      p.Logger,
		dryRun:      false, // Will be set by command line flag
		cleanFirst:  false, // Will be set by command line flag
	}
}

// SetDryRun enables or disables dry-run mode
func (u *ImageUploader) SetDryRun(dryRun bool) {
	u.dryRun = dryRun
}

// SetCleanFirst enables or disables cleanup mode
func (u *ImageUploader) SetCleanFirst(cleanFirst bool) {
	u.cleanFirst = cleanFirst
}

// Upload processes images from the specified directory
func (u *ImageUploader) Upload(sourcePath string) (*UploadResult, error) {
	u.logger.Infof("Starting image upload from path: %s (dry-run: %v)", sourcePath, u.dryRun)

	result := &UploadResult{
		Errors: make([]error, 0),
		DryRun: u.dryRun,
	}

	// Validate source path
	if err := u.validateSourcePath(sourcePath); err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}

	// Scan directory for images grouped by SKU
	imageGroups, err := u.scanDirectory(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(imageGroups) == 0 {
		u.logger.Warn("No product directories found in source path")
		return result, nil
	}

	u.logger.Infof("Found %d product directories to process", len(imageGroups))

	// Process each SKU group
	for sku, images := range imageGroups {
		err := u.processProductImages(sku, images, result)
		if err != nil {
			u.logger.Errorf("Failed to process images for SKU %s: %v", sku, err)
			result.Errors = append(result.Errors, fmt.Errorf("SKU %s: %w", sku, err))
			continue
		}
		result.ProcessedProducts++
	}

	status := "completed"
	if u.dryRun {
		status = "completed (DRY RUN - no changes made)"
	}

	u.logger.Infof("Upload %s. Processed: %d products, Uploaded: %d images, Skipped: %d images, Errors: %d, Total size: %.2f MB",
		status, result.ProcessedProducts, result.UploadedImages, result.SkippedImages, len(result.Errors), float64(result.TotalSizeBytes)/(1024*1024))

	return result, nil
}

// validateSourcePath validates that the source path exists and is readable
func (u *ImageUploader) validateSourcePath(sourcePath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", sourcePath)
		}
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", sourcePath)
	}

	// Test if we can read the directory
	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("cannot read directory: %w", err)
	}

	u.logger.Debugf("Source directory validated: %s (%d entries)", sourcePath, len(entries))
	return nil
}

// scanDirectory scans the source directory and groups images by SKU (directory name)
func (u *ImageUploader) scanDirectory(sourcePath string) (map[string][]string, error) {
	imageGroups := make(map[string][]string)
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			u.logger.Warnf("Error accessing path %s: %v", path, err)
			return nil // Continue processing other files
		}

		// Skip if it's a directory
		if info.IsDir() {
			return nil
		}

		// Check if it's a supported image file
		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			u.logger.Debugf("Skipping unsupported file: %s", path)
			return nil
		}

		// Get the parent directory name as SKU
		parentDir := filepath.Base(filepath.Dir(path))

		// Skip if the parent directory is the source directory itself
		if filepath.Dir(path) == sourcePath {
			u.logger.Warnf("Skipping image in root directory: %s", path)
			return nil
		}

		// Validate file size
		if info.Size() == 0 {
			u.logger.Warnf("Skipping empty file: %s", path)
			return nil
		}

		// Add to the appropriate SKU group
		if imageGroups[parentDir] == nil {
			imageGroups[parentDir] = make([]string, 0)
		}
		imageGroups[parentDir] = append(imageGroups[parentDir], path)

		u.logger.Debugf("Found image: %s (SKU: %s, Size: %d bytes)", filepath.Base(path), parentDir, info.Size())
		return nil
	})

	return imageGroups, err
}

// cleanupExistingImages removes all existing images for a SKU from both Azure and database
func (u *ImageUploader) cleanupExistingImages(ctx context.Context, productID int64, sku string) error {
	blobPrefix := sku + "/"

	if u.dryRun {
		// In dry-run mode, just list what would be deleted
		blobNames, err := u.azureClient.ListBlobsWithPrefix(ctx, azure.ProductImageContainerName, blobPrefix)
		if err != nil {
			return fmt.Errorf("failed to list existing blobs for SKU %s: %w", sku, err)
		}

		if len(blobNames) > 0 {
			u.logger.Infof("[DRY RUN] Would delete %d existing images for SKU %s:", len(blobNames), sku)
			for _, blobName := range blobNames {
				u.logger.Infof("[DRY RUN]   - %s", blobName)
			}
		} else {
			u.logger.Infof("[DRY RUN] No existing images found for SKU %s", sku)
		}
		return nil
	}

	// Delete blobs from Azure
	deletedCount, err := u.azureClient.DeleteBlobsWithPrefix(ctx, azure.ProductImageContainerName, blobPrefix)
	if err != nil {
		return fmt.Errorf("failed to delete existing blobs for SKU %s: %w", sku, err)
	}

	if deletedCount > 0 {
		u.logger.Infof("Deleted %d existing images from Azure for SKU %s", deletedCount, sku)

		// Clean up database records for this product's images
		err = u.cleanupDatabaseRecords(ctx, productID)
		if err != nil {
			u.logger.Warnf("Failed to cleanup database records for product ID %d: %v", productID, err)
			// Don't fail the entire process for database cleanup issues
		}
	} else {
		u.logger.Infof("No existing images found for SKU %s", sku)
	}

	return nil
}

// cleanupDatabaseRecords removes image records from the database for a product
func (u *ImageUploader) cleanupDatabaseRecords(ctx context.Context, productID int64) error {
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get image IDs that belong to this product
	var imageIDs []int64
	query := `SELECT image_id FROM image_entities WHERE entity_id = $1 AND entity_type = 'product'`
	err = tx.SelectContext(ctx, &imageIDs, query, productID)
	if err != nil {
		return fmt.Errorf("failed to get image IDs: %w", err)
	}

	if len(imageIDs) == 0 {
		return tx.Commit() // Nothing to clean up
	}

	// Delete from image_entities first (foreign key constraint)
	deleteEntitiesQuery := `DELETE FROM image_entities WHERE entity_id = $1 AND entity_type = 'product'`
	_, err = tx.ExecContext(ctx, deleteEntitiesQuery, productID)
	if err != nil {
		return fmt.Errorf("failed to delete image entities: %w", err)
	}

	// Delete from images table
	for _, imageID := range imageIDs {
		deleteImageQuery := `DELETE FROM images WHERE id = $1`
		_, err = tx.ExecContext(ctx, deleteImageQuery, imageID)
		if err != nil {
			return fmt.Errorf("failed to delete image %d: %w", imageID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit cleanup transaction: %w", err)
	}

	u.logger.Debugf("Cleaned up %d database records for product ID %d", len(imageIDs), productID)
	return nil
}

// processProductImages processes all images for a single product SKU
func (u *ImageUploader) processProductImages(sku string, imagePaths []string, result *UploadResult) error {
	ctx := context.Background()

	// Check if product exists using raw SQL query
	var product struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}

	query := `SELECT id, name FROM products WHERE sku = $1 LIMIT 1`
	err := u.db.GetContext(ctx, &product, query, sku)
	if err != nil {
		u.logger.Warnf("Product with SKU %s not found, skipping %d images", sku, len(imagePaths))
		result.SkippedImages += len(imagePaths)
		return nil
	}

	u.logger.Infof("Processing %d images for product '%s' (SKU: %s, ID: %d)", len(imagePaths), product.Name, sku, product.ID)

	// Clean up existing images if requested
	if u.cleanFirst {
		err = u.cleanupExistingImages(ctx, product.ID, sku)
		if err != nil {
			return fmt.Errorf("failed to cleanup existing images for SKU %s: %w", sku, err)
		}
	}

	// Process each image
	for i, imagePath := range imagePaths {
		fileInfo, err := os.Stat(imagePath)
		if err != nil {
			u.logger.Errorf("Cannot access image file %s: %v", imagePath, err)
			result.Errors = append(result.Errors, fmt.Errorf("file access %s: %w", imagePath, err))
			result.SkippedImages++
			continue
		}

		result.TotalSizeBytes += fileInfo.Size()

		if u.dryRun {
			u.logger.Infof("[DRY RUN] Would upload: %s (%.2f KB)", filepath.Base(imagePath), float64(fileInfo.Size())/1024)
			result.UploadedImages++
			continue
		}

		err = u.processImage(ctx, product.ID, sku, imagePath, i)
		if err != nil {
			u.logger.Errorf("Failed to process image %s: %v", imagePath, err)
			result.Errors = append(result.Errors, fmt.Errorf("image %s: %w", imagePath, err))
			result.SkippedImages++
			continue
		}
		result.UploadedImages++
	}

	return nil
}

// processImage processes a single image file
func (u *ImageUploader) processImage(ctx context.Context, productID int64, sku, imagePath string, sortOrder int) error {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Generate a new filename with nano ID
	ext := filepath.Ext(imagePath)
	newFileName := u.generateNanoID() + ext
	blobName := fmt.Sprintf("%s/%s", sku, newFileName)

	u.logger.Debugf("Uploading %s as %s", filepath.Base(imagePath), blobName)

	// Upload to Azure Blob Storage
	publicURL, err := u.azureClient.UploadProductImage(ctx, blobName, file)
	if err != nil {
		return fmt.Errorf("failed to upload to Azure: %w", err)
	}

	// Start database transaction for two-table insert
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Step 1: Insert into images table and get the image ID
	var imageID int64
	imageInsertQuery := `INSERT INTO images (url) VALUES ($1) RETURNING id`
	err = tx.GetContext(ctx, &imageID, imageInsertQuery, publicURL)
	if err != nil {
		return fmt.Errorf("failed to insert into images table: %w", err)
	}

	// Step 2: Insert into image_entities table
	altText := fmt.Sprintf("Product image for %s", sku)
	isPrimary := sortOrder == 0 // First image is primary
	entityType := "product"

	entityInsertQuery := `
		INSERT INTO image_entities (entity_id, image_id, alt_text, is_primary, sort_order, entity_type)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.ExecContext(ctx, entityInsertQuery, productID, imageID, altText, isPrimary, sortOrder, entityType)
	if err != nil {
		return fmt.Errorf("failed to insert into image_entities table: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	u.logger.Debugf("Successfully processed image: %s -> %s (image_id: %d)", filepath.Base(imagePath), publicURL, imageID)
	return nil
}

// generateNanoID generates a short unique ID for filenames
func (u *ImageUploader) generateNanoID() string {
	// Generate a 12-character nanoid (good balance of uniqueness and brevity)
	id, err := gonanoid.New(12)
	if err != nil {
		// Fallback to timestamp-based approach if nanoid fails
		u.logger.Warnf("Failed to generate nanoid, using timestamp fallback: %v", err)
		return fmt.Sprintf("%d", time.Now().UnixNano()%1000000000)
	}
	return id
}

// Command line flags
var (
	dryRunFlag = flag.Bool("dry-run", false, "Run in dry-run mode without making actual changes")
	cleanFlag  = flag.Bool("clean-first", false, "Remove all existing images for each SKU before uploading new ones")
	helpFlag   = flag.Bool("help", false, "Show help information")
)

// Upload function to be called by FX
func Upload(uploader *ImageUploader) {
	flag.Parse()

	if *helpFlag {
		fmt.Println("Product Image Upload Tool")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [flags] [path_to_images_directory]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run scripts/upload_image.go ./images")
		fmt.Println("  go run scripts/upload_image.go --dry-run /path/to/images")
		fmt.Println("  go run scripts/upload_image.go --clean-first ./images")
		fmt.Println("  go run scripts/upload_image.go --clean-first --dry-run ./images")
		fmt.Println()
		fmt.Println("See scripts/README.md for detailed documentation")
		return
	}

	// Set dry-run mode and clean-first mode
	uploader.SetDryRun(*dryRunFlag)
	uploader.SetCleanFirst(*cleanFlag)

	// Get source path from command line args or use default
	sourcePath := "./images" // Default path
	args := flag.Args()
	if len(args) > 0 {
		sourcePath = args[0]
	}

	// Check if source directory exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		log.Fatalf("Source directory does not exist: %s", sourcePath)
	}

	result, err := uploader.Upload(sourcePath)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	// Print summary
	fmt.Printf("\n=== Upload Summary ===\n")
	if result.DryRun {
		fmt.Printf("🔍 DRY RUN MODE - No actual changes were made\n")
	}
	fmt.Printf("Processed Products: %d\n", result.ProcessedProducts)
	fmt.Printf("Uploaded Images: %d\n", result.UploadedImages)
	fmt.Printf("Skipped Images: %d\n", result.SkippedImages)
	fmt.Printf("Total Size: %.2f MB\n", float64(result.TotalSizeBytes)/(1024*1024))
	fmt.Printf("Errors: %d\n", len(result.Errors))

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors encountered:\n")
		for _, err := range result.Errors {
			fmt.Printf("- %v\n", err)
		}
	}

	if result.DryRun {
		fmt.Printf("\n💡 Run without --dry-run flag to perform actual upload\n")
	}
}

func main() {
	fx.New(
		logger.TagLogger("image-uploader"),
		appfx.CoreConfigOptions,
		fx.Provide(
			azure.NewSharedKeyCredential,
			azure.NewBlobStorageClient,
			azure.NewBlobStorageWrapperClient,
			NewImageUploader,
		),
		fx.Invoke(Upload),
	)
}
