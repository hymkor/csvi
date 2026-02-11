ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=set
    WHICH=where.exe
    DEL=del
    NUL=nul
else
    SET=export
    WHICH=which
    DEL=rm
    NUL=/dev/null
endif

ifndef GO
    SUPPORTGO=go1.20.14
    GO:=$(shell $(WHICH) $(SUPPORTGO) 2>$(NUL) || echo go)
endif

NAME:=$(notdir $(CURDIR))
VERSION:=$(shell git describe --tags 2>$(NUL) || echo v0.0.0)
GOOPT:=-ldflags "-s -w -X main.version=$(VERSION)"
EXE:=$(shell $(GO) env GOEXE)

all:
	$(GO) fmt ./...
	$(SET) "CGO_ENABLED=0" && $(GO) build -C "cmd/csvi" -o "$(CURDIR)" $(GOOPT)

_dist:
	$(SET) "CGO_ENABLED=0" && $(GO) build -C "cmd/csvi" -o "$(CURDIR)" $(GOOPT)
	zip -9 $(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).zip $(NAME)$(EXE)

dist:
	$(SET) "GOOS=windows" && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=amd64" && $(MAKE) _dist
	$(SET) "GOOS=darwin"  && $(SET) "GOARCH=amd64" && $(MAKE) _dist
	$(SET) "GOOS=darwin"  && $(SET) "GOARCH=arm64" && $(MAKE) _dist
	$(SET) "GOOS=freebsd" && $(SET) "GOARCH=amd64" && $(MAKE) _dist
	$(SET) "GOOS=linux"   && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=linux"   && $(SET) "GOARCH=amd64" && $(MAKE) _dist

bump:
	$(GO) run bump.go -suffix "-goinstall" -gosource main release_note*.md > cmd/csvi/version.go

clean:
	$(DEL) *.zip $(NAME)$(EXE)

release:
	$(GO) run github.com/hymkor/latest-notes@master | gh release create -d --notes-file - -t $(VERSION) $(VERSION) $(wildcard $(NAME)-$(VERSION)-*.zip)

manifest:
	$(GO) run github.com/hymkor/make-scoop-manifest@master -all *-windows-*.zip > $(NAME).json

test:
	$(GO) test -v ./...

benchmark:
	pwsh test/benchmark.ps1

readme:
	$(GO) run github.com/hymkor/example-into-readme@master
	$(GO) run github.com/hymkor/example-into-readme@master -target README_ja.md

docs:
	minipage -title "\"CSVI\" Terminal CSV Editor" -outline-in-sidebar -readme-to-index README.md > docs\index.html
	minipage -title "\"CSVI\" Terminal CSV Editor" -outline-in-sidebar -readme-to-index README_ja.md > docs\index_ja.html
	minipage -title "\"CSVI\" Release notes" -outline-in-sidebar -readme-to-index release_note_en.md > docs\release_note_en.html
	minipage -title "\"CSVI\" Release notes" -outline-in-sidebar -readme-to-index release_note_ja.md > docs\release_note_ja.html

get:
	$(GO) get -u
	$(GO) get golang.org/x/sys@v0.30.0
	$(GO) get golang.org/x/text@v0.22.0
	$(GO) get golang.org/x/term@v0.29.0 
	$(GO) mod tidy

.PHONY: all test dist _dist clean release manifest readme docs
