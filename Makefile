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
	$(SET) "CGO_ENABLED=0" && $(GO) build $(GOOPT)
	$(SET) "CGO_ENABLED=0" && cd "cmd/csvi" && $(GO) build -o "$(CURDIR)" $(GOOPT) && cd "../.."

_dist:
	$(MAKE) all
	zip -9 $(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).zip $(NAME)$(EXE)

dist:
	$(SET) "GOOS=linux" && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=linux" && $(SET) "GOARCH=amd64" && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=amd64" && $(MAKE) _dist
	$(SET) "GOOS=darwin" && $(SET) "GOARCH=amd64"  && $(MAKE) _dist
	$(SET) "GOOS=darwin" && $(SET) "GOARCH=arm64"  && $(MAKE) _dist
	$(SET) "GOOS=freebsd" && $(SET) "GOARCH=amd64"  && $(MAKE) _dist

clean:
	$(DEL) *.zip $(NAME)$(EXE)

release:
	pwsh latest-notes.ps1 | gh release create -d --notes-file - -t $(VERSION) $(VERSION) $(wildcard $(NAME)-$(VERSION)-*.zip)

manifest:
	make-scoop-manifest *-windows-*.zip > $(NAME).json

test:
	$(GO) test -v ./...
	pwsh test/case0.ps1
	pwsh test/case1.ps1
	pwsh test/case2.ps1
	pwsh test/case3.ps1

benchmark:
	pwsh test/benchmark.ps1

readme:
	example-into-readme
	example-into-readme -target README_ja.md

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
