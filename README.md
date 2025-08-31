# Fix Track Bot

A Discord bot for issue tracking and bug reporting built with Go, following clean architecture principles.

## Architecture

This project follows Clean Architecture with clear separation of concerns:

```
fix-track-bot/
├── internal/
│   ├── domain/          # Business entities and rules
│   │   ├── issue.go     # Issue entity and business rules
│   │   ├── interfaces.go # Repository and service interfaces  
│   │   └── errors.go    # Domain-specific errors
│   ├── repository/      # Data access layer
│   │   ├── issue_repository.go # Issue database operations
│   │   └── database.go  # Database connection and migrations
│   ├── service/         # Business logic layer
│   │   └── issue_service.go # Issue business logic
│   ├── transport/       # External interfaces
│   │   └── discord/     # Discord bot handlers
│   │       ├── handler.go   # Discord event handlers
│   │       └── commands.go  # Slash command management
│   └── config/          # Configuration management
│       └── config.go    # Application configuration
├── pkg/
│   └── logger/          # Logging utilities
│       └── logger.go    # Structured logging with Zap
├── data/               # Database files (SQLite)
├── config.example.yaml # Example configuration file
├── main.go            # Application entry point
└── go.mod             # Go module dependencies
```

### Layer Responsibilities

- **Domain Layer**: Contains business entities (`Issue`), interfaces, and domain-specific errors
- **Repository Layer**: Handles database operations and data persistence
- **Service Layer**: Implements business logic and orchestrates domain operations
- **Transport Layer**: Handles external communication (Discord interactions)
- **Config Layer**: Manages application configuration and environment variables

## Features

- ✅ Channel registration with customer and project information
- ✅ Issue creation via Discord slash commands
- ✅ Issue tracking with unique IDs  
- ✅ Thread-based discussions for each issue
- ✅ Priority levels (Low, Medium, High) with visual indicators
- ✅ Issue status management (Open, Closed)
- ✅ Issue listing and searching by channel
- ✅ Detailed issue status checking with partial ID support
- ✅ Interactive priority setting via dropdown menus
- ✅ Comprehensive help system
- ✅ Structured logging with Zap
- ✅ Database persistence with GORM (SQLite)
- ✅ Clean architecture with dependency injection
- ✅ Proper error handling and validation

## Configuration

### Environment Variables

Copy `config.example.yaml` to `config.yaml` and configure:

```yaml
app:
  name: "fix-track-bot"
  version: "1.0.0"
  environment: "development"
  debug: true

discord:
  token: "your_discord_bot_token_here"
  prefix: "!"

database:
  driver: "sqlite"
  file_path: "./data/fix-track.db"

logger:
  level: "info"
  environment: "development"
  output_paths:
    - "stdout"
```

### Environment Variables (Alternative)

You can also use environment variables:

```bash
export DISCORD_TOKEN="your_discord_bot_token_here"
export DATABASE_FILE_PATH="./data/fix-track.db"
export LOG_LEVEL="info"
```

## Setup

### Option 1: Docker Compose (Recommended)

1. **Clone and Setup**
   ```bash
   git clone <repository>
   cd fix-track-bot
   make setup  # Copies environment file
   ```

2. **Configure Environment**
   ```bash
   # Edit .env file and set your Discord bot token
   nano .env
   # Set DISCORD_TOKEN=your_actual_bot_token
   ```

3. **Start Services**
   ```bash
   make docker-up
   # This starts PostgreSQL and the bot
   ```

4. **View Logs**
   ```bash
   make docker-logs
   ```

### Option 2: Local Development

1. **Install Dependencies**
   ```bash
   go mod tidy
   ```

2. **Start PostgreSQL**
   ```bash
   # Option A: Use Docker for just PostgreSQL
   docker-compose up -d postgres
   
   # Option B: Install PostgreSQL locally
   # Create database: fix_track
   # Create user: fix_track_user
   ```

3. **Configure the Bot**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your Discord bot token and database settings
   ```

4. **Run the Bot**
   ```bash
   go run main.go
   ```

### Available Make Commands

```bash
make help          # Show all available commands
make docker-up     # Start all services
make docker-down   # Stop all services
make docker-logs   # View service logs
make build         # Build the application
make test          # Run tests
```

## Discord Bot Setup

1. Create a new application at [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a bot user and copy the token
3. Enable the following bot permissions:
   - Send Messages
   - Use Slash Commands
   - Create Public Threads
   - Manage Threads
   - Embed Links
4. Invite the bot to your server with the required permissions

## Usage

### Slash Commands

- `/register` - Register the current channel for issue tracking with customer and project information
- `/issue` - Create a new issue with a modal form
- `/issues` - List all issues in the current channel (shows up to 10 most recent)
- `/issue-status <id>` - Check the status of a specific issue (accepts full UUID or first 8 characters)
- `/help` - Show comprehensive help information

### Issue Management

1. **Register the channel** using `/register` with customer name and project name
2. Use `/issue` to create a new issue
3. Fill out the modal with title, description, and optional image URL
4. The bot creates a thread for discussion
5. Set priority using the dropdown menu in the thread
6. Close issues using the "🔒 Close Issue" button

## Development

### Code Standards

This project follows Go best practices and clean architecture principles:

- **Idiomatic Go**: Following Effective Go guidelines
- **Clean Architecture**: Separation of concerns across layers
- **SOLID Principles**: Single responsibility, dependency inversion
- **Error Handling**: Proper error wrapping and context
- **Logging**: Structured logging with correlation IDs
- **Testing**: Unit tests with mocks (testify/mockery)

### Running Tests

```bash
go test ./...
```

### Linting

```bash
golangci-lint run
```

## Database Schema

The bot uses PostgreSQL with a normalized schema supporting multi-tenancy:

### Architecture Overview

```
Customers (Organizations)
    ├── Projects (Customer Projects)
    │   ├── Channels (Discord Channel Registrations)
    │   └── Issues (Bug Reports/Features)
    └── Users (Customer Users)
```

### Entity Relationships

- **Customers** can have multiple **Projects** and **Users**
- **Projects** belong to one **Customer** and can have multiple **Channels** and **Issues**
- **Users** can belong to one **Customer** (or be system users)
- **Channels** are registered for one **Project** by one **User**
- **Issues** are reported in one **Project** by one **User**

## Database Schema

### Customers Table
```sql
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

### Projects Table
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id),
    name VARCHAR(255),
    email VARCHAR(255),
    discord_id VARCHAR(100) UNIQUE,
    role VARCHAR(20) DEFAULT 'customer',
    created_at TIMESTAMPTZ DEFAULT now()
);
```

### Channels Table
```sql
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    channel_id VARCHAR(100) NOT NULL UNIQUE,
    guild_id VARCHAR(100) NOT NULL,
    registered_by UUID NOT NULL REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

### Issues Table
```sql
CREATE TABLE issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    reporter_id UUID NOT NULL REFERENCES users(id),
    assignee_id UUID REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    image_url VARCHAR(500),
    priority VARCHAR(10) DEFAULT 'medium',
    status VARCHAR(10) DEFAULT 'open',
    channel_id VARCHAR(100),  -- Discord channel ID (optional)
    thread_id VARCHAR(100),
    message_id VARCHAR(100),
    public_hash VARCHAR(100) UNIQUE, -- For public links
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    closed_at TIMESTAMPTZ
);
```

## Database Management

### Docker Environment

When using Docker Compose, the PostgreSQL database is automatically set up with:
- **Database:** `fix_track`
- **User:** `fix_track_user`
- **Password:** `fix_track_password`
- **Port:** `5432` (mapped to host)

### Database Operations

```bash
# Reset database (destroys all data)
make db-reset

# Connect to PostgreSQL (when running in Docker)
docker exec -it fix-track-postgres psql -U fix_track_user -d fix_track

# View database logs
docker-compose logs postgres

# Backup database
docker exec fix-track-postgres pg_dump -U fix_track_user fix_track > backup.sql

# Restore database
docker exec -i fix-track-postgres psql -U fix_track_user -d fix_track < backup.sql
```

### Migrations

The application automatically runs GORM auto-migrations on startup, creating:
- `channels` table for channel registrations
- `issues` table for issue tracking

## Contributing

1. Fork the repository
2. Create a feature branch
3. Follow the coding standards
4. Write tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License.
