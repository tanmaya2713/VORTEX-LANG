package common

import "fmt"

type Position struct {
	Line   int
	Column int
	File   string
}

func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}

type CompilerError struct {
	Pos     Position
	Message string
	Phase   string
}

func (e *CompilerError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Phase, e.Pos, e.Message)
}

func NewError(phase string, pos Position, msg string) *CompilerError {
	return &CompilerError{Phase: phase, Pos: pos, Message: msg}
}

type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

type Diagnostic struct {
	Pos      Position
	Message  string
	Severity Severity
}

type TokenKind int

const (
	TokenEOF TokenKind = iota
	TokenIdent
	TokenNumber
	TokenString
	TokenKeyword
	TokenSymbol
	TokenComment
)

type Token struct {
	Kind    TokenKind
	Lexeme  string
	Pos     Position
	Literal interface{}
}

func (t Token) String() string {
	return fmt.Sprintf("Token{%s %q @ %s}", t.Kind, t.Lexeme, t.Pos)
}

type Symbol int

const (
	Semi Symbol = iota
	Colon
	Comma
	Dot
	Arrow
	FatArrow
	LParen
	RParen
	LBrace
	RBrace
	LBracket
	RBracket
	Equals
	EqEq
	Bang
	BangEq
	Less
	LessEq
	Greater
	GreaterEq
	Plus
	Minus
	Star
	Slash
	Percent
	Ampersand
	Pipe
	Caret
	Tilde
	At
	Hash
)

func (s Symbol) String() string {
	names := map[Symbol]string{
		Semi: ";", Colon: ":", Comma: ",", Dot: ".", Arrow: "->", FatArrow: "=>",
		LParen: "(", RParen: ")", LBrace: "{", RBrace: "}", LBracket: "[", RBracket: "]",
		Equals: "=", EqEq: "==", Bang: "!", BangEq: "!=",
		Less: "<", LessEq: "<=", Greater: ">", GreaterEq: ">=",
		Plus: "+", Minus: "-", Star: "*", Slash: "/", Percent: "%",
		Ampersand: "&", Pipe: "|", Caret: "^", Tilde: "~", At: "@", Hash: "#",
	}
	if n, ok := names[s]; ok {
		return n
	}
	return fmt.Sprintf("Symbol(%d)", s)
}

func (k TokenKind) String() string {
	names := map[TokenKind]string{
		TokenEOF: "EOF", TokenIdent: "Ident", TokenNumber: "Number",
		TokenString: "String", TokenKeyword: "Keyword", TokenSymbol: "Symbol",
		TokenComment: "Comment",
	}
	if n, ok := names[k]; ok {
		return n
	}
	return fmt.Sprintf("TokenKind(%d)", k)
}

type DataType int

const (
	TypeVoid DataType = iota
	TypeI8
	TypeI16
	TypeI32
	TypeI64
	TypeU8
	TypeU16
	TypeU32
	TypeU64
	TypeF32
	TypeF64
	TypeBool
	TypeString
	TypeTensor
	TypeModel
	TypeLayer
	TypeFn
	TypeStruct
	TypeArray
	TypePtr
	TypeRef
	TypeError
)

func (d DataType) String() string {
	names := map[DataType]string{
		TypeVoid: "void", TypeI8: "i8", TypeI16: "i16", TypeI32: "i32", TypeI64: "i64",
		TypeU8: "u8", TypeU16: "u16", TypeU32: "u32", TypeU64: "u64",
		TypeF32: "f32", TypeF64: "f64", TypeBool: "bool", TypeString: "string",
		TypeTensor: "tensor", TypeModel: "model", TypeLayer: "layer",
		TypeFn: "fn", TypeStruct: "struct", TypeArray: "array",
		TypePtr: "ptr", TypeRef: "ref", TypeError: "error",
	}
	if n, ok := names[d]; ok {
		return n
	}
	return fmt.Sprintf("DataType(%d)", d)
}
