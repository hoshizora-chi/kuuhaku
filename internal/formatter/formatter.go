package formatter

import (
	"fmt"
	"os"

	"github.com/ciii1/kuuhaku/internal/config_reader"
	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_runtime"
)

type FormattedFile struct {
	Content string
	Filename string
}

func Format(filename string, targetName string, formatConfig string, isDebugRuntime bool, isDebugAnalyzer bool, isDebugParser bool, isDebugReader bool, isStatic bool) error {
	targetFile, err := os.ReadFile(filename)
	helper.Check(err)
	formattedFile := FormattedFile{
		Content: string(targetFile),
		Filename: filename,
	}

	if isDebugReader {
		fmt.Println("Format(), content:\n", formattedFile.Content)
		fmt.Println("Formatting " + formattedFile.Filename + "...")
	}

	res, errs := config_reader.ReadConfig(formatConfig, isDebugAnalyzer, isDebugParser, isDebugReader)
	if len(errs) != 0 {
		fmt.Println("Error while reading configuration, file " + formatConfig + ":")
		helper.DisplayAllErrors(errs)
		return err
	}

	if !isStatic {
		strRes, err := kuuhaku_runtime.Format(formattedFile.Content, res, true, isDebugRuntime)
		if err != nil {
			fmt.Println("Error while transforming the target file, file " + formattedFile.Filename + ":")
			fmt.Println(err.Error())
			return err
		}

		f, err := os.OpenFile(targetName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			fmt.Println("Error while writing the file: " + targetName + ":")
			fmt.Println(err.Error())
			return err
		}
		defer f.Close()
		helper.Check(err)

		_, err = f.WriteString(strRes)
		helper.Check(err)
	}
	return nil
}
