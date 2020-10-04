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
	now := time.Now()
	fname := strconv.FormatInt(now.Unix(), 10)

	dateDir := ""
	if FilesDateDir != "" {
		dateDir = now.Format(FilesDateDir)
	}

	fdir := filepath.Join(bname, dateDir)     // relative dir path
	fpath := filepath.Join(fdir, fname)       // relative file path
	fdirabs := filepath.Join(FilesRoot, fdir) // absolute dir path
	fpathabs := filepath.Join(fdirabs, fname) // absolute file path

	var err error
	if err = os.MkdirAll(fdirabs, 0777); err != nil {
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
