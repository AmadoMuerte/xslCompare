.PHONY: build clean

build xmlCompare:
	mkdir -p builds
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o builds/xlsCompare.exe cmd/xmlCompare/main.go
	@echo "Build complete: builds/xlsCompare.exe"

build countLog:
	mkdir -p builds
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o builds/countLog.exe cmd/countLog/main.go
	@echo "Build complete: builds/countLog.exe"

clean:
	rm -rf builds
	@echo "Clean complete"
