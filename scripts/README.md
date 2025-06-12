# Product Image Upload Tool

This tool uploads product images from a local directory structure to Azure Blob Storage and creates corresponding database records.

## Features

- Scans directory structure where folder names represent product SKUs
- Uploads images to Azure Blob Storage maintaining directory hierarchy
- Generates nanoid-based filenames to avoid conflicts
- Creates database records linking images to products
- Supports batch processing with detailed logging
- **Dry-run mode** for testing without making changes
- Graceful error handling with comprehensive reporting
- File validation and size reporting
- Command-line flags for better control

## Directory Structure

The tool expects a directory structure like this:

```
images/
├── PRODUCT-SKU-001/
│   ├── image1.jpg
│   ├── image2.png
│   └── image3.webp
├── PRODUCT-SKU-002/
│   ├── front.jpg
│   └── back.jpg
└── PRODUCT-SKU-003/
    └── main.jpeg
```

Where:
- Each subdirectory name is a product SKU that exists in the `products` table
- Supported image formats: `.jpg`, `.jpeg`, `.png`, `.webp`
- The first image in each directory becomes the primary image (`is_primary = true`)
- Empty files are automatically skipped
- Non-image files are ignored

## Usage

### Basic Usage

```bash
go run scripts/upload_image.go [flags] [path_to_images_directory]
```

### Command Line Flags

- `--dry-run`: Run without making actual changes (test mode)
- `--help`: Show help information

### Examples

```bash
# Use default directory (./images)
go run scripts/upload_image.go

# Specify custom directory
go run scripts/upload_image.go /path/to/product/images

# Test run without making changes
go run scripts/upload_image.go --dry-run ./images

# Show help
go run scripts/upload_image.go --help

# Combine flags and path
go run scripts/upload_image.go --dry-run /Users/username/Downloads/product-images
```

### Using Makefile

```bash
# Quick upload with Makefile
make upload/images PATH=./my-images

# Show help for Makefile usage
make upload/images/help
```

## Environment Variables

Make sure you have the following environment variables set:

```bash
# Database configuration
export DB_HOST="your-db-host"
export DB_PORT="5432"
export DB_USER="your-db-user"
export DB_PASSWORD="your-db-password"
export DB_NAME="your-db-name"

# Azure Blob Storage configuration
export AZURE_BLOB_STORAGE_ACCOUNT_NAME="your-storage-account"
export AZURE_BLOB_STORAGE_KEY="your-storage-key"
```

## Dry-Run Mode

Use the `--dry-run` flag to test the tool without making actual changes:

```bash
go run scripts/upload_image.go --dry-run ./images
```

In dry-run mode, the tool will:
- ✅ Scan and validate directory structure
- ✅ Check product existence in database
- ✅ Validate image files
- ✅ Calculate total file sizes
- ❌ **NOT** upload images to Azure
- ❌ **NOT** create database records

This is perfect for:
- Testing directory structure
- Validating product SKUs
- Checking file sizes before upload
- Identifying potential issues

## Azure Blob Storage Structure

Images are uploaded to the `products` container with the following structure:

```
products/
├── PRODUCT-SKU-001/
│   ├── Xy9z8w7v6u5t.jpg
│   ├── Ab1c2d3e4f5g.png
│   └── Yz8x7w6v5u4t.webp
└── PRODUCT-SKU-002/
    ├── Mn6o7p8q9r0s.jpg
    └── Uv4w5x6y7z8a.jpg
```

Where each filename is a 12-character nanoid to ensure uniqueness.

## Database Records

For each uploaded image, a record is created in the `product_images` table with:

- `product_id`: Found by matching the directory name (SKU) to `products.sku`
- `url`: Public URL of the uploaded image
- `alt_text`: Generated description like "Product image for SKU-001"
- `is_primary`: `true` for the first image in each directory, `false` for others
- `sort_order`: Order based on file processing sequence (0, 1, 2, ...)

## Error Handling

The tool handles various error scenarios gracefully:

- **Missing products**: If a SKU doesn't exist in the database, images are skipped with a warning
- **Upload failures**: Individual image upload failures don't stop the entire process
- **Invalid files**: Non-image files and empty files are ignored
- **Permission errors**: File access issues are logged and reported
- **Directory validation**: Source path is validated before processing

## Output

The tool provides detailed logging during execution and a comprehensive summary report:

### Regular Mode
```
=== Upload Summary ===
Processed Products: 5
Uploaded Images: 23
Skipped Images: 2
Total Size: 15.67 MB
Errors: 1

Errors encountered:
- SKU UNKNOWN-PRODUCT: Product with SKU UNKNOWN-PRODUCT not found, skipping images
```

### Dry-Run Mode
```
=== Upload Summary ===
🔍 DRY RUN MODE - No actual changes were made
Processed Products: 5
Uploaded Images: 23
Skipped Images: 2
Total Size: 15.67 MB
Errors: 0

💡 Run without --dry-run flag to perform actual upload
```

## Dependencies

- Go 1.24+
- Azure SDK for Go
- PostgreSQL database with existing schema
- Valid Azure Blob Storage account
- nanoid library for unique ID generation

## Logging

The tool uses structured logging with different levels:
- `INFO`: General progress information
- `WARN`: Non-fatal issues like missing products
- `ERROR`: Serious errors that prevent processing
- `DEBUG`: Detailed operation information

Set log level via environment: `export LOG_LEVEL=debug`

## Tips

1. **Always test first**: Use `--dry-run` to validate your setup before actual upload
2. **Check SKUs**: Ensure all directory names match existing product SKUs in your database
3. **File sizes**: Monitor the total size output to estimate upload time
4. **Error logs**: Review any errors in the summary to fix issues
5. **Incremental uploads**: The tool is idempotent-safe for re-running on the same directory

## Troubleshooting

### Common Issues

1. **"directory does not exist"**: Check the path you provided exists
2. **"Product with SKU X not found"**: Verify the SKU exists in your products table
3. **Azure upload failures**: Check your Azure credentials and network connectivity
4. **Database connection errors**: Verify your database environment variables

### Getting Help

```bash
# Show command help
go run scripts/upload_image.go --help

# Show Makefile help
make upload/images/help

# Check this documentation
cat scripts/README.md
```