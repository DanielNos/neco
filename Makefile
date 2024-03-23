BIN_NAME=neco
VERSION=1.0

PACKAGE=$(BIN_NAME)_$(VERSION)-amd64
BUILD_COMMAND=go build -ldflags="-w -s" -o

debug: *.go
	GOOS=linux GOARCH=amd64 go build -o bin/$(BIN_NAME)_linux_amd64_debug .

build: *.go
	GOOS=linux GOARCH=amd64 $(BUILD_COMMAND) bin/$(BIN_NAME)_linux_amd64 .
	GOOS=linux GOARCH=386 $(BUILD_COMMAND) bin/$(BIN_NAME)_linux_386 .
	GOOS=linux GOARCH=arm $(BUILD_COMMAND) bin/$(BIN_NAME)_linux_arm .
	GOOS=linux GOARCH=arm64 $(BUILD_COMMAND) bin/$(BIN_NAME)_linux_arm64 .
	GOOS=windows GOARCH=amd64 $(BUILD_COMMAND) bin/$(BIN_NAME)_windows_amd64.exe .
	GOOS=windows GOARCH=arm $(BUILD_COMMAND) bin/$(BIN_NAME)_windows_arm.exe .
	GOOS=darwin GOARCH=amd64 $(BUILD_COMMAND) bin/$(BIN_NAME)_macos_amd64 .
	GOOS=darwin GOARCH=arm64 $(BUILD_COMMAND) bin/$(BIN_NAME)_macos_arm64 .

package: clean build
	mkdir -p $(PACKAGE)/usr/bin
	mkdir $(PACKAGE)/DEBIAN
	echo "Package: $(BIN_NAME)\nVersion: $(VERSION)\nArchitecture: amd64\nMaintainer: Daniel Nos <nos.daniel@pm.me>\nDescription: Programming language." > $(PACKAGE)/DEBIAN/control
	cp bin/$(BIN_NAME)_linux_amd64 $(PACKAGE)/usr/bin/$(BIN_NAME)
	dpkg-deb --build --root-owner-group $(PACKAGE)
	mv $(PACKAGE).deb bin
	rm -rf $(PACKAGE)
	
clean:
	rm -rf bin
