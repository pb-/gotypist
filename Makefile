run:
	go run *.go

build: gotypist

gotypist: *.go
	go build

dep:
	go get -v ./...

clean:
	rm -f gotypist
