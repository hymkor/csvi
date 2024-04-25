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

_dist:
	$(SET) "CGO_ENABLED=0" && go build $(GOOPT)
	zip -9 $(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).zip $(NAME)$(EXE)

dist:
	$(SET) "GOOS=linux" && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=linux" && $(SET) "GOARCH=amd64" && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=amd64" && $(MAKE) _dist

clean:
	$(DEL) *.zip $(NAME)$(EXE)

release:
	gh release create -d --notes "" -t $(VERSION) $(VERSION) $(wildcard $(NAME)-$(VERSION)-*.zip)

manifest:
	make-scoop-manifest *-windows-*.zip > $(NAME).json

test:
	cd test && pwsh case0.ps1 && pwsh case1.ps1 && pwsh case2.ps1

benchmark:
	powershell "if ( -not (Test-Path 27OSAKA.CSV) ){ curl.exe -O https://www.post.japanpost.jp/zipcode/dl/kogaki/zip/27osaka.zip ; unzip.exe 27osaka.zip ; Remove-Item 27osaka.zip} ; $$start = Get-Date ; .\csvi.exe -auto 'l|l|l|l|l|l|l|l|l|l|l|l|l|l|q|y' 27OSAKA.CSV ; $$end = Get-Date ; Write-Host ($$end-$$start).TotalSeconds"

.PHONY: all test dist _dist clean release manifest
