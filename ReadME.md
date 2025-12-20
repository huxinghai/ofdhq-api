## OFDHQ
-  ginskeleton

```bash
go env -w GOARCH=amd64
go env -w GOOS=linux
go env -w CGO_ENABLED=0
go build -o bin/ofdhq-api -ldflags "-w -s"  -trimpath  ./cmd/api/main.go
nohup ./ofdhq-api > /var/www/ofdhq-api/logs/output.log  2>&1 &

go env -u GOARCH
go env -u GOOS
go env -u CGO_ENABLED

go run cmd/api/main.go

kill -2 pid
```


git@github.com:huxinghai/ofdhq-api.git