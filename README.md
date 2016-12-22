## gather
Golang project for downloading files. The attempt is to make a general purpose
downloader for files over http, https, ftp, etc.  We'll see how that goes!

## Test
`go test`
`go test -v -cover`
`go test ./... # etc.`

## Docs
`godoc github.com/pladdy/gather`

## Example
examples/* has .sh scripts that are wrappers to more easily download or scrape.

### Verbose Example
```sh
go build && ./gather download -s ./data.txt -u http://malc0de.com/bl/IP_Blacklist.txt
```
