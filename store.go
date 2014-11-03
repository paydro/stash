package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strings"
)

type Store struct {
	db *sql.DB
}

func NewStore() (*Store, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/tmp"
	}

	configPath := homeDir + "/.stash"
	os.MkdirAll(configPath, 0755)

	var err error
	db, err := sql.Open("sqlite3", configPath+"/"+dbName)
	if err != nil {
		log.Fatalf("Could not open %s\n", configPath+"/"+dbName)
	}

	s := &Store{db: db}
	if err := s.setup(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) setup() error {
	stmt := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER NOT NULL PRIMARY KEY,
		content TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	if _, err := s.db.Exec(stmt); err != nil {
		return err
	}

	return nil
}

func (s *Store) Insert(content string) error {
	sql := "INSERT INTO items (content) VALUES (?)"

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(strings.TrimSpace(content)); err != nil {
		return err
	}
	return nil
}

func (s *Store) Update(i *Item) error {
	sql := "UPDATE items SET content = ?, updated_at = datetime('now') WHERE id = ?"
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(i.Content, i.Id); err != nil {
		return err
	}

	log.Printf("Updated Content to: %s\n", i.Content)

	return nil
}

func (s *Store) Remove(id int) error {
	sql := "DELETE FROM items WHERE id = ?"
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(id); err != nil {
		return err
	}

	return nil
}

func (s *Store) FindAll(yield func(*Item)) error {
	rows, err := s.db.Query(`
		SELECT id, content, created_at, updated_at
		FROM items
		ORDER BY created_at
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := Item{}
	for rows.Next() {
		if err := rows.Scan(&i.Id, &i.Content, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return err
		}

		yield(&i)
	}

	return nil
}

func (s *Store) Find(id int) (*Item, error) {
	query := `
		SELECT id, content, created_at, updated_at
		FROM items
		WHERE id = ?
	`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)

	i := &Item{}
	err = row.Scan(&i.Id, &i.Content, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return i, nil
}
