## gather
Golang project for downloading files. The attempt is to make a general purpose
downloader for files over http, https, ftp, etc.  We'll see how that goes

## Example
```sh
# for this download you'll need 'bgpdump' to convert the file to readable text
go build && ./gather testdata/ripe_rrc_00.json
# or
go run gather.go testdata/malc0de_ip_blacklist.json
```
