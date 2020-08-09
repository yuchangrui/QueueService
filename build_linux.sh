GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server -ldflags "${flags}" cmd/server/*.go
GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o client cmd/client/*.go
