run: build
	./gotypist

build: gotypist

gotypist: *.go
	go build

test:
	go test

dep:
	go get -v ./...

clean:
	rm -f gotypist
