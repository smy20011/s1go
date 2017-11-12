PROTO_DIR := stage1stpb

all: proto bindata install

.PHONY: proto bindata clean test

install:
	go install .

proto:
	protoc -I ${PROTO_DIR} --go_out ${PROTO_DIR} ${PROTO_DIR}/*.proto

bindata:
	go-bindata --pkg=data -o test_util/data/bindata.go --prefix test_util test_util/data

clean:
	rm -rf stage1st.db*

test:
	go test ./...
