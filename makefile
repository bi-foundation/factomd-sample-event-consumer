# To (re)generate source code from the proto files you will need to follow instructions here: https://github.com/golang/protobuf#installation
generate:
	protoc \
	    --gofast_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mopsee/protobuf/opsee.proto=github.com/opsee/protobuf/opseeproto,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:./eventmessages/generated \
	    --proto_path=$(GOPATH)/src:. eventmessages/*.proto
clean:
	rm -r -f eventmessages/generated/*

.PHONY:
	clean
	generate
