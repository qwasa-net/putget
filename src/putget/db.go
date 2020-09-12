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
	Filename string    `json:"filename"`
	Size     int       `json:"size"`
	Ts       time.Time `json:"ts"`
	Clength  int64     `json:"-"`
	Ctype    string    `json:"content_type"`
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

func (s *storage) getBuckets() *map[string]*[]record {
	return &s.buckets
}

func (s *storage) initBucket(bname string) *[]record {
	buck := make([]record, 0)
	s.buckets[bname] = &buck
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
	if _, exists := s.buckets[bname]; !exists {
		return nil
	}
	recs := db.getRecords(bname)
	if len(recs) > 0 {
		return &recs[len(recs)-1]
	}
	return nil
}

func (s *storage) toString() string {
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
	rec := record{Filename: filename, Size: len(content), Ts: time.Now(), Clength: cl, Ctype: ct}
	i := db.addRecord(bname, rec)
	return i
}

//
func getDB(bname string) *record {
	return db.getLastRecord(bname)
}

type listingInfo struct {
	Name string  `json:"name"`
	Size int     `json:"size"`
	Last *record `json:"last,omitempty"`
}

func getBucketsLists() []listingInfo {
	bucks := make([]listingInfo, 0)
	for bname, buck := range *db.getBuckets() {
		blen := len(*buck)
		binfo := listingInfo{Name: bname, Size: blen}
		if blen > 0 {
			binfo.Last = &(*buck)[blen-1]
		}
		bucks = append(bucks, binfo)
	}
	return bucks
}
