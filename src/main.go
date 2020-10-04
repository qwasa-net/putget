package main

import putget "./putget"
import "flag"
import "os"
import "log"

type clArgs struct {
	filesRoot    string
	filesDateDir string
	bindAddress  string
	urlRoot      string
	dbPath       string
}

func main() {

	log.Println("starting putget …")

	params := parseArgs()

	putget.ServerBindAddress = params.bindAddress
	putget.FilesRoot = params.filesRoot
	putget.FilesDateDir = params.filesDateDir
	putget.ServerURLRoot = params.urlRoot
	putget.DBPath = params.dbPath

	log.Printf("files at `%v`, db `%v`, listening to `%v%v`",
		params.filesRoot, params.dbPath, params.bindAddress, params.urlRoot)
	putget.Start()

}

func parseArgs() clArgs {

	var params = clArgs{
		filesRoot:    "./putget.files",
		filesDateDir: "2006.01.02",
		bindAddress:  "localhost:8800",
		urlRoot:      "/",
		dbPath:       "./putget.sqlite",
	}

	if value, exist := os.LookupEnv("PUTGET_FILES_ROOT"); exist {
		params.filesRoot = value
	}
	if value, exist := os.LookupEnv("PUTGET_FILES_DATEDIR"); exist {
		params.filesDateDir = value
	}
	if value, exist := os.LookupEnv("PUTGET_BIND_ADDRESS"); exist {
		params.bindAddress = value
	}
	if value, exist := os.LookupEnv("PUTGET_URL_ROOT"); exist {
		params.urlRoot = value
	}
	if value, exist := os.LookupEnv("PUTGET_DB_PATH"); exist {
		params.dbPath = value
	}

	flag.StringVar(&params.filesRoot, "files-root", params.filesRoot,
		"files storage root, (env: $PUTGET_FILES_ROOT)")
	flag.StringVar(&params.filesDateDir, "files-datedir", params.filesDateDir,
		"files storage root, (env: $PUTGET_FILES_DATEDIR)")
	flag.StringVar(&params.bindAddress, "bind-address", params.bindAddress,
		"server bind address, (env: $PUTGET_BIND_ADDRESS)")
	flag.StringVar(&params.urlRoot, "url-root", params.urlRoot,
		"'secret' URL prefix, (env: $PUTGET_URL_ROOT)")
	flag.StringVar(&params.dbPath, "db-path", params.dbPath,
		"files storage root, (env: $PUTGET_FILES_ROOT)")
	flag.Parse()

	return params

}
