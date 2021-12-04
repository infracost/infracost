package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/infracost/infracost/internal/usage"
	"golang.org/x/tools/go/ast/astutil"
)

var PROVIDER string = "aws"

type duStruct struct {
	fieldType string // Bool, String, Int, Float, Exists
}

/*
	This scripts tries to migrate resources from the old structure to the new one located
	at "internal/resources".
	The script iterates over resource files and does the following stages:
		1. Find all occurrences of d.Get and u.Get methods and extract their key and store them.
			1.1. If the methods are not called with string literals, the script will fail for such resource.
		2. If more than one resource is defined in the file, then the script will fail.
*/
func main() {
	basePath := fmt.Sprintf("internal/providers/terraform/%s/", PROVIDER)
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		log.Fatal(err)
	}

	// Load usage file data
	referenceFile, err := usage.LoadReferenceFile()
	if err != nil {
		log.Fatal(err)
	}
	referenceFile.SetDefaultValues()

	allCount := 0
	migratedCount := 0
	problemFiles := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.Contains(file.Name(), "test") {
			continue
		}
		filePath := fmt.Sprintf("%s%s", basePath, file.Name())

		allCount += 1
		fmt.Printf(" %d  %s\n", allCount, filePath)
		isMigrated, err := migrateFile(filePath, referenceFile, basePath, file.Name())
		// isMigrated, err := migrateFile("internal/providers/terraform/aws/db_instance.go", referenceFile, "internal/providers/terraform/aws/", "db_instance.go")
		// break
		if isMigrated {
			migratedCount += 1
		} else {
			problemFiles = append(problemFiles, filePath)
		}
		if isMigrated && err == nil {
			fmt.Printf("\t %t\n", isMigrated)
		} else {
			fmt.Printf("\t %t  :: %s \n", isMigrated, err)
		}
	}
	fmt.Println()
	fmt.Printf("%d of %d resources can be migrated! The impossible files are: \n%s\n", migratedCount, allCount, strings.Join(problemFiles, "\n"))
}

func migrateFile(filePath string, referenceFile *usage.ReferenceFile, basePath, fileName string) (bool, error) {
	resFilePath := fmt.Sprintf("internal/resources/%s/%s", PROVIDER, fileName)
	if _, err := os.Stat(resFilePath); err == nil {
		return true, errors.New("manually migrated")
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return false, err
	}
	if isImpossibleWithGets(file) {
		return false, errors.New("bad d/u gets")
	}
	if isImpossibleWithDotGets(file) {
		return false, errors.New("dotted d/u gets")
	}
	if isImpossibleWithResourceDefsCount(file) {
		return false, errors.New("multiple resource defs")
	}
	if isImpossibleWithGetsTypes(file) {
		return false, errors.New("unknown d/u gets types")
	}

	_, resourceFile, err := doMigration(fset, file, referenceFile)
	if err != nil {
		return false, err
	}

	// saveFile(fset, pFile)

	err = saveFile(fset, resourceFile, resFilePath)
	if err != nil {
		return false, err
	}

	return true, nil
}

func isImpossibleWithGets(file *ast.File) bool {
	isImpossible := false
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		if isImpossible {
			return false
		}
		n := c.Node()
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selExpr.X.(*ast.Ident); ok {
					if (ident.Name == "d" || ident.Name == "u") && selExpr.Sel.Name == "Get" {
						argLit, ok := callExpr.Args[0].(*ast.BasicLit)
						if ok {
							if argLit.Kind != token.STRING {
								isImpossible = true
							}
						} else {
							isImpossible = true
						}
					}
				}
			}
		} else if pSelExpr, ok := n.(*ast.SelectorExpr); ok {
			if callExpr, ok := pSelExpr.X.(*ast.CallExpr); ok {
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok {
						if (ident.Name == "d" || ident.Name == "u") && pSelExpr.Sel.Name == "Type" {
							argLit, ok := callExpr.Args[0].(*ast.BasicLit)
							if ok {
								if argLit.Kind != token.STRING {
									isImpossible = true
								}
							} else {
								isImpossible = true
							}
						}
					}
				}
			}
		}
		return true
	})
	return isImpossible
}

func isImpossibleWithDotGets(file *ast.File) bool {
	isImpossible := false
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		if isImpossible {
			return false
		}
		n := c.Node()
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selExpr.X.(*ast.Ident); ok {
					if (ident.Name == "d" || ident.Name == "u") && selExpr.Sel.Name == "Get" {
						argLit, ok := callExpr.Args[0].(*ast.BasicLit)
						if ok {
							if argLit.Kind == token.STRING {
								if strings.Contains(argLit.Value, ".") {
									isImpossible = true
								}
							}
						} else {
							isImpossible = true
						}
					}
				}
			}
		} else if pSelExpr, ok := n.(*ast.SelectorExpr); ok {
			if callExpr, ok := pSelExpr.X.(*ast.CallExpr); ok {
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok {
						if (ident.Name == "d" || ident.Name == "u") && pSelExpr.Sel.Name == "Type" {
							argLit, ok := callExpr.Args[0].(*ast.BasicLit)
							if ok {
								if argLit.Kind == token.STRING {
									if strings.Contains(argLit.Value, ".") {
										isImpossible = true
									}
								}
							} else {
								isImpossible = true
							}
						}
					}
				}
			}
		}
		return true
	})
	return isImpossible
}

func isImpossibleWithGetsTypes(file *ast.File) bool {
	isImpossible := false
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		if isImpossible {
			return false
		}
		n := c.Node()
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if callExpr2, ok := selExpr.X.(*ast.CallExpr); ok {
					if selExpr2, ok := callExpr2.Fun.(*ast.SelectorExpr); ok {
						if identExpr, ok := selExpr2.X.(*ast.Ident); ok {
							if identExpr.Name == "d" || identExpr.Name == "u" {
								if argLit, ok := callExpr2.Args[0].(*ast.BasicLit); ok {
									if argLit.Kind == token.STRING {
										switch selExpr.Sel.Name {
										case "Map":
											isImpossible = true
										case "Array":
											isImpossible = true
										case "References":
											isImpossible = true
										}
									}
								}
							}
						}
					}
				}
			}
		} else if selExpr, ok := n.(*ast.SelectorExpr); ok {
			if XIdent, ok := selExpr.X.(*ast.Ident); ok {
				if XIdent.Name == "d" && selExpr.Sel.Name == "References" {
					isImpossible = true
				}
			}
		}
		return true
	})
	return isImpossible
}

func isImpossibleWithResourceDefsCount(file *ast.File) bool {
	isImpossible := false
	numberOfRegistryItems := 0
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		if isImpossible {
			return false
		}
		n := c.Node()
		if declExpr, ok := n.(*ast.FuncDecl); ok {
			if len(declExpr.Type.Results.List) == 1 {
				if starExpr, ok := declExpr.Type.Results.List[0].Type.(*ast.StarExpr); ok {
					if selExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
						if selExpr.Sel.Name == "RegistryItem" {
							numberOfRegistryItems += 1
							if numberOfRegistryItems > 1 {
								isImpossible = true
							}
						}
					}
				}
			}
		}
		return true
	})
	return isImpossible
}

func doMigration(fset *token.FileSet, file *ast.File, referenceFile *usage.ReferenceFile) (*ast.File, *ast.File, error) {
	registryFuncName := getRegistryFuncName(file)
	resourceFuncName, resourceName := getResourceFuncName(registryFuncName, file)
	if registryFuncName == "" || resourceFuncName == "" {
		return nil, nil, errors.New("invalid registry/resource func names")
	}
	resourceCamelName := strings.Replace(registryFuncName, "Get", "", 1)
	resourceCamelName = strings.Replace(resourceCamelName, "RegistryItem", "", 1)

	dsList := getDsList(file)
	usList := getUsList(file)

	resourceFile := &ast.File{
		Package: file.Package,
		Name:    file.Name,
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok:   token.IMPORT,
				Specs: make([]ast.Spec, 0),
			},
		},
		Comments: []*ast.CommentGroup{},
	}
	err := extractResourceFile(registryFuncName, resourceFuncName, resourceCamelName, file, resourceFile)
	if err != nil {
		return nil, nil, err
	}

	err = addResourceSchemaAndFuncs(resourceCamelName, resourceName, resourceFile, dsList, usList, referenceFile)
	if err != nil {
		return nil, nil, err
	}

	replaceDUs(resourceCamelName, resourceFile)
	addSchemaToResource(resourceCamelName, resourceFile)
	migrateImports(file, resourceFile)
	addProviderFunc(resourceCamelName, dsList, usList, file)
	fixRFuncName(resourceCamelName, file)

	return file, resourceFile, nil
}

func getRegistryFuncName(file *ast.File) string {
	funcName := ""
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if declExpr, ok := n.(*ast.FuncDecl); ok {
			if len(declExpr.Type.Results.List) == 1 {
				if starExpr, ok := declExpr.Type.Results.List[0].Type.(*ast.StarExpr); ok {
					if selExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
						if selExpr.Sel.Name == "RegistryItem" {
							funcName = declExpr.Name.Name
						}
					}
				}
			}
		}
		return true
	})
	return funcName
}

func getResourceFuncName(registryFuncName string, file *ast.File) (string, string) {
	funcName := ""
	resourceName := ""
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if declExpr, ok := n.(*ast.FuncDecl); ok {
			if declExpr.Name.Name == registryFuncName {
				astutil.Apply(n, nil, func(c2 *astutil.Cursor) bool {
					n2 := c2.Node()
					if keyValExpr, ok := n2.(*ast.KeyValueExpr); ok {
						if keyIdent, ok := keyValExpr.Key.(*ast.Ident); ok {
							if keyIdent.Name == "RFunc" {
								if valueIdent, ok := keyValExpr.Value.(*ast.Ident); ok {
									funcName = valueIdent.Name
								}
							} else if keyIdent.Name == "Name" {
								if valueLit, ok := keyValExpr.Value.(*ast.BasicLit); ok {
									resourceName = strings.Replace(valueLit.Value, fmt.Sprintf("%s_", PROVIDER), "", -1)
									resourceName = strings.Replace(resourceName, "\"", "", -1)
								}
							}
						}
					}
					return true
				})
			}
		}
		return true
	})
	return funcName, resourceName
}

func getDsList(file *ast.File) map[string]duStruct {
	result := make(map[string]duStruct, 0)
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if callExpr2, ok := selExpr.X.(*ast.CallExpr); ok {
					if selExpr2, ok := callExpr2.Fun.(*ast.SelectorExpr); ok {
						if identExpr, ok := selExpr2.X.(*ast.Ident); ok {
							if identExpr.Name == "d" {
								if argLit, ok := callExpr2.Args[0].(*ast.BasicLit); ok {
									if argLit.Kind == token.STRING {
										keyName := strings.Replace(argLit.Value, "\"", "", -1)
										if val, ok := result[keyName]; ok {
											if val.fieldType != "Exists" {
												return true
											}
										}
										result[keyName] = duStruct{
											fieldType: selExpr.Sel.Name,
										}
									}
								}
							}
						}
					}
				}
			}
		} else if pSelExpr, ok := n.(*ast.SelectorExpr); ok && pSelExpr.Sel.Name == "Type" {
			if callExpr, ok := pSelExpr.X.(*ast.CallExpr); ok {
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if identExpr, ok := selExpr.X.(*ast.Ident); ok {
						if identExpr.Name == "d" {
							if argLit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
								if argLit.Kind == token.STRING {
									keyName := strings.Replace(argLit.Value, "\"", "", -1)
									if val, ok := result[keyName]; ok {
										if val.fieldType != "Exists" {
											return true
										}
									}
									result[keyName] = duStruct{
										fieldType: "Exists",
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

func getUsList(file *ast.File) map[string]duStruct {
	result := make(map[string]duStruct, 0)
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if callExpr2, ok := selExpr.X.(*ast.CallExpr); ok {
					if selExpr2, ok := callExpr2.Fun.(*ast.SelectorExpr); ok {
						if identExpr, ok := selExpr2.X.(*ast.Ident); ok {
							if identExpr.Name == "u" {
								if argLit, ok := callExpr2.Args[0].(*ast.BasicLit); ok {
									if argLit.Kind == token.STRING {
										keyName := strings.Replace(argLit.Value, "\"", "", -1)
										if val, ok := result[keyName]; ok {
											if val.fieldType != "Exists" {
												return true
											}
										}
										result[keyName] = duStruct{
											fieldType: selExpr.Sel.Name,
										}
									}
								}
							}
						}
					}
				}
			}
		} else if pSelExpr, ok := n.(*ast.SelectorExpr); ok && pSelExpr.Sel.Name == "Type" {
			if callExpr, ok := pSelExpr.X.(*ast.CallExpr); ok {
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if identExpr, ok := selExpr.X.(*ast.Ident); ok {
						if identExpr.Name == "u" {
							if argLit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
								if argLit.Kind == token.STRING {
									keyName := strings.Replace(argLit.Value, "\"", "", -1)
									if val, ok := result[keyName]; ok {
										if val.fieldType != "Exists" {
											return true
										}
									}
									result[keyName] = duStruct{
										fieldType: "Exists",
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

func extractResourceFile(registryFuncName, resourceFuncName, resourceCamelName string, file *ast.File, resourceFile *ast.File) error {
	// First, move the body of the resource func to the resource file.
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		if c == nil {
			return true
		}
		if _, ok := c.Parent().(*ast.File); !ok {
			return true
		}
		n := c.Node()
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == resourceFuncName {
				funcDecl.Name.Name = "BuildResource"
				funcDecl.Type.Params = &ast.FieldList{}
				funcDecl.Recv = &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{
									Name: "resP",
									Obj:  &ast.Object{Kind: ast.Var, Name: "resP"},
								},
							},
							Type: &ast.StarExpr{
								X: &ast.Ident{
									Name: resourceCamelName,
									Obj:  &ast.Object{Kind: ast.Typ, Name: resourceCamelName},
								},
							},
						},
					},
				}
				resourceFile.Decls = append(resourceFile.Decls, funcDecl)
				c.Delete()
			}
		}
		return true
	})

	// Second, move other funcs to resource file.
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		shouldContinue := false
		var newDecl ast.Decl
		if c == nil {
			return true
		}
		if _, ok := c.Parent().(*ast.File); !ok {
			return true
		}
		n := c.Node()

		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name != registryFuncName && funcDecl.Name.Name != resourceFuncName {
				shouldContinue = true
				newDecl = funcDecl
			}
		} else if genDecl, ok := n.(*ast.GenDecl); ok {
			if genDecl.Tok != token.IMPORT {
				shouldContinue = true
				newDecl = genDecl
			}
		}

		if shouldContinue {
			resourceFile.Decls = append(resourceFile.Decls, newDecl)
			c.Delete()
		}
		return true
	})

	return nil
}

func addResourceSchemaAndFuncs(resourceCamelName, resourceName string, resourceFile *ast.File, dsList, usList map[string]duStruct, referenceFile *usage.ReferenceFile) error {
	fieldsList := &ast.FieldList{
		List: make([]*ast.Field, 0),
	}
	fieldsList.List = append(fieldsList.List, &ast.Field{
		Type: &ast.StarExpr{X: &ast.Ident{Name: "string"}},
		Names: []*ast.Ident{{
			Name: "Address",
			Obj: &ast.Object{
				Kind: ast.Var,
				Name: "Address",
			},
		}},
	})
	for key, val := range dsList {
		fieldsList.List = append(fieldsList.List, &ast.Field{
			Type: &ast.StarExpr{X: &ast.Ident{Name: duTypeToASTType(val.fieldType)}},
			Names: []*ast.Ident{{
				Name: strcase.ToCamel(key),
				Obj: &ast.Object{
					Kind: ast.Var,
					Name: strcase.ToCamel(key),
				},
			}},
		})
	}

	usageElts := make([]ast.Expr, 0)
	for key, val := range usList {
		fieldsList.List = append(fieldsList.List, &ast.Field{
			Type: &ast.StarExpr{X: &ast.Ident{Name: duTypeToASTType(val.fieldType)}},
			Names: []*ast.Ident{{
				Name: strcase.ToCamel(key),
				Obj: &ast.Object{
					Kind: ast.Var,
					Name: strcase.ToCamel(key),
				},
			}},
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("`infracost_usage:\"%s\"`", key),
			},
		})
		var defaultValue string
		usageDefaultValues := referenceFile.FindMatchingResourceUsage(fmt.Sprintf("%s_%s.foo", PROVIDER, resourceName))
		if usageDefaultValues == nil {
			log.Fatalf("nil usageData for: %s", resourceName)
		}
		for _, usageItem := range usageDefaultValues.Items {
			if usageItem.Key == key {
				switch val.fieldType {
				case "String":
					defaultValue = fmt.Sprintf("\"%s\"", usageItem.DefaultValue.(string))
				case "Int":
					defaultValue = fmt.Sprintf("%d", usageItem.DefaultValue.(int))
				case "Float":
					if fc, ok := usageItem.DefaultValue.(float64); ok {
						defaultValue = fmt.Sprintf("%f", fc)
					} else {
						defaultValue = fmt.Sprintf("%d", usageItem.DefaultValue.(int))
					}
				}
			}
		}
		if defaultValue == "" {
			log.Fatal("Empty default value")
		}
		usageElts = append(usageElts, &ast.CompositeLit{
			Elts: []ast.Expr{
				&ast.KeyValueExpr{
					Key:   &ast.Ident{Name: "Key"},
					Value: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", key)},
				},
				&ast.KeyValueExpr{
					Key: &ast.Ident{Name: "ValueType"},
					Value: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "schema"},
						Sel: &ast.Ident{Name: duTypeToSchemaType(val.fieldType)},
					},
				},
				&ast.KeyValueExpr{
					Key: &ast.Ident{Name: "DefaultValue"},
					Value: &ast.BasicLit{
						Kind:  duTypeToToken(val.fieldType),
						Value: defaultValue,
					},
				},
			},
		})
	}
	schemaDecl := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(resourceCamelName),
				Type: &ast.StructType{
					Fields: fieldsList,
				},
			},
		},
	}
	usageSchemaDecl := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{{Name: fmt.Sprintf("%sUsageSchema", resourceCamelName)}},
				Values: []ast.Expr{
					&ast.CompositeLit{
						Type: &ast.ArrayType{
							Elt: &ast.StarExpr{
								X: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "schema"},
									Sel: &ast.Ident{Name: "UsageItem"},
								},
							},
						},
						Elts: usageElts,
					},
				},
			},
		},
	}

	// Add PopulateUsage func
	popDecl := &ast.FuncDecl{
		Name: &ast.Ident{Name: "PopulateUsage"},
		Recv: &ast.FieldList{
			List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "resP"}},
				Type:  &ast.StarExpr{X: &ast.Ident{Name: resourceCamelName}},
			}},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{{
					Names: []*ast.Ident{{
						Name: "u",
						Obj:  &ast.Object{Kind: ast.Var, Name: "u"},
					}},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "schema"},
							Sel: &ast.Ident{Name: "UsageData"},
						},
					},
				}},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "resources"},
							Sel: &ast.Ident{Name: "PopulateArgsWithUsage"},
						},
						Args: []ast.Expr{
							&ast.Ident{Name: "resP", Obj: &ast.Object{Kind: ast.Var, Name: "resP"}},
							&ast.Ident{Name: "u", Obj: &ast.Object{Kind: ast.Var, Name: "u"}},
						},
					},
				},
			},
		},
	}

	resourceFile.Decls = append([]ast.Decl{resourceFile.Decls[0], schemaDecl, usageSchemaDecl, popDecl}, resourceFile.Decls[1:]...)
	return nil
}

func replaceDUs(resourceCamelName string, resourceFile *ast.File) {
	// Replace Gets
	astutil.Apply(resourceFile, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if callExpr2, ok := selExpr.X.(*ast.CallExpr); ok {
					if selExpr2, ok := callExpr2.Fun.(*ast.SelectorExpr); ok {
						if identExpr, ok := selExpr2.X.(*ast.Ident); ok {
							if identExpr.Name == "u" || identExpr.Name == "d" {
								if argLit, ok := callExpr2.Args[0].(*ast.BasicLit); ok {
									if argLit.Kind == token.STRING {
										keyName := strings.Replace(argLit.Value, "\"", "", -1)
										var replacementNode ast.Node
										if selExpr.Sel.Name == "Exists" {
											replacementNode = &ast.BinaryExpr{
												Op: token.NEQ,
												Y:  &ast.Ident{Name: "nil"},
												X: &ast.SelectorExpr{
													X: &ast.Ident{
														Name: "resP",
														Obj: &ast.Object{
															Kind: ast.Var,
															Name: "resP",
														},
													},
													Sel: &ast.Ident{
														Name: strcase.ToCamel(keyName),
													},
												},
											}
										} else {
											replacementNode = &ast.StarExpr{
												X: &ast.SelectorExpr{
													X: &ast.Ident{
														Name: "resP",
														Obj: &ast.Object{
															Kind: ast.Var,
															Name: "resP",
														},
													},
													Sel: &ast.Ident{
														Name: strcase.ToCamel(keyName),
													},
												},
											}
										}
										c.Replace(replacementNode)
									}
								}
							}
						}
					}
				}
			}
		} else if binExpr, ok := n.(*ast.BinaryExpr); ok && binExpr.Op == token.NEQ {
			if pSelExpr, ok := binExpr.X.(*ast.SelectorExpr); ok && pSelExpr.Sel.Name == "Type" {
				if callExpr, ok := pSelExpr.X.(*ast.CallExpr); ok {
					if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
						if identExpr, ok := selExpr.X.(*ast.Ident); ok {
							if identExpr.Name == "u" || identExpr.Name == "d" {
								if argLit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
									if argLit.Kind == token.STRING {
										keyName := strings.Replace(argLit.Value, "\"", "", -1)
										var replacementNode ast.Node
										replacementNode = &ast.BinaryExpr{
											Op: token.NEQ,
											Y:  &ast.Ident{Name: "nil"},
											X: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "resP",
													Obj: &ast.Object{
														Kind: ast.Var,
														Name: "resP",
													},
												},
												Sel: &ast.Ident{
													Name: strcase.ToCamel(keyName),
												},
											},
										}
										c.Replace(replacementNode)
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

	// Replace d.X like d.Address
	astutil.Apply(resourceFile, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if selExpr, ok := n.(*ast.SelectorExpr); ok {
			if XIdent, ok := selExpr.X.(*ast.Ident); ok {
				if XIdent.Name == "d" {
					XIdent.Name = "resP"
					c.Replace(&ast.StarExpr{X: selExpr})
				}
			}
		}
		return true
	})

	// Replace u != nil
	astutil.Apply(resourceFile, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if binExpr, ok := n.(*ast.BinaryExpr); ok {
			if binX, ok := binExpr.X.(*ast.Ident); ok {
				if binY, ok := binExpr.Y.(*ast.Ident); ok {
					if binY.Name == "nil" && binX.Name == "u" {
						binX.Name = "resP"
					}
				}
			}
		}
		return true
	})

	// Replace d, u method calls
	astutil.Apply(resourceFile, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if len(callExpr.Args) == 1 {
				identExpr1, ok1 := callExpr.Args[0].(*ast.Ident)
				if ok1 && (identExpr1.Name == "u" || identExpr1.Name == "d") {
					callExpr.Args = []ast.Expr{
						&ast.Ident{Name: "resP"},
					}
				}
			}
			if len(callExpr.Args) == 2 {
				identExpr1, ok1 := callExpr.Args[0].(*ast.Ident)
				identExpr2, ok2 := callExpr.Args[1].(*ast.Ident)
				if (ok1 && ok2) && ((identExpr1.Name == "u" && identExpr2.Name == "d") || (identExpr1.Name == "d" && identExpr2.Name == "u")) {
					callExpr.Args = []ast.Expr{
						&ast.Ident{Name: "resP"},
					}
				}
			}
		}
		return true
	})

	// Replace d, u func defs
	astutil.Apply(resourceFile, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "PopulateUsage" {
				return true
			}
			if len(funcDecl.Type.Params.List) == 1 {
				name := funcDecl.Type.Params.List[0].Names[0].Name
				if name == "u" || name == "d" {
					funcDecl.Type.Params.List = []*ast.Field{{
						Names: []*ast.Ident{{Name: "resP"}},
						Type: &ast.StarExpr{
							X: &ast.Ident{Name: resourceCamelName},
						},
					}}
				}
			} else if len(funcDecl.Type.Params.List) == 2 {
				name1 := funcDecl.Type.Params.List[0].Names[0].Name
				name2 := funcDecl.Type.Params.List[1].Names[0].Name
				if (name1 == "u" && name2 == "d") || (name1 == "d" && name2 == "u") {
					funcDecl.Type.Params.List = []*ast.Field{{
						Names: []*ast.Ident{{Name: "resP"}},
						Type: &ast.StarExpr{
							X: &ast.Ident{Name: resourceCamelName},
						},
					}}
				}
			}
		}
		return true
	})

}

func addSchemaToResource(resourceCamelName string, resourceFile *ast.File) {
	astutil.Apply(resourceFile, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if declExpr, ok := n.(*ast.FuncDecl); ok {
			if declExpr.Name.Name == "BuildResource" {
				astutil.Apply(n, nil, func(c2 *astutil.Cursor) bool {
					n2 := c2.Node()
					if compLit, ok := n2.(*ast.CompositeLit); ok {
						if selExpr, ok := compLit.Type.(*ast.SelectorExpr); ok {
							if selExpr.Sel.Name == "Resource" {
								compLit.Elts = append(compLit.Elts, &ast.KeyValueExpr{
									Key:   &ast.Ident{Name: "UsageSchema"},
									Value: &ast.Ident{Name: fmt.Sprintf("%sUsageSchema", resourceCamelName)},
								})
							}
						}
					}
					return true
				})
			}
		}
		return true
	})
}

func migrateImports(file, resourceFile *ast.File) {
	resourceFileImports := resourceFile.Decls[0].(*ast.GenDecl)
	resourceFileImports.Specs = append(resourceFileImports.Specs, &ast.ImportSpec{
		Path: &ast.BasicLit{Kind: token.STRING, Value: "\"github.com/infracost/infracost/internal/resources\""},
	})
	resourceFileImports.Specs = append(resourceFileImports.Specs, &ast.ImportSpec{
		Path: &ast.BasicLit{Kind: token.STRING, Value: "\"github.com/infracost/infracost/internal/schema\""},
	})
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()

		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			astutil.Apply(genDecl, nil, func(c2 *astutil.Cursor) bool {
				n2 := c2.Node()
				if impSpec, ok := n2.(*ast.ImportSpec); ok {
					impPath := strings.Replace(impSpec.Path.Value, "\"", "", -1)
					if shouldRemoveImport(impPath) {
						c2.Delete()
					} else if !isImportNeededForProvider(impPath) {
						resourceFileImports.Specs = append(resourceFileImports.Specs, impSpec)
						c2.Delete()
					}
				}
				return true
			})
		}

		return true
	})
	file.Decls[0].(*ast.GenDecl).Specs = append(file.Decls[0].(*ast.GenDecl).Specs, &ast.ImportSpec{
		Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"github.com/infracost/infracost/internal/resources/%s\"", PROVIDER)},
	})
}

func addProviderFunc(resourceCamelName string, dsList, usList map[string]duStruct, file *ast.File) {
	funcBodyList := make([]ast.Stmt, 0)

	resourceELTs := []ast.Expr{
		&ast.KeyValueExpr{
			Key: &ast.Ident{Name: "Address"},
			Value: &ast.CallExpr{
				Fun: &ast.Ident{Name: "strPtr"},
				Args: []ast.Expr{
					&ast.SelectorExpr{
						X:   &ast.Ident{Name: "d"},
						Sel: &ast.Ident{Name: "Address"},
					},
				},
			},
		},
	}

	for key, val := range dsList {
		resourceELTs = append(resourceELTs, &ast.KeyValueExpr{
			Key: &ast.Ident{Name: strcase.ToCamel(key)},
			Value: &ast.CallExpr{
				Fun: &ast.Ident{Name: duTypeToPtrCall(val.fieldType)},
				Args: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							Sel: &ast.Ident{Name: duTypeToResCall(val.fieldType)},
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									Sel: &ast.Ident{Name: "Get"},
									X:   &ast.Ident{Name: "d"},
								},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", key)},
								},
							},
						},
					},
				},
			},
		})
	}

	funcBodyList = append(funcBodyList, &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{
			&ast.Ident{Name: "resP"},
		},
		Rhs: []ast.Expr{
			&ast.UnaryExpr{
				Op: token.AND,
				X: &ast.CompositeLit{
					Type: &ast.SelectorExpr{
						X:   &ast.Ident{Name: PROVIDER},
						Sel: &ast.Ident{Name: resourceCamelName},
					},
					Elts: resourceELTs,
				},
			},
		},
	})

	funcBodyList = append(funcBodyList, &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "resP"},
				Sel: &ast.Ident{Name: "PopulateUsage"},
			},
			Args: []ast.Expr{
				&ast.Ident{Name: "u"},
			},
		},
	})

	funcBodyList = append(funcBodyList, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "resP"},
					Sel: &ast.Ident{Name: "BuildResource"},
				},
			},
		},
	})

	decl := &ast.FuncDecl{
		Name: &ast.Ident{Name: fmt.Sprintf("New%s", resourceCamelName)},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "d"}},
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "schema"},
								Sel: &ast.Ident{Name: "ResourceData"},
							},
						},
					},
					{
						Names: []*ast.Ident{{Name: "u"}},
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "schema"},
								Sel: &ast.Ident{Name: "UsageData"},
							},
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "schema"},
								Sel: &ast.Ident{Name: "Resource"},
							},
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: funcBodyList,
		},
	}
	file.Decls = append(file.Decls, decl)
}

func fixRFuncName(resourceCamelName string, file *ast.File) {
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if keyValExpr, ok := n.(*ast.KeyValueExpr); ok {
			if keyIdent, ok := keyValExpr.Key.(*ast.Ident); ok {
				if keyIdent.Name == "RFunc" {
					keyValExpr.Value.(*ast.Ident).Name = fmt.Sprintf("New%s", resourceCamelName)
				}
			}
		}
		return true
	})
}

func shouldRemoveImport(importPath string) bool {
	switch importPath {
	case "github.com/tidwall/gjson":
		return true
	default:
		return false
	}
}

func isImportNeededForProvider(importPath string) bool {
	switch importPath {
	case fmt.Sprintf("github.com/infracost/infracost/internal/resources/%s", PROVIDER):
		return true
	case "github.com/infracost/infracost/internal/schema":
		return true
	default:
		return false
	}
}

func duTypeToSchemaType(duType string) string {
	switch duType {
	case "String":
		return "String"
	case "Int":
		return "Int64"
	case "Float":
		return "Float64"
	default:
		panic(fmt.Sprintf("Unsupported duTypeToSchemaType type %s", duType))
	}
}

func duTypeToToken(duType string) token.Token {
	switch duType {
	case "String":
		return token.STRING
	case "Int":
		return token.INT
	case "Float":
		return token.FLOAT
	default:
		panic(fmt.Sprintf("Unsupported duTypeToToken type %s", duType))
	}
}

func duTypeToASTType(duType string) string {
	switch duType {
	case "String":
		return "string"
	case "Bool":
		return "bool"
	case "Int":
		return "int64"
	case "Float":
		return "float64"
	case "Exists":
		return "string"
	default:
		panic(fmt.Sprintf("Unsupported duTypeToASTType type %s", duType))
	}
}

func duTypeToResCall(duType string) string {
	switch duType {
	case "String":
		return "String"
	case "Bool":
		return "Bool"
	case "Int":
		return "Int"
	case "Float":
		return "Float"
	case "Exists":
		return "String"
	default:
		panic(fmt.Sprintf("Unsupported duTypeToResCall type %s", duType))
	}
}

func duTypeToPtrCall(duType string) string {
	switch duType {
	case "String":
		return "strPtr"
	case "Bool":
		return "boolPtr"
	case "Int":
		return "intPtr"
	case "Float":
		return "floatPtr"
	case "Exists":
		return "strPtr"
	default:
		panic(fmt.Sprintf("Unsupported duTypeToPtrCall type %s", duType))
	}
}

func saveFile(fset *token.FileSet, file *ast.File, filePath string) error {
	f, err := os.Create(filePath)
	defer f.Close()

	if err != nil {
		return err
	}

	printer.Fprint(f, fset, file)
	return nil

	// fmt.Println("#################################")
	// printer.Fprint(os.Stdout, fset, file)
	// fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
}
