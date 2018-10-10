package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	est "github.com/zetamatta/experimental/go-estraier"
)

func search(args []string) error {
	if len(args) < 2 {
		return errors.New("too few arguments")
	}
	db, err := est.Open(args[0])
	if err != nil {
		return err
	}
	pages := db.Search(est.Phrase(strings.Join(args[1:], " ")))

	for i, page1 := range pages {
		doc := db.GetDoc(page1)
		fmt.Printf("(%d)\t%d\t%s\n", i+1, page1, doc.Uri())
		doc.Close()
	}
	return db.Close()
}

func id2uri(args []string) error {
	if len(args) < 2 {
		return errors.New("too few arguments")
	}
	db, err := est.Open(args[0])
	if err != nil {
		return err
	}
	for i, idStr := range args[1:] {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", idStr, err)
		} else {
			doc := db.GetDoc(est.DocId(id))
			fmt.Printf("(%d)\t%d\t%s\n", i+1, id, doc.Uri())
			doc.Close()
		}
	}
	return db.Close()
}

func main1() error {
	if len(os.Args) < 2 {
		return errors.New("too few arguments")
	}
	switch os.Args[1] {
	case "search":
		return search(os.Args[2:])
	case "id2uri":
		return id2uri(os.Args[2:])
	}
	return nil
}

func main() {
	if err := main1(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
