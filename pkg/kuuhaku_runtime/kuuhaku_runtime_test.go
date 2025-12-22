package kuuhaku_runtime

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_analyzer"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_runtime/test_format/khk"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_runtime/test_format/khk_array"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestRuntime1(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("SEARCH_MODE E{C D = `\"hello\"`} C{<a>} D{<b>}")
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("abababaa", &res, false, false)

	if len(errs) != 0 {
		println("Expected runtime errors length to be 0, got " + strconv.Itoa(len(errs)))
		println(err.Error())
		t.Fatal()
	}

	if strRes != "[[[a],[b]]][[[a],[b]]][[[a],[b]]]aa" {
		println("Expected the string to be \"[[[a],[b]]][[[a],[b]]][[[a],[b]]]aa\", got " + strRes)
		t.Fatal()
	}
}

func TestRuntime2(t *testing.T) {
	println("TestRuntime2:")
	ast, errs := kuuhaku_parser.Parse("SEARCH_MODE E{C D C = ``return \"1\"``} C{<[A-Za-z0-9]+>} D{<\\.>}")
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("test.Hello test2.Hello3", &res, false, false)

	if err != nil {
		println("Expected runtime errors length to be 0, got " + strconv.Itoa(len(errs)))
		println(err.Error())
		t.Fatal()
	}

	if strRes != "[[[test],[.],[Hello]]] [[[test2],[.],[Hello3]]]" {
		println("Expected the string to be \"[[[test],[.],[Hello]]] [[[[test2],[.],[Hello3]]]\", got \"" + strRes + "\"")
		t.Fatal()
	}
}

func TestRuntime3(t *testing.T) {
	println("TestRuntime3:")
	ast, errs := kuuhaku_parser.Parse("E{l nl} E{E l nl} l{<test>} nl{<hello>}")
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("testhellotesthello", &res, false, false)

	if err != nil {
		println("Expected runtime errors length to be 0")
		println(err.Error())
		t.Fatal()
	}

	if strRes != "[[[[test],[hello]],[test],[hello]]]" {
		println("Expected the string to be \"[[[[test],[hello]],[test],[hello]]]\", got \"" + strRes + "\"")
		t.Fatal()
	}
}

func TestRuntime4(t *testing.T) {
	println("TestRuntime4:")
	ast, errs := kuuhaku_parser.Parse("E{E PLUS B} E{E MUL B} E{B} B{<0>} B{<1>} PLUS{<\\+>} MUL{<\\*>}")
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("0+1*0+1", &res, false, false)

	if err != nil {
		println("Expected runtime errors length to be 0")
		println(err.Error())
		t.Fatal()
	}

	if strRes != "[[[[[[0]],[+],[1]],[*],[0]],[+],[1]]]" {
		println("Expected the string to be \"[[[[[[0]],[+],[1]],[*],[0]],[+],[1]]]\", got \"" + strRes + "\"")
		t.Fatal()
	}
}

func TestRun1(t *testing.T) {
	println("TestRun1:")
	ast, errs := kuuhaku_parser.Parse(
		"E{E PLUS B = `E1 + B1`}" +
		"E{E MUL B = `E1 * B1`}" +
		"E{B = `B1`}" +
		"B{<[0-9]+> = `tonumber(LITERAL1)`}" + 
		"PLUS{<\\+>} MUL{<\\*>}",
	)
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("20+1*1+1", &res, true, false)

	if err != nil {
		println("Expected runtime errors length to be 0")
		println(err.Error())
		t.Fatal()
	}
	
	if strRes != "22" {
		println("Expected the result to be 22, got " + strRes)
		t.Fatal()
	}
}

func TestRun2(t *testing.T) {
	println("TestRun2:")
	ast, errs := kuuhaku_parser.Parse(
		"E{E PLUS B(`3`) = `E1 + B1`}" +
		"E{E MUL B(`2`) = `E1 * B1`}" +
		"E{B(`1`) = `B1`}" +
		"B(offset){<[0-9]+> = `tonumber(LITERAL1) + offset`}" + 
		"PLUS{<\\+>} MUL{<\\*>}",
	)
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("20+1*1+1", &res, true, false)

	if err != nil {
		println("Expected runtime errors length to be 0")
		println(err.Error())
		t.Fatal()
	}
	
	if strRes != "79" {
		println("Expected the result to be 79, got " + strRes)
		t.Fatal()
	}
}

func TestRunEscapes(t *testing.T) {
	println("TestRunEscapes:")
	ast, errs := kuuhaku_parser.Parse(
		"nl{<\\n>}"+
		"test{<test>}"+
		"E{test nl}",
	)
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("test\n", &res, true, false)

	if err != nil {
		println("Expected runtime errors length to be 0")
		println(err.Error())
		t.Fatal()
	}
	
	if strRes != "test\n" {
		println("Expected the result to be test\\n, got " + strRes)
		t.Fatal()
	}
}

func TestRunWeirdRegex(t *testing.T) {
	println("TestRunWeirdRegex:")
	ast, errs := kuuhaku_parser.Parse(
		"test{<test>}"+
		"w{<[ \\n\\t\\r]*>}"+
		"E{w E w test}" +
		"E{test}",
	)
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	_, err := Format("test test", &res, true, false)

	if err == nil {
		println("Expected a runtime error")
	} else {
		println(err.Error())
	}
}

func TestRunStartOfString(t *testing.T) {
	println("TestRunStartOfString1:")
	ast, errs := kuuhaku_parser.Parse(
		"test{C D C}"+
		"C{<A|B>}"+
		"D{<\\.>}",
	)
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	_, err := Format("A. B", &res, true, false)

	if err == nil {
		println("Expected a runtime error")
	} else {
		println(err.Error())
	}
}

func TestRunEOF(t *testing.T) {
	println("TestRunEOF:")
	ast, errs := kuuhaku_parser.Parse(
		"test{C D C}"+
		"C{<A|B>}"+
		"D{<\\.>}",
	)
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	_, err := Format("A.B", &res, true, false)

	if err != nil {
		println("Unexpected runtime error:")
		println(err.Error())
		t.Fatal()
	}
}

func TestRunDoubleRecursive(t *testing.T) {
	println("TestRunDoubleRecursive:")
	ast, errs := kuuhaku_parser.Parse(
		"Es {Es E}"+
		"Es {E}"+
		"E {open A close}"+
		"A {A AE}" +
		"A {AE}" + 
		"AE {<e>}" +
		"open {<{>}" +
		"close {<}>}",
	)
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("{eee}{eee}{eeeeee}", &res, true, false)

	if err != nil {
		println("Unexpected runtime error:")
		println(err.Error())
		t.Fatal()
	}
	
	if strRes != "{eee}{eee}{eeeeee}" {
		println("Expected the result to be {eee}{eee}{eeeeee}, got " + strRes)
		t.Fatal()
	}
}


func TestRunKhk(t *testing.T) {
	println("TestRunKhk:")
	ast, errs := kuuhaku_parser.Parse(khk.KHK)
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format(khk.TEST, &res, true, false)

	if err != nil {
		println("Unexpected runtime error:")
		println(err.Error())
		t.Fatal()
	}
	
	if strRes != khk.CORRECT {
		dmp := diffmatchpatch.New()
		fmt.Println("The resulting string is not as expected:")
		diffs := dmp.DiffMain(strRes, khk.CORRECT, false)
		fmt.Println(dmp.DiffPrettyText(diffs))
		t.Fatal()
	}
}

func TestRunArray(t *testing.T) {
	println("TestRunArray:")
	ast, errs := kuuhaku_parser.Parse(khk_array.ARRAY)
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format(khk_array.TEST, &res, true, false)

	if err != nil {
		println("Unexpected runtime error:")
		println(err.Error())
		t.Fatal()
	}
	
	if strRes != khk_array.CORRECT {
		dmp := diffmatchpatch.New()
		fmt.Println("The resulting string is not as expected:")
		diffs := dmp.DiffMain(strRes, khk_array.CORRECT, false)
		fmt.Println(dmp.DiffPrettyText(diffs))
		t.Fatal()
	}
}
