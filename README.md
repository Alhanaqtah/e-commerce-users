# E-Commerce Users Service

## Overview
The `e-commerce-users-service` is a microservice designed to handle user authentication and management in an e-commerce platform. Built with Go, it adheres to **Clean Architecture** principles to ensure scalability, maintainability, and separation of concerns.

### Key Features
- **User Authentication**:
  - Sign up with email confirmation.
  - Sign in with JWT-based token generation.
  - Logout with token blacklisting.
  - Token refresh functionality.
- **Email Confirmation**:
  - Send confirmation codes to users.
  - Resend confirmation codes.
  - Confirm user accounts via email.
- **Secure Token Management**:
  - Access and refresh tokens with customizable TTL.
  - Blacklist invalid or expired tokens.

---

## Architecture
This service follows Clean Architecture principles:
- **Delivery**: Manages incoming requests and interacts with the service layer.
- **Service**: Contains business logic and orchestrates data flow between layers.
- **Repository**: Handles interactions with the database, cache, and external APIs.

---

## Configuration
The service is configurable using a `.env` file. Below is an example configuration:

```env
# App configuration
ENV=local
PREFIX=USERS

# HTTP Server Configuration
HTTP_SERVER_HOST=localhost
HTTP_SERVER_PORT=8080
HTTP_SERVER_IDLE_TIMEOUT=4s

# PostgreSQL Database Configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_NAME=postgres
POSTGRES_MAX_CONNS=100

# Redis Configuration
REDIS_ADDRESS=localhost:6379
REDIS_PASSWORD=secret-password
REDIS_DB=0

# SMTP Configuration
SMTP_USERNAME=***
SMTP_PASSWORD=***
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_CODE_TTL=15m

# Tokens Configuration
TOKENS_SECRET=secret-password
TOKENS_ACCESS_TTL=5m
TOKENS_REFRESH_TTL=15m
```

---

## Running the Service

### Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/e-commerce-users-service.git
   cd e-commerce-users-service
   ```

2. Set up the `.env` file with your configuration.

3. Start the local development environment with Docker Compose:
   ```bash
   make sandbox-up
   ```

4. To build and run the service inside the Docker container:
   ```bash
   make run
   ```

## Local Development with Docker Compose

The project includes a `docker-compose.yml` file for local development:

1. Start the service:
   ```bash
   make sandbox-up
   ```

2. Stop the service:
   ```bash
   make sandbox-down
   ```

---

## Testing
Run the tests using:
```bash
go test -v ./...
```

---

## Docker Support
A Dockerfile is available to containerize the service:

### Build the Docker Image
```bash
docker build -t e-commerce-users-service .
```

### Run the Docker Container
```bash
docker run -d -p 8080:8080 e-commerce-users-service
```
