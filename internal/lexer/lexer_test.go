package lexer

import "testing"

func TestNextToken(t *testing.T) {
	input := `stream chat_input [level: secure]`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{STREAM, "stream"},
		{IDENT, "chat_input"},
		{LBRACKET, "["},
		{LEVEL, "level"},
		{COLON, ":"},
		{IDENT, "secure"},
		{RBRACKET, "]"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
