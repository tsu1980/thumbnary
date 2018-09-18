# thumbnary

This is [imaginary](https://github.com/h2non/imaginary) fork. Changes from original are below.

- Multiple origins support
  - Origin has image source and URL signature key
  - Origin ID determines from HTTP request(subdomain or URL path or HTTP header)
  - Origin setting fetched from Database(supports MySQL)
  - Origin setting has cached and clear when updated using Redis pub/sub
- Change to single operation endpoints
- Add overlay(composition external image) operation

## Docker

To run docker container
```
set THUMBNARY_CONFIG=C:\thumbnary_config.yml
cd C:\go\src\github.com\tsu1980\thumbnary
docker build -t tsu1980/thumbnary:latest .
docker-compose up
```
