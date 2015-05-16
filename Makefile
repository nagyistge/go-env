.PHONY: \
	all \
	precommit \
	deps \
	updatedeps \
	testdeps \
	updatetestdeps \
	generate \
	build \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	cov \
	checkjet \
	jet \
	clean \

all: test

precommit: test jet

deps:
	go get -d -v ./...

updatedeps:
	go get -d -v -u -f ./...

testdeps:
	go get -d -v -t ./...

updatetestdeps:
	go get -d -v -t -u -f ./...

build: deps
	go build ./...

lint: testdeps
	go get -v github.com/golang/lint/golint
	golint ./...

vet: testdeps
	go get -v golang.org/x/tools/cmd/vet
	go vet ./...

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck ./...

pretest: lint vet errcheck

test: pretest
	go test -test.v ./...

cov: testdeps
	go get -v github.com/axw/gocov/gocov
	go get -v golang.org/x/tools/cmd/cover
	gocov test | gocov report

checkjet:
	@ if ! which jet > /dev/null; then \
			echo "error: jet not installed" >&2; \
			exit 1; \
	  fi

jet: checkjet
	jet steps

clean:
	go clean ./...
