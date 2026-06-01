package parser

import (
	"github.com/vortex-lang/vortex/src/ast"
	"github.com/vortex-lang/vortex/src/common"
)

type memoKey struct {
	rule string
	pos  int
}

type memoEntry struct {
	node  ast.Node
	pos   int
	found bool
}

type Parser struct {
	tokens []common.Token
	pos    int
	memo   map[memoKey]memoEntry
	errs   []*common.CompilerError
	file   string
}

func New(tokens []common.Token, file string) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		memo:   make(map[memoKey]memoEntry),
		file:   file,
	}
}

func (p *Parser) memoize(rule string, fn func() (ast.Node, bool)) (ast.Node, bool) {
	key := memoKey{rule, p.pos}
	if entry, ok := p.memo[key]; ok {
		p.pos = entry.pos
		return entry.node, entry.found
	}
	savePos := p.pos
	node, found := fn()
	p.memo[key] = memoEntry{node: node, pos: p.pos, found: found}
	if !found {
		p.pos = savePos
	}
	return node, found
}

func (p *Parser) current() common.Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return common.Token{Kind: common.TokenEOF, Pos: common.Position{Line: 0, Column: 0, File: p.file}}
}

func (p *Parser) advance() common.Token {
	tok := p.current()
	p.pos++
	return tok
}

func (p *Parser) expect(kind common.TokenKind, lexeme string) (common.Token, bool) {
	tok := p.current()
	if tok.Kind == kind && (lexeme == "" || tok.Lexeme == lexeme) {
		return p.advance(), true
	}
	return tok, false
}

func (p *Parser) expectKeyword(kw string) bool {
	tok := p.current()
	if tok.Kind == common.TokenKeyword && tok.Lexeme == kw {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expectSymbol(sym common.Symbol) bool {
	tok := p.current()
	if tok.Kind == common.TokenSymbol {
		if s, ok := tok.Literal.(common.Symbol); ok && s == sym {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) error(msg string) {
	tok := p.current()
	p.errs = append(p.errs, common.NewError("parse", tok.Pos, msg))
}

func (p *Parser) sync() {
	for p.pos < len(p.tokens) {
		tok := p.current()
		if tok.Kind == common.TokenEOF {
			return
		}
		if tok.Kind == common.TokenKeyword && (tok.Lexeme == "fn" || tok.Lexeme == "let" ||
			tok.Lexeme == "struct" || tok.Lexeme == "model" || tok.Lexeme == "train" ||
			tok.Lexeme == "for" || tok.Lexeme == "while" || tok.Lexeme == "if" ||
			tok.Lexeme == "return" || tok.Lexeme == "import" || tok.Lexeme == "print" ||
			tok.Lexeme == "assert") {
			return
		}
		p.advance()
	}
}

func (p *Parser) Parse() *ast.Program {
	prog := &ast.Program{}
	for p.pos < len(p.tokens) {
		if p.current().Kind == common.TokenEOF {
			break
		}
		if stmt, ok := p.parseStmt(); ok {
			prog.Stmts = append(prog.Stmts, stmt)
		} else {
			if len(p.errs) > 0 {
				break
			}
			p.sync()
		}
	}
	return prog
}

func (p *Parser) parseStmt() (ast.Stmt, bool) {
	if node, ok := p.memoize("let", p.parseLet); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("fn", p.parseFn); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("struct", p.parseStruct); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("model", p.parseModel); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("train", p.parseTrain); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("if", func() (ast.Node, bool) {
		n, ok := p.parseIf()
		return n, ok
	}); ok {
		if stmt, ok := node.(ast.Stmt); ok {
			return stmt, true
		}
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("for", p.parseFor); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("while", p.parseWhile); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("return", p.parseReturn); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("break", p.parseBreak); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("continue", p.parseContinue); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("import", p.parseImport); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("print", p.parsePrint); ok {
		return node.(ast.Stmt), true
	}
	if node, ok := p.memoize("assert", p.parseAssert); ok {
		return node.(ast.Stmt), true
	}
	if p.current().Kind == common.TokenSymbol && p.current().Literal == common.LBrace {
		block, ok := p.parseBlock()
		if ok {
			return block, true
		}
		return nil, false
	}
	if node, ok := p.memoize("expr_stmt", p.parseExprStmt); ok {
		return node.(ast.Stmt), true
	}
	return nil, false
}

func (p *Parser) parseLet() (ast.Node, bool) {
	if !p.expectKeyword("let") {
		return nil, false
	}
	loc := p.current().Pos
	mut := p.expectKeyword("mut")
	nameTok := p.current()
	if nameTok.Kind != common.TokenIdent {
		p.error("expected identifier after let")
		return nil, false
	}
	p.advance()
	name := &ast.Ident{Name: nameTok.Lexeme, Loc: nameTok.Pos}
	var typeExpr ast.Expr
	if p.expectSymbol(common.Colon) {
		typeExpr, _ = p.parseType()
	}
	if !p.expectSymbol(common.Equals) {
		p.error("expected '=' in let assignment")
		return nil, false
	}
	val, ok := p.parseExpr()
	if !ok {
		p.error("expected expression after '='")
		return nil, false
	}
	p.expectSymbol(common.Semi)
	return &ast.LetStmt{Name: name, Type: typeExpr, Value: val, Mut: mut, Loc: loc}, true
}

func (p *Parser) parseFn() (ast.Node, bool) {
	if !p.expectKeyword("fn") {
		return nil, false
	}
	loc := p.current().Pos
	nameTok := p.current()
	if nameTok.Kind != common.TokenIdent {
		p.error("expected function name")
		return nil, false
	}
	p.advance()
	name := &ast.Ident{Name: nameTok.Lexeme, Loc: nameTok.Pos}

	if !p.expectSymbol(common.LParen) {
		p.error("expected '(' after function name")
		return nil, false
	}
	params := p.parseParamList()
	if !p.expectSymbol(common.RParen) {
		p.error("expected ')' after parameters")
		return nil, false
	}
	var returnType ast.Expr
	if p.expectSymbol(common.Arrow) {
		returnType, _ = p.parseType()
	}
	body, ok := p.parseBlock()
	if !ok {
		p.error("expected function body")
		return nil, false
	}
	return &ast.FnDef{Name: name, Params: params, Body: body, Return: returnType, Loc: loc}, true
}

func (p *Parser) parseParamList() []*ast.ParamDef {
	var params []*ast.ParamDef
	for {
		if p.current().Kind == common.TokenSymbol && p.current().Literal == common.RParen {
			break
		}
		nameTok := p.current()
		if nameTok.Kind != common.TokenIdent {
			break
		}
		p.advance()
		name := &ast.Ident{Name: nameTok.Lexeme, Loc: nameTok.Pos}
		var typeExpr ast.Expr
		if p.expectSymbol(common.Colon) {
			typeExpr, _ = p.parseType()
		}
		params = append(params, &ast.ParamDef{Name: name, Type: typeExpr, Loc: nameTok.Pos})
		if !p.expectSymbol(common.Comma) {
			break
		}
	}
	return params
}

func (p *Parser) parseBlock() (*ast.BlockStmt, bool) {
	if !p.expectSymbol(common.LBrace) {
		return nil, false
	}
	loc := p.current().Pos
	stmts := p.parseBlockBody()
	if !p.expectSymbol(common.RBrace) {
		p.error("expected '}'")
		return nil, false
	}
	return &ast.BlockStmt{Stmts: stmts, Loc: loc}, true
}

func (p *Parser) parseBlockBody() []ast.Stmt {
	var stmts []ast.Stmt
	for p.pos < len(p.tokens) {
		if p.current().Kind == common.TokenEOF || p.current().Lexeme == "}" {
			break
		}
		if stmt, ok := p.parseStmt(); ok {
			stmts = append(stmts, stmt)
		} else {
			break
		}
	}
	return stmts
}

func (p *Parser) parseIf() (*ast.IfStmt, bool) {
	if !p.expectKeyword("if") {
		return nil, false
	}
	loc := p.current().Pos
	cond, ok := p.parseExpr()
	if !ok {
		p.error("expected condition after 'if'")
		return nil, false
	}
	thenBlock, ok := p.parseBlock()
	if !ok {
		p.error("expected block after if condition")
		return nil, false
	}
	var elseStmt ast.Stmt
	if p.expectKeyword("else") {
		if p.current().Lexeme == "if" {
			elseStmt, _ = p.parseIf()
		} else {
			elseBlock, ok := p.parseBlock()
			if ok {
				elseStmt = elseBlock
			}
		}
	}
	return &ast.IfStmt{Cond: cond, Then: thenBlock, Else: elseStmt, Loc: loc}, true
}

func (p *Parser) parseFor() (ast.Node, bool) {
	if !p.expectKeyword("for") {
		return nil, false
	}
	loc := p.current().Pos
	varTok := p.current()
	if varTok.Kind != common.TokenIdent {
		p.error("expected loop variable")
		return nil, false
	}
	p.advance()
	varIdent := &ast.Ident{Name: varTok.Lexeme, Loc: varTok.Pos}
	if !p.expectKeyword("in") && !p.expectSymbol(common.Equals) && !p.expectSymbol(common.Colon) {
	}
	rangeExpr, ok := p.parseExpr()
	if !ok {
		p.error("expected range expression")
		return nil, false
	}
	body, ok := p.parseBlock()
	if !ok {
		p.error("expected loop body")
		return nil, false
	}
	return &ast.ForStmt{Var: varIdent, Range: rangeExpr, Body: body, Loc: loc}, true
}

func (p *Parser) parseWhile() (ast.Node, bool) {
	if !p.expectKeyword("while") {
		return nil, false
	}
	loc := p.current().Pos
	cond, ok := p.parseExpr()
	if !ok {
		p.error("expected condition after 'while'")
		return nil, false
	}
	body, ok := p.parseBlock()
	if !ok {
		p.error("expected loop body")
		return nil, false
	}
	return &ast.WhileStmt{Cond: cond, Body: body, Loc: loc}, true
}

func (p *Parser) parseReturn() (ast.Node, bool) {
	if !p.expectKeyword("return") {
		return nil, false
	}
	loc := p.current().Pos
	var val ast.Expr
	if p.current().Lexeme != ";" && p.current().Kind != common.TokenEOF {
		val, _ = p.parseExpr()
	}
	p.expectSymbol(common.Semi)
	return &ast.ReturnStmt{Value: val, Loc: loc}, true
}

func (p *Parser) parseBreak() (ast.Node, bool) {
	if !p.expectKeyword("break") {
		return nil, false
	}
	loc := p.current().Pos
	p.expectSymbol(common.Semi)
	return &ast.BreakStmt{Loc: loc}, true
}

func (p *Parser) parseContinue() (ast.Node, bool) {
	if !p.expectKeyword("continue") {
		return nil, false
	}
	loc := p.current().Pos
	p.expectSymbol(common.Semi)
	return &ast.ContinueStmt{Loc: loc}, true
}

func (p *Parser) parseImport() (ast.Node, bool) {
	if !p.expectKeyword("import") {
		return nil, false
	}
	loc := p.current().Pos
	strTok := p.current()
	if strTok.Kind != common.TokenString {
		p.error("expected string path after import")
		return nil, false
	}
	p.advance()
	p.expectSymbol(common.Semi)
	return &ast.ImportStmt{Path: &ast.StringLit{Value: strTok.Lexeme, Loc: strTok.Pos}, Loc: loc}, true
}

func (p *Parser) parsePrint() (ast.Node, bool) {
	if !p.expectKeyword("print") {
		return nil, false
	}
	loc := p.current().Pos
	expr, ok := p.parseExpr()
	if !ok {
		p.error("expected expression after print")
		return nil, false
	}
	p.expectSymbol(common.Semi)
	return &ast.PrintStmt{Expr: expr, Loc: loc}, true
}

func (p *Parser) parseAssert() (ast.Node, bool) {
	if !p.expectKeyword("assert") {
		return nil, false
	}
	loc := p.current().Pos
	cond, ok := p.parseExpr()
	if !ok {
		p.error("expected condition after assert")
		return nil, false
	}
	var msg ast.Expr
	if p.current().Lexeme == "," {
		p.advance()
		msg, _ = p.parseExpr()
	}
	p.expectSymbol(common.Semi)
	return &ast.AssertStmt{Cond: cond, Msg: msg, Loc: loc}, true
}

func (p *Parser) parseStruct() (ast.Node, bool) {
	if !p.expectKeyword("struct") {
		return nil, false
	}
	loc := p.current().Pos
	nameTok := p.current()
	if nameTok.Kind != common.TokenIdent {
		p.error("expected struct name")
		return nil, false
	}
	p.advance()
	name := &ast.Ident{Name: nameTok.Lexeme, Loc: nameTok.Pos}
	if !p.expectSymbol(common.LBrace) {
		p.error("expected '{' for struct body")
		return nil, false
	}
	var fields []*ast.FieldDef
	for {
		if p.current().Lexeme == "}" || p.current().Kind == common.TokenEOF {
			break
		}
		fieldTok := p.current()
		if fieldTok.Kind != common.TokenIdent {
			break
		}
		p.advance()
		fieldName := &ast.Ident{Name: fieldTok.Lexeme, Loc: fieldTok.Pos}
		var fieldType ast.Expr
		if p.expectSymbol(common.Colon) {
			fieldType, _ = p.parseType()
		}
		fields = append(fields, &ast.FieldDef{Name: fieldName, Type: fieldType, Loc: fieldTok.Pos})
		if !p.expectSymbol(common.Comma) {
			break
		}
	}
	if !p.expectSymbol(common.RBrace) {
		p.error("expected '}' after struct fields")
	}
	return &ast.StructDef{Name: name, Fields: fields, Loc: loc}, true
}

func (p *Parser) parseModel() (ast.Node, bool) {
	if !p.expectKeyword("model") {
		return nil, false
	}
	loc := p.current().Pos
	nameTok := p.current()
	if nameTok.Kind != common.TokenIdent {
		p.error("expected model name")
		return nil, false
	}
	p.advance()
	name := &ast.Ident{Name: nameTok.Lexeme, Loc: nameTok.Pos}
	if !p.expectSymbol(common.LBrace) {
		p.error("expected '{' for model body")
		return nil, false
	}
	var layers []*ast.LayerDef
	for {
		if p.current().Lexeme == "}" || p.current().Kind == common.TokenEOF {
			break
		}
		layerTok := p.current()
		if layerTok.Kind != common.TokenIdent {
			break
		}
		p.advance()
		layerName := &ast.Ident{Name: layerTok.Lexeme, Loc: layerTok.Pos}
		if !p.expectSymbol(common.Equals) {
			p.error("expected '=' after layer name")
			break
		}
		layerKind, ok := p.parseExpr()
		if !ok {
			p.error("expected layer kind")
			break
		}
		var params []*ast.NamedParam
		if p.expectSymbol(common.Comma) {
			params = p.parseNamedParams()
		}
		layers = append(layers, &ast.LayerDef{Name: layerName, Kind: layerKind, Params: params, Loc: layerTok.Pos})
	}
	if !p.expectSymbol(common.RBrace) {
		p.error("expected '}' after model layers")
	}
	return &ast.ModelDef{Name: name, Layers: layers, Loc: loc}, true
}

func (p *Parser) parseNamedParams() []*ast.NamedParam {
	var params []*ast.NamedParam
	for {
		nameTok := p.current()
		if nameTok.Kind != common.TokenIdent {
			break
		}
		p.advance()
		name := &ast.Ident{Name: nameTok.Lexeme, Loc: nameTok.Pos}
		if !p.expectSymbol(common.Equals) {
			break
		}
		val, ok := p.parseExpr()
		if !ok {
			break
		}
		params = append(params, &ast.NamedParam{Name: name, Value: val, Loc: nameTok.Pos})
		if !p.expectSymbol(common.Comma) {
			break
		}
	}
	return params
}

func (p *Parser) parseTrain() (ast.Node, bool) {
	if !p.expectKeyword("train") {
		return nil, false
	}
	loc := p.current().Pos
	modelExpr, ok := p.parseExpr()
	if !ok {
		p.error("expected model expression after train")
		return nil, false
	}
	p.expectSymbol(common.Comma)
	var data ast.Expr
	var epochs ast.Expr
	var lr ast.Expr
	var strategy ast.Expr
	var devices ast.Expr
	for {
		nameTok := p.current()
		if nameTok.Kind != common.TokenIdent {
			break
		}
		p.advance()
		switch nameTok.Lexeme {
		case "data":
			p.expectSymbol(common.Colon)
			data, _ = p.parseExpr()
		case "epochs":
			p.expectSymbol(common.Colon)
			epochs, _ = p.parseExpr()
		case "lr":
			p.expectSymbol(common.Colon)
			lr, _ = p.parseExpr()
		case "strategy":
			p.expectSymbol(common.Colon)
			strategy, _ = p.parseExpr()
		case "devices":
			p.expectSymbol(common.Colon)
			devices, _ = p.parseExpr()
		}
		if !p.expectSymbol(common.Comma) {
			break
		}
	}
	return &ast.TrainStmt{Model: modelExpr, Data: data, Epochs: epochs, LR: lr, Strategy: strategy, Devices: devices, Loc: loc}, true
}

func (p *Parser) parseExprStmt() (ast.Node, bool) {
	if node, ok := p.parseExpr(); ok {
		p.expectSymbol(common.Semi)
		return &ast.ExprStmt{E: node, Loc: node.Pos()}, true
	}
	return nil, false
}

func (p *Parser) parseExpr() (ast.Expr, bool) {
	return p.parseAssign()
}

func (p *Parser) parseAssign() (ast.Expr, bool) {
	left, ok := p.parseOr()
	if !ok {
		return nil, false
	}
	if p.current().Literal == common.Equals {
		p.advance()
		right, ok := p.parseAssign()
		if !ok {
			p.error("expected expression after '='")
			return nil, false
		}
		return &ast.AssignExpr{Left: left, Right: right, Loc: left.Pos()}, true
	}
	return left, true
}

func (p *Parser) parseOr() (ast.Expr, bool) {
	left, ok := p.parseAnd()
	if !ok {
		return nil, false
	}
	for p.current().Lexeme == "or" {
		op := p.advance().Lexeme
		right, ok := p.parseAnd()
		if !ok {
			return nil, false
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Loc: left.Pos()}
	}
	return left, true
}

func (p *Parser) parseAnd() (ast.Expr, bool) {
	left, ok := p.parseComparison()
	if !ok {
		return nil, false
	}
	for p.current().Lexeme == "and" {
		op := p.advance().Lexeme
		right, ok := p.parseComparison()
		if !ok {
			return nil, false
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Loc: left.Pos()}
	}
	return left, true
}

func (p *Parser) parseComparison() (ast.Expr, bool) {
	left, ok := p.parseAddition()
	if !ok {
		return nil, false
	}
	compOps := map[string]bool{"==": true, "!=": true, "<": true, "<=": true, ">": true, ">=": true}
	tok := p.current()
	if tok.Kind == common.TokenSymbol {
		if s, ok := tok.Literal.(common.Symbol); ok {
			op := s.String()
			if compOps[op] {
				p.advance()
				right, ok := p.parseAddition()
				if !ok {
					return nil, false
				}
				left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Loc: left.Pos()}
			}
		}
	}
	return left, true
}

func (p *Parser) parseAddition() (ast.Expr, bool) {
	left, ok := p.parseMultiplication()
	if !ok {
		return nil, false
	}
	for {
		tok := p.current()
		if tok.Kind != common.TokenSymbol {
			break
		}
		s, ok := tok.Literal.(common.Symbol)
		if !ok {
			break
		}
		var op string
		switch s {
		case common.Plus:
			op = "+"
		case common.Minus:
			op = "-"
		default:
			break
		}
		if op == "" {
			break
		}
		p.advance()
		right, ok := p.parseMultiplication()
		if !ok {
			return nil, false
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Loc: left.Pos()}
	}
	return left, true
}

func (p *Parser) parseMultiplication() (ast.Expr, bool) {
	left, ok := p.parseUnary()
	if !ok {
		return nil, false
	}
	for {
		tok := p.current()
		if tok.Kind != common.TokenSymbol {
			break
		}
		s, ok := tok.Literal.(common.Symbol)
		if !ok {
			break
		}
		var op string
		switch s {
		case common.Star:
			op = "*"
		case common.Slash:
			op = "/"
		case common.Percent:
			op = "%"
		default:
			break
		}
		if op == "" {
			break
		}
		p.advance()
		right, ok := p.parseUnary()
		if !ok {
			return nil, false
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Loc: left.Pos()}
	}
	return left, true
}

func (p *Parser) parseUnary() (ast.Expr, bool) {
	tok := p.current()
	if tok.Kind == common.TokenSymbol {
		s, ok := tok.Literal.(common.Symbol)
		if ok && (s == common.Minus || s == common.Bang) {
			op := s.String()
			p.advance()
			right, ok := p.parseUnary()
			if !ok {
				return nil, false
			}
			return &ast.UnaryExpr{Op: op, Right: right, Loc: tok.Pos}, true
		}
	}
	return p.parseCall()
}

func (p *Parser) parseCall() (ast.Expr, bool) {
	expr, ok := p.parsePrimary()
	if !ok {
		return nil, false
	}
	for {
		if p.current().Literal == common.LParen {
			p.advance()
			args := p.parseArgList()
			if !p.expectSymbol(common.RParen) {
				p.error("expected ')' after arguments")
				return expr, true
			}
			expr = &ast.CallExpr{Fn: expr, Args: args, Loc: expr.Pos()}
		} else if p.current().Literal == common.Dot {
			p.advance()
			nameTok := p.current()
			if nameTok.Kind != common.TokenIdent {
				return expr, true
			}
			p.advance()
			expr = &ast.MemberExpr{Target: expr, Name: &ast.Ident{Name: nameTok.Lexeme, Loc: nameTok.Pos}, Loc: expr.Pos()}
		} else if p.current().Literal == common.LBracket {
			p.advance()
			index, ok := p.parseExpr()
			if !ok {
				return expr, true
			}
			if !p.expectSymbol(common.RBracket) {
				p.error("expected ']' after index")
				return expr, true
			}
			expr = &ast.IndexExpr{Target: expr, Index: index, Loc: expr.Pos()}
		} else {
			break
		}
	}
	return expr, true
}

func (p *Parser) parseArgList() []ast.Expr {
	var args []ast.Expr
	for {
		if p.current().Literal == common.RParen || p.current().Kind == common.TokenEOF {
			break
		}
		expr, ok := p.parseExpr()
		if !ok {
			break
		}
		args = append(args, expr)
		if !p.expectSymbol(common.Comma) {
			break
		}
	}
	return args
}

func (p *Parser) parsePrimary() (ast.Expr, bool) {
	tok := p.current()
	switch tok.Kind {
	case common.TokenIdent:
		p.advance()
		return &ast.Ident{Name: tok.Lexeme, Loc: tok.Pos}, true
	case common.TokenNumber:
		p.advance()
		return &ast.NumberLit{Value: tok.Lexeme, Loc: tok.Pos}, true
	case common.TokenString:
		p.advance()
		if lit, ok := tok.Literal.(string); ok {
			return &ast.StringLit{Value: lit, Loc: tok.Pos}, true
		}
		return &ast.StringLit{Value: tok.Lexeme, Loc: tok.Pos}, true
	case common.TokenKeyword:
		switch tok.Lexeme {
		case "true":
			p.advance()
			return &ast.BoolLit{Value: true, Loc: tok.Pos}, true
		case "false":
			p.advance()
			return &ast.BoolLit{Value: false, Loc: tok.Pos}, true
		}
		if tok.Lexeme == "tensor" {
			return p.parseTensorType()
		}
		if tok.Lexeme == "f32" || tok.Lexeme == "f64" || tok.Lexeme == "i32" ||
			tok.Lexeme == "i64" || tok.Lexeme == "i8" || tok.Lexeme == "i16" ||
			tok.Lexeme == "u8" || tok.Lexeme == "u16" || tok.Lexeme == "u32" ||
			tok.Lexeme == "u64" || tok.Lexeme == "bool" || tok.Lexeme == "void" ||
			tok.Lexeme == "string" {
			p.advance()
			return &ast.TypeExpr{Name: tok.Lexeme, Loc: tok.Pos}, true
		}
		p.advance()
		return &ast.Ident{Name: tok.Lexeme, Loc: tok.Pos}, true
	case common.TokenSymbol:
		if tok.Literal == common.LParen {
			p.advance()
			expr, ok := p.parseExpr()
			if !ok {
				return nil, false
			}
			if !p.expectSymbol(common.RParen) {
				p.error("expected ')'")
				return expr, true
			}
			return expr, true
		}
		if tok.Literal == common.LBracket {
			p.advance()
			var elems []ast.Expr
			for {
				if p.current().Literal == common.RBracket || p.current().Kind == common.TokenEOF {
					break
				}
				elem, ok := p.parseExpr()
				if !ok {
					break
				}
				elems = append(elems, elem)
				if !p.expectSymbol(common.Comma) {
					break
				}
			}
			p.expectSymbol(common.RBracket)
			return &ast.ArrayLit{Elems: elems, Loc: tok.Pos}, true
		}
		return nil, false
	default:
		return nil, false
	}
}

func (p *Parser) parseType() (ast.Expr, bool) {
	tok := p.current()
	if tok.Kind == common.TokenIdent || tok.Kind == common.TokenKeyword {
		name := tok.Lexeme
		p.advance()
		if name == "tensor" {
			return p.parseTensorType()
		}
		return &ast.TypeExpr{Name: name, Loc: tok.Pos}, true
	}
	return nil, false
}

func (p *Parser) parseTensorType() (ast.Expr, bool) {
	p.expectKeyword("tensor")
	if !p.expectSymbol(common.Less) {
		p.error("expected '<' after tensor")
		return nil, false
	}
	elemType, ok := p.parseType()
	if !ok {
		p.error("expected element type in tensor")
		return nil, false
	}
	var dims []ast.Expr
	if p.expectSymbol(common.Comma) {
		for {
			if p.current().Literal == common.Greater || p.current().Kind == common.TokenEOF {
				break
			}
			if p.current().Lexeme == "?" {
				p.advance()
				dims = append(dims, nil)
			} else {
				dim, ok := p.parseAddition()
				if ok {
					dims = append(dims, dim)
				}
			}
			if !p.expectSymbol(common.Comma) {
				break
			}
		}
	}
	if !p.expectSymbol(common.Greater) {
		p.error("expected '>' in tensor type")
		return nil, false
	}
	return &ast.TensorTypeExpr{ElemType: elemType, Dims: dims, Loc: elemType.Pos()}, true
}

func (p *Parser) Errors() []*common.CompilerError {
	return p.errs
}

func ParseTokens(tokens []common.Token, file string) (*ast.Program, []*common.CompilerError) {
	p := New(tokens, file)
	prog := p.Parse()
	return prog, p.errs
}
