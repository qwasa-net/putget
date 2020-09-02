package putget

import "regexp"

var DefaultBucketName = "default"

var bucketNameCleanRE = regexp.MustCompile("[^a-zA-Z0-9]")

var FilesRoot = "../files"

var ServerBindAddress = "127.0.0.1:8080"

var ServerURLRoot = "/"

var DBMaxSize = 100
