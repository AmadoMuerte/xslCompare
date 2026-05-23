.PHONY: build clean

build:
	mkdir -p builds
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o builds/xlsCompare.exe cmd/main.go
	@echo "Build complete: builds/xlsCompare.exe"

clean:
	rm -rf builds
	@echo "Clean complete"
