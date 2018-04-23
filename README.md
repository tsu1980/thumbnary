# thumbnary

This is [imaginary](https://github.com/h2non/imaginary) fork. Changes from original are below.

- Multiple origins support
  - Origin has image source and URL signature key
  - Origin ID determines from HTTP request(subdomain or URL path or HTTP header)
  - Origin setting fetched from Database(supports MySQL)
  - Origin setting has cached and clear when updated using Redis pub/sub
- Change to single operation endpoints
- Add overlay(composition external image) operation
