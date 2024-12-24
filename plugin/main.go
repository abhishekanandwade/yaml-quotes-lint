package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v3"
)

var Analyzer = &analysis.Analyzer{
	Name: "yamlquotes",
	Doc:  "checks for unquoted strings in YAML files",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, filename := range pass.OtherFiles {
		if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
			continue
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}

		var rootNode yaml.Node
		if err := yaml.Unmarshal(data, &rootNode); err != nil {
			continue
		}

		issues := checkNode(&rootNode)
		for _, issue := range issues {
			pass.Report(analysis.Diagnostic{
				Pos:     0, // File start
				Message: issue,
			})
		}
	}
	return nil, nil
}

func checkNode(node *yaml.Node) []string {
	var issues []string

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			valueNode := node.Content[i+1]
			issues = append(issues, checkNode(valueNode)...)
		}

	case yaml.SequenceNode:
		for _, child := range node.Content {
			issues = append(issues, checkNode(child)...)
		}

	case yaml.ScalarNode:
		if node.Tag == "!!str" && node.Value != "" && node.Style != yaml.DoubleQuotedStyle {
			issues = append(issues, fmt.Sprintf("Value '%s' is not double-quoted", node.Value))
		}
	}

	return issues
}

// New provides the plugin entry point for golangci-lint.
func New(conf any) ([]*analysis.Analyzer, error) {
	fmt.Printf("My configuration (%[1]T): %#[1]v\n", conf)

	return []*analysis.Analyzer{Analyzer}, nil
}
