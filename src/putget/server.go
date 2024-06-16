package putget

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
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

func encrypt(content []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(content))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], content)

	return ciphertext, nil
}

func decrypt(ciphertext []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	content := make([]byte, len(ciphertext))

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(content, ciphertext)

	return content, nil
}

func put(rsp http.ResponseWriter, req *http.Request) {

	// read headers
	ctype := req.Header.Get("Content-Type")
	cl := req.Header.Get("Content-Length")
	xssec := req.Header.Get("X-SSE-C")
	var clength int64 = 0
	if cl != "" {
		clength, _ = strconv.ParseInt(cl, 10, 64)
	}

	// read content
	content, err := io.ReadAll(req.Body)
	if err != nil || len(content) == 0 {
		fail(rsp, req, err)
		return
	}

	// encrypt content before saving
	if xssec != "" {
		content, _ = encrypt(content, xssec)
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
	rsp.Header().Set("Access-Control-Allow-Origin", "*")
	rsp.WriteHeader(http.StatusOK)
	fmt.Fprint(rsp, string(data))
}

func getRecord(rsp http.ResponseWriter, req *http.Request) {

	var err error

	// get bucket name
	bucket := defaultBucketName
	paths := strings.Split(req.URL.Path, "/")
	if len(paths) > 1 {
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
			rec.Clength = (int64)(len(content))
		}
	}

	// send file
	rsp.Header().Set("Content-Type", (*rec).Ctype)
	rsp.Header().Set("Content-Length", fmt.Sprintf("%d", (*rec).Clength))
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

type server struct {
	bind string
	root string
}

func createServer() *server {

	s := server{bind: ServerBindAddress, root: ServerURLRoot}
	http.HandleFunc(s.root, handler)
	http.ListenAndServe(s.bind, nil)
	return &s

}
