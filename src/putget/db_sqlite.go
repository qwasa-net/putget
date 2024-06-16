package putget

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // a comment justifying a blank import
)

type storageSQLite struct {
	db              *sql.DB
	bucketsIdsCache map[string]int
}

func (s *storageSQLite) init() {
	var err error
	if s.db, err = sql.Open("sqlite3", DBPath); err != nil {
		log.Fatal(err)
	}
	if err = s.createTables(); err != nil {
		log.Fatal(err)
	}
	s.bucketsIdsCache = make(map[string]int)
}

func (s *storageSQLite) createTables() error {

	sqlCreate := `

	CREATE TABLE IF NOT EXISTS buckets (
		id INTEGER NOT NULL PRIMARY KEY,
		name TEXT
	);

	CREATE TABLE IF NOT EXISTS records (
		id INTEGER NOT NULL PRIMARY KEY,
		bucket_id INTEGER NOT NULL,
		ts INTEGER,
		filename TEXT,
		size INTEGER,
		ctype TEXT
	);
	`

	_, err := s.db.Exec(sqlCreate)
	return err

}

func (s *storageSQLite) getBucketID(bname string, create bool) int {

	var bid int
	var exists bool

	if bid, exists = s.bucketsIdsCache[bname]; exists {
		return bid
	}

	qc, _ := s.db.Prepare(`SELECT COUNT(*) FROM buckets WHERE name = ?`)
	defer qc.Close()
	var count int
	qc.QueryRow(bname).Scan(&count)

	if count < 1 {
		if create {
			s.initBucket(bname)
		} else {
			return 0
		}
	}

	q, _ := s.db.Prepare(`SELECT id FROM buckets WHERE name = ?`)
	defer q.Close()
	q.QueryRow(bname).Scan(&bid)

	s.bucketsIdsCache[bname] = bid
	return bid
}

func (s *storageSQLite) getBuckets() []string {
	keys := make([]string, 0)
	rows, _ := s.db.Query(`SELECT name FROM buckets`)
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		keys = append(keys, name)
	}
	return keys
}

func (s *storageSQLite) initBucket(bname string) {
	q, _ := s.db.Prepare(`INSERT INTO buckets (name) VALUES(?)`)
	defer q.Close()
	q.Exec(bname)
}

func (s *storageSQLite) addRecord(bname string, rec record) int {
	var err error

	bid := s.getBucketID(bname, true)

	qi, _ := s.db.Prepare(`INSERT INTO records (bucket_id, ts, filename, size, ctype) VALUES(?, ?, ?, ?, ?)`)
	defer qi.Close()
	if _, err = qi.Exec(bid, time.Now().Unix(), rec.Filename, rec.Size, rec.Ctype); err != nil {
		log.Println(err, bname)
	}

	return s.getBucketSize(bname)

}

func (s *storageSQLite) getBucketSize(bname string) int {

	bid := s.getBucketID(bname, false)
	if bid == 0 {
		return 0
	}

	qc, _ := s.db.Prepare(`SELECT COUNT(*) FROM records WHERE bucket_id=?`)
	defer qc.Close()
	var count int
	qc.QueryRow(bid).Scan(&count)
	return count

}

func (s *storageSQLite) getLastRecord(bname string, before int64) *record {

	bid := s.getBucketID(bname, false)
	if bid == 0 {
		return nil
	}

	var q string

	if before <= 0 {
		before = 5000000000 // Fri Jun 11 08:53:20 UTC 2128
	}

	rec := record{}
	var recTs int64
	var recID int

	q = `SELECT id, ts, filename, size, ctype
	FROM records
	WHERE bucket_id=? AND ts<?
	ORDER BY id DESC LIMIT 1`

	qs, _ := s.db.Prepare(q)
	defer qs.Close()

	qr := qs.QueryRow(bid, before)
	if err := qr.Scan(&recID, &recTs, &rec.Filename, &rec.Size, &rec.Ctype); err != nil {
		return nil
	}

	rec.Ts = time.Unix(recTs, 0)
	rec.Clength = rec.Size
	return &rec
}

func (s *storageSQLite) toString() string {
	return ""
}
