package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"time"
	"errors"
)

const (
	VERSION = "0.1.0"
)

var (
	dbPath string = "./tmp/stash.db"
	store  *Store
)

type Store struct {
	db *sql.DB
}

func NewStore() (*Store, error) {
	var err error
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Could not open %s\n", dbPath)
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

	row := stmt.QueryRow(id)

	i := &Item{}
	err = row.Scan(&i.Id, &i.Content, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return i, nil
}

type Item struct {
	Id        int
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (i *Item) TrimContent() string {
	return strings.TrimSpace(i.Content)
}

func RunCommand(command string, args []string) {
	switch command {
	case "new":
		err := NewCommand(args)
		if err != nil {
			fmt.Println("Could not create item:", err)
			os.Exit(1)
		}

	case "list":
		err := ListCommand()
		if err != nil {
			fmt.Println("Error finding content:", err)
			os.Exit(1)
		}

	case "update":
		err := UpdateCommand(args)
		if err != nil {
			fmt.Println("Error updating content:", err)
			os.Exit(1)
		}

	case "remove":
		err := RemoveCommand(args)
		if err != nil {
			fmt.Println("Error removing item:", err)
			os.Exit(1)
		}

	case "version":
		fmt.Println("stash v" + VERSION)
		os.Exit(0)

	default:
		fmt.Println("Usage:")
		fmt.Println("\tstash new|list|update|remove ...")
		fmt.Println("")
	}

}

func NewCommand(args []string) error {
	var content string

	if len(args) > 0 {
		content = strings.Join(args, " ")
	} else {
		var err error
		var bytes []byte
		bytes, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		content = string(bytes)
	}

	if err := store.Insert(content); err != nil {
		return err
	}

	return nil
}

func ListCommand() error {
	err := store.FindAll(func(item *Item) {
		fmt.Printf("[%d] %s\n", item.Id, item.Content)
	})
	if err != nil {
		return err
	}

	return nil
}

func UpdateCommand(args []string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("Please set the EDITOR environment variable.")
	}

	var (
		item *Item
		err error
		tf *os.File
	)

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	item, err = store.Find(id)
	if err != nil {
		return err
	}

	tf, err = ioutil.TempFile("", "item_edit")

	fmt.Fprintf(tf, "%s", item.Content)

	cmd := exec.Command(editor, tf.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(tf.Name())
	if err != nil {
		return err
	}

	item.Content = strings.TrimSpace(string(content))
	err = store.Update(item)
	if err != nil {
		return err
	}

	return nil
}

func RemoveCommand(args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	err = store.Remove(id)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	flag.Parse()
	command := flag.Arg(0)

	if command == "" {
		RunCommand("", []string{})
		os.Exit(1)
	}

	var err error
	store, err = NewStore()
	if err != nil {
		log.Fatalln("Failed to initialize store:", err)
	}
	defer store.Close()

	RunCommand(command, flag.Args()[1:])
}
