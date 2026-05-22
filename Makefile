run:
	go run ./cmd/ghostd

cli:
	go run ./cmd/ghostmesh

build:
	go build -o bin/ghostd ./cmd/ghostd
	go build -o bin/ghostmesh ./cmd/ghostmesh

test:
	go test ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin
