package putget

import "regexp"

var defaultBucketName = "default"

var bucketNameCleanRE = regexp.MustCompile("[^a-zA-Z0-9]")

// FilesRoot is a files storage path
var FilesRoot = "../putget.files"

// FilesDateDir is a file storage date-stamp sub-directory
var FilesDateDir = "2006.01.02"

// ServerBindAddress is HTTP service bind address
var ServerBindAddress = "127.0.0.1:8080"

// ServerURLRoot defines HTTP service path prefix (possibly secret)
var ServerURLRoot = "/"

// DBPath is a path for sqlite db (otherwise dummy mem map is used)
var DBPath = "../putget.sqlite"

// dbMaxSize is a limit for in-memery map storage
var dbMaxSize = 15000

// serverTimeout is a timeout for HTTP server connections
var serverTimeout = 60
