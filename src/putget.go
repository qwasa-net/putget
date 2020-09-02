package main

import (
	putget "./putget"
	"flag"
	"log"
)

func main() {

	log.Println("starting putget")

	filesRoot := flag.String("files-root", "./files/", "files storage root")
	bindAddress := flag.String("bind-address", "localhost:8800", "server bind address")
	urlRoot := flag.String("url-root", "/", "'secret' URL prefix")
	flag.Parse()

	log.Printf("files at `%v`, listening to `%v%v`", *filesRoot, *bindAddress, *urlRoot)

	putget.ServerBindAddress = *bindAddress
	putget.FilesRoot = *filesRoot
	putget.ServerURLRoot = *urlRoot
	putget.Start()

}
