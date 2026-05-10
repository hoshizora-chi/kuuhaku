package main

import (
	"flag"
	"strconv"

	"github.com/ciii1/kuuhaku/internal/formatter"
)

func main() {
	flag.Usage = PrintHelp
	var isDebugAnalyzer = flag.Bool("debug-analyzer", false, "Print debug messages for the analyzer")
	var isDebugParser = flag.Bool("debug-parser", false, "Print debug messages for the parser")
	var isDebugRuntime = flag.Bool("debug-runtime", false, "Print debug messages for the runtime")
	var isDebugReader = flag.Bool("debug-reader", false, "Print debug messages for the reader")
	var isStatic = flag.Bool("static", false, "Stop after analyzing the config file")
	flag.Parse()

	if len(flag.Args()) == 3 {
		filename := flag.Arg(0)
		targetName := flag.Arg(1)
		configName := flag.Arg(2)
		formatter.Format(filename, targetName, configName, *isDebugRuntime, *isDebugAnalyzer, *isDebugParser, *isDebugReader, *isStatic)
	} else {
		println("Expected exactly 3 argument, got " + strconv.Itoa(len(flag.Args())))
		PrintHelp()
	}
}

func PrintHelp() {
	println("Kuuhaku - A simple compiler builder")
	println("")
	println("Usage:")
	println("kuuhaku <flags> <filename> <target filepath> <config path>")
	println("Filename is the file to be compiled.")
	println("Config path is the filepath to the transformer configuration to be used.")
	println("")
	println("Flags:")
	println("-recursive\t\tProcess directories recursively")
	println("-debug-analyzer\t\tPrint debug messages for the analyzer")
	println("-debug-parser\t\tPrint debug messages for the parser")
	println("-debug-runtime\t\tPrint debug messages for the runtime")
	println("-debug-reader\t\tPrint debug messages for the file reader")
	println("")
	println("Exiting...")
}
