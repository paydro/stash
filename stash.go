package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"fmt"
	"os"
	"io/ioutil"
	"strings"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Stash something.")

	args := os.Args[1:]
	log.Printf("Args: %#v\n", args)

	db, err := sql.Open("sqlite3", "./tmp/stash.db")
	if err != nil {
		log.Fatalln("Could not open ./tmp/stash.db")
	}
	defer db.Close()

	if err := createItemsTable(db); err != nil {
		log.Println("Could not create items table:")
		log.Fatalln(err)
	}

	var item string
	if len(args) > 0 {
		item = strings.Join(args, " ")

	} else {
		fmt.Println("What would you like to store?")
		var err error
		var bytes []byte
		bytes, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalln(err)
		}

		item = string(bytes)
	}

	log.Printf("Item to insert:\n%s", item)

	if err := insertItem(db, item); err != nil {
		log.Println("Could not insert item:")
		log.Fatalln(err)
	}

	log.Println("inserted!")
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

func insertItem(db *sql.DB, item string) error {
	stmt := "INSERT INTO items (desc) VALUES (?)"

	preparedStmt, err := db.Prepare(stmt)
	if err != nil {
		return err
	}
	log.Println("about to insert item", item)
	if _, err := preparedStmt.Exec(item); err != nil {
		return err
	}
	return nil
}
