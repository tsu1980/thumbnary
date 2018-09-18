build: $(wildcard **/*.go)
	go build -o bin/thumbnary
	go test
	./bin/thumbnary -config ./config.yml
