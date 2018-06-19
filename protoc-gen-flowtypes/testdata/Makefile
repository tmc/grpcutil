all: test

scalapb.proto:
	wget https://raw.githubusercontent.com/scalapb/ScalaPB/master/protobuf/scalapb/scalapb.proto

out:
	mkdir out

.PHONY: test
test: scalapb.proto out
	bash test

