package putget

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func put(rsp http.ResponseWriter, req *http.Request) {

	// read headers
	ctype := req.Header.Get("Content-Type")
	xssec := req.Header.Get("X-SSE-C")

	// read content
	content, err := io.ReadAll(req.Body)
	if err != nil || len(content) == 0 {
		fail(rsp, req, err)
		return
	}

	clength := int64(len(content))

	// encrypt content before saving
	if xssec != "" {
		content, _ = encrypt(content, xssec)
	}

	// read bucket name
	bucket := defaultBucketName
	paths := strings.Split(req.URL.Path, "/")
	if len(paths) > 1 && paths[1] != "" {
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
	log.Printf("file=%v size=%d bucket=%d sse=%v", filename, len(content), i, (xssec != ""))
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
	rsp.Header().Set("Access-Control-Allow-Origin", "*")
	rsp.WriteHeader(http.StatusOK)
	fmt.Fprint(rsp, string(data))
}

func getRecord(rsp http.ResponseWriter, req *http.Request) {

	var err error

	// get bucket name
	bucket := defaultBucketName
	paths := strings.Split(req.URL.Path, "/")
	if len(paths) > 1 && paths[1] != "" {
		bucket = paths[1]
		bucket = bucketNameCleanRE.ReplaceAllString(bucket, "_")
	}

	var before int64 = 0
	beforeQS := req.URL.Query().Get("before")
	if beforeQS != "" {
		if before, err = strconv.ParseInt(beforeQS, 10, 64); err != nil {
			fail(rsp, req, err)
			return
		}
	}

	// get the most recent record from the bucket
	rec := getDB(bucket, before)
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
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil || len(content) == 0 {
		fail(rsp, req, err)
		return
	}

	xssec := req.Header.Get("X-SSE-C")
	// decrypt content before sending
	if xssec != "" {
		decrypted, err := decrypt(content, xssec)
		if err == nil {
			content = decrypted
		} else {
			content = []byte("")
		}
	}
	clength := (int64)(len(content))

	// send file
	rsp.Header().Set("Content-Type", (*rec).Ctype)
	rsp.Header().Set("Content-Length", fmt.Sprintf("%d", clength))
	rsp.Header().Set("Access-Control-Allow-Origin", "*")
	rsp.Header().Set("Last-Modified", (*rec).Ts.Format(time.RFC1123Z))
	rsp.WriteHeader(http.StatusOK)
	rsp.Write(content)

}

func fail(rsp http.ResponseWriter, req *http.Request, err error) {
	rsp.Header().Set("Access-Control-Allow-Origin", "*")
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

func createServer() error {

	mux := http.NewServeMux()
	mux.HandleFunc(ServerURLRoot, handler)

	srv := &http.Server{
		Addr:         ServerBindAddress,
		Handler:      mux,
		ReadTimeout:  time.Duration(serverTimeout) * time.Second,
		WriteTimeout: time.Duration(serverTimeout) * time.Second,
		IdleTimeout:  time.Duration(serverTimeout) * time.Second,
	}
	err := srv.ListenAndServe()
	log.Println("server stopped:", err)
	return err
}
