PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)

all: $(GO)

%.pb.go: %.proto
	protoc -I . *.proto --go_out=. $<

clean:
	rm *.pb.go