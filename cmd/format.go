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

type structData struct {
	name    string
	struc   dst.Decl
	methods []dst.Decl
}

// structs complies with sort/Interface
type structs []structData

func (s structs) Decls() []dst.Decl {
	//decls starts with a reasonable amount: since there are at least len(structs) decls
	decls := make([]dst.Decl, 0, len(s))
	for i := range s {
		decls = append(decls, s[i].struc)
		decls = append(decls, s[i].methods...)
	}
	return decls
}

func (s structs) Len() int {
	return len(s)
}

func (s structs) Less(i, j int) bool {
	return strings.Compare(s[i].name, s[j].name) < 0
}

func (s structs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
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
		structs         structs
		destructedDecls []dst.Decl
	)

declloop:
	for _, decl := range astFile.Decls {
		switch t := decl.(type) {
		case *dst.GenDecl:
			if t.Tok != token.TYPE {
				destructedDecls = append(destructedDecls, decl)
				continue
			}

			s, ok := t.Specs[0].(*dst.TypeSpec)
			if !ok {
				return fmt.Errorf("expected a typeSpec after type token got (%v)", t.Specs[0])
			}

			_, ok = s.Type.(*dst.StructType)
			if !ok {
				destructedDecls = append(destructedDecls, decl)
				continue
			}

			for i := range structs {
				if structs[i].name != s.Name.Name {
					continue
				}
				structs[i].struc = decl
				continue declloop
			}

			structs = append(structs, structData{
				name:  s.Name.Name,
				struc: decl,
			})
		case *dst.FuncDecl:
			if t.Recv == nil {
				destructedDecls = append(destructedDecls, decl)
				continue
			}

			funcIdent, ok := t.Recv.List[0].Type.(*dst.Ident)
			if !ok {
				funcIdent, ok = t.Recv.List[0].Type.(*dst.StarExpr).X.(*dst.Ident)
				if !ok {
					return fmt.Errorf("expected a ident func token got (%T)", t.Recv.List[0].Type)
				}
			}

			for i := range structs {
				if structs[i].name != funcIdent.Name {
					continue
				}
				structs[i].methods = append(structs[i].methods, decl)
				continue declloop
			}

			structs = append(structs, structData{name: funcIdent.Name, methods: []dst.Decl{decl}})
		default:
			destructedDecls = append(destructedDecls, decl)
		}

	}

	sort.Sort(structs)

	astFile.Decls = structs.Decls()
	astFile.Decls = append(astFile.Decls, destructedDecls...)

	return decorator.Fprint(writer, astFile)
}
