
tuntuntun: *.go src/tun/*
	GOPATH=$(PWD) go build

test:
	GOPATH=$(PWD) go test
