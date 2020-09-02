package putget

import (
	"fmt"
	"time"
)

type storage struct {
	buckets map[string]*[]record
	maxsize int
}

type record struct {
	filename string
	size     int
	ts       time.Time
	cl       int64
	ct       string
}

func (s *storage) init() {
	s.buckets = make(map[string]*[]record)
	s.maxsize = dbMaxSize
}

func (s *storage) getBucket(bname string) *[]record {
	_, exists := s.buckets[bname]
	if !exists {
		s.initBucket(bname)
	}
	return s.buckets[bname]
}

func (s *storage) initBucket(bname string) *[]record {
	b := make([]record, 0)
	s.buckets[bname] = &b
	return s.buckets[bname]
}

func (s *storage) getRecords(bname string) []record {
	_, exists := s.buckets[bname]
	if !exists {
		s.initBucket(bname)
	}
	return *s.buckets[bname]
}

func (s *storage) addRecord(bname string, rec record) int {
	b := s.getBucket(bname)
	*b = append(*b, rec)
	if len(*b) > s.maxsize {
		*b = (*b)[1:]
	}
	return len(*b)
}

func (s *storage) getLastRecord(bname string) *record {
	recs := db.getRecords(bname)
	if len(recs) > 0 {
		return &recs[len(recs)-1]
	}
	return nil
}

func (s *storage) print() string {
	str := ""
	for k, v := range s.buckets {
		str += fmt.Sprintf("%v: %v\n", k, *v)
	}
	return str
}

//
var db storage

//
func initDB() *storage {
	db = storage{}
	db.init()
	db.initBucket("default")
	return &db
}

//
func saveDB(bname string, filename string, content []byte, ct string, cl int64) int {
	rec := record{filename: filename, size: len(content), ts: time.Now(), cl: cl, ct: ct}
	i := db.addRecord(bname, rec)
	return i
}

//
func getDB(bname string) *record {
	return db.getLastRecord(bname)
}
