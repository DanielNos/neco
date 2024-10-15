.PHONY: build run test package clean

run:
	go run . main.neco

build:
	go build .

package:
	makego

test:
	go clean -testcache
	cd tests && go test .

clean:
	cd tests/src && find . -type f ! -name "*.*" -exec rm {} +
	rm -rf build
