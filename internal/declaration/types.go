package declaration

import . "github.com/catmorte/go-wrap/pkg/wrap"

type (
	BeanType int
	Import   struct {
		Alias string
		Path  string
	}
	Type[T any] struct {
		Code string
		Meta T
	}
	TypeMeta struct {
		Name string
	}
	StructFieldMeta struct {
		Name  string
		Tag   *string
		Index *IndexMeta
	}
	IndexMeta struct {
		Field *Type[TypeMeta]
		Index *Type[TypeMeta]
	}
	ParamMeta struct {
		IsVararg bool
	}
	Func struct {
		Name      string
		Code      string
		Params    []*Type[ParamMeta]
		Results   []*Type[Empty]
		Types     []*Type[TypeMeta]
		Receivers []*Type[Empty]
	}
	Struct struct {
		Name   string
		Code   string
		Bean   *Type[StructFieldMeta]
		Fields []*Type[StructFieldMeta]
	}
	File struct {
		Path    string
		Imports []*Import
		Funcs   []*Func
		Structs []*Struct
	}
	Package struct {
		Name   string
		Files  []*File
		Errors []*Error
	}
	Error struct {
		Message  string
		Position string
	}
)
