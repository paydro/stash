package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION = "0.1.0"
)

var (
	dbName string = "stash.db"
	store  *Store
)

type Item struct {
	Id        int
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (i *Item) TrimContent() string {
	return strings.TrimSpace(i.Content)
}

func (i *Item) ContentTitle() string {
	buf := bytes.NewBuffer([]byte(i.Content))
	line, err := buf.ReadString('\n')
	if err == nil {
		line = line[:len(line)-1]
	}

	if len(line) > 50 {
		return line[0:50]
	} else {
		return line
	}
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

	case "show":
		err := ShowCommand(args)
		if err != nil {
			fmt.Println("Error finding content:", err)
			os.Exit(1)
		}

	case "edit":
		err := EditCommand(args)
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
		fmt.Println("\tstash new|list|show|edit|remove|version ...")
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

func ShowCommand(args []string) error {
	var (
		item *Item
		err error
	)

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	item, err = store.Find(id)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", item.Content)

	return nil
}

func ListCommand() error {
	err := store.FindAll(func(item *Item) {
		fmt.Printf("%4d| %s\n", item.Id, strings.Replace(item.ContentTitle(), "\n", " ", -1))
	})
	if err != nil {
		return err
	}

	return nil
}

func EditCommand(args []string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("Please set the EDITOR environment variable.")
	}

	var (
		item *Item
		err  error
		tf   *os.File
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
