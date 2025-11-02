# pm-tool

A Clean Architecture Go application generated with go-template-gen.

## Features

- Clean Architecture structure
- RESTful API with Gin
- Database integration (postgres)
- Redis caching and rate limiting
- Observability with opentelemetry
- Docker support
- Hot reload with Air
- Database migrations

## Prerequisites

- Go 1.21 or higher
- postgres
- Redis
- Docker & Docker Compose (optional)

## Getting Started

### 1. Install dependencies

```bash
make deps
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Run with Docker Compose

```bash
make docker-up
```

### 4. Or run locally

```bash
# Run migrations
make migrate-up

# Run the application
make run

# Or use hot reload
make dev
```

## Project Structure

```
.
├── main.go                 # Application entry point
├── cmd/                    # CLI commands
├── config/                 # Configuration management
├── handler/                # HTTP handlers and queue handlers
│   ├── api/               # REST API handlers
│   └── queue/             # Queue message handlers
├── business/              # Business logic layer
│   ├── entity/            # Domain entities
│   ├── domain/            # Domain interfaces
│   └── usecase/           # Use cases
├── infra/                 # Infrastructure layer
│   ├── postgres/ # Database implementation
│   ├── redis/             # Redis implementation

├── pkg/                   # Shared packages
│   ├── middleware/        # HTTP middleware
│   ├── telemetry/         # Observability
│   └── transaction/       # Transaction management
└── migration/             # Database migrations
```

## Available Commands

```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run the application
make test          # Run tests
make migrate-up    # Run migrations
make docker-up     # Start with Docker
make dev           # Run with hot reload
```

## API Endpoints

The API will be available at: http://localhost:8080

Generated endpoints will be documented here based on your schema.

## Development

### Running Tests

```bash
make test
make test-coverage
```

### Creating Migrations

```bash
make migrate-create name=create_users_table
```

### Code Formatting

```bash
make fmt
make lint
```

## Deployment

### Build Docker Image

```bash
make docker-build
```

### Deploy

```bash
docker run -p 8080:8080 pm-tool:latest
```

## License

MIT
