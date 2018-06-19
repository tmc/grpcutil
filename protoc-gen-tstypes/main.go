package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/tmc/grpcutil/protoc-gen-tstypes/gentstypes"
)

func main() {
	g := gentstypes.New()
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "reading input"))
	}
	if err := proto.Unmarshal(data, g.Request); err != nil {
		log.Fatalln(errors.Wrap(err, "parsing input"))
	}
	if len(g.Request.FileToGenerate) == 0 {
		log.Fatalln(errors.Wrap(err, "no files to generate"))
	}
	g.GenerateAllFiles()
	// Send back the results.
	data, err = proto.Marshal(g.Response)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to marshal output proto"))
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to write output proto"))
	}
}
