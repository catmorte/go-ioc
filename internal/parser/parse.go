//go:generate go-wrap -mode=priv
package parser

import (
	"bytes"
	"go/ast"
	"go/printer"
	"os"
	"strconv"
	"strings"

	"github.com/catmorte/go-ioc/internal/declaration"

	. "github.com/catmorte/go-wrap/pkg/wrap"
	"golang.org/x/tools/go/packages"
)

type packageParser struct {
	*packages.Package
}

func loadPackages(path string, cfg *packages.Config) ([]*packages.Package, error) {
	return packages.Load(cfg, path+"/...")
}

func (p packageParser) extractRawCode(node any) string {
	b := new(bytes.Buffer)
	printer.Fprint(b, p.Fset, node)
	return b.String()
}

func (p packageParser) parseError(err packages.Error) *declaration.Error {
	return &declaration.Error{
		Message:  err.Msg,
		Position: err.Pos,
	}
}

func (p packageParser) newType(f *ast.Field) *declaration.Type[Empty] {
	return &declaration.Type[Empty]{Code: p.extractRawCode(f.Type)}
}

func (p packageParser) newTypeParams(f *ast.Field) *declaration.Type[declaration.ParamMeta] {
	code := p.extractRawCode(f.Type)
	return &declaration.Type[declaration.ParamMeta]{
		Code: code,
		Meta: declaration.ParamMeta{
			IsVararg: strings.HasPrefix(code, "..."),
		},
	}
}

func (p packageParser) newTypeTypeArg(f *ast.Field) *declaration.Type[declaration.TypeMeta] {
	return &declaration.Type[declaration.TypeMeta]{
		Code: p.extractRawCode(f.Type),
		Meta: declaration.TypeMeta{
			Name: f.Names[0].Name,
		},
	}
}

func (p packageParser) createBean(x ast.Expr, index ast.Expr) *declaration.IndexMeta {
	return &declaration.IndexMeta{
		Field: &declaration.Type[declaration.TypeMeta]{
			Code: p.extractRawCode(x),
		},
		Index: &declaration.Type[declaration.TypeMeta]{
			Code: p.extractRawCode(index),
		},
	}
}

func (p packageParser) newStructType(f *ast.Field) *declaration.Type[declaration.StructFieldMeta] {
	name := ""
	if len(f.Names) > 0 {
		name = f.Names[0].Name
	}

	var index *declaration.IndexMeta
	if idx, ok := f.Type.(*ast.IndexExpr); ok {
		if bn, ok := idx.X.(*ast.Ident); ok {
			if len(name) == 0 {
				name = bn.Name
			}
		}
		if bn, ok := idx.X.(*ast.SelectorExpr); ok {
			if len(name) == 0 {
				name = bn.Sel.Name
			}
		}
		index = p.createBean(idx.X, idx.Index)
	}

	return &declaration.Type[declaration.StructFieldMeta]{
		Code: p.extractRawCode(f.Type),
		Meta: declaration.StructFieldMeta{
			Name:  name,
			Tag:   getBeanTagValue(f.Tag, declaration.IocTag),
			Index: index,
		},
	}
}

func getBeanTagValue(tagObj *ast.BasicLit, key string) *string {
	if tagObj == nil {
		return nil
	}
	tag := strings.Trim(tagObj.Value, "`")
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 && strings.Trim(kv[0], " ") == key {
			v := strings.Trim(kv[1], "\"")
			return &v
		}
	}
	return nil
}

func newPackageParser(p *packages.Package) packageParser {
	return packageParser{p}
}

func unquote(v string) (string, error) {
	return strconv.Unquote(v)
}

func parseFields[T any](l *ast.FieldList, fn func(f *ast.Field) *declaration.Type[T]) []*declaration.Type[T] {
	if l == nil {
		return nil
	}
	res := make([]*declaration.Type[T], 0, len(l.List))
	for _, f := range l.List {
		res = append(res, fn(f))
	}
	return res
}

func filterFuncDecls(decls []ast.Decl) []*ast.FuncDecl {
	res := []*ast.FuncDecl{}
	for _, v := range decls {
		if f, ok := v.(*ast.FuncDecl); ok {
			res = append(res, f)
		}
	}
	return res
}

func filterStructDecls(decls []ast.Decl) []*ast.TypeSpec {
	res := []*ast.TypeSpec{}
	for _, decl := range decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						res = append(res, typeSpec)
					}
				}
			}
		}
	}
	return res
}

func (p packageParser) newStruct(s *ast.TypeSpec) *declaration.Struct {
	return &declaration.Struct{
		Name:   s.Name.Name,
		Code:   p.extractRawCode(s),
		Fields: parseFields(s.Type.(*ast.StructType).Fields, p.newStructType),
	}
}

func (p packageParser) newFunc(fn *ast.FuncDecl) *declaration.Func {
	return &declaration.Func{
		Name:      fn.Name.Name,
		Code:      p.extractRawCode(fn),
		Params:    parseFields(fn.Type.Params, p.newTypeParams),
		Receivers: parseFields(fn.Recv, p.newType),
		Results:   parseFields(fn.Type.Results, p.newType),
		Types:     parseFields(fn.Type.TypeParams, p.newTypeTypeArg),
	}
}

func (p packageParser) newFile(fPath string, funcs []*declaration.Func, imports []*declaration.Import, structs []*declaration.Struct) *declaration.File {
	return &declaration.File{
		Path:    fPath,
		Funcs:   funcs,
		Imports: imports,
		Structs: structs,
	}
}

func (p packageParser) newPackage(errors []*declaration.Error, files []*declaration.File) *declaration.Package {
	return &declaration.Package{
		Files:  files,
		Name:   p.Name,
		Errors: errors,
	}
}

func (p packageParser) newImport(v *ast.ImportSpec) (*declaration.Import, error) {
	pathUnquoted, err := unquote(v.Path.Value)
	if err != nil {
		return nil, err
	}
	alias := ""
	if v.Name != nil {
		alias = v.Name.String()
	}
	return &declaration.Import{Path: pathUnquoted, Alias: alias}, nil
}

func (p packageParser) plantBean(v *declaration.Struct, imports []*declaration.Import) {
	fields := []*declaration.Type[declaration.StructFieldMeta]{}
	for _, vv := range v.Fields {
		if p.isIocBean(vv, imports) {
			v.Bean = vv
		} else {
			fields = append(fields, vv)
		}
	}
	v.Fields = fields
}

func (p packageParser) isIocBean(f *declaration.Type[declaration.StructFieldMeta], imports []*declaration.Import) bool {
	if f.Meta.Index == nil || f.Meta.Name != declaration.IocBeanStructName {
		return false
	}
	for _, imp := range imports {
		switch imp.Alias {
		case "_":
			continue
		case "":
			if strings.HasPrefix(f.Meta.Index.Field.Code, declaration.IocPkgSingletonAlias+".") ||
				strings.HasPrefix(f.Meta.Index.Field.Code, declaration.IocPkgPrototypeAlias+".") {
				return true
			}
		case ".":
			if strings.HasPrefix(f.Meta.Index.Field.Code, declaration.IocBeanStructName) {
				return true
			}
		default:
			if strings.HasPrefix(f.Meta.Index.Field.Code, imp.Alias+".") {
				return true
			}
		}
	}
	return false
}

func Parse(path string) []Out[*declaration.Package] {
	cfg := &packages.Config{
		Mode:  packages.NeedExportFile | packages.NeedModule | packages.NeedName | packages.NeedDeps | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedTypes | packages.NeedImports | packages.NeedFiles,
		Dir:   ".",
		Env:   os.Environ(),
		Tests: false,
	}
	packagesLoaded := DisJoin(loadPackagesWrap(path, cfg))
	convertedToPackageParsers := EachAsync(packagesLoaded, newPackageParserWrap)
	return Each(convertedToPackageParsers, func(p packageParser) Out[*declaration.Package] {
		errorsParsed := EachAsync(OKVargs(p.Errors...), p.parseErrorWrap)
		filesDisJoined := DisJoin(OK(p.Syntax))
		filesParsed := EachAsync(filesDisJoined, func(f *ast.File) Out[*declaration.File] {
			fullPath := OK(p.Fset.Position(f.Pos()).Filename)

			importsProcessed := EachAsync(OKSlice(f.Imports), p.newImportWrap)
			importsJoined := JoinAsync(importsProcessed)

			funcsConverted := EachAsync(OKSlice(filterFuncDecls(f.Decls)), p.newFuncWrap)
			funcsJoined := JoinAsync(funcsConverted)

			structsConverted := EachAsync(OKSlice(filterStructDecls(f.Decls)), p.newStructWrap)
			structsJoined := JoinAsync(structsConverted)

			structsPlanted := AndX2(importsJoined, structsJoined, func(imports []*declaration.Import, structs []*declaration.Struct) Out[[]*declaration.Struct] {
				for _, s := range structs {
					p.plantBeanWrap(s, imports)
				}
				return OK(structs)
			})

			return AndX4Async(fullPath, funcsJoined, importsJoined, structsPlanted, p.newFileWrap)
		})
		errorsJoined := JoinAsync(errorsParsed)
		filesJoined := JoinAsync(filesParsed)
		return AndX2(errorsJoined, filesJoined, p.newPackageWrap)
	})
}
