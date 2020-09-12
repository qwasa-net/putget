package putget

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func put(rsp http.ResponseWriter, req *http.Request) {

	// read headers
	ctype := req.Header.Get("Content-Type")
	cl := req.Header.Get("Content-Length")
	var clength int64 = 0
	if cl != "" {
		clength, _ = strconv.ParseInt(cl, 10, 64)
	}

	// read content
	content, err := ioutil.ReadAll(req.Body)
	if err != nil || len(content) == 0 {
		fail(rsp, req, err)
		return
	}

	// read bucket name
	bucket := defaultBucketName
	paths := strings.Split(req.URL.Path, "/")
	if len(paths) > 1 {
		bucket = paths[1]
		bucket = bucketNameCleanRE.ReplaceAllString(bucket, "_")
	}

	// save file
	filename, err := saveFile(bucket, content)
	if err != nil {
		fail(rsp, req, err)
		return
	}

	// create record
	i := saveDB(bucket, filename, content, ctype, clength)

	// done
	log.Printf("file=`%v` size=%d bucket size=%d", filename, len(content), i)
	rsp.WriteHeader(http.StatusOK)
	fmt.Fprintf(rsp, "ok|%v|%d|%d\n", filename, len(content), i)

}

func get(rsp http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		getListing(rsp, req) // get list of buckets
	} else {
		getRecord(rsp, req) // get the most recent record from the bucket
	}
}

func getListing(rsp http.ResponseWriter, req *http.Request) {
	listing := getBucketsLists()
	data, err := json.Marshal(&listing)
	if err != nil {
		fail(rsp, req, err)
		return
	}
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(http.StatusOK)
	fmt.Fprint(rsp, string(data))
}

func getRecord(rsp http.ResponseWriter, req *http.Request) {

	// get bucket name
	bucket := defaultBucketName
	paths := strings.Split(req.URL.Path, "/")
	if len(paths) > 1 {
		bucket = paths[1]
		bucket = bucketNameCleanRE.ReplaceAllString(bucket, "_")
	}

	// get the most recent record from the bucket
	rec := getDB(bucket)
	if rec == nil {
		fail(rsp, req, errors.New("no records"))
		return
	}

	// get file
	file, err := getFile((*rec).Filename)
	if err != nil {
		fail(rsp, req, err)
		return
	}

	// send file
	rsp.Header().Set("Content-Type", (*rec).Ctype)
	rsp.Header().Set("Content-Length", fmt.Sprintf("%d", (*rec).Clength))
	rsp.WriteHeader(http.StatusOK)
	io.Copy(rsp, file)

}

func fail(rsp http.ResponseWriter, req *http.Request, err error) {
	http.Error(rsp, err.Error(), http.StatusBadRequest)
	log.Println(err)
}

func handler(rsp http.ResponseWriter, req *http.Request) {

	clientIP := req.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = req.RemoteAddr
	}

	log.Println(clientIP, req.Method, req.URL)

	switch req.Method {
	case "GET":
		get(rsp, req)
	case "PUT":
		put(rsp, req)
	case "POST":
		put(rsp, req)
	default:
		fail(rsp, req, errors.New("bad method"))
	}

}

type server struct {
	bind string
	root string
}

//
func createServer() *server {

	s := server{bind: ServerBindAddress, root: ServerURLRoot}
	http.HandleFunc(s.root, handler)
	http.ListenAndServe(s.bind, nil)
	return &s

}
