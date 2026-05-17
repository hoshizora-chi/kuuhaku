package kuuhaku_analyzer

import (
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/dlclark/regexp2/v2/compat"
)

type StateTransition struct {
	SymbolGroups *[]*SymbolGroup
}

type SymbolGroup struct {
	Title   SymbolTitle
	Symbols *[]*Symbol
}

type Symbol struct {
	Position   int
	Title      SymbolTitle
	Rule       *kuuhaku_parser.Rule
	Lookahead  SymbolTitle
}

type SymbolTitleType int

const (
	REGEX_LITERAL_TITLE = iota
	IDENTIFIER_TITLE
	EMPTY_TITLE
)

type SymbolTitle struct {
	String string
	Type   SymbolTitleType
}

type AnalyzerResult struct {
	ParseTables  []ParseTable
	IsSearchMode bool
	GlobalLua 	 *kuuhaku_parser.LuaLiteral
}

type ParseTable struct {
	States    []ParseTableState
	Terminals []TerminalList
	Lhss      []string
}

type TerminalList struct {
	Terminal   string 
	Precedence int
	Regexp     *compat.Regexp
}

type ParseTableState struct {
	ActionTable map[string]*ActionCell //map[kuuhaku_parser.RegexLiteral.Content]ActionCell
	GotoTable   map[string]*GotoCell   //map[kuuhaku_parser.Rule.Name]GotoCell

	//How the parser would read the following field: test all terminals inside action table, if no match
	//then use EndReduceRule. If it's a nil, then return error
	EndReduceRule *ActionCell
}

type Action int

const (
	REDUCE = iota
	SHIFT
	ACCEPT
)

type ActionCell struct {
	LookaheadTerminal string
	Action            Action
	ReduceRule        *kuuhaku_parser.Rule
	ShiftState        int
}

type GotoCell struct {
	Lhs       string
	GotoState int
}
