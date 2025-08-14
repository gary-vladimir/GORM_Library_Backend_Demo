// Package main provides a GORM-based library management system with PostgreSQL backend.
// This application demonstrates database operations including book management,
// review system, and proper database connection handling with connection pooling.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Author represents a book author with biographical information.
type Author struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Biography string `gorm:"type:text"`
	BirthYear int    `gorm:"type:smallint"`
	Books     []Book `gorm:"many2many:book_authors;"`
}

// Book represents a book entity with metadata and relationships.
type Book struct {
	ID              uint      `gorm:"primaryKey"`
	ISBN            string    `gorm:"uniqueIndex;not null;size:13"`
	Title           string    `gorm:"size:200;not null"`
	PublicationYear int       `gorm:"type:smallint"`
	Copies          int       `gorm:"default:0"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	PublisherID     uint
	Publisher       Publisher
	Authors         []Author    `gorm:"many2many:book_authors;"`
	Categories      []Category  `gorm:"many2many:book_categories;"`
}

// BookService handles business logic for book-related operations.
type BookService struct {
	db *gorm.DB
}

// Category represents a book category for classification.
type Category struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null;unique"`
	Books []Book `gorm:"many2many:book_categories;"`
}

// Publisher represents a book publisher with contact information.
type Publisher struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"not null"`
	Address string `gorm:"type:text"`
}

// Review represents a customer review for a product.
type Review struct {
	ID         int  `gorm:"primaryKey"`
	Rating     int  `gorm:"check:rating >= 1 AND rating <= 5"`
	Comment    string
	CustomerID uint
	ProductID  uint
}

// AddBook creates a new book record in the database.
// Returns an error if the operation fails.
func (s *BookService) AddBook(book *Book) error {
	result := s.db.Create(book)
	if result.Error != nil {
		return fmt.Errorf("failed to add book: %w", result.Error)
	}
	return nil
}

// FindBook retrieves a book by its ISBN from the database.
// Returns the book if found, or an error if not found or on database error.
func (s *BookService) FindBook(isbn string) (*Book, error) {
	var book Book
	result := s.db.Where("isbn = ?", isbn).First(&book)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("book not found")
		}
		return nil, fmt.Errorf("error finding book: %w", result.Error)
	}
	return &book, nil
}

// RemoveBook deletes a book from the database by ISBN.
// Returns an error if the book is not found or on database error.
func (s *BookService) RemoveBook(isbn string) error {
	result := s.db.Where("isbn = ?", isbn).Delete(&Book{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove book: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("book not found")
	}
	return nil
}

// setupDB initializes and configures the database connection.
// Sets up connection pooling and returns a configured GORM database instance.
func setupDB() (*gorm.DB, error) {
	dsn := os.Getenv("GO_DATABASE_URL")
	fmt.Println(dsn, "<< dsn")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

// UpdateBookCopies updates the number of copies for a book by ISBN.
// Returns an error if the book is not found or on database error.
func (s *BookService) UpdateBookCopies(isbn string, copies int) error {
	result := s.db.Model(&Book{}).
		Where("isbn = ?", isbn).
		Update("copies", copies)

	if result.Error != nil {
		return fmt.Errorf("failed to update copies: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("book not found")
	}
	return nil
}

// main is the entry point of the application.
// It sets up the database, migrates schemas, and demonstrates
// the book service functionality with sample data.
func main() {
	db, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database")

	// Get the underlying sql.DB and defer its close here
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if err := db.AutoMigrate(&Review{}, &Book{}, &Author{}, &Publisher{}, &Category{}); err != nil {
		log.Fatal("Error migrating database: ", err)
	}
	log.Println("Database migrated")

	// Create a book service instance
	bookService := &BookService{db: db}

	// Test the book service
	book := &Book{
		ISBN:            "978-0-123456-47-2",
		Title:           "The Go Programming Language",
		PublicationYear: 2015,
		Copies:          10,
		PublisherID:     1,
	}

	if err := bookService.AddBook(book); err != nil {
		log.Printf("Failed to add book: %v", err)
	} else {
		fmt.Println("Book added successfully!")
	}

	// Test finding the book
	foundBook, err := bookService.FindBook("978-0-123456-47-2")
	if err != nil {
		log.Printf("Failed to find book: %v", err)
	} else {
		fmt.Printf("Found book: %s by ISBN %s\n", foundBook.Title, foundBook.ISBN)
	}

	// Test updating copies
	if err := bookService.UpdateBookCopies("978-0-123456-47-2", 15); err != nil {
		log.Printf("Failed to update copies: %v", err)
	} else {
		fmt.Println("Book copies updated successfully!")
	}

	review := Review{
		Rating:     5,
		Comment:    "Great product!",
		CustomerID: 1,
		ProductID:  1,
	}
	result := db.Create(&review)
	fmt.Printf("Review created? %v\n", result.Error == nil)
}
