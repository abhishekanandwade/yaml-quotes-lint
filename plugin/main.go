package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v3"
)

type analyzerPlugin struct{}

func (analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		{
			Name: "yamlquotes",
			Doc:  "checks for unquoted strings in YAML files",
			Run:  run,
		},
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.OtherFiles {
		if !strings.HasSuffix(file, ".yaml") && !strings.HasSuffix(file, ".yml") {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var rootNode yaml.Node
		if err := yaml.Unmarshal(data, &rootNode); err != nil {
			continue
		}

		issues := checkNode(&rootNode)
		for _, issue := range issues {
			pass.Reportf(0, "YAML Quote Issue: %s in file %s", issue, file)
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

// Required export for golangci-lint plugin
var AnalyzerPlugin analyzerPlugin
