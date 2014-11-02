package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strings"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Stash something.")

	args := os.Args[1:]
	log.Printf("Args: %#v\n", args)

	db, err := sql.Open("sqlite3", "stash.db")
	if err != nil {
		log.Fatalln("Could not open ./stash.db")
	}
	defer db.Close()

	if err := createItemsTable(db); err != nil {
		log.Println("Could not create items table:")
		log.Fatalln(err)
	}

	if err := insertItem(db, args); err != nil {
		log.Println("Could not insert item:")
		log.Fatalln(err)
	}


}

func createItemsTable(db *sql.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER NOT NULL PRIMARY KEY,
		desc TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	if _, err := db.Exec(stmt); err != nil {
		return err
	}

	return nil
}

func insertItem(db *sql.DB, args []string) error {
	item := strings.Join(args, " ")

	stmt := "INSERT INTO items (desc) VALUES (?)"

	preparedStmt, err := db.Prepare(stmt)
	if err != nil {
		return err
	}


	if _, err := preparedStmt.Exec(item); err != nil {
		return err
	}
	return nil
}
