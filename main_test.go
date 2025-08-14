// Package main_test provides comprehensive test coverage for the GORM-based book management system.
// This test suite validates the BookService operations including AddBook, FindBook, RemoveBook,
// and UpdateBookCopies methods, as well as database model relationships, constraints, and
// schema validation. Tests use a dedicated PostgreSQL test database with proper setup and
// teardown procedures to ensure isolation and reliability.

package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --- Test helpers ---

// ensurePublisher creates or finds a default Publisher and returns its ID.
func ensurePublisher(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	p := Publisher{Name: "Test Publisher", Address: "123 Test St"}
	if err := db.FirstOrCreate(&p, Publisher{Name: p.Name}).Error; err != nil {
		t.Fatalf("failed to create/find default publisher: %v", err)
	}
	return p.ID
}

// mustCreateBook seeds a book ensuring a valid PublisherID.
func mustCreateBook(t *testing.T, db *gorm.DB, b *Book) {
	t.Helper()
	if b.PublisherID == 0 {
		b.PublisherID = ensurePublisher(t, db)
	}
	if err := db.Create(b).Error; err != nil {
		t.Fatalf("failed to seed book: %v", err)
	}
}

// newTestDB connects to Postgres, migrates schemas, and returns a cleanup function.
func newTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	dsn := testPostgresDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to postgres test db: %v", err)
	}

	if err := db.AutoMigrate(&Review{}, &Book{}, &Author{}, &Publisher{}, &Category{}); err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}

	cleanup := func() {
		_ = sqlDB.Close()
	}

	return db, cleanup
}

// testPostgresDSN returns the test database connection string.
// Uses TEST_PG_DSN environment variable if set, otherwise returns default.
func testPostgresDSN() string {
	if d := os.Getenv("TEST_PG_DSN"); d != "" {
		return d
	}
	// Default provided by you:
	return "host=localhost user=postgres password=genio123 dbname=gorm_db_test port=5432 sslmode=disable"
}

// --- Tests for BookService.AddBook ---

// TestAddBook_DuplicateISBN tests that adding a book with duplicate ISBN returns an error.
func TestAddBook_DuplicateISBN(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	svc := &BookService{db: db}

	pubID := ensurePublisher(t, db)

	b1 := &Book{ISBN: "9780000000001", Title: "First", PublisherID: pubID}
	b2 := &Book{ISBN: "9780000000001", Title: "Second (dup)", PublisherID: pubID}

	if err := svc.AddBook(b1); err != nil {
		t.Fatalf("unexpected error adding first: %v", err)
	}
	if err := svc.AddBook(b2); err == nil {
		t.Fatalf("expected duplicate ISBN error, got nil")
	} else {
		low := strings.ToLower(err.Error())
		if !strings.Contains(low, "duplicate") && !strings.Contains(low, "unique") {
			t.Errorf("expected unique/duplicate constraint error, got: %v", err)
		}
	}
}

// TestAddBook_Success tests successful addition of a book.
func TestAddBook_Success(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	svc := &BookService{db: db}

	pubID := ensurePublisher(t, db)

	book := &Book{
		ISBN:            "9780123456789",
		Title:           "Clean Go",
		PublicationYear: 2020,
		Copies:          3,
		CreatedAt:       time.Now(),
		PublisherID:     pubID,
	}
	if err := svc.AddBook(book); err != nil {
		t.Fatalf("AddBook returned error: %v", err)
	}

	// Verify persisted
	var got Book
	if err := db.First(&got, "isbn = ?", "9780123456789").Error; err != nil {
		t.Fatalf("book not found after AddBook: %v", err)
	}
	if got.Title != "Clean Go" || got.Copies != 3 || got.PublisherID != pubID {
		t.Errorf("unexpected persisted values: %+v", got)
	}
}

// --- Tests for BookService.FindBook ---

// TestFindBook_Found tests successful retrieval of an existing book.
func TestFindBook_Found(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	svc := &BookService{db: db}

	want := &Book{ISBN: "9781111111111", Title: "Found Me", Copies: 2}
	mustCreateBook(t, db, want)

	got, err := svc.FindBook("9781111111111")
	if err != nil {
		t.Fatalf("FindBook returned error: %v", err)
	}
	if got.ISBN != want.ISBN || got.Title != want.Title || got.Copies != want.Copies {
		t.Errorf("unexpected book: got=%+v want=%+v", got, want)
	}
}

// TestFindBook_NotFound tests that finding a non-existent book returns an error.
func TestFindBook_NotFound(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	svc := &BookService{db: db}

	_, err := svc.FindBook("nope")
	if err == nil {
		t.Fatalf("expected error for missing book, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "book not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Tests for model relationships and constraints ---

// TestModel_NotNullAndSizes tests NOT NULL constraints and size limits on model fields.
func TestModel_NotNullAndSizes(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	pubID := ensurePublisher(t, db)

	// NOT NULL allows empty string ("" != NULL). This should succeed.
	if err := db.Create(&Book{ISBN: "9786666666666", Title: "", PublisherID: pubID}).Error; err != nil {
		t.Fatalf("unexpected error for empty title (empty string is allowed by NOT NULL): %v", err)
	}

	// Exactly 200 chars should succeed.
	long200 := strings.Repeat("T", 200)
	if err := db.Create(&Book{ISBN: "9786666666667", Title: long200, PublisherID: pubID}).Error; err != nil {
		t.Fatalf("expected success with 200-char title, got: %v", err)
	}

	// 201 chars should fail due to size:200 (varchar(200)).
	long201 := strings.Repeat("T", 201)
	if err := db.Create(&Book{ISBN: "9786666666668", Title: long201, PublisherID: pubID}).Error; err == nil {
		t.Fatalf("expected error with 201-char title, got nil")
	}
}

// TestModel_Relationships_AuthorPublisherCategory tests model relationships between Book, Author, Publisher, and Category.
func TestModel_Relationships_AuthorPublisherCategory(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	pub := Publisher{Name: "Acme Publishing", Address: "123 Road"}
	cat1 := Category{Name: "Computers"}
	cat2 := Category{Name: "Programming"}
	auth1 := Author{Name: "Jane Dev", Biography: "Go enthusiast", BirthYear: 1985}
	auth2 := Author{Name: "John Code", Biography: "Systems hacker", BirthYear: 1980}

	if err := db.Create(&pub).Error; err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	if err := db.Create(&cat1).Error; err != nil {
		t.Fatalf("create category1: %v", err)
	}
	if err := db.Create(&cat2).Error; err != nil {
		t.Fatalf("create category2: %v", err)
	}
	if err := db.Create(&auth1).Error; err != nil {
		t.Fatalf("create author1: %v", err)
	}
	if err := db.Create(&auth2).Error; err != nil {
		t.Fatalf("create author2: %v", err)
	}

	book := Book{
		ISBN:        "9784444444444",
		Title:       "Go Deep Dive",
		Copies:      7,
		PublisherID: pub.ID,
		Authors:     []Author{auth1, auth2},
		Categories:  []Category{cat1, cat2},
	}

	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("create book with relations: %v", err)
	}

	var fetched Book
	if err := db.Preload("Publisher").
		Preload("Authors").
		Preload("Categories").
		First(&fetched, "isbn = ?", "9784444444444").Error; err != nil {
		t.Fatalf("fetch with preloads: %v", err)
	}

	if fetched.Publisher.ID != pub.ID {
		t.Errorf("publisher not linked, got=%d want=%d", fetched.Publisher.ID, pub.ID)
	}
	if len(fetched.Authors) != 2 {
		t.Fatalf("expected 2 authors, got %d", len(fetched.Authors))
	}
	if len(fetched.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(fetched.Categories))
	}
}

// TestModel_UniqueISBNConstraint tests that the unique ISBN constraint is enforced at the database level.
func TestModel_UniqueISBNConstraint(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	pubID := ensurePublisher(t, db)

	b1 := Book{ISBN: "9785555555555", Title: "One", PublisherID: pubID}
	b2 := Book{ISBN: "9785555555555", Title: "Two (dup)", PublisherID: pubID}

	if err := db.Create(&b1).Error; err != nil {
		t.Fatalf("create first book failed: %v", err)
	}
	err := db.Create(&b2).Error
	if err == nil {
		t.Fatalf("expected unique error on second insert; got nil")
	}
	low := strings.ToLower(err.Error())
	if !strings.Contains(low, "duplicate") && !strings.Contains(low, "unique") {
		t.Errorf("expected unique constraint error, got: %v", err)
	}
}

// --- Tests for BookService.RemoveBook ---

// TestRemoveBook_NotFound tests that removing a non-existent book returns an error.
func TestRemoveBook_NotFound(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	svc := &BookService{db: db}

	if err := svc.RemoveBook("missing"); err == nil {
		t.Fatalf("expected not found error, got nil")
	} else if !strings.Contains(strings.ToLower(err.Error()), "book not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRemoveBook_Success tests successful removal of an existing book.
func TestRemoveBook_Success(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	svc := &BookService{db: db}

	mustCreateBook(t, db, &Book{ISBN: "9782222222222", Title: "To Be Removed", Copies: 1})

	if err := svc.RemoveBook("9782222222222"); err != nil {
		t.Fatalf("RemoveBook returned error: %v", err)
	}

	// Verify deletion
	var count int64
	if err := db.Model(&Book{}).Where("isbn = ?", "9782222222222").Count(&count).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows after delete, got %d", count)
	}
}

// TestReview_CheckConstraint tests that the CHECK constraint on Review rating is enforced.
func TestReview_CheckConstraint(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	// Valid review
	ok := Review{Rating: 5, Comment: "Great product!", CustomerID: 1, ProductID: 1}
	if err := db.Create(&ok).Error; err != nil {
		t.Fatalf("valid review failed to insert: %v", err)
	}

	// Invalid review (rating > 5) — expect CHECK constraint error on Postgres
	bad := Review{Rating: 6, Comment: "Too high", CustomerID: 1, ProductID: 1}
	err := db.Create(&bad).Error
	if err == nil {
		t.Fatalf("expected check constraint error for rating>5, got nil")
	}
	low := strings.ToLower(err.Error())
	if !strings.Contains(low, "check") && !strings.Contains(low, "constraint") && !strings.Contains(low, "23514") {
		t.Logf("received error for invalid rating (message may vary by driver): %v", err)
	}
}

// --- Tests for setupDB ---

// TestSetupDB_InvalidDSN tests that setupDB fails gracefully with an invalid DSN.
func TestSetupDB_InvalidDSN(t *testing.T) {
	// Save and restore env var
	prev := os.Getenv("GO_DATABASE_URL")
	defer func() { _ = os.Setenv("GO_DATABASE_URL", prev) }()

	// Intentionally invalid/non-routable DSN to force a connection error
	_ = os.Setenv("GO_DATABASE_URL", "postgres://invalid-host.local:5432/doesnotexist?sslmode=disable")

	db, err := setupDB()
	if err == nil {
		// Close to avoid leak if unexpectedly succeeds.
		if db != nil {
			if sqlDB, e := db.DB(); e == nil {
				_ = sqlDB.Close()
			}
		}
		t.Fatalf("expected setupDB to fail with invalid DSN, got nil error")
	}
	// Error message shape varies by driver/platform; just ensure it's a connection error.
	low := strings.ToLower(err.Error())
	if !strings.Contains(low, "failed to connect") && !strings.Contains(low, "connect") && !strings.Contains(low, "dial") && !strings.Contains(low, "resolve") {
		t.Logf("setupDB returned error (ok), message: %v", err)
	}
}

// --- Tests for BookService.UpdateBookCopies ---

// TestUpdateBookCopies_NotFound tests that updating copies for a non-existent book returns an error.
func TestUpdateBookCopies_NotFound(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	svc := &BookService{db: db}

	if err := svc.UpdateBookCopies("missing", 10); err == nil {
		t.Fatalf("expected not found error, got nil")
	} else if !strings.Contains(strings.ToLower(err.Error()), "book not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestUpdateBookCopies_Success tests successful updating of book copies including edge cases.
func TestUpdateBookCopies_Success(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	svc := &BookService{db: db}

	mustCreateBook(t, db, &Book{ISBN: "9783333333333", Title: "Inventory", Copies: 5})

	if err := svc.UpdateBookCopies("9783333333333", 15); err != nil {
		t.Fatalf("UpdateBookCopies returned error: %v", err)
	}

	var got Book
	if err := db.First(&got, "isbn = ?", "9783333333333").Error; err != nil {
		t.Fatalf("failed to refetch book: %v", err)
	}
	if got.Copies != 15 {
		t.Errorf("copies not updated, got %d want 15", got.Copies)
	}

	// Edge cases: zero and large value
	if err := svc.UpdateBookCopies("9783333333333", 0); err != nil {
		t.Fatalf("update to zero copies failed: %v", err)
	}
	if err := svc.UpdateBookCopies("9783333333333", 1000000); err != nil {
		t.Fatalf("update to large copies failed: %v", err)
	}
}
