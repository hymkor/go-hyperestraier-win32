package estraier

import (
	"errors"
	"strings"
	"syscall"
	"unsafe"
)

var estraier = syscall.NewLazyDLL("estraier.dll")
var estOpen = estraier.NewProc("est_db_open")
var estClose = estraier.NewProc("est_db_close")
var estCondNew = estraier.NewProc("est_cond_new")
var estCondSetPhrase = estraier.NewProc("est_cond_set_phrase")
var estCondDelete = estraier.NewProc("est_cond_delete")
var estDbSearch = estraier.NewProc("est_db_search")
var estDbGetDoc = estraier.NewProc("est_db_get_doc")
var estDocDelete = estraier.NewProc("est_doc_delete")
var estDocAttr = estraier.NewProc("est_doc_attr")
var estErrMsg = estraier.NewProc("est_err_msg")

var msvcrt = syscall.NewLazyDLL("msvcrt.dll")
var free = msvcrt.NewProc("free")
var memcpy = msvcrt.NewProc("memcpy")

type Database uintptr

const (
	forRead = 1
)

type EstError uint32

const (
	ESTENOERR  EstError = iota
	ESTEINVAL                  /* invalid argument */
	ESTEACCES                  /* access forbidden */
	ESTELOCK                   /* lock failure */
	ESTEDB                     /* database problem */
	ESTEIO                     /* I/O problem */
	ESTENOITEM                 /* no item */
	ESTEMISC   EstError = 9999 /* miscellaneous */
)

func (this *EstError) Address() uintptr {
	return uintptr(unsafe.Pointer(this))
}

func cstr2string(cstr uintptr) string {
	if cstr == 0 {
		return ""
	}
	var buffer strings.Builder
	for {
		c := *(*byte)(unsafe.Pointer(cstr))
		if c == 0 {
			break
		}
		buffer.WriteByte(c)
		cstr++
	}
	return buffer.String()
}

func LastError(ecode EstError) error {
	if ecode == ESTENOERR {
		return nil
	}
	msg, _, _ := estErrMsg.Call(uintptr(ecode))
	return errors.New(cstr2string(msg))
}

func (db Database) Close() error {
	ecode := ESTEMISC
	estClose.Call(uintptr(db), ecode.Address())
	return LastError(ecode)
}

func Address(s string) uintptr {
	bin := []byte(s)
	return uintptr(unsafe.Pointer(&bin[0]))
}

func NewDatabase(dbPath string) (Database, error) {
	ecode := ESTEMISC
	db, _, _ := estOpen.Call(
		Address(dbPath),
		forRead,
		ecode.Address())

	return Database(db), LastError(ecode)
}

type Cond uintptr

func NewCond() Cond {
	cond, _, _ := estCondNew.Call()
	return Cond(cond)
}

func (cond Cond) SetPhrase(expr string) {
	estCondSetPhrase.Call(uintptr(cond), Address(expr))
}

func (cond Cond) Close() {
	estCondDelete.Call(uintptr(cond))
}

type DocId int

func (db Database) Search(cond Cond) []DocId {
	var num int32

	pages, _, _ := estDbSearch.Call(
		uintptr(db),
		uintptr(cond),
		uintptr(unsafe.Pointer(&num)),
		0)

	if num <= 0 {
		return []DocId{}
	}

	result := make([]DocId, num)
	memcpy.Call(uintptr(unsafe.Pointer(&result[0])), pages, uintptr(4*num))
	free.Call(pages)
	return result
}

type Doc uintptr

func (db Database) GetDoc(id DocId) Doc {
	doc, _, _ := estDbGetDoc.Call(
		uintptr(db),
		uintptr(id),
		0)

	return Doc(doc)
}

func (doc Doc) Close() {
	if doc != 0 {
		estDocDelete.Call(uintptr(doc))
	}
}

func (doc Doc) Attr(attr string) string {
	if doc == 0 {
		return ""
	}
	value, _, _ := estDocAttr.Call(uintptr(doc), Address(attr))
	return cstr2string(value)
}

func (doc Doc) Uri() string {
	return doc.Attr("@uri")
}
