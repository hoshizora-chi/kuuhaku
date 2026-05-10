package config_reader

import (
	"fmt"
	"os"
	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_analyzer"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
)

var ErrUnrecognizedExtension = fmt.Errorf("Extension is unrecognized")

func ReadConfig(configFilePath string, isDebugAnalyzer bool, isDebugParser bool, isDebugReader bool) (*kuuhaku_analyzer.AnalyzerResult, []error) {
	configGrammar, err := os.ReadFile(configFilePath)
	helper.Check(err)
	if isDebugReader {
		fmt.Println(string(configGrammar))
	}
	ast, errs := kuuhaku_parser.Parse(string(configGrammar))
	if len(errs) != 0 {
		return nil, errs
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, isDebugAnalyzer)
	if len(errs) != 0 {
		return nil, errs
	}
	return &res, []error{}
}
