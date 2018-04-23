APPNAME=app.companion
DATEVERSION=`date -u +%Y%m%d`
COMMIT=`git --no-pager log --pretty="%h" -n 1`

.PHONY:all
all:clean build

linux_build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${APPNAME}

.PHONY:build
build:
	go build -o ${APPNAME} -ldflags "-X main.version=${DATEVERSION} -X main.gitCommit=${COMMIT}"

.PHONY:install
install:build
	mv ${APPNAME} ${GOPATH}/bin/

.PHONY:clean
clean:
	-rm -f ${APPNAME}

.PHONY:rebuild
clean:
	-rm -f ${APPNAME}