
tuntuntun: *.go src/tun/* src/socks/*
	GOPATH=$(PWD) go build

test:
	GOPATH=$(PWD) go test
