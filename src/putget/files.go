package putget

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

//
func saveFile(bname string, content []byte) (string, error) {
	if bname == "" {
		bname = defaultBucketName
	}
	fname := strconv.FormatInt(time.Now().Unix(), 10)
	fdir := filepath.Join(FilesRoot, bname)
	fpath := filepath.Join(bname, fname)
	fpathabs := filepath.Join(fdir, fname)
	var err error
	if err = os.MkdirAll(fdir, 0777); err != nil {
		return "", err
	}
	err = ioutil.WriteFile(fpathabs, content, 0644)
	return fpath, err
}

//
func getFile(fpath string) (*os.File, error) {
	fpathabs := filepath.Join(FilesRoot, fpath)
	f, err := os.Open(fpathabs)
	return f, err
}
