NAME=$(lastword $(subst /, ,$(abspath .)))
VERSION=$(shell git.exe describe --tags)
GOOPT=-ldflags "-s -w -X main.version=$(VERSION)"
ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=set
    DEL=del
else
    SET=export
    DEL=rm
endif

all:
	go fmt
	$(SET) "CGO_ENABLED=0" && go build $(GOOPT)

test:
	go test -v

_package:
	$(SET) "CGO_ENABLED=0" && go build $(GOOPT) && \
	zip -9 $(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).zip $(NAME)$(EXE)

package:
	$(SET) "GOOS=linux" && $(SET) "GOARCH=386"   && $(MAKE) _package EXE=
	$(SET) "GOOS=linux" && $(SET) "GOARCH=amd64" && $(MAKE) _package EXE=
	$(SET) "GOOS=windows" && $(SET) "GOARCH=386"   && $(MAKE) _package EXE=.exe
	$(SET) "GOOS=windows" && $(SET) "GOARCH=amd64" && $(MAKE) _package EXE=.exe

clean:
	$(DEL) *.zip *.tar.gz $(NAME) $(NAME).exe
