package gentstypes

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jhump/protoreflect/desc"
)

/*
provide the following part in the import syntax

import * as google_api from "../../google/api/annotations"
		    ^^^^^^^^^^
*/
func formatImportModule(f *desc.FileDescriptor) string {
	pkg := f.GetPackage()
	pkg = strings.ReplaceAll(pkg, "/", "_")
	pkg = strings.ReplaceAll(pkg, ".", "_")
	return pkg
}

/*
provide the following part in the import syntax

import * as google_api from "../../google/api/annotations"
		    			     ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
*/
func formatImportFile(currentFile *desc.FileDescriptor, externalFile *desc.FileDescriptor) string {

	var (
		baseFilePath      = currentFile.GetName()                     // "backend/enterprise/proto/company.proto"
		baseDirPath, _    = filepath.Split(baseFilePath)              // "backend/enterprise/proto/"
		targetFilePath    = externalFile.GetName()                    // "backend/common/proto/paging.proto"
		relativePath, err = filepath.Rel(baseDirPath, targetFilePath) // "../../common/proto/paging.proto"
	)

	if err != nil {
		panic(err)
	}

	// remove ".proto"
	s := regexp.MustCompile(`^(.*)(\.proto)$`).ReplaceAllString(relativePath, `$1`)
	return s
}
