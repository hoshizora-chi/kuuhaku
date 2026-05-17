package kuuhaku_analyzer

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/dlclark/regexp2/v2/compat"
)

func TestErrorInvalidArgumentLength(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test2(`'test'`)}\ntest2( test ){identifier(`'hello'`)}")
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		t.Fatal()
	}

	analyzer := initAnalyzer(&ast, false)
	_ = analyzer.analyzeStart()

	println("TestErrorInvalidArgumentLength - Errors:")
	helper.DisplayAllErrors(analyzer.Errors)

	if len(analyzer.Errors) != 1 {
		println("Expected analyzer Errors length to be 1")
		t.Fatal()
	}

	var analyzeError *AnalyzeError
	if errors.As(analyzer.Errors[0], &analyzeError) {
		if analyzeError.Type != INVALID_ARG_LENGTH {
			println("Expected InvalidArgLength error")
			t.Fail()
		}
		if analyzeError.Position.Column != 15 || analyzeError.Position.Line != 2 {
			col := strconv.Itoa(analyzeError.Position.Column)
			line := strconv.Itoa(analyzeError.Position.Line)
			println("Expected InvalidArgLength error with column 15 and line 2, got (" + col + ", " + line + ")")
			t.Fail()
		}
		if analyzeError.Message[0] != '1' {
			println("Expected the message to start with a 1")
		}
	} else {
		println("Expected AnalyzeError")
		t.Fail()
	}
}

func TestErrorUndefinedVariable(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest2{identifier}\ntest34{test4}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}

	analyzer := initAnalyzer(&ast, false)
	_ = analyzer.analyzeStart()

	println("TestErrorUndefinedVariable - Errors:")
	helper.DisplayAllErrors(analyzer.Errors)

	if len(analyzer.Errors) != 2 {
		println("Expected analyzer Errors length to be 2")
		t.Fatal()
	}

	var analyzeError *AnalyzeError
	if errors.As(analyzer.Errors[0], &analyzeError) {
		if analyzeError.Type != UNDEFINED_VARIABLE {
			println("Expected UndefinedVariableError error")
			t.Fail()
		}
		if (analyzeError.Position.Column != 12 || analyzeError.Position.Line != 1) && (analyzeError.Position.Column != 8 || analyzeError.Position.Line != 3) {
			col := strconv.Itoa(analyzeError.Position.Column)
			line := strconv.Itoa(analyzeError.Position.Line)
			println("Expected UndefinedVariableError error with column 12 and line 1, got (" + col + ", " + line + ")")
			t.Fail()
		}
	} else {
		println("Expected AnalyzeError")
		t.Fail()
	}

	if errors.As(analyzer.Errors[1], &analyzeError) {
		if analyzeError.Type != UNDEFINED_VARIABLE {
			println("Expected UndefinedVariableError error")
			t.Fail()
		}
		if (analyzeError.Position.Column != 8 || analyzeError.Position.Line != 3) && (analyzeError.Position.Column != 12 || analyzeError.Position.Line != 1) {
			col := strconv.Itoa(analyzeError.Position.Column)
			line := strconv.Itoa(analyzeError.Position.Line)
			println("Expected UndefinedVariableError error with column 8 and line 3, got (" + col + ", " + line + ")")
			t.Fail()
		}
	} else {
		println("Expected AnalyzeError")
		t.Fail()
	}

}

func TestStartSymbols(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{<\\.>}\nidentifier{<\\.>}\ntest3{<\\.>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	startSymbols := analyzer.analyzeStart()
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(startSymbols) != 2 {
		println("Expected startSymbols length to be 2")
		t.Fatal()
	}

	if startSymbols[0] == "Sidentifier" {
		if startSymbols[1] != "Stest3" {
			println("Expected startSymbols[1] to be \"Stest3\", got" + startSymbols[1])
			t.Fail()
		}
	} else if startSymbols[1] == "Sidentifier" {
		if startSymbols[0] != "Stest3" {
			println("Expected startSymbols[1] to be \"Stest3\", got" + startSymbols[1])
			t.Fail()
		}
	} else {
		println("Expected startSymbols[0] or [1] to be \"identifier\"")
		t.Fail()
	}
}

func TestExpandSymbol(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{test3}\nidentifier{<\\.>}\ntest3{<\\.>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	rules := ast.Rules["identifier"]
	expandedSymbols := analyzer.expandSymbol(&rules, 0, &[]*Symbol{}, SymbolTitle{Type: EMPTY_TITLE}, true)
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(*expandedSymbols) != 4 {
		println("Expected expandedSymbols length to be 4, got " + strconv.Itoa(len(*expandedSymbols)))
		t.Fatal()
	}

	firstSymbol := (*(*expandedSymbols)[0])
	title1 := firstSymbol.Title
	if title1.Type != IDENTIFIER_TITLE {
		println("Expected expandedSymbols[0].Title to be an identifier")
		t.Fail()
	}
	if title1.String != "test" {
		println("Expected expandedSymbols[0].Title.String to be \"test\"")
		t.Fail()
	}
	if firstSymbol.Lookahead.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[0].Lookahead.Type to be EMPTY_TITLE")
		t.Fail()
	}

	if !reflect.DeepEqual(*firstSymbol.Rule, *ast.Rules["identifier"][0]) {
		println("The first symbol's rule is not matching")
	}

	if firstSymbol.Position != 0 {
		println("Expected expandedSymbols[0].Position to be 0")
	}

	secondSymbol := (*(*expandedSymbols)[1])
	title2 := secondSymbol.Title
	if title2.Type != IDENTIFIER_TITLE {
		println("Expected expandedSymbols[1].Title to be a regex literal")
		t.Fail()
	}
	if title2.String != "test3" {
		println("Expected expandedSymbols[1].Title.String to be \"test3\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*secondSymbol.Rule, *ast.Rules["test"][0]) {
		println("The second symbol's rule is not matching")
		t.Fail()
	}

	if secondSymbol.Position != 0 {
		println("Expected expandedSymbols[1].Position to be 0")
		t.Fail()
	}

	if secondSymbol.Lookahead.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[1].Lookahead.Type to be REGEX_LITERAL_TITLE")
		t.Fail()
	}

	if secondSymbol.Lookahead.String != "\\." {
		println("Expected expandedSymbols[1].Lookahead.String to be \"\\.\"")
		t.Fail()
	}

	thirdSymbol := (*(*expandedSymbols)[3])
	title3 := thirdSymbol.Title
	if title3.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[3].Title to be a regex literal")
		t.Fail()
	}

	if title3.String != "\\." {
		println("Expected expandedSymbols[3].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*thirdSymbol.Rule, *ast.Rules["identifier"][1]) {
		println("The third symbol's rule is not matching")
	}

	if thirdSymbol.Lookahead.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[3].Lookahead.Type to be EMPTY_TITLE")
		t.Fail()
	}

	if thirdSymbol.Position != 0 {
		println("Expected expandedSymbols[2].Position to be 0")
	}

	fourthSymbol := (*(*expandedSymbols)[2])
	title4 := thirdSymbol.Title
	if title4.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[2].Title to be a regex literal")
		t.Fail()
	}

	if title4.String != "\\." {
		println("Expected expandedSymbols[2].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*fourthSymbol.Rule, *ast.Rules["test3"][0]) {
		println("The fourth symbol's rule is not matching")
		t.Fail()
	}

	if fourthSymbol.Lookahead.String != "\\." {
		println("Expected expandedSymbols[2].Lookahead.String to be \"\\.\"")
		t.Fail()
	}
}

func TestExpandSymbol2(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{<\\.>test}\ntest{<\\.>}\nidentifier{<\\.>}\ntest3{<\\.>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	rules := ast.Rules["identifier"]
	expandedSymbols := analyzer.expandSymbol(&rules, 1, &[]*Symbol{}, SymbolTitle{Type: EMPTY_TITLE}, true)
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(*expandedSymbols) != 3 {
		println("Expected expandedSymbols length to be 3, got " + strconv.Itoa(len(*expandedSymbols)))
		t.Fatal()
	}

	firstSymbol := (*(*expandedSymbols)[0])
	title1 := firstSymbol.Title
	if title1.Type != IDENTIFIER_TITLE {
		println("Expected expandedSymbols[0].Title to be an identifier")
		t.Fail()
	}

	if title1.String != "test" {
		println("Expected expandedSymbols[0].Title.String to be \"test\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*firstSymbol.Rule, *ast.Rules["identifier"][0]) {
		println("The first symbol's rule is not matching")
	}

	if firstSymbol.Position != 1 {
		println("Expected expandedSymbols[0].Position to be 1")
	}

	secondSymbol := (*(*expandedSymbols)[1])
	title2 := secondSymbol.Title
	if title2.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[1].Title to be a regex literal")
		t.Fail()
	}
	if title2.String != "\\." {
		println("Expected expandedSymbols[1].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*secondSymbol.Rule, *ast.Rules["test"][0]) {
		println("The second symbol's rule is not matching")
		t.Fail()
	}

	if secondSymbol.Position != 0 {
		println("Expected expandedSymbols[1].Position to be 0")
		t.Fail()
	}

	if secondSymbol.Lookahead.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[1].Lookahead.Type to be EMPTY TITLE")
		t.Fail()
	}

	thirdSymbol := (*(*expandedSymbols)[2])
	title3 := thirdSymbol.Title
	if title3.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[2].Title to be an EMPTY_TITLE")
		t.Fail()
	}

	if !reflect.DeepEqual(*thirdSymbol.Rule, *ast.Rules["identifier"][1]) {
		println("The third symbol's rule is not matching")
	}

	if thirdSymbol.Lookahead.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[2].Lookahead.Type to be EMPTY_TITLE")
		t.Fail()
	}

	if thirdSymbol.Position != 1 {
		println("Expected expandedSymbols[2].Position to be 0")
	}
}

func TestExpandSymbol3(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{<\\.>test}\ntest{<\\.>}\nidentifier{<\\.>}\ntest3{<\\.>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	rules := ast.Rules["test"]
	expandedSymbols := analyzer.expandSymbol(&rules, 2, &[]*Symbol{}, SymbolTitle{Type: EMPTY_TITLE}, true)
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(*expandedSymbols) != 1 {
		println("Expected expandedSymbols length to be 1, got " + strconv.Itoa(len(*expandedSymbols)))
		t.Fatal()
	}

	firstSymbol := (*(*expandedSymbols)[0])
	title1 := firstSymbol.Title
	if title1.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[0].Title to be empty")
		t.Fail()
	}
}

func TestGroupSymbols(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{<\\.>}\nidentifier{<\\.>}\ntest3{<\\.>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	rules := ast.Rules["identifier"]
	expandedSymbols := analyzer.expandSymbol(&rules, 0, &[]*Symbol{}, SymbolTitle{Type: EMPTY_TITLE}, true)
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(*expandedSymbols) != 3 {
		println("Expected expandedSymbols length to be 3, got " + strconv.Itoa(len(*expandedSymbols)))
		t.Fatal()
	}

	groupedSymbols := analyzer.groupSymbols(expandedSymbols)

	if len(*groupedSymbols) != 2 {
		println("Expected expandedSymbols length to be 2, got " + strconv.Itoa(len(*groupedSymbols)))
		t.Fatal()
	}

	comparedTitle1 := SymbolTitle{
		String: "\\.",
		Type:   REGEX_LITERAL_TITLE,
	}
	comparedTitle2 := SymbolTitle{
		String: "test",
		Type:   IDENTIFIER_TITLE,
	}
	var regexLitGroup *SymbolGroup
	var identifierGroup *SymbolGroup
	if (*groupedSymbols)[0].Title == comparedTitle1 {
		if (*groupedSymbols)[1].Title != comparedTitle2 {
			println("Expected groupedSymbols[1].Title to be \"test\" with the type identifier")
		}
		regexLitGroup = (*groupedSymbols)[0]
		identifierGroup = (*groupedSymbols)[1]
	} else if (*groupedSymbols)[1].Title == comparedTitle1 {
		if (*groupedSymbols)[0].Title != comparedTitle2 {
			println("Expected groupedSymbols[0] to be \"test\" with the type identifier")
		}
		regexLitGroup = (*groupedSymbols)[1]
		identifierGroup = (*groupedSymbols)[0]
	} else {
		println("Expected groupedSymbols[0] or [1] to contain the string \"\\.\" with the type regex literal")
		t.Fatal()
	}

	firstSymbol := *(*identifierGroup.Symbols)[0]
	title1 := firstSymbol.Title
	if title1.Type != IDENTIFIER_TITLE {
		println("Expected expandedSymbols[0].Title to be an identifier")
		t.Fail()
	}
	if title1.String != "test" {
		println("Expected expandedSymbols[0].Title.String to be \"test\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*firstSymbol.Rule, *ast.Rules["identifier"][0]) {
		println("The first symbol's rule is not matching")
	}

	if firstSymbol.Position != 0 {
		println("Expected expandedSymbols[0].Position to be 0")
	}

	secondSymbol := *(*regexLitGroup.Symbols)[0]
	title2 := secondSymbol.Title
	if title2.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[1].Title to be a regex literal")
		t.Fail()
	}
	if title2.String != "\\." {
		println("Expected expandedSymbols[1].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*secondSymbol.Rule, *ast.Rules["test"][0]) {
		println("The second symbol's rule is not matching")
	}

	if secondSymbol.Position != 0 {
		println("Expected expandedSymbols[1].Position to be 0")
	}

	thirdSymbol := *(*regexLitGroup.Symbols)[1]
	title3 := thirdSymbol.Title
	if title3.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[2].Title to be a regex literal")
		t.Fail()
	}

	if title3.String != "\\." {
		println("Expected expandedSymbols[2].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*thirdSymbol.Rule, *ast.Rules["identifier"][1]) {
		println("The third symbol's rule is not matching")
	}

	if thirdSymbol.Position != 0 {
		println("Expected expandedSymbols[2].Position to be 0")
	}
}

func TestBuildParseTableStateTransition(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{<hello>}\nidentifier{<\\.>}\ntest3{<\\.>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	analyzer.parseTables = append(analyzer.parseTables, analyzer.makeEmptyParseTable("identifier"))
	stateTransitions := analyzer.buildParseTable("identifier")

	if len(*stateTransitions) != 5 {
		println("Expected stateTransitions length to be 5, got " + strconv.Itoa(len(*stateTransitions)))
		t.Fatal()
	}

	if len(*(*stateTransitions)[0].SymbolGroups) != 3 {
		println("Expected the first state transition to contain exactly three groups")
		t.Fail()
	}

	titles := []SymbolTitle{
		{
			String: "test",
			Type:   IDENTIFIER_TITLE,
		},
		{
			String: "\\.",
			Type:   REGEX_LITERAL_TITLE,
		},
		{
			String: "hello",
			Type:   REGEX_LITERAL_TITLE,
		},
	}

	for _, title := range titles {
		isExist := false
		for _, group := range *(*stateTransitions)[0].SymbolGroups {
			if title == group.Title {
				isExist = true
			}
		}
		if !isExist {
			println("Expected the first state transition to contain groups with title \"test\", \"\\.\", and \"hello\"")
			t.Fail()
		}
	}

	if len(*(*stateTransitions)[1].SymbolGroups) != 1 {
		println("Expected the second state transition to contain exactly one group")
		t.Fail()
	}

	if len(*(*stateTransitions)[2].SymbolGroups) != 1 {
		println("Expected the third state transition to contain exactly one group")
		t.Fail()
	}

	if len(*(*stateTransitions)[3].SymbolGroups) != 1 {
		println("Expected the fourth state transition to contain exactly one group")
		t.Fail()
	}

	middleTransitions := []*StateTransition{
		(*stateTransitions)[1],
		(*stateTransitions)[2],
		(*stateTransitions)[3],
	}
	titles2 := []SymbolTitle{
		{
			String: "\\.",
			Type:   REGEX_LITERAL_TITLE,
		},
		{
			String: "<end>",
			Type:   EMPTY_TITLE,
		},
		{
			String: "<end>",
			Type:   EMPTY_TITLE,
		},
	}
	for _, title := range titles2 {
		isExist := false
		for _, transition := range middleTransitions {
			if title == (*(*transition).SymbolGroups)[0].Title {
				isExist = true
			}
		}
		if !isExist {
			println("Expected the second, third, and fourth state transition to contain groups with title {\"test\", 1}, and two empty titles")
			t.Fail()
		}
	}

	lastSymbol := SymbolTitle{
		String: "<end>",
		Type:   EMPTY_TITLE,
	}
	if (*(*stateTransitions)[4].SymbolGroups)[0].Title != lastSymbol {
		println("Expected the fifth state transition to contain group with an empty title")
		t.Fail()
	}

	if len(*(*stateTransitions)[4].SymbolGroups) != 1 {
		println("Expected the fifth state transition to contain exactly one group")
		t.Fail()
	}
}

func TestBuildParseTable(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{<\\.>}\nidentifier{<\\.>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	analyzer.parseTables = append(analyzer.parseTables, analyzer.makeEmptyParseTable("identifier"))
	stateTransitions := analyzer.buildParseTable("identifier")

	if len(*stateTransitions) != 4 {
		println("Expected stateTransitions length to be 4, got " + strconv.Itoa(len(*stateTransitions)))
		t.Fatal()
	}

	if len(analyzer.Errors) != 0 {
		println("Expected analyzer errors to be 0")
		t.Fatal()
	}

	if len(analyzer.parseTables[0].States) != 4 {
		println("Expected parse table states length to be 4, got " + strconv.Itoa(len(analyzer.parseTables[0].States)))
		t.Fatal()
	}

	firstRow := analyzer.parseTables[0].States[0]
	if firstRow.ActionTable["\\."].Action != SHIFT {
		println("Expected the first state row to have SHIFT on column \"\\.\"")
		t.Fail()
	}

	secondRow := analyzer.parseTables[0].States[firstRow.ActionTable["\\."].ShiftState]
	if secondRow.ActionTable["\\."].Action != REDUCE {
		println("Expected the second state row to have REDUCE on column \"\\.\"")
		t.Fail()
	}
	if secondRow.ActionTable["\\."].ReduceRule != ast.Rules["test"][0] {
		println("Expected the second state row to have the reduce rule 2 on column \"\\.\"")
		t.Fail()
	}
	if secondRow.EndReduceRule.ReduceRule != ast.Rules["identifier"][1] {
		println("Expected the second state row to have the end reduce rule 3 on column \"\\.\"")
		t.Fail()
	}
	if secondRow.EndReduceRule.Action != ACCEPT {
		println("Expected the second state row to have the ACCEPT action")
		t.Fail()
	}

	thirdRow := analyzer.parseTables[0].States[firstRow.GotoTable["test"].GotoState]
	if thirdRow.ActionTable["\\."].Action != SHIFT {
		println("Expected the second state row to have SHIFT on column \"\\.\"")
		t.Fail()
	}

	fourthRow := analyzer.parseTables[0].States[thirdRow.ActionTable["\\."].ShiftState]
	if fourthRow.EndReduceRule.ReduceRule != ast.Rules["identifier"][0] {
		println("Expected the second state row to have the end reduce rule 1 on column \"\\.\"")
		t.Fail()
	}
	if fourthRow.EndReduceRule.Action != ACCEPT {
		println("Expected the second state row to have the ACCEPT action")
		t.Fail()
	}
}

func TestBuildParseTableErrorMultiplePartialReduce(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("E{B <1>} E{<1> B C} B{<1> <2>} B{<2>} C{<2>} C{<1>}")
	if len(errs) != 3 {
		println("Expected parser errors length to be 3")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	analyzer.parseTables = append(analyzer.parseTables, analyzer.makeEmptyParseTable("E"))
	analyzer.buildParseTable("E")

	println("TestBuildParseTableErrorMultiplePartialReduce - Errors:")
	helper.DisplayAllErrors(analyzer.Errors)

	if len(analyzer.Errors) != 1 {
		println("Expected analyzer.Error length to be 1, got " + strconv.Itoa(len(analyzer.Errors)))
		t.Fatal()
	}

	var conflictError *ConflictError
	if errors.As(analyzer.Errors[0], &conflictError) {
		if conflictError.Symbol1.Rule.Order == 3 {
			if conflictError.Symbol2.Rule.Order != 2 {
				println("Expected the rule order to be 2, 3 or 3, 2")
				t.Fail()
			}
		} else if conflictError.Symbol1.Rule.Order == 2 {
			if conflictError.Symbol2.Rule.Order != 3 {
				println("Expected the rule order to be 2, 3 or 3, 2")
				t.Fail()
			}
		} else {
			println("Expected the rule order to be 2, 3 or 3, 2")
			t.Fail()
		}

		if conflictError.Position1.Line == 1 && conflictError.Position1.Column == 27 {
			if conflictError.Position2.Line != 1 || conflictError.Position2.Column != 34 {
				println("Expected the rule position to be (1, 34) and (1, 27) or reversed")
				t.Fail()
			}
		} else if conflictError.Position1.Line == 1 && conflictError.Position1.Column == 34 {
			if conflictError.Position2.Line != 1 || conflictError.Position2.Column != 27 {
				println("Expected the rule position to be (1, 34) and (1, 27) or reversed")
				t.Fail()
			}
		} else {
			println("Expected the rule position to be (1, 34) and (1, 27) or reversed")
			t.Fail()
		}

		if conflictError.Symbol1.Lookahead.String == "C" {
			if conflictError.Symbol2.Lookahead.String != "1" {
				println("Expected the lookaheads to be (1, C) or reversed")
				t.Fail()
			}
		} else if conflictError.Symbol1.Lookahead.String == "1" {
			if conflictError.Symbol2.Lookahead.String != "C" {
				println("Expected the lookaheads to be (C, 1) or reversed")
				t.Fail()
			}
		} else {
			println("Expected the lookaheads to be (C, 1) or reversed")
			t.Fail()
		}
	} else {
		println("Expected a conflict error")
		t.Fail()
	}
}

func TestBuildParseTable2(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("E{E <\\*> B} E{E <\\+> B} E{B} B{<0>} B{<1>}")
	if len(errs) != 2 {
		println("Expected parser errors length to be 2")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	analyzer.parseTables = append(analyzer.parseTables, analyzer.makeEmptyParseTable("E"))
	analyzer.makeAugmentedGrammar("E")
	stateTransitions := analyzer.buildParseTable("SE")
	PrintParseTable(&analyzer.parseTables[0])

	if len(analyzer.Errors) != 0 {
		println("Expected analyzer.Error length to be 0, got " + strconv.Itoa(len(analyzer.Errors)))
		helper.DisplayAllErrors(analyzer.Errors)
		t.Fatal()
	}

	if len(analyzer.parseTables[0].States) != 9 {
		println("Expected stateTransitions length to be 9, got " + strconv.Itoa(len(*stateTransitions)))
		t.Fatal()
	}

	if analyzer.parseTables[0].States[5].ActionTable["0"].Action != SHIFT {
		println("Expected the fifth state row to have the shift on column \"0\"")
		t.Fail()
	}
	if analyzer.parseTables[0].States[5].ActionTable["0"].ShiftState != 3 && analyzer.parseTables[0].States[5].ActionTable["0"].ShiftState != 4 && analyzer.parseTables[0].States[5].ActionTable["0"].ShiftState != 2 && analyzer.parseTables[0].States[5].ActionTable["0"].ShiftState != 1 {
		println("Expected the fifth state row to have the shift 3 or 4 on column \"0\"")
		t.Fail()
	}
	if analyzer.parseTables[0].States[5].ActionTable["1"].Action != SHIFT {
		println("Expected the fifth state row to have the shift on column \"1\"")
		t.Fail()
	}
	if analyzer.parseTables[0].States[5].ActionTable["1"].ShiftState != 3 && analyzer.parseTables[0].States[5].ActionTable["1"].ShiftState != 4 && analyzer.parseTables[0].States[5].ActionTable["1"].ShiftState != 2 && analyzer.parseTables[0].States[5].ActionTable["1"].ShiftState != 1 {
		println("Expected the fifth state row to have the shift 3 or 4 on column \"1\"")
		t.Fail()
	}

	if analyzer.parseTables[0].States[6].ActionTable["0"].Action != SHIFT {
		println("Expected the sixth state row to have the shift on column \"0\"")
		t.Fail()
	}
	if analyzer.parseTables[0].States[6].ActionTable["0"].ShiftState != 3 && analyzer.parseTables[0].States[5].ActionTable["0"].ShiftState != 4 && analyzer.parseTables[0].States[5].ActionTable["0"].ShiftState != 2 && analyzer.parseTables[0].States[5].ActionTable["0"].ShiftState != 1 {
		println("Expected the sixth state row to have the shift 3 or 4 on column \"0\"")
		t.Fail()
	}
	if analyzer.parseTables[0].States[6].ActionTable["1"].Action != SHIFT {
		println("Expected the sixth state row to have the shift on column \"1\"")
		t.Fail()
	}
	if analyzer.parseTables[0].States[6].ActionTable["1"].ShiftState != 3 && analyzer.parseTables[0].States[5].ActionTable["1"].ShiftState != 4 && analyzer.parseTables[0].States[5].ActionTable["1"].ShiftState != 1 && analyzer.parseTables[0].States[5].ActionTable["1"].ShiftState != 2 {
		println("Expected the sixth state row to have the shift 3, 4, 2, 1 on column \"1\"")
		t.Fail()
	}
	for _, state := range analyzer.parseTables[0].States {
		if state.EndReduceRule != nil && state.EndReduceRule.Action == ACCEPT {
			if state.ActionTable["+"] != nil {
				if state.ActionTable["+"].ShiftState != 6 && state.ActionTable["+"].ShiftState != 5 {
					println("Expected ACCEPT only on a state where there is a 5 or a 6 on its ActionTable[\"+\"]")
					t.Fail()
				}
			}
		}
	}
}

func TestBuildParseTableErrorPartialReduceAndShift(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("E{B <1>} E{<1> B C} B{<2> <1>} B{<1>} C{<2>} C{<1>}")
	if len(errs) != 3 {
		println("Expected parser errors length to be 3")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	analyzer.parseTables = append(analyzer.parseTables, analyzer.makeEmptyParseTable("E"))
	analyzer.buildParseTable("E")

	println("TestBuildParseTableErrorPartialReduceAndShift - Errors:")
	helper.DisplayAllErrors(analyzer.Errors)

	if len(analyzer.Errors) != 1 {
		println("Expected analyzer.Error length to be 1, got " + strconv.Itoa(len(analyzer.Errors)))
		t.Fatal()
	}

	var conflictError *ConflictError
	if errors.As(analyzer.Errors[0], &conflictError) {
		if conflictError.Symbol1.Rule.Order == 2 {
			if conflictError.Symbol2.Rule.Order != 3 || conflictError.Symbol1.Rule.Order != 3 {
				println("Expected the rule order to be both 3")
				t.Fail()
			}
		}

		if conflictError.Position1.Line != 1 || conflictError.Position1.Column != 34 || conflictError.Position2.Line != 1 || conflictError.Position2.Column != 34 {
			println("Expected the rule position to be (1, 34) and (1, 34)")
			t.Fail()
		}

		if conflictError.Symbol1.Lookahead.String == "C" {
			if conflictError.Symbol2.Lookahead.String != "1" {
				println("Expected the lookaheads to be (1, C) or reversed")
				t.Fail()
			}
		} else if conflictError.Symbol1.Lookahead.String == "1" {
			if conflictError.Symbol2.Lookahead.String != "C" {
				println("Expected the lookaheads to be (C, 1) or reversed")
				t.Fail()
			}
		} else {
			println("Expected the lookaheads to be (C, 1) or reversed")
			t.Fail()
		}
	} else {
		println("Expected a conflict error")
		t.Fail()
	}
}

func TestBuildParseTableErrorMultipleEndReduce(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("E{C} E{B} B{<0>} B{<1>} C{<1>}")
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	analyzer.parseTables = append(analyzer.parseTables, analyzer.makeEmptyParseTable("E"))
	analyzer.buildParseTable("E")

	println("TestBuildParseTableErrorMultipleEndReduce - Errors:")
	helper.DisplayAllErrors(analyzer.Errors)

	if len(analyzer.Errors) != 1 {
		println("Expected analyzer.Error length to be 1, got " + strconv.Itoa(len(analyzer.Errors)))
		t.Fatal()
	}

	var conflictError *ConflictError
	if errors.As(analyzer.Errors[0], &conflictError) {
		if conflictError.Symbol1.Rule.Order == 3 {
			if conflictError.Symbol2.Rule.Order != 4 {
				println("Expected the rule order to be 4, 3 or 3, 4")
				t.Fail()
			}
		} else if conflictError.Symbol1.Rule.Order == 4 {
			if conflictError.Symbol2.Rule.Order != 3 {
				println("Expected the rule order to be 4, 3 or 3, 4")
				t.Fail()
			}
		} else {
			println("Expected the rule order to be 4, 3 or 3, 4")
			t.Fail()
		}

		if conflictError.Position1.Line == 1 && conflictError.Position1.Column == 27 {
			if conflictError.Position2.Line != 1 || conflictError.Position2.Column != 20 {
				println("Expected the rule position to be (1, 20) and (1, 27) or reversed")
				t.Fail()
			}
		} else if conflictError.Position1.Line == 1 && conflictError.Position1.Column == 20 {
			if conflictError.Position2.Line != 1 || conflictError.Position2.Column != 27 {
				println("Expected the rule position to be (1, 20) and (1, 27) or reversed")
				t.Fail()
			}
		} else {
			println("Expected the rule position to be (1, 20) and (1, 27) or reversed")
			t.Fail()
		}

		if conflictError.Symbol1.Lookahead.Type != EMPTY_TITLE || conflictError.Symbol2.Lookahead.Type != EMPTY_TITLE {
			println("Expected the lookaheads to be (<end>, <end>)")
			t.Fail()
		}
	} else {
		println("Expected a conflict error")
		t.Fail()
	}
}

func TestGetAllTerminalsAndLhs(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("E{C} E{B} B{<0>} B{<1>} C{<1>} D{<3> F} F{<1>} B{<5>} B{<10>} B{<9>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	terminalsMapInput := make(map[string]*TerminalList)
	var terminalsMap *map[string]*TerminalList
	lhsMapInput := make(map[string]bool)
	var lhsMap *map[string]bool
	terminalsMap, lhsMap = analyzer.getAllTerminalsAndLhs("E", &terminalsMapInput, &lhsMapInput)

	lhsMapCorrect := map[string]bool{
		"E": true,
		"C": true,
		"B": true,
	}

	regexCompiled1, _ := compat.Compile("0")
	regexCompiled2, _ := compat.Compile("1")
	regexCompiled3, _ := compat.Compile("5")
	regexCompiled4, _ := compat.Compile("10")
	regexCompiled5, _ := compat.Compile("9")
	terminalsMapCorrect := map[string]TerminalList{
		"0": {
			Terminal:   "0",
			Precedence: 2,
			Regexp:     regexCompiled1,
		},
		"1": {
			Terminal:   "1",
			Precedence: 3,
			Regexp:     regexCompiled2,
		},
		"5": {
			Terminal:   "5",
			Precedence: 7,
			Regexp:     regexCompiled3,
		},
		"10": {
			Terminal:   "10",
			Precedence: 8,
			Regexp:     regexCompiled4,
		},
		"9": {
			Terminal:   "9",
			Precedence: 9,
			Regexp:     regexCompiled5,
		},
	}

	if !reflect.DeepEqual(lhsMapCorrect, *lhsMap) {
		t.Fail()
	}

	for terminal, terminalList := range *terminalsMap {
		if terminalsMapCorrect[terminal].Terminal != terminalList.Terminal {
			println("terminalsMap[" + terminal + "].Terminal (" + terminalList.Terminal + ") != terminalsMapCorrect[" +
				terminal + "].Terminal (" + terminalsMapCorrect[terminal].Terminal + ")")
			t.Fail()
		}
		if terminalsMapCorrect[terminal].Precedence != terminalList.Precedence {
			println("terminalsMap[" + terminal + "].Precedence (" + strconv.Itoa(terminalList.Precedence) +
				") != terminalsMapCorrect[" + terminal + "].Precedence (" +
				strconv.Itoa(terminalsMapCorrect[terminal].Precedence) + ")")
			t.Fail()
		}
		if terminalsMapCorrect[terminal].Regexp.String() != terminalList.Regexp.String() {
			println("terminalsMap[" + terminal + "].Regexp.String != terminalsMapCorrect[" + terminal + "].Regexp")
			println(terminalsMapCorrect[terminal].Regexp.String() + " != " + terminalList.Regexp.String())
			t.Fail()
		}
	}

	terminalsCorrect := []TerminalList{
		{
			Terminal:   "0",
			Precedence: 2,
			Regexp:     regexCompiled1,
		},
		{
			Terminal:   "1",
			Precedence: 3,
			Regexp:     regexCompiled2,
		},
		{
			Terminal:   "5",
			Precedence: 7,
			Regexp:     regexCompiled3,
		},
		{
			Terminal:   "10",
			Precedence: 8,
			Regexp:     regexCompiled4,
		},
		{
			Terminal:   "9",
			Precedence: 9,
			Regexp:     regexCompiled5,
		},
	}

	terminals := sortTerminalsMaptoArray(terminalsMap)

	for i, terminalList := range *terminals {
		if terminalsCorrect[i].Terminal != terminalList.Terminal {
			println("terminals[" + strconv.Itoa(i) + "].Terminal (" +
				terminalList.Terminal + ") != terminalsMapCorrect[" + strconv.Itoa(i) + "].Terminal (" +
				terminalsCorrect[i].Terminal + ")")
			t.Fail()
		}
		if terminalsCorrect[i].Precedence != terminalList.Precedence {
			println("terminalsMap[" + strconv.Itoa(i) + "].Precedence (" + strconv.Itoa(terminalList.Precedence) +
				") != terminalsMapCorrect[" + strconv.Itoa(i) + "].Precedence (" +
				strconv.Itoa(terminalsCorrect[i].Precedence) + ")")
			t.Fail()
		}
		if terminalsCorrect[i].Regexp.String() != terminalList.Regexp.String() {
			println("terminalsMap[" + strconv.Itoa(i) + "].Regexp != terminalsMapCorrect[" + strconv.Itoa(i) + "].Regexp")
			println(terminalsCorrect[i].Regexp.String() + " != " + terminalList.Regexp.String())
			t.Fail()
		}
	}
}

func TestGetAllTerminalsAndLhsRegexError(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("E{C} E{B} B{<0>} B{<1>} C{<1>} D{<3> F} F{<1>} B{<5>} B{<10>} B{<[>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast, false)
	terminalsMapInput := make(map[string]*TerminalList)
	lhsMapInput := make(map[string]bool)
	_, _ = analyzer.getAllTerminalsAndLhs("E", &terminalsMapInput, &lhsMapInput)

	if len(analyzer.Errors) != 1 {
		println("Expected analyzer Errors length to be 1, got " + strconv.Itoa(len(analyzer.Errors)))
		t.Fatal()
	}

	println("TestGetAllTerminalsAndLhsRegexError - Errors:")
	helper.DisplayAllErrors(analyzer.Errors)

	var analyzeError *AnalyzeError
	if errors.As(analyzer.Errors[0], &analyzeError) {
		if analyzeError.Type != INVALID_REGEX {
			println("Expected invalid regex error")
			t.Fail()
		}
		if analyzeError.Position.Column != 65 || analyzeError.Position.Line != 1 {
			println("Expected position to be (1, 65), got (" +
				strconv.Itoa(analyzeError.Position.Line) + ", " +
				strconv.Itoa(analyzeError.Position.Column) + ")")
			t.Fail()
		}
	} else {
		println("Expected AnalyzeError")
		t.Fail()
	}

}

func TestAnalyze(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("E{C} E{B} B{<0>} B{<1>} C{<1>} D{<3> F} F{<1>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	_, errs = Analyze(&ast, false)
	if len(errs) != 1 {
		println("Expected analyzer Errors length to be 1, got " + strconv.Itoa(len(errs)))
		t.Fatal()
	}

	println("TestAnalyze - Errors:")
	helper.DisplayAllErrors(errs)

	var analyzeError *AnalyzeError
	if errors.As(errs[0], &analyzeError) {
		if analyzeError.Type != MULTIPLE_START_SYMBOLS {
			println("Expected multiple start symbols error")
			t.Fail()
		}
	} else {
		println("Expected AnalyzeError")
		t.Fail()
	}
}

func TestAnalyze2(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("SEARCH_MODE ``Hello`` E{C} E{B} B{<0>} B{<1>} C{<3>} D{<3> F} F{<1>} G{<2>}")
	if len(errs) != 1 {
		println("Expected parser errors length to be 1")
		t.Fatal()
	}
	res, errs := Analyze(&ast, false)
	if len(errs) != 0 {
		println("Expected analyzer Errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	if len(res.ParseTables) != 3 {
		println("Expected parse tables length to be 3, got " + strconv.Itoa(len(res.ParseTables)))
		t.Fatal()
	}

	if res.GlobalLua == nil {
		println("Expected GlobalLua != nil")
		t.Fatal()
	} else if res.GlobalLua.LuaString != "Hello" {
		println("Expected GlobalLua to contain \"Hello\", got " + res.GlobalLua.LuaString)
		t.Fatal()
	}
}
