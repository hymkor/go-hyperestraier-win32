[![GoDoc](https://godoc.org/github.com/zetamatta/go-hyperestraier-win32?status.svg)](https://godoc.org/github.com/zetamatta/go-hyperestraier-win32)

go-hyperestraier-win32
======================

This is the dll-wrapper of HyperEstraier for the programming language Go.

* It supports GOOS=windows and GOARCH=386.
* It requires these libraries.
    * estraier.dll
    * libgnurx-0.dll
    * libiconv-2.dll
    * pthreadGC2.dll
    * qdbm.dll
    * zlib.dll
    * zlib1.dll

Sample
======

[Full source](https://github.com/zetamatta/go-hyperestraier-win32/tree/master/goestcmd)

```
import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	est "github.com/zetamatta/go-hyperestraier-win32"
)

func search(args []string) error {
	if len(args) < 2 {
		return errors.New("too few arguments")
	}
	db, err := est.Open(args[0])
	if err != nil {
		return err
	}
	pages := db.Search(est.Phrase(strings.Join(args[1:], " ")), est.Simple)

	for i, page1 := range pages {
		doc := db.GetDoc(page1)
		fmt.Printf("(%d)\t%d\t%s\n", i+1, page1, doc.URI())
		doc.Close()
	}
	return db.Close()
}
```
