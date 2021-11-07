package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
)

var pkgsToRecord = hashset.New(
	"crypto/x509",
	"net/http",
	"context",
)

var pkgsToCheckImport = hashset.New(
	"crypto/x509",
)

var funToCheckUsed = hashset.New(
	"net/http#StripPrefix",
	"net/http#SameSiteDefaultMode",
)

var contextWithFuns = hashset.New(
	"context#WithValue",
	"context#WithDeadline",
	"context#WithCancel",
	"context#WithTimeout",
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage:\n\t%s dir\n", os.Args[0])
		os.Exit(1)
	}

	directory := os.Args[1]

	directory = strings.TrimRight(directory, ".")
	goFiles, err := GoFileWalk(directory)
	fmt.Println("goFiles:", goFiles)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fs := token.NewFileSet()

	fmt.Println("parse file loop start")
	for _, filename := range goFiles {

		astFile, err := parser.ParseFile(fs, filename, nil, parser.AllErrors)
		if err != nil {
			log.Printf("could not parse %s: %v", filename, err)
			continue
		}

		importMap := extractImport(astFile)
		fmt.Println("importMap:", importMap)
		if len(importMap) > 0 {
			result := checkImport(importMap)
			if len(result) > 0 {
				// Todo report
				fmt.Printf("File %s should check the usage of %s\n", filename, result)
			}
		}

		v := &visitor{
			fset:      fs,
			importMap: importMap,
		}

		// get all function
		fmt.Println("Functions:")

		var funNodes []*ast.FuncDecl
		for _, f := range astFile.Decls {
			fn, ok := f.(*ast.FuncDecl)
			if !ok {
				continue
			}
			funNodes = append(funNodes, fn)
			fmt.Printf("%v function name: %s\n", fs.Position(fn.Pos()), fn.Name.Name)
		}

		v.funNodes = funNodes
		// check context in function one by one
		for _, funNode := range funNodes {
			// check  parent context may nil
			var issueNodes = findBadContextPos(v, funNode)
			if issueNodes == nil || len(issueNodes) != 0 {
				for _, pos := range issueNodes {
					// at the context some parentContext may be nil
					fmt.Printf("%s context may with nil parent", fs.Position(pos.Pos()))
				}
			}
		}

		ast.Walk(v, astFile)

		fmt.Println()
		ast.Print(fs, astFile)

	}
	// fmt.Println(v.locals)
}

func findBadContextPos(v *visitor, funNode *ast.FuncDecl) []ast.Node {
	// check variable
	// TODO
	const contextType = "context#Context"

	result := make([]ast.Node, 0)
	contextsNotInit := hashset.New()
	ast.Inspect(funNode, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.AssignStmt:
			// check the right hand side
			if assignStmt, ok := n.(*ast.AssignStmt); ok {
				for idx, expr := range assignStmt.Rhs {
					if callExpr, ok := expr.(*ast.CallExpr); ok {
						if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
							if x, ok := selectorExpr.X.(*ast.Ident); ok {
								realPkgName := v.importMap[x.Name]
								if realPkgName != "" {
									fullExprName := realPkgName + "#" + selectorExpr.Sel.Name
									// fmt.Println(fullExprName)
									if contextWithFuns.Contains(fullExprName) {
										if argId, ok := callExpr.Args[0].(*ast.Ident); ok {
											if argId.Name == "nil" {
												// context parent should not be nil
												fmt.Println("context parent should not be nil")
												result = append(result, argId)
											} else if contextsNotInit.Contains(argId.Name) {
												// TODO
												fmt.Printf("parent context %s is nil", argId.Name)
												result = append(result, argId)
											}
										}
									} else {
										if varId, ok := assignStmt.Lhs[idx].(*ast.Ident); ok {
											if contextsNotInit.Contains(varId.Name) {
												contextsNotInit.Remove(varId.Name)
												fmt.Println("remove uninit context variable:", varId.Name)
											}
										}
									}
								}
							}
						}
					} else if valueId, ok := expr.(*ast.Ident); ok {
						if valueId.Name != "nil" {
							if varId, ok := assignStmt.Lhs[idx].(*ast.Ident); ok {
								if contextsNotInit.Contains(varId.Name) {
									contextsNotInit.Remove(varId.Name)
									fmt.Println("remove uninit context variable:", varId.Name)
								}
							}
						}
					}
				}
			}
		case *ast.ValueSpec:
			// valueSpec not assign value to variable
			if valueSpec, ok := n.(*ast.ValueSpec); ok {
				if selectorExpr, ok := valueSpec.Type.(*ast.SelectorExpr); ok {
					if x, ok := selectorExpr.X.(*ast.Ident); ok {
						realPkgName := v.importMap[x.Name]
						if realPkgName != "" {
							fullExprName := realPkgName + "#" + selectorExpr.Sel.Name
							if contextType == fullExprName {
								if valueSpec.Values == nil {
									for _, name := range valueSpec.Names {
										contextsNotInit.Add(name.Name)
									}
								} else {
									for idx, value := range valueSpec.Values {
										if value, ok := value.(*ast.Ident); ok {
											if value.Name == "nil" {
												contextsNotInit.Add(valueSpec.Names[idx].Name)
												fmt.Println("add uninit context variable:", valueSpec.Names[idx].Name)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		return true
	})
	return result
}

// extractImport return map[alia-package-name] real-package-name
func extractImport(astFile *ast.File) map[string]string {
	var pkgNames = make(map[string]string)
	for _, importSpec := range astFile.Imports {
		fmt.Println("extracting import ", importSpec.Path.Value)
		pathValue := strings.Trim(importSpec.Path.Value, "\"")
		// fmt.Println(pathValue)
		if pkgsToRecord.Contains(pathValue) {
			if importSpec.Name == nil {
				items := strings.Split(pathValue, "/")
				pkgNames[items[len(items)-1]] = pathValue
			} else {
				// alias pkg name
				pkgNames[importSpec.Name.Name] = pathValue
			}
		}
	}
	return pkgNames
}

// checkImport return the string list which not pass the check
func checkImport(importMap map[string]string) []string {
	pkgs := make([]string, 0)
	for _, element := range importMap {
		if pkgsToCheckImport.Contains(element) {
			pkgs = append(pkgs, element)
		}
	}
	return pkgs
}

// func checkFunCall()

func GoFileWalk(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

type visitor struct {
	fset      *token.FileSet
	importMap map[string]string
	funNodes  []*ast.FuncDecl
}

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	switch d := n.(type) {
	case *ast.SelectorExpr:
		if x, ok := d.X.(*ast.Ident); ok {
			// fmt.Println("CallExpr.Fun.X.Name:", x.Name)
			realPkgName := v.importMap[x.Name]
			if realPkgName != "" {
				// fmt.Println("realPkgName:", realPkgName)
				fullExprName := realPkgName + "#" + d.Sel.Name
				// simple function used
				if funToCheckUsed.Contains(fullExprName) {
					// TODO report
					fmt.Printf("File %s should check the usage of %s\n", v.fset.Position(d.Pos()), fullExprName)
				}
			}
		}
	}
	return v
}
