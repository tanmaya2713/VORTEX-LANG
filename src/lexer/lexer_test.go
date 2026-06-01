package lexer

import (
	"testing"

	"github.com/vortex-lang/vortex/src/common"
)

func TestBasicTokens(t *testing.T) {
	tokens, err := Lex("let x = 42;", "test.vtx")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) < 5 {
		t.Fatalf("expected at least 5 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != common.TokenKeyword || tokens[0].Lexeme != "let" {
		t.Errorf("expected keyword 'let', got %v", tokens[0])
	}
	if tokens[1].Kind != common.TokenIdent || tokens[1].Lexeme != "x" {
		t.Errorf("expected ident 'x', got %v", tokens[1])
	}
	if tokens[2].Lexeme != "=" {
		t.Errorf("expected '=' symbol, got %v", tokens[2])
	}
	if tokens[3].Kind != common.TokenNumber || tokens[3].Lexeme != "42" {
		t.Errorf("expected number 42, got %v", tokens[3])
	}
	if tokens[len(tokens)-1].Kind != common.TokenEOF {
		t.Errorf("expected EOF, got %v", tokens[len(tokens)-1])
	}
}

func TestNoiseFiltering(t *testing.T) {
	input := "let x the is be to = 42 and or not;"
	tokens, err := Lex(input, "test.vtx")
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range tokens {
		if tok.Kind == common.TokenIdent {
			switch tok.Lexeme {
			case "the", "is", "be", "to", "and", "or", "not":
				t.Errorf("noise word '%s' should have been filtered", tok.Lexeme)
			}
		}
	}
}

func TestKeywords(t *testing.T) {
	tokens, err := Lex("fn if else for while return break continue struct model train", "test.vtx")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"fn", "if", "else", "for", "while", "return", "break", "continue", "struct", "model", "train"}
	idents := 0
	for _, tok := range tokens {
		if tok.Kind == common.TokenKeyword {
			if idents >= len(expected) {
				t.Errorf("unexpected keyword: %s", tok.Lexeme)
				continue
			}
			if tok.Lexeme != expected[idents] {
				t.Errorf("expected keyword %s, got %s", expected[idents], tok.Lexeme)
			}
			idents++
		}
	}
	if idents != len(expected) {
		t.Errorf("expected %d keywords, found %d", len(expected), idents)
	}
}

func TestNumbers(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"42", "42"},
		{"3.14", "3.14"},
		{"1e10", "1e10"},
		{"2.5e-3", "2.5e-3"},
	}
	for _, tt := range tests {
		tokens, err := Lex(tt.input, "test.vtx")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, tok := range tokens {
			if tok.Kind == common.TokenNumber && tok.Lexeme == tt.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected number %s in %s", tt.want, tt.input)
		}
	}
}

func TestStrings(t *testing.T) {
	tokens, err := Lex("print \"hello world\";", "test.vtx")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tok := range tokens {
		if tok.Kind == common.TokenString {
			found = true
			if tok.Lexeme != "hello world" {
				t.Errorf("expected 'hello world', got '%s'", tok.Lexeme)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected string token")
	}
}

func TestComments(t *testing.T) {
	tokens, err := Lex("// this is a comment\nlet x = 1;", "test.vtx")
	if err != nil {
		t.Fatal(err)
	}
	hasLet := false
	hasComment := false
	for _, tok := range tokens {
		if tok.Kind == common.TokenKeyword && tok.Lexeme == "let" {
			hasLet = true
		}
		if tok.Kind == common.TokenComment {
			hasComment = true
		}
	}
	if !hasLet {
		t.Errorf("expected 'let' keyword after comment")
	}
	if hasComment {
		t.Errorf("expected comments to be stripped")
	}
}

func TestSymbols(t *testing.T) {
	input := "+ - * / = == ! != < > <= >= ( ) { } [ ] , ; : . ->"
	tokens, err := Lex(input, "test.vtx")
	if err != nil {
		t.Fatal(err)
	}
	symbols := []string{"+", "-", "*", "/", "=", "==", "!", "!=", "<", ">", "<=", ">=", "(", ")", "{", "}", "[", "]", ",", ";", ":", ".", "->"}
	symIdx := 0
	for _, tok := range tokens {
		if tok.Kind == common.TokenSymbol {
			if symIdx >= len(symbols) {
				continue
			}
			if tok.Lexeme != symbols[symIdx] {
				t.Fatalf("expected symbol %s, got %s at index %d", symbols[symIdx], tok.Lexeme, symIdx)
			}
			symIdx++
		}
	}
}

func TestFloats(t *testing.T) {
	tokens, err := Lex("let pi = 3.14159;", "test.vtx")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tok := range tokens {
		if tok.Kind == common.TokenNumber && tok.Lexeme == "3.14159" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected float number 3.14159")
	}
}

func TestMLKeywords(t *testing.T) {
	input := "model MNISTClassifier:\n    hidden = dense, inputs:784\n"
	tokens, err := Lex(input, "test.vtx")
	if err != nil {
		t.Fatal(err)
	}
	hasModel := false
	hasDense := false
	for _, tok := range tokens {
		if tok.Kind == common.TokenKeyword {
			switch tok.Lexeme {
			case "model":
				hasModel = true
			case "dense":
				hasDense = true
			}
		}
	}
	if !hasModel {
		t.Errorf("expected 'model' keyword")
	}
	if !hasDense {
		t.Errorf("expected 'dense' keyword")
	}
}
