package parser

import (
	"testing"
	"vigil/internal/ast"
	"vigil/internal/lexer"
)

func TestParsing(t *testing.T) {
	input := `
	stream bank_kyc [level: secure]

	on bank_kyc as client_data {
		mask(client_data)
	}
	`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements. got=%d", len(program.Statements))
	}

	// Проверяем stream
	streamStmt, ok := program.Statements[0].(*ast.StreamStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.StreamStatement. got=%T", program.Statements[0])
	}
	if streamStmt.Name.Value != "bank_kyc" {
		t.Errorf("stream name wrong. got=%s", streamStmt.Name.Value)
	}
	if streamStmt.Level.Value != "secure" {
		t.Errorf("stream level wrong. got=%s", streamStmt.Level.Value)
	}

	// Проверяем on ... as ...
	reactorStmt, ok := program.Statements[1].(*ast.ReactorStatement)
	if !ok {
		t.Fatalf("program.Statements[1] is not ast.ReactorStatement. got=%T", program.Statements[1])
	}
	if reactorStmt.Variable.Value != "client_data" {
		t.Errorf("reactor variable wrong. got=%s", reactorStmt.Variable.Value)
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
