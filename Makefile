ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=set
    DEL=del
    NUL=nul
else
    SET=export
    DEL=rm
    NUL=/dev/null
endif

NAME:=$(notdir $(CURDIR))
VERSION:=$(shell git describe --tags 2>$(NUL) || echo v0.0.0)
GOOPT:=-ldflags "-s -w -X main.version=$(VERSION)"
EXE:=$(shell go env GOEXE)

all:
	go fmt ./...
	$(SET) "CGO_ENABLED=0" && go build $(GOOPT)
	$(SET) "CGO_ENABLED=0" && pushd "cmd/csvi" && go build -o "$(CURDIR)" $(GOOPT) && popd

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

clean:
	$(DEL) *.zip $(NAME)$(EXE)

release:
	gh release create -d --notes "" -t $(VERSION) $(VERSION) $(wildcard $(NAME)-$(VERSION)-*.zip)

manifest:
	make-scoop-manifest *-windows-*.zip > $(NAME).json

test:
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
	goawk "{ gsub(/README_ja/,'index_ja');print }" README.md | minipage -title "\"CSVI\" Terminal CSV Editor" - > docs\index.html
	goawk "{ gsub(/README/,'index');print }" README_ja.md | minipage -title "\"CSVI\" Terminal CSV Editor" - > docs\index_ja.html
	minipage -title "\"CSVI\" Release notes" release_note_en.md > docs\release_note_en.html
	minipage -title "\"CSVI\" Release notes" release_note_ja.md > docs\release_note_ja.html

.PHONY: all test dist _dist clean release manifest readme docs
