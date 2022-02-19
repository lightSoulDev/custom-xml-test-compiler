package xmlParser

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var (
	modulePathRegex = regexp.MustCompile(`^([a-zA-z0-9_-]*\/)+([a-zA-z0-9]*)$`)
	commonPathRegex = regexp.MustCompile(`^common\/([a-zA-z0-9]*)$`)
)

type TestConfig struct {
    XMLName 			xml.Name `xml:"tests"`
	TestNodes   		[]Node `xml:",any"`
	Declarations 		[]Node
	UnresolvedTests		[]Node
	ResolvedTests		[]Node
	FilePath			string
}

type XmlParser struct {
	config *Config
	ConfigXML string
}

func New(config *Config) *XmlParser {

	config.CommonPath = config.AppData + "/Common"
	config.ConfigsPath = config.AppData + "/Configs"

	return &XmlParser{
		config: config,
	}
}

func (x *XmlParser) ResolveModulePath(id string) (string, error) {

	if len(modulePathRegex.FindStringSubmatch(id)) == 0 {
		return "", fmt.Errorf("%s module path was not resolved", id)
	}

	var path string

	testName := id[strings.LastIndex(id, "/")+1:]

	mathes := commonPathRegex.FindStringSubmatch(id)
	if len(mathes) > 0 {
		path = x.config.CommonPath + "/common.xml"
	} else if strings.HasPrefix(id, "common/") {
		path = strings.Replace(strings.Replace(id, "/" + testName, ".xml", 1), "common/", x.config.CommonPath + "/", 1)
	} else {
		path = x.config.ConfigsPath + "/" + strings.Replace(id, "/" + testName, ".xml", 1)
	}

	return path, nil
}
 
func (x *XmlParser) GetTestConfig(path string, root ...bool) (TestConfig, error) {
	xmlFile, err := os.Open(path)
	if err != nil {
		return TestConfig{}, err
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)
	if root != nil {
		x.ConfigXML = string(byteValue)
	}

	var testConfig TestConfig
	xml.Unmarshal(byteValue, &testConfig)

	testConfig.FilePath = path

	for _, n := range testConfig.TestNodes {
		tag := n.XMLName.Local
		if tag == "import" {

			moduleName, err := x.GetXmlAttribute(n.Attrs, "module")
			if err != nil {
				return TestConfig{}, fmt.Errorf("node %s %s", tag, err.Error())
			}

			modulePath, err := x.ResolveModulePath(moduleName + "/All")
			if err != nil {
				return TestConfig{}, err
			}

			declarations, err := x.GetDeclarations(modulePath)
			if err != nil {
				return TestConfig{}, err
			}

			testConfig.Declarations = append(testConfig.Declarations, declarations...)

		} else if tag == "test" {
			testConfig.UnresolvedTests = append(testConfig.UnresolvedTests, n)
		} else {
			replace := false
			index := -1

			for i, d := range testConfig.Declarations {
				if d.XMLName.Local == n.XMLName.Local {
					replace = true
					index = i
				}
			}
			if replace {
				testConfig.Declarations = append(testConfig.Declarations[:index], testConfig.Declarations[index+1:]...)
				testConfig.Declarations = append(testConfig.Declarations, n)
			} else {
				testConfig.Declarations = append(testConfig.Declarations, n)
			}
		}
	}

	return testConfig, nil
}

func (x *XmlParser) GetDeclarations(path string) ([]Node, error) {
	testConfig, err := x.GetTestConfig(path)
	if err != nil {
		return []Node{}, err
	}

	return testConfig.Declarations, nil
}

func (x *XmlParser) GetTestDeclaration(path string, testName string) (Node, error) {
	testConfig, err := x.GetTestConfig(path)
	if err != nil {
		return Node{}, err
	}

	for _, n := range testConfig.Declarations {
		if n.XMLName.Local == testName {
			return n, nil
		}
	}

	return Node{}, fmt.Errorf("no declarations found for %s in %s", testName, path)
}

func (x *XmlParser) ResolveConfig(path string) (TestConfig, error) {

	testConfig, err := x.GetTestConfig(path, true)
	if err != nil {
		return TestConfig{}, err
	}

	for _, n := range testConfig.UnresolvedTests {
		var id string

		for _, a := range n.Attrs {
			if a.Name.Local == "id" {
				id = a.Value
				break
			}
		}

		pathMatches := modulePathRegex.FindStringSubmatch(id)

		if len(pathMatches) > 0 {
			modulePath, err := x.ResolveModulePath(id)
			if err != nil {
				return TestConfig{}, err
			}

			testName := id[strings.LastIndex(id, "/")+1:]
			testDeclaration, err := x.GetTestDeclaration(modulePath, testName)
			if err != nil {
				return TestConfig{}, err
			}

			testDeclaration.XMLName = xml.Name{ Local: strings.ReplaceAll(id, "/", "."), Space: "" }
			testDeclaration.Module = modulePath
			testConfig.ResolvedTests = append(testConfig.ResolvedTests, testDeclaration)
		} else {
			resolved := false
			for _, n := range testConfig.Declarations {
				if n.XMLName.Local == id {
					n.Module = path
					testConfig.ResolvedTests = append(testConfig.ResolvedTests, n)
					resolved = true
					break
				}
			}
			if !resolved {
				return TestConfig{}, fmt.Errorf("<test id=\"%s\" ... /> was not resolved", id)
			}
		}
	}

	for _, n := range testConfig.ResolvedTests {

		modulePath := n.Module

		err := x.NestedPointerWalk(&n, func (n *Node, p *Node) (bool, error) {

			if n.XMLName.Local == "test" {
				id, err := x.GetXmlAttribute(n.Attrs, "id")
				if err != nil {
					return false, err
				}

				if modulePath == path {
					pathMatches := modulePathRegex.FindStringSubmatch(id)
					if len(pathMatches) == 0 {
						found := false
						for _, d := range testConfig.Declarations {
							if d.XMLName.Local == id {
								if p.XMLName.Local != d.XMLName.Local && string(p.Content) != string(d.Content) {
									(*n) = d
									found = true
								} else {
									return false, fmt.Errorf("recursion links are not allowed")
								}
								break
							}
						}
						if !found {
							return false, fmt.Errorf("no declarations found for %s in %s", id, modulePath)
						}
					} else {
						actualModulePath, err := x.ResolveModulePath(id)
						if err != nil {
							return false, err
						}

						testName := id[strings.LastIndex(id, "/")+1:]
						d, err := x.GetTestDeclaration(actualModulePath, testName)
						if err != nil {
							return false, err
						}
			
						d.XMLName = xml.Name{ Local: strings.ReplaceAll(id, "/", "."), Space: "" }
						d.Module = actualModulePath

						if p.XMLName.Local != d.XMLName.Local && string(p.Content) != string(d.Content) {

							err := x.NestedWalk(d, func (dn Node, dp Node) (bool, error) {
								if dn.XMLName.Local == "test" {
									dnId, err := x.GetXmlAttribute(dn.Attrs, "id")

									if err != nil {
										return false, err
									}

									if dnId == strings.ReplaceAll(p.XMLName.Local, ".", "/") {
										return false, fmt.Errorf(
											"nested recoursion from %s to %s with child <test id=\"%s\" .../>",
											p.XMLName.Local,
											d.XMLName.Local,
											d.XMLName.Local,
										)
									}
								}
								return len(dn.Nodes) > 0, nil
							})

							if err != nil {
								return false, err
							} else {
								(*n) = d
							}
						} else {
							return false, fmt.Errorf("recursion links are not allowed")
						}
					}
				} else {
					pathMatches := modulePathRegex.FindStringSubmatch(id)
					if len(pathMatches) > 0 {
						subModulePath, err := x.ResolveModulePath(id)
						if err != nil {
							return false, err
						}
			
						testName := id[strings.LastIndex(id, "/")+1:]
						d, err := x.GetTestDeclaration(subModulePath, testName)
						if err != nil {
							return false, err
						}

						d.XMLName = xml.Name{ Local: strings.ReplaceAll(id, "/", "."), Space: "" }
						d.Module = subModulePath

						if p.XMLName.Local != d.XMLName.Local && string(p.Content) != string(d.Content) {
							(*n) = d
						} else {
							return false, fmt.Errorf("recursion links are not allowed")
						}
					} else {
						actualModulePath := modulePath
						if p.Module != "" && p.Module != actualModulePath {
							actualModulePath = p.Module
						}

						declarations, err := x.GetDeclarations(actualModulePath)
						if err != nil {
							return false, err
						}
	
						for _, d := range declarations {
							if d.XMLName.Local == id {
								if p.XMLName.Local != d.XMLName.Local && string(p.Content) != string(d.Content) {
									(*n) = d
								} else {
									return false, fmt.Errorf("recursion links are not allowed")
								}
								break
							}
						}
					}
				}
			}

			return len(n.Nodes) > 0, nil
		})
		
		if err != nil {
			return TestConfig{}, err
		}
	}

	return testConfig, nil
}

func (x *XmlParser) XmlTestConfig(testConfig TestConfig) (string) {

	result := ""

	for _, n := range testConfig.ResolvedTests {
		result += x.nodeXml(&n, 1)
	}

	return fmt.Sprintf("<tests>%s</tests>", result)
}

func (x *XmlParser) ReprTestConfig(testConfig TestConfig) (string) {

	result := ""
	result += "============= TestConfig ================\n"

	result += "| FilePath: " + testConfig.FilePath + "\n"

	result += "=========== Declared Tests ==============\n"

	for _, n := range testConfig.Declarations {
		result += "| " + n.XMLName.Local + "\n"
	}

	result += "=========== Resolved Tests ==============\n"

	for _, n := range testConfig.ResolvedTests {
		result += "| " + n.XMLName.Local + " (module: " + n.Module + ")\n"

		depth := 1

		x.Walk(n.Nodes, &depth, func (n Node) (bool, error) {
			result += strings.Repeat(">	", depth) + n.XMLName.Local + " " + n.JsonAttributes() + "\n"
			return len(n.Nodes) > 0, nil
		})
	}

	return result
}

type JsonNodeAttr struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type JsonInstructionNode struct {
	Name 			string `json:"name"`
	Atrrs			[]JsonNodeAttr `json:"attr"`
	Instructions	[]JsonInstructionNode `json:"instructions"`
}

type JsonTestNode struct {
	Name 			string `json:"name"`
	Module			string `json:"module"`
	Instructions	[]JsonInstructionNode `json:"instructions"`
}

type JsonTestConfig struct {
	Path 	string `json:"path"`
	Tests 	[]JsonTestNode `json:"tests"`
}

func (x *XmlParser) JsonTestConfig(testConfig TestConfig) string {

	result := JsonTestConfig{}
	result.Path = testConfig.FilePath

	for _, tn := range testConfig.ResolvedTests {

		jtn := JsonTestNode{ Name: tn.XMLName.Local, Module: tn.Module } 

		for _, in := range tn.Nodes {
			jna := []JsonNodeAttr{}
			
			for _, a := range in.Attrs {
				jna = append(jna, JsonNodeAttr{ Name: a.Name.Local, Value: a.Value})
			}

			jin := JsonInstructionNode{ Name: in.XMLName.Local, Atrrs: jna}
			
			x.ParseJsonInstructionTree(in.Nodes, &jin)

			jtn.Instructions = append(jtn.Instructions, jin)
		}

		result.Tests = append(result.Tests, jtn)
	}

	b, err := json.Marshal(result)
	if err != nil {
		return "{}"
	}

	return string(b)
}

func (x *XmlParser) ParseJsonNodeAtrributes(attrs []xml.Attr) []JsonNodeAttr {
	result := []JsonNodeAttr{}
		
	for _, a := range attrs {
		result = append(result, JsonNodeAttr{Name: a.Name.Local, Value: a.Value})
	}

	return result
}

func (x *XmlParser) ParseJsonInstructionTree(nodes []Node, target *JsonInstructionNode) {

	for _, n := range nodes {
		temp := &JsonInstructionNode{ Name: n.XMLName.Local, Atrrs: x.ParseJsonNodeAtrributes(n.Attrs)}

		if len(n.Nodes) > 0 {
			x.ParseJsonInstructionTree(n.Nodes, temp)
		}

		target.Instructions = append((*target).Instructions, *temp)
	}
}