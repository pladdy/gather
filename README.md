## gather
Golang project for downloading files. The attempt is to make a general purpose
downloader for files over http, https, ftp, etc.  We'll see how that goes!

## Test
`go test`

## Docs
`godoc github.com/pladdy/gather`

## Example
```sh
# for this download you'll need 'bgpdump' to convert the file to readable text
go build && ./gather config/malc0de_ip_blacklist.json
```
