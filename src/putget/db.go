package putget

import "time"

type record struct {
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	Ts       time.Time `json:"ts"`
	Clength  int64     `json:"-"`
	Ctype    string    `json:"content_type"`
}

type storage interface {
	init()
	getBuckets() []string
	addRecord(bname string, rec record) int
	getBucketSize(bname string) int
	getLastRecord(bname string, before int64) *record
	toString() string
}

var db storage

func initDB() *storage {
	if len(DBPath) > 0 {
		db = &storageSQLite{}
	} else {
		db = &storageMap{}
	}
	db.init()
	return &db
}

func saveDB(bname string, filename string, content []byte, ct string, cl int64) int {
	rec := record{Filename: filename, Size: int64(len(content)), Ts: time.Now(), Clength: cl, Ctype: ct}
	i := db.addRecord(bname, rec)
	return i
}

func getDB(bname string, before int64) *record {
	return db.getLastRecord(bname, before)
}

type listingInfo struct {
	Name string  `json:"name"`
	Size int     `json:"size"`
	Last *record `json:"last,omitempty"`
}

func getBucketsLists() []listingInfo {
	bucks := make([]listingInfo, 0)
	for _, bname := range db.getBuckets() {
		binfo := listingInfo{}
		binfo.Name = bname
		binfo.Size = db.getBucketSize(bname)
		binfo.Last = db.getLastRecord(bname, 0)
		bucks = append(bucks, binfo)
	}
	return bucks
}
