# Webhooks Handler

This package handles incoming webhooks from external services, primarily Clerk authentication events.

## Overview

The webhooks handler is responsible for:
- Receiving webhook events from Clerk
- Processing `user.created` events
- Creating user records in the database
- Handling duplicate user creation (idempotency)

## API Endpoints

### POST `/v1/webhooks/clerk/create-user`

Processes Clerk `user.created` webhook events.

#### Request

- **Content-Type**: `application/json`
- **Body**: Clerk webhook event payload

```json
{
  "data": {
    "id": "user_29w83sxmDNGwOuEthce5gg56FcC",
    "first_name": "John",
    "last_name": "Doe",
    "email_addresses": [
      {
        "email_address": "john.doe@example.org",
        "verification": {
          "status": "verified"
        }
      }
    ],
    "created_at": 1654012591514,
    "updated_at": 1654012591835
  },
  "type": "user.created",
  "object": "event",
  "timestamp": 1654012591835
}
```

#### Response

**Success (200)**:
```json
{
  "message": "User created successfully",
  "user_id": 123,
  "clerk_id": "user_29w83sxmDNGwOuEthce5gg56FcC"
}
```

**Error Responses**:
- `400` - Invalid payload, unsupported event type, or missing required fields
- `500` - Database error during user creation

## Features

### Email Handling
- Email is optional (nullable in database)
- Users can be created without email addresses
- Primary email is extracted from verified emails first, then any email

### Name Handling
- Combines `first_name` and `last_name` from Clerk
- Defaults to "User" if no name provided

### Idempotency
- Checks if user already exists by `auth_provider_id`
- Returns existing user if found (no duplicate creation)
- Logs appropriate messages for both scenarios

### Error Handling
- Comprehensive error logging
- Proper HTTP status codes
- Detailed error messages for debugging

## Database Schema

Users are stored with:
- `name`: Combined first and last name
- `email`: Primary email address (nullable)
- `auth_provider`: Set to 'clerk'
- `auth_provider_id`: Clerk user ID

## Testing

Use the HTTP test file at `http/webhooks.http` to test various scenarios:
- User with email
- User without email
- Minimal user data
- Unsupported event types

## Error Codes

- `FAILED_TO_DECODE_WEBHOOK`: JSON parsing failed
- `FAILED_TO_VERIFY_WEBHOOK`: Webhook signature verification failed
- `UNSUPPORTED_EVENT_TYPE`: Event type is not 'user.created'
- `FAILED_TO_CREATE_USER`: Database error during user creation
- `INVALID_WEBHOOK_PAYLOAD`: Missing required fields
- `USER_ALREADY_EXISTS`: User with same auth_provider_id exists

## Configuration

No additional configuration required beyond database connection.