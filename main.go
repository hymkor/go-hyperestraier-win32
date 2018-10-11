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
var estCondSetOptions = estraier.NewProc("est_cond_set_options")
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
	_ESTENOERR  EstError = iota
	_ESTEINVAL                  /* invalid argument */
	_ESTEACCES                  /* access forbidden */
	_ESTELOCK                   /* lock failure */
	_ESTEDB                     /* database problem */
	_ESTEIO                     /* I/O problem */
	_ESTENOITEM                 /* no item */
	_ESTEMISC   EstError = 9999 /* miscellaneous */
)

func (this *EstError) address() uintptr {
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

func lastError(ecode EstError) error {
	if ecode == _ESTENOERR {
		return nil
	}
	msg, _, _ := estErrMsg.Call(uintptr(ecode))
	return errors.New(cstr2string(msg))
}

func (db Database) Close() error {
	ecode := _ESTEMISC
	estClose.Call(uintptr(db), ecode.address())
	return lastError(ecode)
}

func address(s string) uintptr {
	bin := []byte(s)
	return uintptr(unsafe.Pointer(&bin[0]))
}

func Open(dbPath string) (Database, error) {
	ecode := _ESTEMISC
	db, _, _ := estOpen.Call(
		address(dbPath),
		forRead,
		ecode.address())

	return Database(db), lastError(ecode)
}

type Cond uintptr

func newCond() Cond {
	cond, _, _ := estCondNew.Call()
	return Cond(cond)
}

func (cond Cond) setPhrase(expr string) {
	estCondSetPhrase.Call(uintptr(cond), address(expr))
}

func (cond Cond) setOptions(options uintptr) {
	estCondSetOptions.Call(uintptr(cond), options)
}

func (cond Cond) Close() {
	estCondDelete.Call(uintptr(cond))
}

type DocId int

func (db Database) search(cond Cond) []DocId {
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

type Phrase string

func (phrase Phrase) Join(cond Cond) {
	cond.setPhrase(string(phrase))
}

type Option uintptr

const (
	Sure   Option = 1 << 0  /* check every N-gram key */
	Usual  Option = 1 << 1  /* check N-gram keys skipping by one */
	Fast   Option = 1 << 2  /* check N-gram keys skipping by two */
	Agito  Option = 1 << 3  /* check N-gram keys skipping by three */
	Noidf  Option = 1 << 4  /* without TF-IDF tuning */
	Simple Option = 1 << 10 /* with the simplified phrase */
	Rough  Option = 1 << 11 /* with the rough phrase */
	Union  Option = 1 << 15 /* with the union phrase */
	Isect  Option = 1 << 16 /* with the intersection phrase */
	Scfb   Option = 1 << 30 /* feed back scores (for debug) */
)

func (option Option) Join(cond Cond) {
	cond.setOptions(uintptr(option))
}

type ICond interface {
	Join(Cond)
}

func (db Database) Search(conds ...ICond) []DocId {
	cond := newCond()
	for _, c1 := range conds {
		c1.Join(cond)
	}
	rc := db.search(cond)
	cond.Close()
	return rc
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
	value, _, _ := estDocAttr.Call(uintptr(doc), address(attr))
	return cstr2string(value)
}

func (doc Doc) Uri() string {
	return doc.Attr("@uri")
}
