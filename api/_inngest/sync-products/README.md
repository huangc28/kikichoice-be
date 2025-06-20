# Sync Inventory Feature Plan

## Overview

The sync-inventory feature is a Vercel Edge Function that automatically synchronizes product data from a Google Sheet to a PostgreSQL database every 15 minutes using Inngest for cron job scheduling.

## Architecture

### Components

1. **Vercel Edge Function** (`api/sync-inventory.ts`)
   - Entry point for the sync operation
   - Handles HTTP requests and responses
   - Integrates with Inngest for job scheduling

2. **Inngest Function** (`api/_inngest/sync-inventory/main.ts`)
   - Core business logic for syncing inventory
   - Scheduled to run every 15 minutes
   - Handles data transformation and database operations

3. **Google Sheets Integration**
   - Fetches product data from specified Google Sheet
   - Handles authentication and rate limiting
   - Parses sheet data into structured format

4. **PostgreSQL Database Operations**
   - Upserts product records
   - Handles inventory quantity updates
   - Maintains data consistency

## Technical Stack

- **Runtime**: Vercel Edge Runtime
- **Scheduling**: Inngest
- **Database**: PostgreSQL
- **Google Sheets API**: Google Sheets API v4
- **Language**: TypeScript

## Data Flow

```
Google Sheet → Google Sheets API → Inngest Function → PostgreSQL Database → Website
```

1. Inngest triggers the sync function every 15 minutes
2. Function authenticates with Google Sheets API
3. Fetches product data from the specified sheet
4. Transforms data into database-compatible format
5. Performs upsert operations on PostgreSQL
6. Logs sync results and any errors

## Database Schema

### Products Table Structure
```sql
CREATE TABLE products (
  id SERIAL PRIMARY KEY,
  sheet_row_id VARCHAR(50) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  price DECIMAL(10,2) NOT NULL,
  quantity INTEGER NOT NULL DEFAULT 0,
  sku VARCHAR(100) UNIQUE,
  category VARCHAR(100),
  image_url VARCHAR(500),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Google Sheet Structure

Expected columns in the Google Sheet:
- A: Product Name
- B: Description
- C: Price
- D: Quantity
- E: SKU
- F: Category
- G: Image URL
- H: Active Status (TRUE/FALSE)

## Environment Variables

```bash
# Google Sheets API
GOOGLE_SHEETS_API_KEY=your_api_key
GOOGLE_SHEET_ID=your_sheet_id
GOOGLE_SHEET_RANGE=Sheet1!A2:H1000

# Database
DATABASE_URL=postgresql://user:password@host:port/database

# Inngest
INNGEST_EVENT_KEY=your_inngest_event_key
INNGEST_SIGNING_KEY=your_inngest_signing_key
```

## API Endpoints

### `/api/sync-inventory`
- **Method**: GET/POST
- **Purpose**: Trigger manual sync or handle Inngest webhooks
- **Response**: JSON with sync status and results

## Error Handling

1. **Google Sheets API Errors**
   - Rate limiting (exponential backoff)
   - Authentication failures
   - Sheet not found or permissions issues

2. **Database Errors**
   - Connection failures
   - Constraint violations
   - Transaction rollbacks

3. **Data Validation Errors**
   - Missing required fields
   - Invalid data types
   - Duplicate SKUs

## Logging and Monitoring

- Sync operation start/end times
- Number of products processed
- Database operation results
- Error details and stack traces
- Performance metrics

## Security Considerations

1. **API Key Management**
   - Store Google Sheets API key securely
   - Rotate keys regularly
   - Use least privilege access

2. **Database Security**
   - Use connection pooling
   - Implement proper SQL injection prevention
   - Encrypt sensitive data

3. **Rate Limiting**
   - Respect Google Sheets API limits
   - Implement exponential backoff
   - Monitor API usage

## Testing Strategy

1. **Unit Tests**
   - Data transformation functions
   - Validation logic
   - Error handling

2. **Integration Tests**
   - Google Sheets API integration
   - Database operations
   - End-to-end sync flow

3. **Load Testing**
   - Large dataset handling
   - Concurrent sync operations
   - Performance under load

## Deployment

1. **Vercel Configuration**
   - Update `vercel.json` with TypeScript function configuration
   - Set environment variables in Vercel dashboard
   - Configure edge function regions

2. **Inngest Setup**
   - Register Inngest functions
   - Configure cron schedule (every 15 minutes)
   - Set up monitoring and alerts

## Implementation Phases

### Phase 1: Core Infrastructure
- [ ] Set up Vercel Edge Function structure
- [ ] Implement basic Inngest integration
- [ ] Create database connection utilities

### Phase 2: Google Sheets Integration
- [ ] Implement Google Sheets API client
- [ ] Add authentication and error handling
- [ ] Create data transformation layer

### Phase 3: Database Operations
- [ ] Implement product upsert logic
- [ ] Add transaction handling
- [ ] Create data validation

### Phase 4: Scheduling and Monitoring
- [ ] Configure Inngest cron job
- [ ] Implement comprehensive logging
- [ ] Add error monitoring and alerts

### Phase 5: Testing and Optimization
- [ ] Add comprehensive test suite
- [ ] Optimize performance
- [ ] Implement rate limiting

## Configuration Examples

### Inngest Function Configuration
```typescript
export default inngest.createFunction(
  {
    id: "sync-inventory",
    name: "Sync Inventory from Google Sheets",
  },
  { cron: "*/15 * * * *" }, // Every 15 minutes
  async ({ event, step }) => {
    // Implementation
  }
);
```

### Google Sheets API Setup
```typescript
const sheets = google.sheets({
  version: 'v4',
  auth: process.env.GOOGLE_SHEETS_API_KEY
});
```

## Success Metrics

- Sync completion rate > 99%
- Data consistency between sheet and database
- Sync operation time < 30 seconds
- Zero data loss incidents
- Error rate < 1%

## Future Enhancements

1. **Multi-sheet Support**
   - Handle multiple product sheets
   - Category-specific sheets
   - Vendor-specific sheets

2. **Advanced Features**
   - Inventory alerts for low stock
   - Price change notifications
   - Bulk operations support

3. **Performance Optimizations**
   - Delta sync (only changed records)
   - Batch processing improvements
   - Caching layer implementation

## Support and Maintenance

- Monitor sync logs daily
- Review error patterns weekly
- Update dependencies monthly
- Performance review quarterly