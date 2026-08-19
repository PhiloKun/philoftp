# PhiloFTP 构建与管理 Makefile
# 版本号单一真源：version.txt（发布时修改该文件并打对应 vX.Y.Z tag）

BINARY      := philoftp
DIST        := dist
# 从 version.txt 读取版本（去除空白），缺失时回退 git describe / dev
VERSION_RAW := $(shell cat version.txt 2>/dev/null | tr -d '[:space:]')
VERSION     := $(if $(VERSION_RAW),$(VERSION_RAW),$(shell git describe --tags --always 2>/dev/null || echo dev))
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
# 统一注入版本/提交信息：handler 包（关于页）+ main 包（启动日志）
VERSION_LDFLAGS := -X github.com/philoftp/handler.version=$(VERSION) -X github.com/philoftp/handler.gitCommit=$(GIT_COMMIT) -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT)

.PHONY: build run clean dist-macos dist-windows dist-linux dist installer-windows test tag-version

## build: 编译到 dist/macos/$(BINARY)（本地运行用，文件名带版本）
build:
	go build -ldflags "$(VERSION_LDFLAGS)" -o $(DIST)/macos/$(BINARY)-$(VERSION) .

## run: 直接编译并运行（默认从 configs/ 读取配置）
run:
	go run .

## dist-macos: 构建 macOS 双架构二进制（arm64 / amd64），产物名携带版本号
dist-macos:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(VERSION_LDFLAGS)" -o $(DIST)/macos/$(BINARY)-$(VERSION)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(VERSION_LDFLAGS)" -o $(DIST)/macos/$(BINARY)-$(VERSION)-darwin-amd64 .
	cd $(DIST)/macos && zip -r $(BINARY)-$(VERSION)-macos.zip $(BINARY)-$(VERSION)-darwin-arm64 $(BINARY)-$(VERSION)-darwin-amd64

## dist-windows: 构建 Windows 二进制（系统托盘 GUI，需 mingw-w64：brew install mingw-w64），产物名携带版本号
dist-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags "-H=windowsgui $(VERSION_LDFLAGS)" -o $(DIST)/windows/$(BINARY)-$(VERSION)-windows-amd64.exe .
	cd $(DIST)/windows && zip $(BINARY)-$(VERSION)-windows.zip $(BINARY)-$(VERSION)-windows-amd64.exe

## installer-windows: 生成 Windows 安装程序（需 NSIS + mingw-w64 + python3+Pillow），安装包名携带版本号
installer-windows: dist-windows
	python3 build/windows/gen_icon.py $(DIST)/windows
	cp build/windows/installer.nsi $(DIST)/windows/installer.nsi
	cp build/windows/README.txt $(DIST)/windows/README.txt
	cp build/windows/LICENSE.txt $(DIST)/windows/LICENSE.txt
	# 将版本占位符注入 NSIS 脚本：__VERSION__ / __VERSION4__（x.y.z.0）/ __EXE_NAME__
	sed -i '' -e 's/__VERSION__/$(VERSION)/g' \
	          -e 's/__VERSION4__/$(VERSION).0/g' \
	          -e 's/__EXE_NAME__/$(BINARY)-$(VERSION)-windows-amd64.exe/g' \
	          $(DIST)/windows/installer.nsi
	cd $(DIST)/windows && makensis installer.nsi
	cd $(DIST)/windows && mv PhiloFTP-Setup.exe $(BINARY)-$(VERSION)-windows-setup.exe

## dist-linux: 构建 Linux 二进制，产物名携带版本号
dist-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(VERSION_LDFLAGS)" -o $(DIST)/linux/$(BINARY)-$(VERSION)-linux-amd64 .
	cd $(DIST)/linux && zip $(BINARY)-$(VERSION)-linux.zip $(BINARY)-$(VERSION)-linux-amd64

## dist: 构建全部平台产物
dist: dist-macos dist-windows dist-linux

## tag-version: 根据 version.txt 创建并推送对应 vX.Y.Z tag（双远端）
tag-version:
	git tag -a v$(VERSION) -m "v$(VERSION)"
	git push gitee v$(VERSION)
	git push origin v$(VERSION)

## test: 运行测试
test:
	go test ./...

## clean: 删除本地编译产物
clean:
	rm -f $(DIST)/macos/$(BINARY)-*
	rm -f $(BINARY)
