package putget

import "fmt"

type storageMap struct {
	buckets map[string]*[]record
	maxsize int
}

func (s *storageMap) init() {
	s.buckets = make(map[string]*[]record)
	s.maxsize = dbMaxSize
}

func (s *storageMap) getBucket(bname string) *[]record {
	_, exists := s.buckets[bname]
	if !exists {
		s.initBucket(bname)
	}
	return s.buckets[bname]
}

func (s *storageMap) getBuckets() []string {
	keys := make([]string, 0)
	for k := range s.buckets {
		keys = append(keys, k)
	}
	return keys
}

func (s *storageMap) getBucketSize(bname string) int {
	recs, exists := s.buckets[bname]
	if !exists {
		return 0
	}
	return len(*recs)
}

func (s *storageMap) initBucket(bname string) {
	buck := make([]record, 0)
	s.buckets[bname] = &buck
}

func (s *storageMap) addRecord(bname string, rec record) int {
	b := s.getBucket(bname)
	*b = append(*b, rec)
	if len(*b) > s.maxsize {
		*b = (*b)[1:]
	}
	return len(*b)
}

func (s *storageMap) getLastRecord(bname string) *record {
	recs, exists := s.buckets[bname]
	if !exists {
		return nil
	}
	if len(*recs) > 0 {
		return &(*recs)[len(*recs)-1]
	}
	return nil
}

func (s *storageMap) toString() string {
	str := ""
	for k, v := range s.buckets {
		str += fmt.Sprintf("%v: %v\n", k, *v)
	}
	return str
}
