DISTPATH = ./dist

.PHONY:all run clean help

all: run

run:
	npm run serve

clean:
	@rm -rf ${DISTPATH}

build:
	npm run build


help:
	@echo "make all - 生成linux执行文件"
	@echo "make run - 直接运行 Go 代码"
	@echo "make clean - 移除编译文件夹"
	@echo "make help - 查看帮助"

