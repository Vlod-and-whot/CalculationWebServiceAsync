package application

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type ASTNode struct {
	Operator  string
	Value     float64
	Left      *ASTNode
	Right     *ASTNode
	TaskID    int
	Evaluated bool
}

func ParseExpression(input string) (*ASTNode, error) {
	tokens := tokenize(input)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("пустое выражение")
	}
	node, remaining, err := parseExpr(tokens)
	if err != nil {
		return nil, err
	}
	if len(remaining) > 0 {
		return nil, fmt.Errorf("неожиданные токены: %v", remaining)
	}
	return node, nil
}

func tokenize(input string) []string {
	var tokens []string
	var number strings.Builder
	for _, ch := range input {
		if unicode.IsDigit(ch) || ch == '.' {
			number.WriteRune(ch)
		} else if ch == ' ' {
			if number.Len() > 0 {
				tokens = append(tokens, number.String())
				number.Reset()
			}
		} else if ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '(' || ch == ')' {
			if number.Len() > 0 {
				tokens = append(tokens, number.String())
				number.Reset()
			}
			tokens = append(tokens, string(ch))
		}
	}
	if number.Len() > 0 {
		tokens = append(tokens, number.String())
	}
	return tokens
}

func parseExpr(tokens []string) (*ASTNode, []string, error) {
	node, tokens, err := parseTerm(tokens)
	if err != nil {
		return nil, tokens, err
	}
	for len(tokens) > 0 && (tokens[0] == "+" || tokens[0] == "-") {
		op := tokens[0]
		tokens = tokens[1:]
		right, rem, err := parseTerm(tokens)
		if err != nil {
			return nil, rem, err
		}
		node = &ASTNode{
			Operator: op,
			Left:     node,
			Right:    right,
		}
		tokens = rem
	}
	return node, tokens, nil
}

func parseTerm(tokens []string) (*ASTNode, []string, error) {
	node, tokens, err := parseFactor(tokens)
	if err != nil {
		return nil, tokens, err
	}
	for len(tokens) > 0 && (tokens[0] == "*" || tokens[0] == "/") {
		op := tokens[0]
		tokens = tokens[1:]
		right, rem, err := parseFactor(tokens)
		if err != nil {
			return nil, rem, err
		}
		node = &ASTNode{
			Operator: op,
			Left:     node,
			Right:    right,
		}
		tokens = rem
	}
	return node, tokens, nil
}

func parseFactor(tokens []string) (*ASTNode, []string, error) {
	if len(tokens) == 0 {
		return nil, tokens, fmt.Errorf("неожиданный конец выражения")
	}
	token := tokens[0]
	if token == "(" {
		node, rem, err := parseExpr(tokens[1:])
		if err != nil {
			return nil, rem, err
		}
		if len(rem) == 0 || rem[0] != ")" {
			return nil, rem, fmt.Errorf("ожидалась закрывающая скобка")
		}
		return node, rem[1:], nil
	}
	value, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return nil, tokens, fmt.Errorf("некорректное число: %s", token)
	}
	node := &ASTNode{
		Value: value,
	}
	return node, tokens[1:], nil
}
