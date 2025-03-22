# Subscription Service API

A subscription management backend service built with Go and Gin. This project implements a complete subscription API with product management, subscriptions, vouchers, and trial periods.

## What's Included

This service handles several key features:

- **Product catalog management**: Create, view, and manage subscription products
- **Subscription lifecycle**: Let users subscribe, pause, unpause, and cancel
- **Voucher system**: Apply fixed or percentage discounts to subscriptions
- **Trial periods**: Allow users to try subscriptions before paying
- **User authentication**: Full JWT-based auth for protecting endpoints

## Tech Stack

- **Language**: Go 1.23.1+
- **Web Framework**: Gin-Gonic
- **Database**: PostgreSQL
- **Auth**: JWT tokens
- **Containerization**: Docker & Docker Compose

## Getting Started

### Option 1: Using Docker (Recommended)

The easiest way to get started is using Docker Compose:

```bash
# Clone the repository
git clone https://github.com/assylzhan-a/subscription-service.git
cd subscription-service

# Start the service
docker-compose up -d
```

This will start both the API service and a PostgreSQL database instance. The API will be available at http://localhost:8080.

### Option 2: Manual Setup

If you prefer to run the service directly:

1. Install Go 1.21+
2. Set up PostgreSQL

```bash
# Create a database
createdb subscription_service
```

3. Clone and run the service

```bash
git clone https://github.com/assylzhan-a/subscription-service.git
cd subscription-service

# Install dependencies
go mod tidy

# Run the service
go run cmd/api/main.go
```

The API will be available at http://localhost:8080.

## API Documentation

### Auth Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/auth/register | Register a new user |
| POST | /api/v1/auth/login | Login and get a JWT token |
| GET | /api/v1/auth/me | Get current user info (requires auth) |

### Product Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/v1/products | List all products |
| GET | /api/v1/products/:id | Get product details |
| POST | /api/v1/products | Create a product |
| PUT | /api/v1/products/:id | Update a product |
| DELETE | /api/v1/products/:id | Delete a product |

### Subscription Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/v1/subscriptions | List user's subscriptions |
| GET | /api/v1/subscriptions/:id | Get subscription details |
| POST | /api/v1/subscriptions | Create a subscription |
| PATCH | /api/v1/subscriptions/:id/pause | Pause a subscription |
| PATCH | /api/v1/subscriptions/:id/unpause | Unpause a subscription |
| PATCH | /api/v1/subscriptions/:id/cancel | Cancel a subscription |

### Voucher Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/vouchers/validate | Validate a voucher code |
| GET | /api/v1/admin/vouchers | List all vouchers (admin) |
| GET | /api/v1/admin/vouchers/:id | Get voucher details (admin) |
| GET | /api/v1/admin/vouchers/product/:id | List vouchers for a product (admin) |
| POST | /api/v1/admin/vouchers | Create a voucher (admin) |
| PUT | /api/v1/admin/vouchers/:id | Update a voucher (admin) |
| DELETE | /api/v1/admin/vouchers/:id | Delete a voucher (admin) |

## Authentication

Protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

You can obtain a token by registering a user and then logging in with that user's credentials.

# Complete Testing Guide for Subscription Service API

This guide provides a comprehensive set of curl commands to test all features of the subscription service API. The commands use placeholders like `YOUR_TOKEN` and `PRODUCT_ID` which you'll need to replace with actual values as you test.

## Authentication Tests

### 1. Register a new user

```bash
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "securepassword123",
    "name": "Test User"
  }'
```

### 2. Login with the registered user

```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "securepassword123"
  }'
```

You'll receive a response with a JWT token. Save this token and use it in subsequent requests:
```
{
  "token": "YOUR_TOKEN",
  ...
}
```

### 3. Get current user info

```bash
curl -X GET "http://localhost:8080/api/v1/auth/me" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Product Management Tests

### 4. Create a product (admin operation)

```bash
curl -X POST "http://localhost:8080/api/v1/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Premium Plan",
    "description": "Our best subscription plan with all features",
    "price": "19.99",
    "billing_period": "monthly",
    "duration_months": 1,
    "features": ["Feature 1", "Feature 2", "Feature 3"],
    "is_active": true
  }'
```

The response will include a product ID that you'll need for subsequent requests:
```
{
  "id": "PRODUCT_ID",
  ...
}
```

### 5. List all products

```bash
curl -X GET "http://localhost:8080/api/v1/products"
```

### 6. Get product details

```bash
curl -X GET "http://localhost:8080/api/v1/products/PRODUCT_ID"
```

### 7. Update a product

```bash
curl -X PUT "http://localhost:8080/api/v1/products/PRODUCT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Premium Plan Plus",
    "description": "Our enhanced premium plan with extra features",
    "price": "24.99",
    "billing_period": "monthly",
    "duration_months": 1,
    "features": ["Feature 1", "Feature 2", "Feature 3", "Feature 4"],
    "is_active": true
  }'
```

## Voucher Management Tests

### 8. Create a voucher (admin operation)

```bash
curl -X POST "http://localhost:8080/api/v1/admin/vouchers" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "code": "SUMMER25",
    "discount_type": "percentage",
    "discount_value": "25",
    "product_id": "PRODUCT_ID",
    "valid_from": "2025-03-22T00:00:00Z",
    "valid_to": "2025-04-22T00:00:00Z",
    "expires_at": "2025-04-22T00:00:00Z",
    "max_uses": 100,
    "is_active": true
  }'
```

The response will include a voucher ID:
```
{
  "id": "VOUCHER_ID",
  ...
}
```

### 9. List all vouchers (admin)

```bash
curl -X GET "http://localhost:8080/api/v1/admin/vouchers" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 10. Get voucher details (admin)

```bash
curl -X GET "http://localhost:8080/api/v1/admin/vouchers/VOUCHER_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 11. List vouchers for a product (admin)

```bash
curl -X GET "http://localhost:8080/api/v1/admin/vouchers/product/PRODUCT_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 12. Validate a voucher code

```bash
curl -X POST "http://localhost:8080/api/v1/vouchers/validate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "code": "SUMMER25",
    "product_id": "PRODUCT_ID"
  }'
```

### 13. Update a voucher (admin)

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/vouchers/VOUCHER_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "code": "SUMMER25",
    "discount_type": "percentage",
    "discount_value": "30",
    "product_id": "PRODUCT_ID",
    "valid_from": "2025-03-22T00:00:00Z",
    "valid_to": "2025-05-22T00:00:00Z",
    "expires_at": "2025-05-22T00:00:00Z",
    "max_uses": 200,
    "is_active": true
  }'
```

## Subscription Management Tests

### 14. Create a subscription (with trial and voucher)

```bash
curl -X POST "http://localhost:8080/api/v1/subscriptions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "product_id": "PRODUCT_ID",
    "voucher_code": "SUMMER25",
    "with_trial": true
  }'
```

The response will include a subscription ID:
```
{
  "id": "SUBSCRIPTION_ID",
  ...
}
```

### 15. List user's subscriptions

```bash
curl -X GET "http://localhost:8080/api/v1/subscriptions" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 16. Get subscription details

```bash
curl -X GET "http://localhost:8080/api/v1/subscriptions/SUBSCRIPTION_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 17. Pause a subscription

```bash
curl -X PATCH "http://localhost:8080/api/v1/subscriptions/SUBSCRIPTION_ID/pause" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 18. Unpause a subscription

```bash
curl -X PATCH "http://localhost:8080/api/v1/subscriptions/SUBSCRIPTION_ID/unpause" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 19. Cancel a subscription

```bash
curl -X PATCH "http://localhost:8080/api/v1/subscriptions/SUBSCRIPTION_ID/cancel" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Create Another Subscription (without trial or voucher)

### 20. Create a regular subscription

```bash
curl -X POST "http://localhost:8080/api/v1/subscriptions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "product_id": "PRODUCT_ID",
    "with_trial": false
  }'
```

## Clean Up (Optional)

### 21. Delete a voucher

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/vouchers/VOUCHER_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 22. Delete a product

```bash
curl -X DELETE "http://localhost:8080/api/v1/products/PRODUCT_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Manual Placeholder Replacement**: You'll need to manually replace placeholders like `YOUR_TOKEN`, `PRODUCT_ID`, etc., with actual values from previous responses.

## Project Structure

The project follows a clean architecture approach with domain-driven design:

```
subscription-service/
├── cmd/api/                  # Application entry point
├── configs/                  # Configuration loading
├── internal/
│   ├── app/                  # Application services
│   │   ├── auth/             # Authentication logic
│   │   ├── product/          # Product business logic
│   │   ├── subscription/     # Subscription business logic
│   │   └── voucher/          # Voucher business logic
│   ├── domain/               # Domain models and errors
│   ├── handlers/             # HTTP request handlers
│   ├── middleware/           # HTTP middleware
│   ├── repository/           # Data access interfaces
│   │   └── postgres/         # PostgreSQL implementations
│   └── transport/            # Transport/presentation layer
│       ├── dto/              # Data transfer objects
│       └── http/             # HTTP routing
└── pkg/                      # Shared utilities
    └── jwt/                  # JWT utils
```

## Design Decisions

- **Decimal for Currency**: Used the decimal package for all money values to avoid floating-point precision issues
- **Service Layer**: Business logic is contained in service packages, separate from handlers
- **Repository Pattern**: Data access is abstracted through repository interfaces
- **Domain-Driven Design**: Code is organized around the business domain
- **Clean Architecture**: Distinct layers with clear dependencies


