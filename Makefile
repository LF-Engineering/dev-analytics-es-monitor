GO_BIN_FILES=esmonitor.go fixture.go exec.go error.go
GO_BIN_CMD=esmonitor
# race
# GO_ENV=CGO_ENABLED=1
# GO_BUILD=go build -ldflags '-s -w' -race
# no race
GO_ENV=CGO_ENABLED=0
GO_BUILD=go build -ldflags '-s -w'
# end
GO_INSTALL=go install -ldflags '-s'
GO_FMT=gofmt -s -w
GO_LINT=golint -set_exit_status
GO_VET=go vet
GO_IMPORTS=goimports -w
GO_USEDEXPORTS=usedexports
GO_ERRCHECK=errcheck -asserts -ignore '[FS]?[Pp]rint*'
BINARIES=esmonitor
STRIP=strip

all: check ${BINARIES}

esmonitor: ${GO_BIN_FILES}
	 ${GO_ENV} ${GO_BUILD} -o ${GO_BIN_CMD} ${GO_BIN_FILES}

fmt: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_FMT}"

lint: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_LINT}"

vet: ${GO_BIN_FILES}
	${GO_ENV} ${GO_VET} ${GO_BIN_FILES}

imports: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_IMPORTS}"

usedexports: ${GO_BIN_FILES}
	${GO_USEDEXPORTS} ./...

errcheck: ${GO_BIN_FILES}
	${GO_ERRCHECK} ./...

check: fmt lint imports vet usedexports errcheck

install: check ${BINARIES}
	${GO_INSTALL} ${GO_BIN_CMD}

strip: ${BINARIES}
	${STRIP} ${BINARIES}

clean:
	rm -f ${BINARIES}
