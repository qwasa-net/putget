package main

import putget "./putget"
import "flag"
import "os"
import "log"

type clArgs struct {
	filesRoot   string
	bindAddress string
	urlRoot     string
}

func main() {

	log.Println("starting putget …")

	params := parseArgs()

	putget.ServerBindAddress = params.bindAddress
	putget.FilesRoot = params.filesRoot
	putget.ServerURLRoot = params.urlRoot

	log.Printf("files at `%v`, listening to `%v%v`", params.filesRoot, params.bindAddress, params.urlRoot)
	putget.Start()

}

func parseArgs() clArgs {

	var params = clArgs{filesRoot: "./files", bindAddress: "localhost:8800", urlRoot: "/"}

	if value, exist := os.LookupEnv("PUTGET_FILES_ROOT"); exist {
		params.filesRoot = value
	}
	if value, exist := os.LookupEnv("PUTGET_BIND_ADDRESS"); exist {
		params.bindAddress = value
	}
	if value, exist := os.LookupEnv("PUTGET_URL_ROOT"); exist {
		params.urlRoot = value
	}

	flag.StringVar(&params.filesRoot, "files-root", params.filesRoot,
		"files storage root, (env: $PUTGET_FILES_ROOT)")
	flag.StringVar(&params.bindAddress, "bind-address", params.bindAddress,
		"server bind address, (env: $PUTGET_BIND_ADDRESS)")
	flag.StringVar(&params.urlRoot, "url-root", params.urlRoot,
		"'secret' URL prefix, (env: $PUTGET_URL_ROOT)")
	flag.Parse()

	return params

}
