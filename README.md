# GORM Library Management System

A robust library management system built with Go and GORM, featuring PostgreSQL backend with comprehensive database operations, relationship management, and extensive test coverage.

## Features

- **Book Management**: Add, find, update, and remove books with ISBN-based operations
- **Database Relationships**: Many-to-many relationships between books, authors, and categories
- **Publisher Integration**: Book-publisher relationships with contact information
- **Review System**: Customer review functionality with rating constraints
- **Connection Pooling**: Optimized database connections with configurable pool settings
- **Comprehensive Testing**: Full test suite with PostgreSQL test database integration
- **Schema Validation**: Database constraints and validation at the model level

## Prerequisites

- Go 1.16 or higher
- PostgreSQL 12 or higher
- GORM v2.x

## Installation

1. **Clone the repository**:

   ```bash
   git clone <repository-url>
   cd go-fundamentals/GORM
   ```

2. **Install dependencies**:

   ```bash
   go mod tidy
   ```

3. **Set up PostgreSQL database**:

   ```sql
   CREATE DATABASE gorm_db;
   CREATE DATABASE gorm_db_test;
   ```

4. **Configure environment variables**:
   ```bash
   export GO_DATABASE_URL="postgres://username:password@localhost:5432/gorm_db?sslmode=disable"
   export TEST_PG_DSN="host=localhost user=postgres password=genio123 dbname=gorm_db_test port=5432 sslmode=disable"
   ```

## Quick Start

### Running the Application

```bash
go run main.go
```

The application will:

- Connect to the PostgreSQL database
- Auto-migrate all schemas
- Demonstrate book service operations
- Create sample data

### Running Tests

```bash
go test -v
```

For comprehensive test coverage with detailed output:

```bash
go test -v -cover
```

## Database Schema

### Core Models

#### Book

- **ISBN**: Unique identifier (13 characters)
- **Title**: Book title (max 200 characters)
- **PublicationYear**: Year of publication
- **Copies**: Number of available copies
- **PublisherID**: Foreign key to Publisher
- **CreatedAt**: Automatic timestamp

#### Author

- **Name**: Author name (required)
- **Biography**: Text field for author bio
- **BirthYear**: Author's birth year
- **Books**: Many-to-many relationship with Book

#### Publisher

- **Name**: Publisher name (required)
- **Address**: Publisher address (text field)

#### Category

- **Name**: Category name (unique, required)
- **Books**: Many-to-many relationship with Book

#### Review

- **Rating**: Customer rating (1-5, with CHECK constraint)
- **Comment**: Review text
- **CustomerID**: Customer identifier
- **ProductID**: Product identifier

## API Reference

### BookService

The `BookService` provides business logic for book operations:

#### AddBook(book \*Book) error

Adds a new book to the database.

```go
book := &Book{
    ISBN:            "978-0-123456-47-2",
    Title:           "The Go Programming Language",
    PublicationYear: 2015,
    Copies:          10,
    PublisherID:     1,
}
err := bookService.AddBook(book)
```

#### FindBook(isbn string) (\*Book, error)

Retrieves a book by ISBN.

```go
book, err := bookService.FindBook("978-0-123456-47-2")
if err != nil {
    // Handle error
}
```

#### UpdateBookCopies(isbn string, copies int) error

Updates the number of copies for a book.

```go
err := bookService.UpdateBookCopies("978-0-123456-47-2", 15)
```

#### RemoveBook(isbn string) error

Removes a book from the database.

```go
err := bookService.RemoveBook("978-0-123456-47-2")
```

## Testing

The project includes comprehensive tests covering:

### Unit Tests

- **BookService Operations**: Add, find, update, remove functionality
- **Database Constraints**: Unique ISBN, NOT NULL fields, size limits
- **Model Relationships**: Many-to-many associations
- **Error Handling**: Proper error responses for edge cases

### Test Database Setup

- Uses dedicated PostgreSQL test database
- Automatic schema migration
- Proper cleanup and isolation
- Environment variable configuration

### Running Specific Tests

```bash
# Test book operations
go test -v -run TestAddBook
go test -v -run TestFindBook
go test -v -run TestUpdateBookCopies
go test -v -run TestRemoveBook

# Test database constraints
go test -v -run TestModel_UniqueISBNConstraint
go test -v -run TestModel_NotNullAndSizes

# Test relationships
go test -v -run TestModel_Relationships
```

## Database Constraints

### Unique Constraints

- **ISBN**: Each book must have a unique ISBN
- **Category Name**: Category names must be unique

### Check Constraints

- **Review Rating**: Must be between 1 and 5

### Foreign Key Constraints

- **Book.PublisherID**: References Publisher.ID
- **Many-to-Many Relationships**: Proper junction tables for book-authors and book-categories

## Performance Features

### Connection Pooling

- **Max Idle Connections**: 10
- **Max Open Connections**: 100
- **Connection Lifetime**: 1 hour
- **Automatic Connection Management**

### Database Optimization

- **Indexed Fields**: ISBN, foreign keys
- **Efficient Queries**: Optimized GORM queries
- **Transaction Support**: ACID compliance

## Error Handling

The application implements comprehensive error handling:

- **Database Connection Errors**: Graceful handling with descriptive messages
- **Constraint Violations**: Proper error messages for unique/check constraints
- **Record Not Found**: Clear "book not found" responses
- **Validation Errors**: Field-level validation with helpful messages

## Environment Variables

| Variable          | Description                     | Default                                                                                        |
| ----------------- | ------------------------------- | ---------------------------------------------------------------------------------------------- |
| `GO_DATABASE_URL` | Main database connection string | Required                                                                                       |
| `TEST_PG_DSN`     | Test database connection string | `host=localhost user=postgres password=genio123 dbname=gorm_db_test port=5432 sslmode=disable` |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:

- Create an issue in the repository
- Check the test files for usage examples
- Review the GORM documentation for advanced features

---

**Built with ❤️ using Go and GORM**
