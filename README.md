# rag-sql

Project focused on understanding the database to create RAG (Retrieval-Augmented Generation) for LLMs response to generate accurate SQL queries.

## Overview

RAG-SQL is a sophisticated system that combines database schema understanding with Large Language Models to generate contextually accurate SQL queries. The system uses retrieval-augmented generation techniques to provide LLMs with relevant database context, enabling more precise and reliable SQL code generation.

## Features

- **Database Schema Analysis**: Automatically analyzes and understands database structures
- **Graph-based Schema Representation**: Uses graph algorithms for efficient schema navigation
- **Fuzzy Search Capabilities**: Intelligent matching of natural language queries to database entities
- **Multiple LLM Support**: Compatible with various language models including Llama and Natural-SQL models
- **Alias Generation**: Automatic generation of database aliases for improved query readability
- **RESTful API**: HTTP API for easy integration with external applications

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Layer    │────│  Graph Engine   │────│  LLM Client     │
│  (router.go)   │    │ (graph/*.go)    │    │ (client.go)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Database       │
                    │  (db/*.go)      │
                    └─────────────────┘
```

## Project Structure

### Core Directories

- **`cmd/`**: Application entry points
  - `server.go`: Main HTTP server
  - `generate-aliases/`: Utility for generating database aliases

- **`internal/api/`**: HTTP API layer
  - `router.go`: API route definitions and handlers

- **`internal/config/`**: Configuration management
  - `config.go`: Application configuration

- **`internal/db/`**: Database interaction layer
  - `conn.go`: Database connection management
  - `contextbuilder/`: Database context building utilities
  - `dbschema/`: Schema introspection services
  - `exec/`: Query execution utilities
  - `schemautil/`: Schema manipulation utilities

- **`internal/graph/`**: Graph-based schema representation
  - `driver.go`: Graph database driver
  - `fuzzy.go`: Fuzzy matching algorithms
  - `loader.go`: Schema loading mechanisms
  - `memory.go`: In-memory graph operations
  - `search.go`: Graph search algorithms
  - `types.go`: Graph type definitions
  - `utils.go`: Graph utility functions

- **`internal/llm/`**: Large Language Model integration
  - `client.go`: LLM client interface and implementations

- **`internal/tools/`**: Utility tools
  - `aliasgen.go`: Database alias generation

- **`train/`**: Model files and training data
  - Pre-trained models for SQL generation

## Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd rag-sql
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

## Configuration

Create a `.env` file based on `.env.example` with the following settings:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=your_database
DB_USER=your_username
DB_PASSWORD=your_password

# LLM Configuration
LLM_MODEL_PATH=./train/llama-3-sqlcoder-8b.Q8_0.gguf
LLM_API_ENDPOINT=http://localhost:11434

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
```

## Usage

### Running the Server

```bash
go run cmd/server.go
```

The server will start on the configured port (default: 8080).

### Generating Database Aliases

```bash
go run cmd/generate-aliases/main.go
```

### API Endpoints

#### Generate SQL Query
```http
POST /api/query
Content-Type: application/json

{
  "question": "Show me all users who registered last month",
  "context": "users table with registration_date column"
}
```

#### Get Schema Information
```http
GET /api/schema
```

#### Search Schema Elements
```http
GET /api/search?q=user
```

## Models

The system supports multiple pre-trained models:

- **Llama-3 SQLCoder 8B**: High-accuracy SQL generation model
- **Natural-SQL 7B**: Lightweight model for natural language to SQL conversion

Models are stored in the `train/` directory and can be configured via environment variables.

## Development

### Prerequisites

- Go 1.21 or higher
- Database system (PostgreSQL, MySQL, etc.)
- Ollama (for local LLM inference)

### Building

```bash
go build -o rag-sql cmd/server.go
```

### Testing

```bash
go test ./...
```

### Code Structure

The codebase follows clean architecture principles:

1. **API Layer**: Handles HTTP requests and responses
2. **Business Logic**: Graph algorithms and schema analysis
3. **Data Layer**: Database connections and schema introspection
4. **External Services**: LLM integration and model inference

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Performance Considerations

- **Memory Management**: Graph operations are optimized for large schemas
- **Caching**: Schema information is cached to reduce database queries
- **Concurrent Processing**: Multiple requests can be processed simultaneously
- **Model Loading**: Models are loaded once and reused for better performance

## Troubleshooting

### Common Issues

1. **Model Loading Errors**: Ensure model files are present in the `train/` directory
2. **Database Connection**: Verify database credentials and network connectivity
3. **Memory Issues**: Increase available memory for large schemas
4. **LLM Timeout**: Adjust timeout settings for complex queries

## Support

For questions and support, please [open an issue] or contact the development team Email: diogosgn@gmail.com.