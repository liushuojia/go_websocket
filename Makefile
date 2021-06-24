GOFILE = ./main.go
IMAGE_TAG = 1.0
DOCKER_IMAGE_NAME = mx

.PHONY:all linux run clean mod help git
all: gotool linux

linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${GOFILE}

run:
	go run ${GOFILE}

gotool:
	go fmt ${GOFILE}
	go vet ${GOFILE}

clean:
	@rm -rf ./build
	@rm -rf ./logs

mod:
	go mod download
	go mod tidy
	go mod vendor
	go mod verify

git:
	git add .
	git commit -m "change"
	git push

help:
	@echo "make all - 生成linux执行文件"
	@echo "make linux - 生成linux执行文件"
	@echo "make run - 直接运行 Go 代码"
	@echo "make clean - 移除编译文件夹"
	@echo "make mod - 运行 Go mod"
	@echo "make help - 查看帮助"
	@echo "make git - git 备份"

