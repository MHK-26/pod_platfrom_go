

## Project Structure

The platform follows a microservices architecture with the following components:

- **Authentication Service**: Handles user registration, login, and token management
- **Content Service**: Manages podcasts, episodes, categories, and user interactions
- **Analytics Service**: Collects and processes listening data and provides insights
- **Recommendation Service**: Generates personalized content recommendations
- **Payment Service**: Handles subscriptions, donations, and monetization

## Technology Stack

- **Backend**: Go with Clean Architecture
- **Database**: PostgreSQL
- **API**: REST with potential for gRPC between services
- **Authentication**: JWT-based with support for social logins
- **Storage**: Local file storage for media files
- **Deployment**: Docker for containerization, can be deployed to Hostinger

## Getting Started

### Prerequisites

- Go 1.20 or later
- PostgreSQL 13 or later
- Docker and Docker Compose (optional)
- Make (for using Makefile commands)

### Setup

1. Clone the repository:

```bash
git clone https://github.com/MHK-26/pod_platfrom_go.git
cd podcast-platform
```

2. Create a `.env` file based on `.env.example`:

```bash
cp .env.example .env
```

3. Update the `.env` file with your configuration values.

4. Create the storage directory for media files:

```bash
mkdir -p storage/podcasts storage/episodes
```

5. Create the database:

```bash
createdb podcast_platform
```

6. Run database migrations:

```bash
make migrateup
```

7. Seed the database with initial data (optional):

```bash
psql podcast_platform < scripts/seed/seed_data.sql
```

### Running the Services

You can run individual services or all services together:

```bash
# Run all services
make run

# Run a specific service
make run-auth-service
```

### Docker Support

To run the entire stack with Docker Compose:

```bash
docker-compose up -d
```

This will start PostgreSQL, all microservices, and an Nginx reverse proxy.

### Building the Services

```bash
# Build all services
make build

# Build a specific service
make build-auth-service
```

## API Documentation

The API documentation is available using Swagger. After starting the services, you can access the Swagger UI at:

```
http://localhost:8080/swagger/index.html
```

To generate the Swagger documentation:

```bash
make swag
```

## Database Schema

The database schema includes tables for:

- Users (listeners, podcasters, admins)
- Podcasts and episodes
- Categories and tags
- Subscriptions and playback history
- Analytics data
- Monetization features

## Deployment Guide

For deploying to Hostinger, please see the [Deployment Guide](DEPLOYMENT.md).

## Development Guidelines

- Follow Go best practices and idiomatic Go
- Use Clean Architecture with separation of concerns
- Write tests for all functionality
- Document all APIs and functions
- Use proper error handling and logging

## Directory Structure

```
/podcast-platform
├── cmd                       # Application entry points
│   ├── auth-service          # Authentication service
│   ├── content-service       # Content service
│   ├── analytics-service     # Analytics service
│   ├── recommendation-service# Recommendation service
│   └── payment-service       # Payment service
├── pkg                       # Library code
│   ├── common                # Common utilities, middleware, config
│   ├── auth                  # Authentication domain
│   ├── content               # Content domain
│   ├── analytics             # Analytics domain
│   ├── recommendation        # Recommendation domain
│   └── payment               # Payment domain
├── api                       # API definitions (gRPC, Swagger)
├── deployments               # Deployment configurations
├── scripts                   # Scripts for migrations, seeding, etc.
└── storage                   # Storage for uploaded files
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin feature/your-feature-name`
5. Submit a pull request

