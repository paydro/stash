package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func OpenInEditor(input string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return "", errors.New("Please set the EDITOR environment variable.")
	}

	var tf *os.File
	var err error

	tf, err = ioutil.TempFile("", "item_edit")

	fmt.Fprintf(tf, "%s", input)

	cmd := exec.Command(editor, tf.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadFile(tf.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}
