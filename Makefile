BIN_NAME=neko
VERSION=1.0

PACKAGE=$(BIN_NAME)_$(VERSION)-amd64

build:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BIN_NAME)_linux_amd64 .
	GOOS=windows GOARCH=amd64 go build -o bin/$(BIN_NAME)_windows_amd64.exe .
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
