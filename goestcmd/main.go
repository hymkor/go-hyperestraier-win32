package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	est "github.com/zetamatta/go-hyperestraier-win32"
)

var estIndex = flag.String("db", "", "Set directory of index")
var condAddAttr = flag.String("a", "", "Add attribute as condition")

func search(args []string) error {
	var _estIndex string
	if *estIndex != "" {
		_estIndex = *estIndex
	} else {
		_estIndex = args[0]
		args = args[1:]
	}
	if len(args) < 1 {
		return errors.New("too few arguments")
	}
	db, err := est.Open(_estIndex)
	if err != nil {
		return err
	}
	conds := []est.Condition{est.Phrase(strings.Join(args, " "))}
	if *condAddAttr != "" {
		conds = append(conds, est.CondAttr(*condAddAttr))
	}
	pages := db.Search(conds...)

	for i, page1 := range pages {
		doc := db.GetDoc(page1)
		fmt.Printf("(%d)\t%d\t%s\n", i+1, page1, doc.URI())
		doc.Close()
	}
	return db.Close()
}

func id2uri(args []string) error {
	var _estIndex string
	if *estIndex != "" {
		_estIndex = *estIndex
	} else {
		_estIndex = args[0]
		args = args[1:]
	}
	if len(args) < 1 {
		return errors.New("too few arguments")
	}
	db, err := est.Open(_estIndex)
	if err != nil {
		return err
	}
	for i, idStr := range args {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", idStr, err)
		} else {
			doc := db.GetDoc(est.DocID(id))
			fmt.Printf("(%d)\t%d\t%s\n", i+1, id, doc.URI())
			doc.Close()
		}
	}
	return db.Close()
}

func main1() error {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		return errors.New("too few arguments")
	}
	switch args[0] {
	case "search":
		return search(args[1:])
	case "id2uri":
		return id2uri(args[1:])
	}
	return nil
}

func main() {
	if err := main1(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
