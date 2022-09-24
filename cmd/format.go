package cmd

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"sort"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func formatPackage() {} // todo: add ability to format an entire package

type typeGroup struct {
	name    string
	parent  dst.Decl
	methods []dst.Decl
}

// typeGroups complies with sort/Interface
type typeGroups []typeGroup

func (s typeGroups) Decls() []dst.Decl {
	var noMethodTypes []dst.Spec

	//decls starts with a reasonable amount: since there are at least len(structs) decls
	decls := make([]dst.Decl, 0, len(s))
	for i := range s {
		// if a type Group has no methods, lets put them together in a type block at the top
		if len(s[i].methods) == 0 {
			gd, ok := s[i].parent.(*dst.GenDecl)
			if !ok {
				panic("expected genDecl")
			}
			ts, ok := gd.Specs[0].(*dst.TypeSpec)
			if !ok {
				panic("expected typeSpec")
			}
			_, ok = ts.Type.(*dst.InterfaceType)
			if ok {
				decls = append(decls, s[i].parent)
				decls = append(decls, s[i].methods...)
				continue
			}

			ts.Decs.NodeDecs = gd.Decs.NodeDecs

			noMethodTypes = append(noMethodTypes, gd.Specs...)
			continue
		}

		decls = append(decls, s[i].parent)
		decls = append(decls, s[i].methods...)
	}

	d := dst.GenDecl{
		Tok:    token.TYPE,
		Lparen: true,
		Specs:  noMethodTypes,
		Rparen: true,
	}

	newDecls := make([]dst.Decl, 0, len(decls)+1)
	newDecls = append(newDecls, &d)

	return append(newDecls, decls...)
}

func (s typeGroups) Len() int {
	return len(s)
}

func (s typeGroups) Less(i, j int) bool {
	return strings.Compare(s[i].name, s[j].name) < 0
}

func (s typeGroups) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func decorationsEmpty(d *dst.NodeDecs) bool {
	if len(d.End.All()) != 0 {
		return false
	}
	if len(d.Start.All()) != 0 {
		return false
	}
	return true
}

// format reader and output formatted version to writer
// formatting actions:
// * sorts structures by name, grouping methods on structure
// todo: sort by usage in c-like fashion in order of use, a object should be defined before being used
func format(filename string, reader io.Reader, writer io.Writer) error {
	fileSet := token.NewFileSet()
	astFile, err := decorator.ParseFile(fileSet, filename, reader, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("could not parse ast (%w)", err)
	}

	var (
		imports  []dst.Decl
		types    typeGroups
		nonTypes []dst.Decl
	)

declloop:
	for _, decl := range astFile.Decls {
		switch t := decl.(type) {
		case *dst.GenDecl:
			if t.Tok != token.TYPE {
				if t.Tok == token.IMPORT {
					imports = append(imports, decl)
					continue
				}
				nonTypes = append(nonTypes, decl)
				continue
			}

			for _, spec := range t.Specs {
				s, ok := spec.(*dst.TypeSpec)
				if !ok {
					return fmt.Errorf("expected a typeSpec after type token got (%v)", t.Specs[0])
				}

				specDecoration := t.Decs.NodeDecs

				if decorationsEmpty(&specDecoration) {
					specDecoration = *s.Decorations()
					s.Decs = dst.TypeSpecDecorations{}
				}

				// todo: replace with a search function on types
				for i := range types {
					if types[i].name != s.Name.Name {
						continue
					}

					types[i].parent = &dst.GenDecl{
						Tok: token.TYPE,
						Specs: []dst.Spec{
							spec,
						},
						Decs: dst.GenDeclDecorations{
							NodeDecs: specDecoration,
						},
					}

					continue declloop
				}

				types = append(types, typeGroup{
					name: s.Name.Name,
					parent: &dst.GenDecl{
						Tok: token.TYPE,
						Specs: []dst.Spec{
							spec,
						},
						Decs: dst.GenDeclDecorations{
							NodeDecs: specDecoration,
						},
					},
				})
			}
		case *dst.FuncDecl:
			if t.Recv == nil {
				nonTypes = append(nonTypes, decl)
				continue
			}

			funcIdent, ok := t.Recv.List[0].Type.(*dst.Ident)
			if !ok {
				funcIdent, ok = t.Recv.List[0].Type.(*dst.StarExpr).X.(*dst.Ident)
				if !ok {
					return fmt.Errorf("expected a ident func token got (%T)", t.Recv.List[0].Type)
				}
			}

			for i := range types {
				if types[i].name != funcIdent.Name {
					continue
				}
				types[i].methods = append(types[i].methods, decl)
				continue declloop
			}

			types = append(types, typeGroup{name: funcIdent.Name, methods: []dst.Decl{decl}})
		default:
			nonTypes = append(nonTypes, decl)
		}
	}

	sort.Sort(types)

	astFile.Decls = imports
	astFile.Decls = append(astFile.Decls, types.Decls()...)
	astFile.Decls = append(astFile.Decls, nonTypes...)

	return decorator.Fprint(writer, astFile)
}
