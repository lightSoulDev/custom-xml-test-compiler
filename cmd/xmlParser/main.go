package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/lightSoulDev/pixi-xml-test-compiler/internal/xmlParser"
)

var (
	configPath string
	inputPath  string
	outputPath string
)

func init() {
	flag.StringVar(&configPath, "config", "configs/config.toml", "Path to parser config.toml file.")
	flag.StringVar(&inputPath, "input", "appData/Configs/full.xml", "Path to input xml file.")
	flag.StringVar(&outputPath, "out", "out/config.xml", "Path to output xml file.")
}

func main() {
	flag.Parse()

	parserConfig := xmlParser.NewConfig()
	_, err := toml.DecodeFile(configPath, parserConfig)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	parser := xmlParser.New(parserConfig)
	testConfig, err := parser.ResolveConfig(inputPath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	payload := []byte(parser.XmlTestConfig(testConfig))

	err = os.WriteFile(outputPath, payload, 0644)
	if err != nil {
		panic(err)
	}
}
