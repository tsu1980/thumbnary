build: $(wildcard **/*.go)
	go build -o bin/thumbnary
	./thumbnary -config ./config.yml
