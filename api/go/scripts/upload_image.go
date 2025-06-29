package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/azure"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/logger"

	"github.com/huangc28/kikichoice-be/api/go/_internal/configs"
	appfx "github.com/huangc28/kikichoice-be/api/go/_internal/fx"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Read images from a folder.
// Generate a nanoid for each image
// Upload the image to azure blob storage
// Create records in images table and image_entities table. You can relate the images by directory name since the directory name is the product sku.

// DirectoryType represents the type of directory (product or variant)
type DirectoryType int

const (
	ProductDirectory DirectoryType = iota
	VariantDirectory
)

// DirectoryInfo contains information about a directory and its images
type DirectoryInfo struct {
	Path      string
	SKU       string
	Type      DirectoryType
	ParentSKU string // Only for variants
	Images    []string
}

// VariantInfo contains database information about a product variant
type VariantInfo struct {
	ID        int64  `json:"id"`
	SKU       string `json:"sku"`
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
}

// LocalDBConfig contains local database connection information
type LocalDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Enabled  bool
}

// readLocalDBConfig reads local database configuration from environment variables
func readLocalDBConfig() *LocalDBConfig {
	config := &LocalDBConfig{
		Host:     os.Getenv("LOCAL_DB_HOST"),
		Port:     os.Getenv("LOCAL_DB_PORT"),
		User:     os.Getenv("LOCAL_DB_USER"),
		Password: os.Getenv("LOCAL_DB_PASSWORD"),
		Name:     os.Getenv("LOCAL_DB_NAME"),
		Enabled:  os.Getenv("LOCAL_DB_ENABLED") == "true",
	}

	// Set defaults for missing values
	if config.Port == "" {
		config.Port = "5432"
	}
	if config.User == "" {
		config.User = "postgres"
	}
	if config.Name == "" {
		config.Name = "postgres"
	}

	return config
}

// buildLocalDBURL constructs the PostgreSQL connection URL from config
func (cfg *LocalDBConfig) buildLocalDBURL() string {
	if !cfg.Enabled || cfg.Host == "" {
		return ""
	}

	// Build connection string
	dbURL := fmt.Sprintf("postgres://%s", cfg.User)
	if cfg.Password != "" {
		dbURL += ":" + cfg.Password
	}
	dbURL += fmt.Sprintf("@%s:%s/%s", cfg.Host, cfg.Port, cfg.Name)

	return dbURL
}

// validateLocalDBConfig checks if the local DB config is complete and valid
func (cfg *LocalDBConfig) validateLocalDBConfig() error {
	if !cfg.Enabled {
		return nil // Not enabled, no validation needed
	}

	if cfg.Host == "" {
		return fmt.Errorf("LOCAL_DB_HOST is required when LOCAL_DB_ENABLED=true")
	}

	return nil
}

type ImageUploader struct {
	cfg           *configs.Config
	prodDB        *sqlx.DB // Production database
	localDB       *sqlx.DB // Local database (optional)
	azureClient   *azure.BlobStorageWrapperClient
	logger        *zap.SugaredLogger
	dryRun        bool
	cleanFirst    bool
	syncLocal     bool // Enable local database sync
	localDBConfig *LocalDBConfig
}

type ImageUploaderParams struct {
	fx.In

	Cfg         *configs.Config
	ProdDB      *sqlx.DB
	LocalDB     *sqlx.DB `optional:"true"`
	AzureClient *azure.BlobStorageWrapperClient
	Logger      *zap.SugaredLogger
}

type UploadResult struct {
	ProcessedProducts     int
	ProcessedVariants     int
	UploadedImages        int
	UploadedProductImages int
	UploadedVariantImages int
	SkippedImages         int
	SkippedProductImages  int
	SkippedVariantImages  int
	Errors                []error
	TotalSizeBytes        int64
	DryRun                bool
	LocalSyncEnabled      bool
	LocalSyncErrors       int
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
		prodDB:      p.ProdDB,
		localDB:     p.LocalDB,
		azureClient: p.AzureClient,
		logger:      p.Logger,
		dryRun:      false, // Will be set by command line flag
		cleanFirst:  false, // Will be set by command line flag
		syncLocal:   false, // Will be set by command line flag
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

// SetSyncLocal enables or disables local database sync
func (u *ImageUploader) SetSyncLocal(syncLocal bool) {
	u.syncLocal = syncLocal
}

// Upload processes images from the specified directory
func (u *ImageUploader) Upload(sourcePath string) (*UploadResult, error) {
	u.logger.Infof("Starting image upload from path: %s (dry-run: %v)", sourcePath, u.dryRun)

	result := &UploadResult{
		Errors:           make([]error, 0),
		DryRun:           u.dryRun,
		LocalSyncEnabled: u.syncLocal,
		LocalSyncErrors:  0,
	}

	// Validate source path
	if err := u.validateSourcePath(sourcePath); err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}

	// Scan directory for images grouped by directory type and SKU
	directories, err := u.scanDirectory(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(directories) == 0 {
		u.logger.Warn("No directories found in source path")
		return result, nil
	}

	// Separate products and variants
	productDirs := make(map[string]DirectoryInfo)
	variantDirs := make(map[string]DirectoryInfo)

	for sku, dirInfo := range directories {
		if dirInfo.Type == ProductDirectory {
			productDirs[sku] = dirInfo
		} else {
			variantDirs[sku] = dirInfo
		}
	}

	u.logger.Infof("Found %d product directories and %d variant directories to process",
		len(productDirs), len(variantDirs))

	// Validate variant directories against database
	if len(variantDirs) > 0 {
		err = u.validateVariantDirectories(context.Background(), variantDirs)
		if err != nil {
			return nil, fmt.Errorf("variant validation failed: %w", err)
		}
	}

	// Process product directories
	for sku, dirInfo := range productDirs {
		err := u.processDirectoryImages(context.Background(), dirInfo, result)
		if err != nil {
			u.logger.Errorf("Failed to process product images for SKU %s: %v", sku, err)
			result.Errors = append(result.Errors, fmt.Errorf("product SKU %s: %w", sku, err))
			continue
		}
		result.ProcessedProducts++
	}

	// Process variant directories
	for sku, dirInfo := range variantDirs {
		err := u.processDirectoryImages(context.Background(), dirInfo, result)
		if err != nil {
			u.logger.Errorf("Failed to process variant images for SKU %s: %v", sku, err)
			result.Errors = append(result.Errors, fmt.Errorf("variant SKU %s: %w", sku, err))
			continue
		}
		result.ProcessedVariants++
	}

	status := "completed"
	if u.dryRun {
		status = "completed (DRY RUN - no changes made)"
	}

	u.logger.Infof("Upload %s. Products: %d, Variants: %d, Total Images: %d (Product: %d, Variant: %d), Skipped: %d, Errors: %d, Total size: %.2f MB",
		status, result.ProcessedProducts, result.ProcessedVariants, result.UploadedImages,
		result.UploadedProductImages, result.UploadedVariantImages,
		result.SkippedImages, len(result.Errors), float64(result.TotalSizeBytes)/(1024*1024))

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

// scanDirectory scans the source directory and classifies directories as product or variant
func (u *ImageUploader) scanDirectory(sourcePath string) (map[string]DirectoryInfo, error) {
	directories := make(map[string]DirectoryInfo)
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	// First pass: identify all directories and their images
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

		// Get the parent directory name as potential SKU
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

		// Initialize directory info if not exists
		if _, exists := directories[parentDir]; !exists {
			// Determine if this is a variant directory
			dirType, parentSKU := u.classifyDirectory(parentDir)

			directories[parentDir] = DirectoryInfo{
				Path:      filepath.Dir(path),
				SKU:       parentDir,
				Type:      dirType,
				ParentSKU: parentSKU,
				Images:    make([]string, 0),
			}
		}

		// Add image to the directory
		dirInfo := directories[parentDir]
		dirInfo.Images = append(dirInfo.Images, path)
		directories[parentDir] = dirInfo

		u.logger.Debugf("Found image: %s (SKU: %s, Type: %v, Size: %d bytes)",
			filepath.Base(path), parentDir, dirInfo.Type, info.Size())
		return nil
	})

	return directories, err
}

// classifyDirectory determines if a directory is for a product or variant
func (u *ImageUploader) classifyDirectory(dirName string) (DirectoryType, string) {
	// Check if directory name contains hyphens (potential variant)
	parts := strings.Split(dirName, "-")
	if len(parts) <= 1 {
		// No hyphens, must be a product directory
		return ProductDirectory, ""
	}

	// For variant directories, the parent SKU is everything except the last part
	// e.g., "kivy-007-dog" -> parent SKU is "kivy-007", variant suffix is "dog"
	parentSKU := strings.Join(parts[:len(parts)-1], "-")

	// If parent SKU is empty, treat as product directory
	if parentSKU == "" {
		return ProductDirectory, ""
	}

	// We'll validate this assumption later against the database
	return VariantDirectory, parentSKU
}

// validateVariantDirectories validates that variant SKUs exist in the database
func (u *ImageUploader) validateVariantDirectories(ctx context.Context, variantDirs map[string]DirectoryInfo) error {
	if len(variantDirs) == 0 {
		return nil
	}

	// Collect all variant SKUs to validate
	variantSKUs := make([]string, 0, len(variantDirs))
	for sku := range variantDirs {
		variantSKUs = append(variantSKUs, sku)
	}

	// Query database for these SKUs
	query := `SELECT sku, id, product_id, name FROM product_variants WHERE sku = ANY($1)`
	rows, err := u.prodDB.QueryxContext(ctx, query, variantSKUs)
	if err != nil {
		return fmt.Errorf("failed to query variant SKUs: %w", err)
	}
	defer rows.Close()

	validVariants := make(map[string]VariantInfo)
	for rows.Next() {
		var variant VariantInfo
		err := rows.StructScan(&variant)
		if err != nil {
			return fmt.Errorf("failed to scan variant row: %w", err)
		}
		validVariants[variant.SKU] = variant
	}

	// Check which variant directories are valid
	validCount := 0
	for sku, dirInfo := range variantDirs {
		if _, exists := validVariants[sku]; exists {
			validCount++
			u.logger.Debugf("Validated variant directory: %s", sku)
		} else {
			u.logger.Warnf("Variant SKU not found in database, treating as product: %s", sku)
			// Convert to product directory
			dirInfo.Type = ProductDirectory
			dirInfo.ParentSKU = ""
			variantDirs[sku] = dirInfo
		}
	}

	u.logger.Infof("Validated %d variant directories out of %d", validCount, len(variantDirs))
	return nil
}

// processDirectoryImages routes to appropriate processing method based on directory type
func (u *ImageUploader) processDirectoryImages(ctx context.Context, dirInfo DirectoryInfo, result *UploadResult) error {
	if dirInfo.Type == ProductDirectory {
		return u.processProductImages(dirInfo.SKU, dirInfo.Images, result)
	} else {
		return u.processVariantImages(ctx, dirInfo.SKU, dirInfo.Images, result)
	}
}

// cleanupExistingImages removes all existing images for a SKU from both Azure and database
func (u *ImageUploader) cleanupExistingImages(ctx context.Context, entityID int64, sku string, entityType string) error {
	blobPrefix := sku + "/"

	if u.dryRun {
		// In dry-run mode, just list what would be deleted
		blobNames, err := u.azureClient.ListBlobsWithPrefix(ctx, azure.ProductImageContainerName, blobPrefix)
		if err != nil {
			return fmt.Errorf("failed to list existing blobs for SKU %s: %w", sku, err)
		}

		if len(blobNames) > 0 {
			u.logger.Infof("[DRY RUN] Would delete %d existing images for %s %s:", len(blobNames), entityType, sku)
			for _, blobName := range blobNames {
				u.logger.Infof("[DRY RUN]   - %s", blobName)
			}
		} else {
			u.logger.Infof("[DRY RUN] No existing images found for %s %s", entityType, sku)
		}
		return nil
	}

	// Delete blobs from Azure
	deletedCount, err := u.azureClient.DeleteBlobsWithPrefix(ctx, azure.ProductImageContainerName, blobPrefix)
	if err != nil {
		return fmt.Errorf("failed to delete existing blobs for SKU %s: %w", sku, err)
	}

	if deletedCount > 0 {
		u.logger.Infof("Deleted %d existing images from Azure for %s %s", deletedCount, entityType, sku)

		// Clean up database records for this entity's images
		err = u.cleanupDatabaseRecords(ctx, entityID, entityType)
		if err != nil {
			u.logger.Warnf("Failed to cleanup database records for %s ID %d: %v", entityType, entityID, err)
			// Don't fail the entire process for database cleanup issues
		}
	} else {
		u.logger.Infof("No existing images found for %s %s", entityType, sku)
	}

	return nil
}

// cleanupDatabaseRecords removes image records from the database for an entity
func (u *ImageUploader) cleanupDatabaseRecords(ctx context.Context, entityID int64, entityType string) error {
	tx, err := u.prodDB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get image IDs that belong to this entity
	var imageIDs []int64
	query := `SELECT image_id FROM image_entities WHERE entity_id = $1 AND entity_type = $2`
	err = tx.SelectContext(ctx, &imageIDs, query, entityID, entityType)
	if err != nil {
		return fmt.Errorf("failed to get image IDs: %w", err)
	}

	if len(imageIDs) == 0 {
		return tx.Commit() // Nothing to clean up
	}

	// Delete from image_entities first (foreign key constraint)
	deleteEntitiesQuery := `DELETE FROM image_entities WHERE entity_id = $1 AND entity_type = $2`
	_, err = tx.ExecContext(ctx, deleteEntitiesQuery, entityID, entityType)
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

	u.logger.Debugf("Cleaned up %d database records for %s ID %d", len(imageIDs), entityType, entityID)
	return nil
}

// processProductImages processes all images for a single product SKU
func (u *ImageUploader) processProductImages(sku string, imagePaths []string, result *UploadResult) error {
	ctx := context.Background()

	// Check if product exists using raw SQL query
	var product struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	query := `SELECT id, name FROM products WHERE sku = $1 LIMIT 1`
	err := u.prodDB.GetContext(ctx, &product, query, sku)
	if err != nil {
		u.logger.Warnf("Product with SKU %s not found, skipping %d images", sku, len(imagePaths))
		result.SkippedImages += len(imagePaths)
		result.SkippedProductImages += len(imagePaths)
		return nil
	}

	u.logger.Infof("Processing %d images for product '%s' (SKU: %s, ID: %d)", len(imagePaths), product.Name, sku, product.ID)

	// Clean up existing images if requested
	if u.cleanFirst {
		err = u.cleanupExistingImages(ctx, product.ID, sku, "product")
		if err != nil {
			return fmt.Errorf("failed to cleanup existing images for product SKU %s: %w", sku, err)
		}
	}

	// Process each image
	for i, imagePath := range imagePaths {
		fileInfo, err := os.Stat(imagePath)
		if err != nil {
			u.logger.Errorf("Cannot access image file %s: %v", imagePath, err)
			result.Errors = append(result.Errors, fmt.Errorf("file access %s: %w", imagePath, err))
			result.SkippedImages++
			result.SkippedProductImages++
			continue
		}

		result.TotalSizeBytes += fileInfo.Size()

		if u.dryRun {
			u.logger.Infof("[DRY RUN] Would upload product image: %s (%.2f KB)", filepath.Base(imagePath), float64(fileInfo.Size())/1024)
			result.UploadedImages++
			result.UploadedProductImages++
			continue
		}

		err = u.processImage(ctx, product.ID, sku, imagePath, i, "product", result)
		if err != nil {
			u.logger.Errorf("Failed to process product image %s: %v", imagePath, err)
			result.Errors = append(result.Errors, fmt.Errorf("product image %s: %w", imagePath, err))
			result.SkippedImages++
			result.SkippedProductImages++
			continue
		}
		result.UploadedImages++
		result.UploadedProductImages++
	}

	return nil
}

// processVariantImages processes all images for a single product variant SKU
func (u *ImageUploader) processVariantImages(ctx context.Context, sku string, imagePaths []string, result *UploadResult) error {
	// Check if variant exists using raw SQL query
	var variant VariantInfo
	query := `SELECT id, product_id, name FROM product_variants WHERE sku = $1 LIMIT 1`
	err := u.prodDB.GetContext(ctx, &variant, query, sku)
	if err != nil {
		u.logger.Warnf("Product variant with SKU %s not found, skipping %d images", sku, len(imagePaths))
		result.SkippedImages += len(imagePaths)
		result.SkippedVariantImages += len(imagePaths)
		return nil
	}

	u.logger.Infof("Processing %d images for variant '%s' (SKU: %s, ID: %d)", len(imagePaths), variant.Name, sku, variant.ID)

	// Clean up existing images if requested
	if u.cleanFirst {
		err = u.cleanupExistingImages(ctx, variant.ID, sku, "product_variant")
		if err != nil {
			return fmt.Errorf("failed to cleanup existing images for variant SKU %s: %w", sku, err)
		}
	}

	// Process each image
	for i, imagePath := range imagePaths {
		fileInfo, err := os.Stat(imagePath)
		if err != nil {
			u.logger.Errorf("Cannot access image file %s: %v", imagePath, err)
			result.Errors = append(result.Errors, fmt.Errorf("file access %s: %w", imagePath, err))
			result.SkippedImages++
			result.SkippedVariantImages++
			continue
		}

		result.TotalSizeBytes += fileInfo.Size()

		if u.dryRun {
			u.logger.Infof("[DRY RUN] Would upload variant image: %s (%.2f KB)", filepath.Base(imagePath), float64(fileInfo.Size())/1024)
			result.UploadedImages++
			result.UploadedVariantImages++
			continue
		}

		err = u.processImage(ctx, variant.ID, sku, imagePath, i, "product_variant", result)
		if err != nil {
			u.logger.Errorf("Failed to process variant image %s: %v", imagePath, err)
			result.Errors = append(result.Errors, fmt.Errorf("variant image %s: %w", imagePath, err))
			result.SkippedImages++
			result.SkippedVariantImages++
			continue
		}
		result.UploadedImages++
		result.UploadedVariantImages++
	}

	return nil
}

// processImage processes a single image file and syncs to both databases
func (u *ImageUploader) processImage(ctx context.Context, entityID int64, sku, imagePath string, sortOrder int, entityType string, result *UploadResult) error {
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

	u.logger.Debugf("Uploading %s as %s (entity_type: %s)", filepath.Base(imagePath), blobName, entityType)

	// Upload to Azure Blob Storage (only once)
	publicURL, err := u.azureClient.UploadProductImage(ctx, blobName, file)
	if err != nil {
		return fmt.Errorf("failed to upload to Azure: %w", err)
	}

	// Create records in production database
	err = u.syncImageToDatabase(ctx, u.prodDB, sku, publicURL, sortOrder, entityType, "production")
	if err != nil {
		return fmt.Errorf("failed to sync to production database: %w", err)
	}

	// Create records in local database if enabled and available
	if u.syncLocal && u.localDB != nil {
		err = u.syncImageToDatabase(ctx, u.localDB, sku, publicURL, sortOrder, entityType, "local")
		if err != nil {
			// Log warning but don't fail the entire process for local DB issues
			u.logger.Warnf("Failed to sync image to local database for SKU %s: %v", sku, err)
			result.LocalSyncErrors++
		} else {
			u.logger.Debugf("Successfully synced image to local database for SKU %s", sku)
		}
	}

	u.logger.Debugf("Successfully processed %s image: %s -> %s", entityType, filepath.Base(imagePath), publicURL)
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

// createLocalDBConnection creates a connection to the local database
func (u *ImageUploader) createLocalDBConnection(dbURL string) (*sqlx.DB, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("local database URL is required")
	}

	// Validate the URL format
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}

	if parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql" {
		return nil, fmt.Errorf("unsupported database scheme: %s (only postgres/postgresql supported)", parsedURL.Scheme)
	}

	// Create database connection
	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to local database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping local database: %w", err)
	}

	u.logger.Infof("Successfully connected to local database: %s", parsedURL.Host)
	return db, nil
}

// resolveProductID finds product ID by SKU in the specified database
func (u *ImageUploader) resolveProductID(ctx context.Context, sku string, db *sqlx.DB) (int64, error) {
	var productID int64
	query := `SELECT id FROM products WHERE sku = $1 LIMIT 1`
	err := db.GetContext(ctx, &productID, query, sku)
	if err != nil {
		return 0, fmt.Errorf("product with SKU %s not found: %w", sku, err)
	}
	return productID, nil
}

// resolveVariantID finds variant ID by SKU in the specified database
func (u *ImageUploader) resolveVariantID(ctx context.Context, sku string, db *sqlx.DB) (int64, error) {
	var variantID int64
	query := `SELECT id FROM product_variants WHERE sku = $1 LIMIT 1`
	err := db.GetContext(ctx, &variantID, query, sku)
	if err != nil {
		return 0, fmt.Errorf("variant with SKU %s not found: %w", sku, err)
	}
	return variantID, nil
}

// createImageRecordsForEntity creates image records in the specified database for an entity
func (u *ImageUploader) createImageRecordsForEntity(ctx context.Context, db *sqlx.DB, entityID int64, publicURL, sku string, sortOrder int, entityType string) error {
	// Start database transaction for two-table insert
	tx, err := db.BeginTxx(ctx, nil)
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
	altText := fmt.Sprintf("%s image for %s", strings.Title(strings.Replace(entityType, "_", " ", -1)), sku)
	isPrimary := sortOrder == 0 // First image is primary

	entityInsertQuery := `
		INSERT INTO image_entities (entity_id, image_id, alt_text, is_primary, sort_order, entity_type)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.ExecContext(ctx, entityInsertQuery, entityID, imageID, altText, isPrimary, sortOrder, entityType)
	if err != nil {
		return fmt.Errorf("failed to insert into image_entities table: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// syncImageToDatabase creates image records in the specified database
func (u *ImageUploader) syncImageToDatabase(ctx context.Context, db *sqlx.DB, sku, publicURL string, sortOrder int, entityType string, dbName string) error {
	// Resolve entity ID for this specific database
	var entityID int64
	var err error

	if entityType == "product" {
		entityID, err = u.resolveProductID(ctx, sku, db)
	} else {
		entityID, err = u.resolveVariantID(ctx, sku, db)
	}

	if err != nil {
		return fmt.Errorf("failed to resolve entity ID in %s database: %w", dbName, err)
	}

	// Create image records
	err = u.createImageRecordsForEntity(ctx, db, entityID, publicURL, sku, sortOrder, entityType)
	if err != nil {
		return fmt.Errorf("failed to create image records in %s database: %w", dbName, err)
	}

	u.logger.Debugf("Successfully synced image to %s database: %s (entity_id: %d)", dbName, publicURL, entityID)
	return nil
}

// getPostgresqlDSN builds the PostgreSQL connection string from config
func getPostgresqlDSN(cfg *configs.Config) string {
	pgdsn := "postgres://%s:%s@%s:%s/%s"
	params := ""

	if cfg.ENV == configs.Production {
		params = "?sslmode=require&pool_mode=transaction"
	}

	return fmt.Sprintf(pgdsn,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	) + params
}

// Command line flags
var (
	dryRunFlag    = flag.Bool("dry-run", false, "Run in dry-run mode without making actual changes")
	cleanFlag     = flag.Bool("clean-first", false, "Remove all existing images for each SKU before uploading new ones")
	syncLocalFlag = flag.Bool("sync-local", false, "Enable local database sync (requires LOCAL_DB_* environment variables)")
	helpFlag      = flag.Bool("help", false, "Show help information")
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
		fmt.Println("Directory Structure:")
		fmt.Println("  product-sku/          # Product images")
		fmt.Println("    image1.jpg")
		fmt.Println("    image2.png")
		fmt.Println("  product-sku-variant/  # Variant images (must have matching parent SKU)")
		fmt.Println("    variant1.jpg")
		fmt.Println("    variant2.png")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run scripts/upload_image.go ./images")
		fmt.Println("  go run scripts/upload_image.go --dry-run /path/to/images")
		fmt.Println("  go run scripts/upload_image.go --clean-first ./images")
		fmt.Println("  LOCAL_DB_ENABLED=true LOCAL_DB_HOST=localhost LOCAL_DB_PORT=55322 go run scripts/upload_image.go --sync-local ./images")
		fmt.Println("  LOCAL_DB_ENABLED=true LOCAL_DB_HOST=localhost LOCAL_DB_PORT=55322 go run scripts/upload_image.go --clean-first --dry-run --sync-local ./images")
		fmt.Println()
		fmt.Println("Environment Variables for Local Database Sync:")
		fmt.Println("  LOCAL_DB_ENABLED=true    # Enable local database sync")
		fmt.Println("  LOCAL_DB_HOST=localhost  # Local database host")
		fmt.Println("  LOCAL_DB_PORT=55322      # Local database port (default: 5432)")
		fmt.Println("  LOCAL_DB_USER=postgres   # Local database user (default: postgres)")
		fmt.Println("  LOCAL_DB_PASSWORD=pass   # Local database password (optional)")
		fmt.Println("  LOCAL_DB_NAME=postgres   # Local database name (default: postgres)")
		fmt.Println()
		fmt.Println("See scripts/README.md for detailed documentation")
		return
	}

	// Set dry-run mode, clean-first mode, and sync-local mode
	uploader.SetDryRun(*dryRunFlag)
	uploader.SetCleanFirst(*cleanFlag)
	uploader.SetSyncLocal(*syncLocalFlag)

	// Setup local database connection if sync is enabled
	if *syncLocalFlag {
		// Read local database configuration from environment variables
		localDBConfig := readLocalDBConfig()

		// Validate configuration
		if err := localDBConfig.validateLocalDBConfig(); err != nil {
			log.Fatalf("Local database configuration error: %v", err)
		}

		// Check if local DB is actually enabled
		if !localDBConfig.Enabled {
			log.Fatalf("LOCAL_DB_ENABLED must be set to 'true' when --sync-local flag is used")
		}

		// Build connection URL and connect
		localDBURL := localDBConfig.buildLocalDBURL()
		localDB, err := uploader.createLocalDBConnection(localDBURL)
		if err != nil {
			log.Fatalf("Failed to connect to local database: %v", err)
		}
		uploader.localDB = localDB
		uploader.localDBConfig = localDBConfig

		// Ensure cleanup
		defer func() {
			if localDB != nil {
				localDB.Close()
				uploader.logger.Debugf("Closed local database connection")
			}
		}()
	}

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
	fmt.Printf("Processed Variants: %d\n", result.ProcessedVariants)
	fmt.Printf("Total Uploaded Images: %d\n", result.UploadedImages)
	fmt.Printf("  - Product Images: %d\n", result.UploadedProductImages)
	fmt.Printf("  - Variant Images: %d\n", result.UploadedVariantImages)
	fmt.Printf("Total Skipped Images: %d\n", result.SkippedImages)
	fmt.Printf("  - Product Images: %d\n", result.SkippedProductImages)
	fmt.Printf("  - Variant Images: %d\n", result.SkippedVariantImages)
	fmt.Printf("Total Size: %.2f MB\n", float64(result.TotalSizeBytes)/(1024*1024))
	fmt.Printf("Errors: %d\n", len(result.Errors))

	if result.LocalSyncEnabled {
		fmt.Printf("Local Database Sync: ✅ Enabled")
		if result.LocalSyncErrors > 0 {
			fmt.Printf(" (⚠️  %d sync errors)", result.LocalSyncErrors)
		}
		fmt.Printf("\n")
	}

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
			// Provide production DB with name annotation
			fx.Annotate(
				func(cfg *configs.Config, lc fx.Lifecycle) *sqlx.DB {
					// Use the existing NewSQLXPostgresDB logic for production DB
					dsn := getPostgresqlDSN(cfg)

					pgxConfig, err := pgx.ParseConfig(dsn)
					if err != nil {
						panic(fmt.Errorf("failed to parse pgx config: %w", err))
					}

					pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
					sqlDB := stdlib.OpenDB(*pgxConfig)
					driver := sqlx.NewDb(sqlDB, "pgx")
					driver.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)

					lc.Append(fx.Hook{
						OnStop: func(ctx context.Context) error { return driver.Close() },
					})

					return driver
				},
				fx.ResultTags(`name:"prod"`),
			),
			NewImageUploader,
		),
		fx.Invoke(Upload),
	)
}
