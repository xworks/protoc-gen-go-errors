package main

import (
	"fmt"
	"strings"

	"github.com/driftingboy/protoc-gen-go-errors/gerr"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

const (
	errorsPackage = protogen.GoImportPath("github.com/driftingboy/protoc-gen-go-errors/gerr") // "github.com/driftingboy/protoc-gen-go-errors/errors"
	fmtPackage    = protogen.GoImportPath("fmt")
)

// generateFile generates a _errors.pb.go file containing kratos errors definitions.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Enums) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_errors.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-errors. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	g.QualifiedGoIdent(fmtPackage.Ident(""))
	generateFileContent(gen, file, g)
	return g
}

// generateFileContent generates the kratos errors definitions, excluding the package statement.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Enums) == 0 {
		return
	}

	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the kratos package it is being compiled against.")
	g.P("const _ = ", errorsPackage.Ident("SupportPackageIsVersion1"))
	g.P()
	index := 0
	for _, enum := range file.Enums {
		skip := genErrorsReason(gen, file, g, enum)
		if !skip {
			index++
		}
	}
	// If all enums do not contain 'errors.code', the current file is skipped
	if index == 0 {
		g.Skip()
	}
}

// return (isSkip bool)
func genErrorsReason(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, enum *protogen.Enum) bool {
	settingsI := proto.GetExtension(enum.Desc.Options(), gerr.E_Settings)
	defaultHttpCode, startBizCode := 0, 100001
	if s, ok := settingsI.(*gerr.Settings); ok && s != nil {
		defaultHttpCode, startBizCode = int(s.DefaultHttpCode), int(s.StartBizCode)
	}
	if defaultHttpCode > 600 || defaultHttpCode < 0 {
		panic(fmt.Sprintf("Enum '%s' httpCode range must in (0,600)", string(enum.Desc.Name())))
	}
	if startBizCode > 999999 || startBizCode < 100001 {
		panic(fmt.Sprintf("Enum '%s' errorNo range must in [100001,999999]", string(enum.Desc.Name())))
	}

	var ew errorWrapper
	nextErrorNo := startBizCode
	for _, v := range enum.Values {
		httpCode, curErrorNo := defaultHttpCode, 0
		eCode := proto.GetExtension(v.Desc.Options(), gerr.E_Code)
		if status, ok := eCode.(*gerr.StatusCode); ok && status != nil {
			httpCode = int(status.HttpCode)
			curErrorNo = int(status.BizCode)
		}
		// 未填写
		if curErrorNo == 0 {
			curErrorNo = nextErrorNo
		}
		nextErrorNo = curErrorNo + 1
		// If the current enumeration does not contain 'errors.code'
		// or the code value exceeds the range, the current enum will be skipped
		if httpCode > 600 || httpCode < 0 {
			panic(fmt.Sprintf("Enum '%s' httpCode range must in (0,600)", string(v.Desc.Name())))
		}
		if curErrorNo > 999999 || curErrorNo < 100001 {
			panic(fmt.Sprintf("Enum '%s' errorNo range must in [100001,999999]", string(v.Desc.Name())))
		}
		if httpCode == 0 || curErrorNo == 0 {
			continue
		}

		// packageName := string(file.Desc.Package())
		err := &errorInfo{
			Name:         string(enum.Desc.Name()),
			Value:        string(v.Desc.Name()),
			CamelValue:   case2Camel(string(v.Desc.Name())),
			HTTPCode:     httpCode,
			BizErrorCode: curErrorNo,
		}
		ew.Errors = append(ew.Errors, err)
	}
	if len(ew.Errors) == 0 {
		return true
	}
	g.P(ew.generateTemp())
	return false
}

func case2Camel(name string) string {
	if !strings.Contains(name, "_") {
		return name
	}
	// name = strings.ToLower(name)
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}
