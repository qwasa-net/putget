package putget

import "regexp"

var defaultBucketName = "default"

var bucketNameCleanRE = regexp.MustCompile("[^a-zA-Z0-9]")

// FilesRoot is a files storage path
var FilesRoot = "../putget.files"

// ServerBindAddress is HTTP service bind address
var ServerBindAddress = "127.0.0.1:8080"

// ServerURLRoot defines HTTP service path prefix (possibly secret)
var ServerURLRoot = "/"

// DBPath is a path for sqlite db (otherwise dummy mem map is used)
var DBPath = "../putget.sqlite"

// dbMaxSize is a limit for in-memery map storage
var dbMaxSize = 15000
