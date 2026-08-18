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

## dist-windows: 构建 Windows 二进制（系统托盘 GUI，需 mingw-w64：brew install mingw-w64）
dist-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags "-H=windowsgui" -o $(DIST)/windows/$(BINARY)-windows-amd64.exe .
	cd $(DIST)/windows && zip $(BINARY)-windows.zip $(BINARY)-windows-amd64.exe

## installer-windows: 生成 Windows 安装程序（需 NSIS + mingw-w64 + python3+Pillow）
installer-windows: dist-windows
	python3 build/windows/gen_icon.py $(DIST)/windows
	cp build/windows/installer.nsi $(DIST)/windows/installer.nsi
	cp build/windows/README.txt $(DIST)/windows/README.txt
	cp build/windows/LICENSE.txt $(DIST)/windows/LICENSE.txt
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
