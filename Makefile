run: build
	./gotypist

demo: build
	./gotypist -d

build: gotypist

gotypist: *.go
	go build

test:
	go test

dep:
	go get -v ./...

clean:
	rm -f gotypist
