package cmd

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"sort"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func formatPackage() {} // todo: add ability to format an entire package

type typeGroup struct {
	name    string
	parent  dst.Decl
	methods []dst.Decl
}

// typeGroups is a slice of typeGroup lexigraphically sorted by typeGroup.name
type typeGroups []typeGroup

func (s *typeGroups) GetAndRemoveNoParentMethods() []dst.Decl {
	var noParentIndexes []int
	for i := range *s {
		if (*s)[i].parent != nil {
			continue
		}
		noParentIndexes = append(noParentIndexes, i)
	}

	if len(noParentIndexes) == 0 {
		return nil
	}

	ns := make([]typeGroup, 0, len(*s)-len(noParentIndexes))
	decls := make([]dst.Decl, 0, len(noParentIndexes))
	for i := range *s {
		if sort.SearchInts(noParentIndexes, i) != len(noParentIndexes) {
			continue
		}
		ns = append(ns, (*s)[i])
	}

	*s = ns

	return decls
}

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
		Lparen: len(noMethodTypes) > 1, // no need to type block if its only one type
		Specs:  noMethodTypes,
		Rparen: len(noMethodTypes) > 1, // no need to type block if its only one type
	}

	newDecls := make([]dst.Decl, 0, len(decls)+1)
	newDecls = append(newDecls, &d)

	return append(newDecls, decls...)
}

// Index in typeGroups where name is, -1 if non-existant
func (s typeGroups) Index(name string) int {
	i := sort.Search(len(s), func(i int) bool {
		return s[i].name >= name
	})
	if i == len(s) {
		return -1
	}
	return i
}

func (s *typeGroups) Insert(tg typeGroup) {
	i := sort.Search(len(*s), func(i int) bool {
		return (*s)[i].name >= tg.name
	})
	*s = append(*s, tg)
	copy((*s)[i+1:], (*s)[i:])
	(*s)[i] = tg
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

func sortASTFile(astFile *dst.File) error {
	var (
		imports  []dst.Decl
		types    typeGroups
		nonTypes []dst.Decl
	)
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
					return fmt.Errorf("expected a typeSpec after type token got (%T)", spec)
				}

				specDecoration := t.Decs.NodeDecs

				if decorationsEmpty(&specDecoration) {
					specDecoration = *s.Decorations()
					s.Decs = dst.TypeSpecDecorations{}
				}

				i := types.Index(s.Name.Name)
				if i == -1 {
					types.Insert(typeGroup{
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
					continue
				}
				types.Insert(typeGroup{
					name: s.Name.Name,
					parent: &dst.GenDecl{
						Tok: token.TYPE,
						Specs: []dst.Spec{
							spec,
						},
						Decs: dst.GenDeclDecorations{
							NodeDecs: specDecoration,
						},
					}})
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

			i := types.Index(funcIdent.Name)
			if i == -1 {
				types.Insert(typeGroup{
					name:    funcIdent.Name,
					methods: []dst.Decl{decl},
				})

				continue
			}
			types[i].methods = append(types[i].methods, decl)
		default:
			nonTypes = append(nonTypes, decl)
		}
	}

	noParentMethods := types.GetAndRemoveNoParentMethods()
	typeDecls := types.Decls()

	astFile.Decls = make([]dst.Decl, 0, len(noParentMethods)+len(imports)+len(typeDecls)+len(nonTypes))

	// imports first
	astFile.Decls = append(astFile.Decls, imports...)
	// sorted types next
	astFile.Decls = append(astFile.Decls, typeDecls...)
	// methods without types in this file
	astFile.Decls = append(astFile.Decls, noParentMethods...)
	// everything else
	astFile.Decls = append(astFile.Decls, nonTypes...)

	return nil
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
	err = sortASTFile(astFile)
	if err != nil {
		return err
	}
	return decorator.Fprint(writer, astFile)
}
