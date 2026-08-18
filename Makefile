# PhiloFTP 构建与管理 Makefile

BINARY      := philoftp
DIST        := dist

.PHONY: build run clean dist-macos dist-windows dist-linux dist installer-windows test

## build: 编译到 dist/macos/$(BINARY)（本地运行用）
build:
	go build -o $(DIST)/macos/$(BINARY) .

## run: 直接编译并运行（默认从 configs/ 读取配置）
run:
	go run .

## dist-macos: 构建 macOS 双架构二进制（arm64 / amd64）
dist-macos:
	GOOS=darwin GOARCH=arm64 go build -o $(DIST)/macos/$(BINARY)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o $(DIST)/macos/$(BINARY)-darwin-amd64 .
	cd $(DIST)/macos && zip -r $(BINARY)-macos.zip $(BINARY)-darwin-arm64 $(BINARY)-darwin-amd64 PhiloFTP.app

## dist-windows: 构建 Windows 二进制
dist-windows:
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/windows/$(BINARY)-windows-amd64.exe .
	cd $(DIST)/windows && zip $(BINARY)-windows.zip $(BINARY)-windows-amd64.exe

## installer-windows: 生成 Windows 安装程序（需 NSIS：brew install nsis）
installer-windows: dist-windows
	cp build/windows/installer.nsi $(DIST)/windows/installer.nsi
	cd $(DIST)/windows && makensis installer.nsi

## dist-linux: 构建 Linux 二进制
dist-linux:
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/linux/$(BINARY)-linux-amd64 .
	cd $(DIST)/linux && zip $(BINARY)-linux.zip $(BINARY)-linux-amd64

## dist: 构建全部平台产物
dist: dist-macos dist-windows dist-linux

## test: 运行测试
test:
	go test ./...

## clean: 删除本地编译产物（保留已发布的 dist 各平台包）
clean:
	rm -f $(DIST)/macos/$(BINARY)
	rm -f $(BINARY)
