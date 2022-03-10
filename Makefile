PROG_NAME := "yum-package-diff"
IMAGE_NAME := "pschou/yum-package-diff"
VERSION = 0.1.$(shell date +%Y%m%d.%H%M)
FLAGS := "-s -w -X main.version=${VERSION}"


build:
	CGO_ENABLED=0 go build -ldflags=${FLAGS} -o ${PROG_NAME} main.go repo.go
