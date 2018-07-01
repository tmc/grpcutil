package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/tmc/grpcutil/protoc-gen-tstypes/gentstypes"
)

var (
	flagVerbose               = flag.Int("v", 0, "verbosity level")
	flagDeclareNamespace      = flag.Bool("declare_namespace", true, "if true, generate a namespace declaration")
	flagAsyncIterators        = flag.Bool("async_iterators", false, "if true, user async iterators")
	flagEnumsAsInts           = flag.Bool("int_enums", false, "if true, generate numeric enums")
	flagOutputFilenamePattern = flag.String("outpattern", "{{.Dir}}/{{.Descriptor.GetPackage | default \"none\"}}.{{.BaseName}}.d.ts", "output filename pattern")
	flagDumpDescriptor        = flag.Bool("dump_request_descriptor", false, "if true, dump request descriptor")
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
	parseFlags(g.Request.Parameter)
	g.GenerateAllFiles(&gentstypes.Parameters{
		AsyncIterators:        *flagAsyncIterators,
		DeclareNamespace:      *flagDeclareNamespace,
		Verbose:               *flagVerbose,
		OutputNamePattern:     *flagOutputFilenamePattern,
		DumpRequestDescriptor: *flagDumpDescriptor,
	})
	data, err = proto.Marshal(g.Response)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to marshal output proto"))
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to write output proto"))
	}
}

func parseFlags(s *string) {
	if s == nil {
		return
	}
	for _, p := range strings.Split(*s, ",") {
		spec := strings.SplitN(p, "=", 2)
		if len(spec) == 1 {
			if err := flag.CommandLine.Set(spec[0], ""); err != nil {
				log.Fatalln("Cannot set flag", p, err)
			}
			continue
		}
		name, value := spec[0], spec[1]
		// TODO: consider supporting package mapping (M flag)
		if err := flag.CommandLine.Set(name, value); err != nil {
			log.Fatalln("Cannot set flag", p)
		}
	}
}
