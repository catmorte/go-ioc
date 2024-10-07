package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/catmorte/go-ioc/internal/declaration"
	"github.com/catmorte/go-ioc/internal/generator"
	"github.com/catmorte/go-ioc/internal/parser"
	. "github.com/catmorte/go-wrap/pkg/wrap"
)

var errEx = regexp.MustCompile(`"(.+)" imported and not used`)

func findUnusedDotImportsForFile(filePath string, errors []*declaration.Error) []string {
	res := []string{}
	for _, e := range errors {
		if !strings.HasPrefix(e.Position, filePath) {
			continue
		}
		subMatches := errEx.FindStringSubmatch(e.Message)
		if len(subMatches) == 0 {
			continue
		}
		res = append(res, subMatches[1])
	}
	return res
}

func removeUnusedDotImports(unusedDotImports []string, imports []*declaration.Import) []*declaration.Import {
	res := []*declaration.Import{}
	for _, v := range imports {
		if v.Alias == "." && slices.Contains(unusedDotImports, v.Path) {
			continue
		}
		res = append(res, v)
	}
	return res
}

func findImportByAlias(i []*declaration.Import, alias string) bool {
	for _, v := range i {
		if v.Alias == alias {
			return true
		}
	}
	return false
}

func findImportByPath(i []*declaration.Import, pkg string) bool {
	for _, v := range i {
		if v.Path == pkg && v.Alias != "" && v.Alias != "_" {
			return true
		}
	}
	return false
}

func findFile(f string, vc []*declaration.File) *declaration.File {
	for _, v := range vc {
		if v.Path == f {
			return v
		}
	}
	return nil
}

func findPackageAndFileByPath(fullPath string, ps []*declaration.Package) (*declaration.Package, *declaration.File) {
	var p *declaration.Package
	var f *declaration.File
	for _, v := range ps {
		f = findFile(fullPath, v.Files)
		if f == nil {
			continue
		}
		p = v
		break
	}
	return p, f
}

func filterStructs(structs []*declaration.Struct) []*declaration.Struct {
	res := []*declaration.Struct{}
	for _, v := range structs {
		if v.Bean == nil {
			continue
		}
		res = append(res, v)
	}

	return res
}

func main() {
	fileFlag := flag.String("file", "", "file")
	flag.Parse()
	file := os.Getenv("GOFILE")

	if file == "" {
		if fileFlag == nil || *fileFlag == "" {
			log.Fatal("file is not specified")
		}
		file = *fileFlag
	}

	pathGot := Wrap(os.Getwd())
	fileSaved := And(pathGot, func(path string) Out[string] {
		fullPath := filepath.Join(path, file)
		packagesParsed := parser.Parse(path)
		packagesJoined := JoinAsync(packagesParsed)
		return And(packagesJoined, func(ps []*declaration.Package) Out[string] {
			p, f := findPackageAndFileByPath(fullPath, ps)
			if p == nil || f == nil {
				return Err[string](fmt.Errorf("file %v not found", fullPath))
			}
			f.Structs = filterStructs(f.Structs)

			ok := findImportByPath(f.Imports, declaration.IocPkgContextPath)
			if !ok {
				i := 0
				alias := ""
				for {
					alias = fmt.Sprintf("%s%d", declaration.IocPkgAlias, i)
					ok = findImportByAlias(f.Imports, alias)
					if !ok {
						f.Imports = append(f.Imports, &declaration.Import{Alias: alias, Path: declaration.IocPkgContextPath})
						break
					}
					i++
				}
			}
			fileName := fmt.Sprintf("%s.ioc.gen.go", strings.TrimSuffix(file, filepath.Ext(file)))
			fullPath := filepath.Join(path, fileName)
			codeGenerated := generator.Generate(p.Name, *f, false)
			return And(codeGenerated, func(raw []byte) Out[string] {
				return Wrap(fullPath, os.WriteFile(fullPath, raw, 0o644))
			})
		})
	})
	AndX2(pathGot, fileSaved, func(path, filePath string) Out[Empty] {
		packagesParsed := parser.Parse(path)
		packagesJoined := JoinAsync(packagesParsed)
		return And(packagesJoined, func(ps []*declaration.Package) Out[Empty] {
			p, f := findPackageAndFileByPath(filePath, ps)
			if p == nil || f == nil {
				return Err[Empty](fmt.Errorf("file %v not found", filePath))
			}
			unusedImports := findUnusedDotImportsForFile(filePath, p.Errors)
			f.Imports = removeUnusedDotImports(unusedImports, f.Imports)

			codeGenerated := generator.Generate(p.Name, *f, true)
			return And(codeGenerated, func(raw []byte) Out[Empty] {
				return Void(os.WriteFile(filePath, raw, 0o644))
			})
		})
	}).IfError(func(err error) {
		log.Fatal(err)
	})
}
