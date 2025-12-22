package kuuhaku_runtime

import (
	"fmt"
	"strconv"

	"github.com/ciii1/kuuhaku/pkg/kuuhaku_analyzer"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
	lua "github.com/yuin/gopher-lua"
)

type ParseStackElementType int

const (
	PARSE_STACK_ELEMENT_TYPE_TREE ParseStackElementType = iota
	PARSE_STACK_ELEMENT_TYPE_TERMINAL
)

type ParseStackElement interface {
	GetType() ParseStackElementType
	GetString() string
	GetState() int
}

type ParseStackTree struct {
	Children *[]ParseStackElement
	Rule *kuuhaku_parser.Rule
	State  int
}

func (_ *ParseStackTree) GetType() ParseStackElementType {
	return PARSE_STACK_ELEMENT_TYPE_TREE;
}

func (p *ParseStackTree) GetState() int {
	return p.State;
}

func (p *ParseStackTree) GetString() string {
	out := "["
	for i, child := range *p.Children {
		if i != 0 {
			out += ","
		}
		out += child.GetString()
	}
	out += "]"
	return out
}

type ParseStackTerminal struct {
	String string
	State  int
}

func (_ *ParseStackTerminal) GetType() ParseStackElementType {
	return PARSE_STACK_ELEMENT_TYPE_TERMINAL;
}

func (p *ParseStackTerminal) GetString() string {
	return p.String
}

func (p *ParseStackTerminal) GetState() int {
	return p.State;
}

type RuntimeErrorType int

const (
	PARSE_STACK_IS_NOT_EMPTY RuntimeErrorType = iota
	REDUCE_RULE_IS_NOT_MATCHING
)

type EvalErrorType int

const (
	START_SYMBOL_WITH_PARAMS EvalErrorType = iota
	INVALID_ARG_LENGTH
	EXEC_ERROR
)

type RuntimeError struct {
	Position kuuhaku_tokenizer.Position
	Message  string
	Type     RuntimeErrorType
}

func (e RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

type EvalError struct {
	Message  string
	Type     EvalErrorType
}

func (e EvalError) Error() string {
	return fmt.Sprintf("Eval error: %s", e.Message)
}

type RuntimeSyntaxError struct {
	Position kuuhaku_tokenizer.Position
	Message  string
	Expected *[]string
}

func (e RuntimeSyntaxError) Error() string {
	return fmt.Sprintf("Syntax formatting error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

func ErrSyntaxError(position kuuhaku_tokenizer.Position, expected *[]string) *RuntimeSyntaxError {
	expectedCombined := ""
	for _, expectedE := range *expected {
		expectedCombined += "\n\t" + expectedE
	}

	return &RuntimeSyntaxError{
		Message:  "Syntax is invalid. Expected one of the following:" + expectedCombined,
		Expected: expected,
		Position: position,
	}
}

func ErrExpectedEOFError(position kuuhaku_tokenizer.Position) *RuntimeSyntaxError {
	return &RuntimeSyntaxError{
		Message:  "Syntax is invalid. Expected EOF.",
		Position: position,
		Expected: &[]string{"<end>"}, //TODO: make a struct for the expected[]
	}
}

func ErrParseStackIsNotEmpty(position kuuhaku_tokenizer.Position) *RuntimeError {
	return &RuntimeError{
		Message:  "Parse stack is not empty at the end of parsing.",
		Position: position,
		Type:     PARSE_STACK_IS_NOT_EMPTY,
	}
}

func ErrReduceRuleIsNotMatching(position kuuhaku_tokenizer.Position) *RuntimeError {
	return &RuntimeError{
		Message:  "The reduce rule on the parse table doesn't match the rule on the parse stack",
		Position: position,
		Type:     REDUCE_RULE_IS_NOT_MATCHING,
	}
}

func ErrLua(luaError string) *EvalError {
	return &EvalError{
		Message:  "Encountered an error while executing Lua chunk:\n\t" + luaError,
		Type:     EXEC_ERROR,
	}
}

func ErrInvalidArgLength(callee string, caller string) *EvalError {
	return &EvalError{
		Message:  "The argument length passed is not matching the parameter's length when calling rule " + callee + " in rule " + caller,
		Type:     INVALID_ARG_LENGTH,
	}
}

func ErrStartSymbolWithParams(startSymbol string) *EvalError {
	return &EvalError{
		Message:  "Start symbol " + startSymbol + " cannot have parameters.",
		Type:     INVALID_ARG_LENGTH,
	}
}

func Format(input string, format *kuuhaku_analyzer.AnalyzerResult, isRun bool, isDebug bool) (string, error) {
	var currPos kuuhaku_tokenizer.Position
	currPos.Line = 1
	currPos.Column = 1
	out := ""
	for currPos.Raw < len(input){
		isThereSuccess := false
		//TODO: change this to only one parse table
		for _, parseTable := range format.ParseTables {
			var globalLua kuuhaku_parser.LuaLiteral
			if format.GlobalLua != nil {
				globalLua = *format.GlobalLua
			}
			res, resPos, err := runParseTable(input, currPos, &parseTable, isRun, globalLua, isDebug)
			if err == nil {
				isThereSuccess = true
				currPos = resPos
				out += res
				break
			} else {
				if !format.IsSearchMode {
					return "", err
				}
			}
		}
		if !isThereSuccess {
			out += string(input[currPos.Raw])
			currPos.Raw++
		}
		if !format.IsSearchMode && currPos.Raw < len(input)-1 {
			return out, ErrExpectedEOFError(currPos)
		}
		if !format.IsSearchMode && currPos.Raw == len(input)-1 {
			break
		}
	}
	return out, nil
}

func addToPositionFromSlicedString(prevPos kuuhaku_tokenizer.Position, sliced string) kuuhaku_tokenizer.Position {
	raw := prevPos.Raw + len(sliced)

	col := prevPos.Column
	colIfContainsNewLine := 1

	i := len(sliced) - 1
	for i >= 0 {
		if sliced[i] == '\n' {
			break
		}
		col++
		colIfContainsNewLine++
		i--
	}

	//i is always -1 if a \n wasn't found. This is because the condition i >= 0 will not be satisfied untill
	//i == -1. sliced[i] != '\n' will only be satisfied if i > -1. So it is safe to check if i > -1 to
	//check for newlines

	if i >= 0 {
		col = colIfContainsNewLine
	}

	line := prevPos.Line
	for _, char := range sliced {
		if char == '\n' {
			line++
		}
	}

	return kuuhaku_tokenizer.Position{
		Raw:    raw,
		Column: col,
		Line:   line,
	}
}

func runParseTable(input string, pos kuuhaku_tokenizer.Position, parseTable *kuuhaku_analyzer.ParseTable, isRun bool, globalLua kuuhaku_parser.LuaLiteral, printCompiled bool) (string, kuuhaku_tokenizer.Position, error) {
	if printCompiled {
		fmt.Println("Input length: " + strconv.Itoa(len(input)))
	}
	var parseStack []ParseStackElement
	currState := 0
	lookahead := ""
	lookaheadRegex := ""

	var expected []string
	for true {
		lookaheadFound := false
		currRow := parseTable.States[currState]

		if pos.Raw > len(input) {
			for _, terminal := range parseTable.Terminals {
				if currRow.ActionTable[terminal.Terminal] != nil && terminal.Regexp != nil {
					expected = append(expected, terminal.Terminal)
				}
			}
			//TODO: might return all of the strings inside the parse stack combined on error in the future
			return "", pos, ErrSyntaxError(pos, &expected)
		}


		slicedInput := input[pos.Raw:]
		tmpPos := pos

		expected = []string{}
		if printCompiled {
			fmt.Println("[")
			for _, terminal := range parseTable.Terminals {
				fmt.Println("\t" + terminal.Terminal + " === " + strconv.Itoa(terminal.Precedence))
			}
			fmt.Println("]")
		}
		for _, terminal := range parseTable.Terminals {
			if currRow.ActionTable[terminal.Terminal] != nil && terminal.Regexp != nil {
				expected = append(expected, terminal.Terminal)
				loc := terminal.Regexp.FindStringIndex(slicedInput)
				if loc == nil || loc[0] != 0 {
					continue
				} else {
					lookahead = slicedInput[0:loc[1]]
					lookaheadRegex = terminal.Terminal
					tmpPos = addToPositionFromSlicedString(pos, lookahead)
					lookaheadFound = true
					break
				}
			}
		}
		slicedInput = input[pos.Raw:]
		if printCompiled {
			fmt.Println("Position: " + strconv.Itoa(pos.Raw))
			slicedInputTo3 := ""
			if len(slicedInput) > 4 {
				slicedInputTo3 = slicedInput[:3]
			} else {
				slicedInputTo3 = slicedInput
			}
			fmt.Println("Character[Position:Position+3]: "  + string(slicedInputTo3))
			fmt.Println("Lookahead found: " + strconv.FormatBool(lookaheadFound))
			fmt.Println("Lookahead: " + lookahead)
			fmt.Println("LookaheadRegex: " + lookaheadRegex)
		}
		if (lookaheadFound && pos.Raw < len(input)) || (lookaheadFound && pos.Raw >= len(input) && currRow.EndReduceRule == nil) {
			currActionCell := currRow.ActionTable[lookaheadRegex]
			if currActionCell != nil {
				if currActionCell.Action == kuuhaku_analyzer.SHIFT {
					if printCompiled {
						fmt.Println("Shifted: " + lookahead + " with the regex " + lookaheadRegex)
						fmt.Println("Shifting to state " + strconv.Itoa(currActionCell.ShiftState))
					}
					content := strconv.Quote(lookahead)
					content = content[1:len(content)-1]
					parseStack = append(parseStack, &ParseStackTerminal {
						String: content,
						State:  currState,
					})
					currState = currActionCell.ShiftState
					pos = tmpPos
				} else if currActionCell.Action == kuuhaku_analyzer.REDUCE {
					var err error
					currState, err = applyRule(parseTable, currActionCell.ReduceRule, &parseStack, pos, false)
					if printCompiled {
						fmt.Println("Reducing rule " + strconv.Itoa(currActionCell.ReduceRule.Order) + " with lhs: " + currActionCell.ReduceRule.Name)
						fmt.Println("New state: " + strconv.Itoa(currState))
					}
					if err != nil {
						return "", pos, err
					}
				}
			} else {
				return "", pos, ErrSyntaxError(pos, &expected)
			}
		} else {
			if currRow.EndReduceRule != nil {
				var err error
				if currRow.EndReduceRule.Action == kuuhaku_analyzer.ACCEPT {
					currState, err = applyRule(parseTable, currRow.EndReduceRule.ReduceRule, &parseStack, pos, true)
					break
				} else if currRow.EndReduceRule.Action == kuuhaku_analyzer.REDUCE {
					currState, err = applyRule(parseTable, currRow.EndReduceRule.ReduceRule, &parseStack, pos, false)
					if printCompiled {
						fmt.Println("End reducing rule " + strconv.Itoa(currRow.EndReduceRule.ReduceRule.Order) + " with lhs: " + currRow.EndReduceRule.ReduceRule.Name)
						fmt.Println("New state: " + strconv.Itoa(currState))
					}
				}
				if err != nil {
					return "", pos, err
				}
			} else {
				//printParseStack(&parseStack)
				return "", pos, ErrSyntaxError(pos, &expected)
			}
		}
	}
	if len(parseStack) != 1 {
		return "", pos, ErrParseStackIsNotEmpty(pos)
	}
	out := ""
	if isRun {
		var err error
		out, err = runParseStack(&parseStack, globalLua, printCompiled)
		if err != nil {
			return "", pos, err
		}
	} else {
		out = parseStackToString(&parseStack)
	}
	return out, pos, nil
}

func printParseStack(parseStack *[]ParseStackElement) {
	fmt.Print("Parse stack: ")
	for i, parseStackElement := range *parseStack {
		if i != 0 {
			fmt.Print(",")
		}
		fmt.Print(parseStackElement.GetString())
	}
	fmt.Println("")
}

func parseStackToString(parseStack *[]ParseStackElement) string {
	out := ""
	for i, parseStackElement := range *parseStack {
		if i != 0 {
			out += ","
		}
		out += parseStackElement.GetString()
	}
	return out
}

func runParseStack(parseStack *[]ParseStackElement, globalLua kuuhaku_parser.LuaLiteral, printCompiled bool) (string, error) {
	compiled := globalLua.LuaString + "\nret = tostring("
	compiledNodes, err := compileNode(&(*parseStack)[0], true, "") 
	compiled += compiledNodes
	compiled += ")"
	if printCompiled {
		fmt.Println("Compiled Lua code: " + compiled)
	}
	if err != nil {
		return "", err
	}
	L := lua.NewState()
	defer L.Close()
	err = L.DoString(compiled)
	if err != nil {
		fmt.Println("Error executing Lua code:", err)
		return "", ErrLua(err.Error())
	}
	ret := L.GetGlobal("ret").String()
	return ret, nil
}

func compileNode(node *ParseStackElement, isFirst bool, passedArgs string) (string, error) {
	out := ""
	if (*node).GetType() == PARSE_STACK_ELEMENT_TYPE_TERMINAL {
		terminal, _ := (*node).(*ParseStackTerminal)
		out += "\"" + terminal.String + "\""
	} else if (*node).GetType() == PARSE_STACK_ELEMENT_TYPE_TREE {
		tree, _ := (*node).(*ParseStackTree)
		out += "(function(\n"
		for i, params := range tree.Rule.ArgList {
			if i != 0 {
				out += ","
			}	
			out += params.Name 
		}

		out += ")\n"
		
		//we put the parameters that will be passed to the match rule functions here

		identifierCounts := make(map[string]int)
		var allVar []string
		for i, child := range *tree.Children {
			identifier, ok := tree.Rule.MatchRules[i].(kuuhaku_parser.Identifier)
			if ok {
				identifierCounts[identifier.Name] += 1
				varName := identifier.Name + strconv.Itoa(identifierCounts[identifier.Name])
				allVar = append(allVar, varName)
				var passingArgs string
				childTree, ok2 := child.(*ParseStackTree)
				if ok2 {
					if len(childTree.Rule.ArgList) != len(identifier.ArgList) {
						if isFirst {
							return "", ErrStartSymbolWithParams(childTree.Rule.Name)
						} else {
							return "", ErrInvalidArgLength(childTree.Rule.Name, tree.Rule.Name)
						}
					}
					for j, arg := range identifier.ArgList {
						if j > 0 {
							passingArgs += ",\n"
						}
						passingArgs += "(function()\n" + arg.LuaString + "\nend)()"
					}
				}
				compiledNode, err := compileNode(&child, false, passingArgs)
				if err != nil {
					return "", err
				}
				out += "local " + varName + " = " + compiledNode
			} else {
				varName := "LITERAL" + strconv.Itoa(i+1)
				allVar = append(allVar, varName)
				compiledNode, err := compileNode(&child, false, "")
				if err != nil {
					return "", err
				}
				out += "local " + varName + " = " + compiledNode
			}
		}

		out += "\n"
		if tree.Rule.ReplaceRule != nil {
			out += tree.Rule.ReplaceRule.LuaString
		} else {
			out += "return "
			for i, varName := range allVar {
				if i != 0 {
					out += ".."	
				}
				out += varName
			}
		}
		out += "\nend)(\n"
		out += passedArgs
		out += ")"
	}
	return out, nil
}

func copyParseStack(parseStack []ParseStackElement) *[]ParseStackElement {
	var newParseStack []ParseStackElement
	for _, e := range parseStack {
		if e.GetType() == PARSE_STACK_ELEMENT_TYPE_TREE {
			newParseStack = append(newParseStack, copyParseStackTreeRecursive(&e))
		} else if  e.GetType() == PARSE_STACK_ELEMENT_TYPE_TERMINAL {
			newParseStack = append(newParseStack, copyParseStackTerminal(&e))
		}
	}
	return &newParseStack
}

func copyParseStackTreeRecursive(e *ParseStackElement) *ParseStackTree {
	parseStackTree, ok := (*e).(*ParseStackTree) 
	var children []ParseStackElement
	if ok {
		for _, child := range *parseStackTree.Children {
			if child.GetType() == PARSE_STACK_ELEMENT_TYPE_TREE {
				children = append(children, copyParseStackTreeRecursive(&child))
			} else if  child.GetType() == PARSE_STACK_ELEMENT_TYPE_TERMINAL {
				newChild := child
				children = append(children, newChild)
			}
		}
		return &ParseStackTree{
			Children: &children,
			Rule: parseStackTree.Rule,
			State: parseStackTree.State,
		}
	}
	return nil
}

func copyParseStackTerminal(e *ParseStackElement) *ParseStackTerminal {
	parseStackTerminal, ok := (*e).(*ParseStackTerminal)
	if ok {
		newTerminal := parseStackTerminal
		return newTerminal
	}
	return nil
}

func applyRule(parseTable *kuuhaku_analyzer.ParseTable, rule *kuuhaku_parser.Rule, parseStack *[]ParseStackElement, pos kuuhaku_tokenizer.Position, isAccept bool) (int, error) {
	nextState := 0
	lhs := rule.Name
	ruleLength := len(rule.MatchRules)

	if len(*parseStack)-ruleLength < 0 {
		return 0, ErrReduceRuleIsNotMatching(pos)
	}
	targetStack := copyParseStack((*parseStack)[len(*parseStack)-ruleLength:])
	*parseStack = (*parseStack)[:len(*parseStack)-ruleLength]	

	if !isAccept {
		backState := 0
		if len(*parseStack)-1 >= 0 {
			backState = (*parseStack)[len(*parseStack)-1].GetState()
		}
		nextState = parseTable.States[backState].GotoTable[lhs].GotoState
	} else {
		nextState = 0
	}

	*parseStack = append(*parseStack, &ParseStackTree{
		Children: targetStack,
		Rule: rule,
		State:  nextState,
	})

	return nextState, nil
}
