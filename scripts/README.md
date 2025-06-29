# Image Upload Scripts

This directory contains scripts for uploading product images to Azure Blob Storage and managing image records in the database.

## Upload Image Script

The `upload_image.go` script processes images from a local directory, uploads them to Azure Blob Storage, and creates corresponding database records.

### Features

- **Dual Database Sync**: Upload images to Azure once and sync metadata to both production and local databases
- **Directory Structure Recognition**: Automatically classifies directories as product or variant based on naming
- **Batch Processing**: Process multiple products and variants in a single run
- **Dry-Run Mode**: Preview changes without making actual modifications
- **Clean-First Option**: Remove existing images before uploading new ones
- **Comprehensive Error Handling**: Continue processing on individual failures
- **Progress Reporting**: Detailed statistics and error reporting

### Usage

#### Basic Usage

```bash
# Upload images to production database only
go run scripts/upload_image.go ./images

# Upload with dry-run (no actual changes)
go run scripts/upload_image.go --dry-run ./images

# Clean existing images first
go run scripts/upload_image.go --clean-first ./images
```

#### Local Database Sync

```bash
# Upload images and sync to local database
LOCAL_DB_ENABLED=true \
LOCAL_DB_HOST=localhost \
LOCAL_DB_PORT=55322 \
LOCAL_DB_USER=postgres \
LOCAL_DB_PASSWORD=postgres \
LOCAL_DB_NAME=postgres \
go run scripts/upload_image.go --sync-local ./images

# Dry-run with local sync preview
LOCAL_DB_ENABLED=true \
LOCAL_DB_HOST=localhost \
LOCAL_DB_PORT=55322 \
go run scripts/upload_image.go --dry-run --sync-local ./images

# Clean and sync to both databases
LOCAL_DB_ENABLED=true \
LOCAL_DB_HOST=localhost \
LOCAL_DB_PORT=55322 \
go run scripts/upload_image.go --clean-first --sync-local ./images
```

#### Using Make Commands

```bash
# Upload images from current directory
make upload/images

# Upload images from custom path
make upload/images PATH=./my-images

# Show help for Makefile usage
make upload/images/help
```

## Command Line Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--dry-run` | Preview mode without making changes | `--dry-run` |
| `--clean-first` | Remove existing images before upload | `--clean-first` |
| `--sync-local` | Enable local database synchronization (requires LOCAL_DB_* env vars) | `--sync-local` |
| `--help` | Show help information | `--help` |

## Local Database Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `LOCAL_DB_ENABLED` | Enable local database sync | `false` | Yes (must be `true`) |
| `LOCAL_DB_HOST` | Local database host | - | Yes |
| `LOCAL_DB_PORT` | Local database port | `5432` | No |
| `LOCAL_DB_USER` | Local database user | `postgres` | No |
| `LOCAL_DB_PASSWORD` | Local database password | - | No |
| `LOCAL_DB_NAME` | Local database name | `postgres` | No |

## Environment Variables

Make sure you have the following environment variables set:

```bash
# Database configuration (Production)
export DB_HOST="your-db-host"
export DB_PORT="5432"
export DB_USER="your-db-user"
export DB_PASSWORD="your-db-password"
export DB_NAME="your-db-name"

# Azure Blob Storage configuration
export AZURE_BLOB_STORAGE_ACCOUNT_NAME="your-storage-account"
export AZURE_BLOB_STORAGE_KEY="your-storage-key"
```

**Note**: Local database connection is configured via LOCAL_DB_* environment variables when using --sync-local flag.

## Directory Structure

The script expects images to be organized in directories named after their SKUs:

```
images/
├── PRODUCT-SKU-001/          # Product images
│   ├── image1.jpg
│   ├── image2.png
│   └── image3.webp
├── PRODUCT-SKU-002/          # Another product
│   └── main.jpg
└── PRODUCT-SKU-001-variant/  # Variant images (parent: PRODUCT-SKU-001)
    ├── variant1.jpg
    └── variant2.png
```

### Supported Image Formats
- `.jpg`, `.jpeg`
- `.png`
- `.webp`

## Dry-Run Mode

Use the `--dry-run` flag to test the tool without making actual changes:

```bash
go run scripts/upload_image.go --dry-run ./images
```

In dry-run mode, the tool will:
- ✅ Scan and validate directory structure
- ✅ Check product existence in databases
- ✅ Validate image files
- ✅ Calculate total file sizes
- ✅ Preview local database sync operations
- ❌ **NOT** upload images to Azure
- ❌ **NOT** create database records

This is perfect for:
- Testing directory structure
- Validating product SKUs
- Checking file sizes before upload
- Verifying local database connectivity
- Identifying potential issues

## Local Database Sync

The local database sync feature allows you to synchronize image records to your local development database while uploading to production Azure storage.

### Key Benefits

1. **Development Consistency**: Your local database has the same image URLs as production
2. **Independent Operations**: Local sync failures don't affect production uploads
3. **Flexible Configuration**: Enable/disable via command-line flags
4. **Error Isolation**: Local database issues are logged but don't stop the process

### Requirements

- Local database must have the same schema as production
- Products/variants must exist in local database with matching SKUs
- Local database connection must be accessible

### Common Local Database Configurations

```bash
# Supabase Local Development
export LOCAL_DB_ENABLED=true
export LOCAL_DB_HOST=localhost
export LOCAL_DB_PORT=55322
export LOCAL_DB_USER=postgres
export LOCAL_DB_PASSWORD=postgres
export LOCAL_DB_NAME=postgres

# Docker PostgreSQL
export LOCAL_DB_ENABLED=true
export LOCAL_DB_HOST=localhost
export LOCAL_DB_PORT=5432
export LOCAL_DB_USER=myuser
export LOCAL_DB_PASSWORD=mypassword
export LOCAL_DB_NAME=mydatabase

# Local PostgreSQL
export LOCAL_DB_ENABLED=true
export LOCAL_DB_HOST=localhost
export LOCAL_DB_PORT=5432
export LOCAL_DB_USER=postgres
export LOCAL_DB_PASSWORD=localpass
export LOCAL_DB_NAME=local_db
```

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

For each uploaded image, records are created in both production and local databases (if sync enabled):

### `images` table
- `id`: Auto-generated unique identifier
- `url`: Public URL of the uploaded image
- `created_at`, `updated_at`: Timestamps

### `image_entities` table
- `entity_id`: Product or variant ID (resolved by SKU in each database)
- `image_id`: References the `images` table
- `alt_text`: Generated description like "Product image for SKU-001"
- `is_primary`: `true` for the first image in each directory
- `sort_order`: Order based on file processing sequence (0, 1, 2, ...)
- `entity_type`: Either "product" or "product_variant"

## Error Handling

The script provides comprehensive error handling:

### Production Database Errors
- **Fatal**: Process stops if production database operations fail
- **Rollback**: Failed transactions are automatically rolled back
- **Logging**: All errors are logged with context

### Local Database Errors
- **Non-Fatal**: Local sync errors don't stop the process
- **Warning**: Errors are logged as warnings
- **Statistics**: Local sync error count is reported in summary

### Azure Upload Errors
- **Retry Logic**: Built into Azure SDK
- **Detailed Logging**: Upload failures include file paths and sizes
- **Continuation**: Process continues with other files

### File System Errors
- **Validation**: Files are validated before processing
- **Skip Invalid**: Invalid/corrupted files are skipped with warnings
- **Size Limits**: Very large files are handled gracefully

## Output Examples

### Successful Upload with Local Sync
```
=== Upload Summary ===
Processed Products: 2
Processed Variants: 1
Total Uploaded Images: 5
  - Product Images: 4
  - Variant Images: 1
Total Skipped Images: 0
  - Product Images: 0
  - Variant Images: 0
Total Size: 2.34 MB
Errors: 0
Local Database Sync: ✅ Enabled
```

### Dry-Run Mode
```
🔍 DRY RUN MODE - No actual changes were made
[DRY RUN] Would upload product image: image1.jpg (245.67 KB)
[DRY RUN] Would sync to local database: PRODUCT-SKU-001

💡 Run without --dry-run flag to perform actual upload
```

### Error Reporting
```
Local Database Sync: ✅ Enabled (⚠️ 2 sync errors)

Errors encountered:
- variant SKU variant-not-found: variant with SKU variant-not-found not found in local database
- product image corrupted.jpg: file access error
```

## Troubleshooting

### Common Issues

1. **Local Database Connection Failed**
   ```
   Failed to connect to local database: connection refused
   ```
   - Check if local database is running
   - Verify LOCAL_DB_* environment variables are set correctly
   - Ensure network connectivity

2. **SKU Not Found in Local Database**
   ```
   product with SKU PROD-001 not found: no rows in result set
   ```
   - Ensure products exist in local database
   - Check SKU spelling and case sensitivity
   - Verify database schema is up to date

3. **Azure Upload Failures**
   ```
   failed to upload to Azure: authentication failed
   ```
   - Check Azure credentials in environment variables
   - Verify storage account permissions
   - Check network connectivity

### Debug Mode

Enable debug logging by setting the log level:

```bash
export LOG_LEVEL=debug
export LOCAL_DB_ENABLED=true
export LOCAL_DB_HOST=localhost
export LOCAL_DB_PORT=55322
go run scripts/upload_image.go --sync-local ./images
```

## Security Considerations

1. **Database Credentials**: Never commit database URLs with credentials to version control
2. **Azure Keys**: Store Azure storage keys securely in environment variables
3. **Local Development**: Use separate credentials for local and production databases
4. **Network Security**: Ensure secure connections to databases (SSL/TLS)

## Performance Tips

1. **Batch Size**: Process images in smaller batches for better memory usage
2. **Network**: Use stable internet connection for Azure uploads
3. **Database**: Ensure database connections are stable
4. **File Size**: Optimize images before upload to reduce transfer time