PROTO_DIR := stage1stpb

.PHONY: install proto

all: install proto

install: proto
	go install .

proto:
	protoc -I ${PROTO_DIR} --go_out ${PROTO_DIR} ${PROTO_DIR}/*.proto

clean:
	rm -rf stage1st.db*
