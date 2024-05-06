package sqlite

import (
	"ReadLaterBot/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type Storage struct {
	db *sql.DB
}

// New creates new SQLite storage.
func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) PickAll(ctx context.Context, userName string) ([]*storage.Page, error) {
	// Query to select URL and title from the pages table for the specified user name
	q := `SELECT url, title FROM pages WHERE user_name = ?`

	// Execute the query with the provided context and user name
	rows, err := s.db.QueryContext(ctx, q, userName)
	if err != nil {
		return nil, fmt.Errorf("can't load data from db: %w", err)
	}

	// Ensure rows is closed after use to release resources
	defer rows.Close()

	// Declare a slice to hold the pages
	var pages []*storage.Page

	// Iterate through the rows returned by the query
	for rows.Next() {
		var url string
		var title string

		// Scan the current row's data into the variables
		if err := rows.Scan(&url, &title); err != nil {
			return nil, fmt.Errorf("can't scan row: %w", err)
		}

		// Create a new Page struct and append it to the pages slice
		page := &storage.Page{UserName: userName, URL: url, Title: title}
		pages = append(pages, page)
		log.Printf("User: %s has link: %s with title: %s", page.UserName, page.URL, page.Title)
	}

	// Check for any errors that may have occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Return the slice of pages
	return pages, nil
}

// Save saves page to storage.
func (s *Storage) Save(ctx context.Context, p *storage.Page) error {
	q := `INSERT INTO pages (url, user_name, title) VALUES (?, ?, ?)`

	if _, err := s.db.ExecContext(ctx, q, p.URL, p.UserName, p.Title); err != nil {
		return fmt.Errorf("can't save page: %w", err)
	}

	return nil
}

// PickRandom picks random page from storage.
func (s *Storage) PickRandom(ctx context.Context, userName string) (*storage.Page, error) {
	q := `SELECT url, title FROM pages WHERE user_name = ? ORDER BY RANDOM() LIMIT 1`

	var url string
	var title string

	err := s.db.QueryRowContext(ctx, q, userName).Scan(&url, &title)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, fmt.Errorf("can't pick random page: %w", err)
	}

	return &storage.Page{
		URL:      url,
		UserName: userName,
		Title:    title,
	}, nil
}

// RemoveByPage removes page from storage.
func (s *Storage) RemoveByPage(ctx context.Context, page *storage.Page) error {
	q := `DELETE FROM pages WHERE url = ? AND user_name = ?`
	if _, err := s.db.ExecContext(ctx, q, page.URL, page.UserName); err != nil {
		return fmt.Errorf("can't remove page: %w", err)
	}

	return nil
}

func (s *Storage) RemoveByIndex(ctx context.Context, userName string, index int) error {
	// Отримати всі сторінки користувача
	pages, err := s.PickAll(ctx, userName)
	if err != nil {
		return fmt.Errorf("can't retrieve pages: %w", err)
	}

	// Перевірити, чи індекс знаходиться в межах списку
	if index < 0 || index >= len(pages) {
		return fmt.Errorf("index out of range: %d", index)
	}

	// Отримати сторінку за індексом
	pageToRemove := pages[index]

	// Видалити сторінку з бази даних
	err = s.RemoveByPage(ctx, pageToRemove)
	if err != nil {
		return fmt.Errorf("can't remove page: %w", err)
	}

	return nil
}

// IsExists checks if page exists in storage.
func (s *Storage) IsExists(ctx context.Context, page *storage.Page) (bool, error) {
	q := `SELECT COUNT(*) FROM pages WHERE url = ? AND user_name = ?`

	var count int

	if err := s.db.QueryRowContext(ctx, q, page.URL, page.UserName).Scan(&count); err != nil {
		return false, fmt.Errorf("can't check if page exists: %w", err)
	}

	return count > 0, nil
}

func (s *Storage) Init(ctx context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS pages (url TEXT, user_name TEXT, title TEXT)`

	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}

	return nil
}

/*func (s *Storage) Init(ctx context.Context) error {
	q := `
    CREATE TABLE IF NOT EXISTS pages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT NOT NULL,
        user_name TEXT NOT NULL,
        title VARCHAR(80)
    )`

	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("can't initialize database tables: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}
*/
