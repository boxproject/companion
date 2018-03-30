package main

// go build -ldflags "-X main.version=`date -u +.%Y%m%d.%H%M%S`" main.go
var (
	stage     string = "dev"
	version   string
	gitCommit string
)
