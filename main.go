package estraier

import (
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
var estCondAddAttr = estraier.NewProc("est_cond_add_attr")
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
	// ESTENOERR means no error
	ESTENOERR EstError = iota
	// ESTEINVAL means invalid argument
	ESTEINVAL
	// ESTEACCES means access forbidden
	ESTEACCES
	// ESTELOCK means lock failure
	ESTELOCK
	// ESTEDB means database problem
	ESTEDB
	// ESTEIO means I/O problem
	ESTEIO
	// ESTENOITEM means no item
	ESTENOITEM
	// ESTEMISC means miscellaneous
	ESTEMISC EstError = 9999
)

func (ecode EstError) Error() string {
	msg, _, _ := estErrMsg.Call(uintptr(ecode))
	return cstr2string(msg)
}

func (ecode *EstError) address() uintptr {
	return uintptr(unsafe.Pointer(ecode))
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
	if ecode == ESTENOERR {
		return nil
	}
	return ecode
}

func (db Database) Close() error {
	ecode := ESTEMISC
	estClose.Call(uintptr(db), ecode.address())
	return lastError(ecode)
}

func address(s string) uintptr {
	bin := []byte(s)
	return uintptr(unsafe.Pointer(&bin[0]))
}

func Open(dbPath string) (Database, error) {
	ecode := ESTEMISC
	db, _, _ := estOpen.Call(
		address(dbPath),
		forRead,
		ecode.address())

	return Database(db), lastError(ecode)
}

type ConditionsContainer uintptr

func newCond() ConditionsContainer {
	cond, _, _ := estCondNew.Call()
	return ConditionsContainer(cond)
}

func (cond ConditionsContainer) setPhrase(expr string) {
	if len(expr) > 0 {
		estCondSetPhrase.Call(uintptr(cond), address(expr))
	}
}

func (cond ConditionsContainer) setOptions(options uintptr) {
	estCondSetOptions.Call(uintptr(cond), options)
}

func (cond ConditionsContainer) addAttr(options string) {
	estCondAddAttr.Call(uintptr(cond), address(options))
}

func (cond ConditionsContainer) close() {
	estCondDelete.Call(uintptr(cond))
}

type DocID int

func (db Database) search(cond ConditionsContainer) []DocID {
	var num int32

	pages, _, _ := estDbSearch.Call(
		uintptr(db),
		uintptr(cond),
		uintptr(unsafe.Pointer(&num)),
		0)

	if num <= 0 {
		return []DocID{}
	}

	result := make([]DocID, num)
	memcpy.Call(uintptr(unsafe.Pointer(&result[0])), pages, uintptr(4*num))
	free.Call(pages)
	return result
}

type Phrase string

func (phrase Phrase) Join(cond ConditionsContainer) {
	cond.setPhrase(string(phrase))
}

type Option uintptr

const (
	// Sure means check every N-gram key
	Sure Option = 1 << 0
	// Usual means check N-gram keys skipping by one
	Usual Option = 1 << 1
	// Fast means check N-gram keys skipping by two
	Fast Option = 1 << 2
	// Agito means check N-gram keys skipping by three
	Agito Option = 1 << 3
	// Noidf means without TF-IDF tuning
	Noidf Option = 1 << 4
	// Simple menas with the simplified phrase
	Simple Option = 1 << 10
	// Rough menas with the rough phrase
	Rough Option = 1 << 11
	// Union means with the union phrase
	Union Option = 1 << 15
	// Isect means with the intersection phrase
	Isect Option = 1 << 16
	// Scfb means feed back scores (for debug)
	Scfb Option = 1 << 30
)

func (option Option) Join(cond ConditionsContainer) {
	cond.setOptions(uintptr(option))
}

type CondAttr string

func (condAttr CondAttr) Join(cond ConditionsContainer) {
	cond.addAttr(string(condAttr))
}

// Condition is the interface for conditions.
// (Database)Search can receive objects satisfying Condition.
type Condition interface {
	Join(ConditionsContainer)
}

// Search searches documents satisfying conditions.
func (db Database) Search(conditions ...Condition) []DocID {
	cond := newCond()
	for _, c1 := range conditions {
		c1.Join(cond)
	}
	rc := db.search(cond)
	cond.close()
	return rc
}

type Doc uintptr

func (db Database) GetDoc(id DocID) Doc {
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

func (doc Doc) URI() string {
	return doc.Attr("@uri")
}
