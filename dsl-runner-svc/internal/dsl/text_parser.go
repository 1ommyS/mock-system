package dsl

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokenEOF tokenKind = iota
	tokenIdent
	tokenString
	tokenNumber

	tokenLBrace
	tokenRBrace
	tokenLBracket
	tokenRBracket
	tokenLParen
	tokenRParen
	tokenComma
	tokenColon
	tokenSemicolon
	tokenDot
	tokenEqual
	tokenPlus

	tokenVersion
	tokenLet
	tokenIf
	tokenElse
	tokenRepeat
	tokenFor
	tokenIn
	tokenEmit
	tokenAssert
	tokenTrue
	tokenFalse
	tokenNull
)

type token struct {
	kind tokenKind
	lit  string
	line int
	col  int
}

type textLexer struct {
	src  []rune
	pos  int
	line int
	col  int
}

func newTextLexer(data []byte) *textLexer {
	return &textLexer{
		src:  []rune(string(data)),
		line: 1,
		col:  1,
	}
}

func (l *textLexer) lexAll() ([]token, error) {
	out := make([]token, 0, 128)
	for {
		if err := l.skipSpaceAndComments(); err != nil {
			return nil, err
		}
		startLine, startCol := l.line, l.col
		ch := l.peek()
		if ch == 0 {
			out = append(out, token{kind: tokenEOF, line: startLine, col: startCol})
			return out, nil
		}
		switch ch {
		case '{':
			l.advance()
			out = append(out, token{kind: tokenLBrace, lit: "{", line: startLine, col: startCol})
		case '}':
			l.advance()
			out = append(out, token{kind: tokenRBrace, lit: "}", line: startLine, col: startCol})
		case '[':
			l.advance()
			out = append(out, token{kind: tokenLBracket, lit: "[", line: startLine, col: startCol})
		case ']':
			l.advance()
			out = append(out, token{kind: tokenRBracket, lit: "]", line: startLine, col: startCol})
		case '(':
			l.advance()
			out = append(out, token{kind: tokenLParen, lit: "(", line: startLine, col: startCol})
		case ')':
			l.advance()
			out = append(out, token{kind: tokenRParen, lit: ")", line: startLine, col: startCol})
		case ',':
			l.advance()
			out = append(out, token{kind: tokenComma, lit: ",", line: startLine, col: startCol})
		case ':':
			l.advance()
			out = append(out, token{kind: tokenColon, lit: ":", line: startLine, col: startCol})
		case ';':
			l.advance()
			out = append(out, token{kind: tokenSemicolon, lit: ";", line: startLine, col: startCol})
		case '.':
			l.advance()
			out = append(out, token{kind: tokenDot, lit: ".", line: startLine, col: startCol})
		case '=':
			l.advance()
			out = append(out, token{kind: tokenEqual, lit: "=", line: startLine, col: startCol})
		case '+':
			l.advance()
			out = append(out, token{kind: tokenPlus, lit: "+", line: startLine, col: startCol})
		case '"':
			str, err := l.readString()
			if err != nil {
				return nil, err
			}
			out = append(out, token{kind: tokenString, lit: str, line: startLine, col: startCol})
		default:
			if isIdentStart(ch) {
				ident := l.readIdent()
				out = append(out, token{kind: keywordKind(ident), lit: ident, line: startLine, col: startCol})
				continue
			}
			if isDigit(ch) {
				num := l.readNumber()
				out = append(out, token{kind: tokenNumber, lit: num, line: startLine, col: startCol})
				continue
			}
			return nil, fmt.Errorf("line %d:%d: unexpected character %q", startLine, startCol, ch)
		}
	}
}

func (l *textLexer) skipSpaceAndComments() error {
	for {
		for {
			ch := l.peek()
			if ch == 0 || !unicode.IsSpace(ch) {
				break
			}
			l.advance()
		}
		if l.peek() == '#' {
			for ch := l.peek(); ch != 0 && ch != '\n'; ch = l.peek() {
				l.advance()
			}
			continue
		}
		if l.peek() == '/' && l.peekN(1) == '/' {
			l.advance()
			l.advance()
			for ch := l.peek(); ch != 0 && ch != '\n'; ch = l.peek() {
				l.advance()
			}
			continue
		}
		if l.peek() == '/' && l.peekN(1) == '*' {
			l.advance()
			l.advance()
			for {
				if l.peek() == 0 {
					return fmt.Errorf("line %d:%d: unterminated block comment", l.line, l.col)
				}
				if l.peek() == '*' && l.peekN(1) == '/' {
					l.advance()
					l.advance()
					break
				}
				l.advance()
			}
			continue
		}
		return nil
	}
}

func (l *textLexer) readString() (string, error) {
	startLine, startCol := l.line, l.col
	var b strings.Builder
	b.WriteRune(l.advance()) // opening quote
	for {
		ch := l.peek()
		if ch == 0 {
			return "", fmt.Errorf("line %d:%d: unterminated string", startLine, startCol)
		}
		b.WriteRune(l.advance())
		if ch == '\\' {
			next := l.peek()
			if next == 0 {
				return "", fmt.Errorf("line %d:%d: unterminated string escape", startLine, startCol)
			}
			b.WriteRune(l.advance())
			continue
		}
		if ch == '"' {
			break
		}
	}
	out, err := strconv.Unquote(b.String())
	if err != nil {
		return "", fmt.Errorf("line %d:%d: invalid string: %w", startLine, startCol, err)
	}
	return out, nil
}

func (l *textLexer) readIdent() string {
	var b strings.Builder
	for {
		ch := l.peek()
		if ch == 0 || !isIdentPart(ch) {
			break
		}
		b.WriteRune(l.advance())
	}
	return b.String()
}

func (l *textLexer) readNumber() string {
	var b strings.Builder
	hasDot := false
	for {
		ch := l.peek()
		if isDigit(ch) {
			b.WriteRune(l.advance())
			continue
		}
		if ch == '.' && !hasDot {
			hasDot = true
			b.WriteRune(l.advance())
			continue
		}
		break
	}
	return b.String()
}

func (l *textLexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *textLexer) peekN(n int) rune {
	idx := l.pos + n
	if idx < 0 || idx >= len(l.src) {
		return 0
	}
	return l.src[idx]
}

func (l *textLexer) advance() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	ch := l.src[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func isIdentStart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isIdentPart(ch rune) bool {
	return isIdentStart(ch) || isDigit(ch)
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func keywordKind(ident string) tokenKind {
	switch ident {
	case "version":
		return tokenVersion
	case "let":
		return tokenLet
	case "if":
		return tokenIf
	case "else":
		return tokenElse
	case "repeat":
		return tokenRepeat
	case "for":
		return tokenFor
	case "in":
		return tokenIn
	case "emit":
		return tokenEmit
	case "assert":
		return tokenAssert
	case "true":
		return tokenTrue
	case "false":
		return tokenFalse
	case "null":
		return tokenNull
	default:
		return tokenIdent
	}
}

type textParser struct {
	tokens []token
	pos    int
}

func parseTextScript(data []byte) (Script, error) {
	lexer := newTextLexer(data)
	tokens, err := lexer.lexAll()
	if err != nil {
		return Script{}, err
	}
	p := &textParser{tokens: tokens}
	return p.parseProgram()
}

func (p *textParser) parseProgram() (Script, error) {
	if _, err := p.expect(tokenVersion, "expected 'version'"); err != nil {
		return Script{}, err
	}
	versionTok, err := p.expect(tokenNumber, "expected version number")
	if err != nil {
		return Script{}, err
	}
	version, err := strconv.Atoi(versionTok.lit)
	if err != nil {
		return Script{}, p.errorf(versionTok, "invalid version")
	}
	if version != 1 {
		return Script{}, p.errorf(versionTok, "unsupported version: %d", version)
	}
	p.match(tokenSemicolon)
	steps, err := p.parseStepsUntil(tokenEOF)
	if err != nil {
		return Script{}, err
	}
	if len(steps) == 0 {
		return Script{}, p.errorf(p.current(), "steps are required")
	}
	if _, err := p.expect(tokenEOF, "unexpected trailing input"); err != nil {
		return Script{}, err
	}
	return Script{Version: 1, Steps: steps}, nil
}

func (p *textParser) parseStepsUntil(stop tokenKind) ([]Step, error) {
	steps := make([]Step, 0)
	for {
		for p.match(tokenSemicolon) {
		}
		if p.current().kind == stop {
			break
		}
		if p.current().kind == tokenEOF {
			if stop != tokenEOF {
				return nil, p.errorf(p.current(), "unexpected end of input")
			}
			break
		}
		step, err := p.parseStep()
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
		for p.match(tokenSemicolon) {
		}
	}
	return steps, nil
}

func (p *textParser) parseStep() (Step, error) {
	switch p.current().kind {
	case tokenLet:
		return p.parseLet()
	case tokenIf:
		return p.parseIf()
	case tokenRepeat:
		return p.parseRepeat()
	case tokenFor:
		return p.parseFor()
	case tokenEmit:
		return p.parseEmit()
	case tokenAssert:
		return p.parseAssert()
	default:
		return nil, p.errorf(p.current(), "unknown statement: %s", p.current().lit)
	}
}

func (p *textParser) parseLet() (Step, error) {
	if _, err := p.expect(tokenLet, "expected 'let'"); err != nil {
		return nil, err
	}
	name, err := p.expect(tokenIdent, "expected variable name")
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(tokenEqual, "expected '=' in let statement"); err != nil {
		return nil, err
	}
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	raw, err := marshalExpr(value)
	if err != nil {
		return nil, err
	}
	return LetStep{
		OpName: "let",
		Name:   name.lit,
		Value:  raw,
	}, nil
}

func (p *textParser) parseIf() (Step, error) {
	if _, err := p.expect(tokenIf, "expected 'if'"); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	thenSteps, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	elseSteps := []Step{}
	if p.match(tokenElse) {
		elseSteps, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}
	condRaw, err := marshalExpr(cond)
	if err != nil {
		return nil, err
	}
	return IfStep{
		OpName: "if",
		Cond:   condRaw,
		Then:   thenSteps,
		Else:   elseSteps,
	}, nil
}

func (p *textParser) parseRepeat() (Step, error) {
	if _, err := p.expect(tokenRepeat, "expected 'repeat'"); err != nil {
		return nil, err
	}
	count, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	doSteps, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	countRaw, err := marshalExpr(count)
	if err != nil {
		return nil, err
	}
	return RepeatStep{
		OpName: "repeat",
		Count:  countRaw,
		Do:     doSteps,
	}, nil
}

func (p *textParser) parseFor() (Step, error) {
	if _, err := p.expect(tokenFor, "expected 'for'"); err != nil {
		return nil, err
	}
	itemName, err := p.expect(tokenIdent, "expected item variable name")
	if err != nil {
		return nil, err
	}
	indexName := ""
	if p.match(tokenComma) {
		idx, err := p.expect(tokenIdent, "expected index variable name")
		if err != nil {
			return nil, err
		}
		indexName = idx.lit
	}
	if _, err := p.expect(tokenIn, "expected 'in' in for statement"); err != nil {
		return nil, err
	}
	itemsExpr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	loopItemRaw, err := marshalExpr(pathExpr("loop.item"))
	if err != nil {
		return nil, err
	}
	doSteps := make([]Step, 0, len(body)+2)
	doSteps = append(doSteps, LetStep{
		OpName: "let",
		Name:   itemName.lit,
		Value:  loopItemRaw,
	})
	if indexName != "" {
		loopIndexRaw, err := marshalExpr(pathExpr("loop.index"))
		if err != nil {
			return nil, err
		}
		doSteps = append(doSteps, LetStep{
			OpName: "let",
			Name:   indexName,
			Value:  loopIndexRaw,
		})
	}
	doSteps = append(doSteps, body...)

	itemsRaw, err := marshalExpr(itemsExpr)
	if err != nil {
		return nil, err
	}
	return ForEachStep{
		OpName: "forEach",
		Items:  itemsRaw,
		Do:     doSteps,
	}, nil
}

func (p *textParser) parseEmit() (Step, error) {
	if _, err := p.expect(tokenEmit, "expected 'emit'"); err != nil {
		return nil, err
	}
	nameExpr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(tokenLBrace, "expected '{' after emit name"); err != nil {
		return nil, err
	}

	var (
		requestMatch any
		response     any
		meta         any
		hasRequest   bool
		hasResponse  bool
		hasMeta      bool
	)
	for p.current().kind != tokenRBrace {
		fieldTok, err := p.expect(tokenIdent, "expected emit field name")
		if err != nil {
			return nil, err
		}
		if p.match(tokenColon) {
		}
		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		switch fieldTok.lit {
		case "requestMatch":
			requestMatch = value
			hasRequest = true
		case "responseTemplate":
			response = value
			hasResponse = true
		case "meta":
			meta = value
			hasMeta = true
		default:
			return nil, p.errorf(fieldTok, "unknown emit field: %s", fieldTok.lit)
		}
		if p.match(tokenComma) || p.match(tokenSemicolon) {
		}
	}
	if _, err := p.expect(tokenRBrace, "expected '}' to close emit block"); err != nil {
		return nil, err
	}
	if !hasRequest || !hasResponse || !hasMeta {
		return nil, p.errorf(p.current(), "emit requires requestMatch, responseTemplate and meta")
	}

	nameRaw, err := marshalExpr(nameExpr)
	if err != nil {
		return nil, err
	}
	requestRaw, err := marshalExpr(requestMatch)
	if err != nil {
		return nil, err
	}
	responseRaw, err := marshalExpr(response)
	if err != nil {
		return nil, err
	}
	metaRaw, err := marshalExpr(meta)
	if err != nil {
		return nil, err
	}
	return EmitStep{
		OpName:           "emit",
		Name:             nameRaw,
		RequestMatch:     requestRaw,
		ResponseTemplate: responseRaw,
		Meta:             metaRaw,
	}, nil
}

func (p *textParser) parseAssert() (Step, error) {
	if _, err := p.expect(tokenAssert, "expected 'assert'"); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.match(tokenComma)
	msg, err := p.expect(tokenString, "expected assertion error message string")
	if err != nil {
		return nil, err
	}
	condRaw, err := marshalExpr(cond)
	if err != nil {
		return nil, err
	}
	return AssertStep{
		OpName: "assert",
		Cond:   condRaw,
		Error:  msg.lit,
	}, nil
}

func (p *textParser) parseBlock() ([]Step, error) {
	if _, err := p.expect(tokenLBrace, "expected '{'"); err != nil {
		return nil, err
	}
	steps, err := p.parseStepsUntil(tokenRBrace)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(tokenRBrace, "expected '}'"); err != nil {
		return nil, err
	}
	return steps, nil
}

func (p *textParser) parseExpr() (any, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	if p.current().kind != tokenPlus {
		return left, nil
	}
	args := []any{left}
	for p.match(tokenPlus) {
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		args = append(args, right)
	}
	return map[string]any{
		"$fn":  "concat",
		"args": args,
	}, nil
}

func (p *textParser) parsePrimary() (any, error) {
	tok := p.current()
	switch tok.kind {
	case tokenString:
		p.pos++
		return tok.lit, nil
	case tokenNumber:
		p.pos++
		n, err := strconv.ParseFloat(tok.lit, 64)
		if err != nil {
			return nil, p.errorf(tok, "invalid number: %s", tok.lit)
		}
		return n, nil
	case tokenTrue:
		p.pos++
		return true, nil
	case tokenFalse:
		p.pos++
		return false, nil
	case tokenNull:
		p.pos++
		return nil, nil
	case tokenIdent:
		return p.parsePathOrCall()
	case tokenLParen:
		p.pos++
		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tokenRParen, "expected ')'"); err != nil {
			return nil, err
		}
		return value, nil
	case tokenLBrace:
		return p.parseObject()
	case tokenLBracket:
		return p.parseArray()
	default:
		return nil, p.errorf(tok, "unexpected token in expression: %s", tok.lit)
	}
}

func (p *textParser) parsePathOrCall() (any, error) {
	base, err := p.expect(tokenIdent, "expected identifier")
	if err != nil {
		return nil, err
	}
	if p.match(tokenLParen) {
		args := make([]any, 0, 4)
		if p.current().kind != tokenRParen {
			for {
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.match(tokenComma) {
					continue
				}
				break
			}
		}
		if _, err := p.expect(tokenRParen, "expected ')' after function arguments"); err != nil {
			return nil, err
		}
		return map[string]any{
			"$fn":  base.lit,
			"args": args,
		}, nil
	}

	var b strings.Builder
	b.WriteString(base.lit)
	for {
		switch p.current().kind {
		case tokenDot:
			p.pos++
			part, err := p.expect(tokenIdent, "expected path segment after '.'")
			if err != nil {
				return nil, err
			}
			b.WriteByte('.')
			b.WriteString(part.lit)
		case tokenLBracket:
			p.pos++
			idxTok, err := p.expect(tokenNumber, "expected array index")
			if err != nil {
				return nil, err
			}
			if strings.Contains(idxTok.lit, ".") {
				return nil, p.errorf(idxTok, "array index must be integer")
			}
			if _, err := strconv.Atoi(idxTok.lit); err != nil {
				return nil, p.errorf(idxTok, "array index must be integer")
			}
			if _, err := p.expect(tokenRBracket, "expected ']'"); err != nil {
				return nil, err
			}
			b.WriteByte('[')
			b.WriteString(idxTok.lit)
			b.WriteByte(']')
		default:
			path := b.String()
			if base.lit != "input" && base.lit != "vars" && base.lit != "loop" {
				path = "vars." + path
			}
			return pathExpr(path), nil
		}
	}
}

func (p *textParser) parseObject() (any, error) {
	if _, err := p.expect(tokenLBrace, "expected '{'"); err != nil {
		return nil, err
	}
	obj := make(map[string]any)
	if p.current().kind == tokenRBrace {
		p.pos++
		return obj, nil
	}
	for {
		keyTok := p.current()
		var key string
		switch keyTok.kind {
		case tokenIdent, tokenString:
			key = keyTok.lit
			p.pos++
		default:
			return nil, p.errorf(keyTok, "expected object key")
		}
		if _, err := p.expect(tokenColon, "expected ':' after object key"); err != nil {
			return nil, err
		}
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		obj[key] = val
		if p.match(tokenComma) {
			if p.current().kind == tokenRBrace {
				break
			}
			continue
		}
		if p.current().kind == tokenRBrace {
			break
		}
		// Optional comma between object fields.
		if p.current().kind == tokenIdent || p.current().kind == tokenString {
			continue
		}
		return nil, p.errorf(p.current(), "expected ',' or '}' in object literal")
	}
	if _, err := p.expect(tokenRBrace, "expected '}'"); err != nil {
		return nil, err
	}
	return obj, nil
}

func (p *textParser) parseArray() (any, error) {
	if _, err := p.expect(tokenLBracket, "expected '['"); err != nil {
		return nil, err
	}
	arr := make([]any, 0, 8)
	if p.current().kind == tokenRBracket {
		p.pos++
		return arr, nil
	}
	for {
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		arr = append(arr, val)
		if p.match(tokenComma) {
			if p.current().kind == tokenRBracket {
				break
			}
			continue
		}
		if p.current().kind == tokenRBracket {
			break
		}
		return nil, p.errorf(p.current(), "expected ',' or ']' in array literal")
	}
	if _, err := p.expect(tokenRBracket, "expected ']'"); err != nil {
		return nil, err
	}
	return arr, nil
}

func (p *textParser) current() token {
	if p.pos >= len(p.tokens) {
		return token{kind: tokenEOF, line: 1, col: 1}
	}
	return p.tokens[p.pos]
}

func (p *textParser) match(kind tokenKind) bool {
	if p.current().kind != kind {
		return false
	}
	p.pos++
	return true
}

func (p *textParser) expect(kind tokenKind, msg string) (token, error) {
	tok := p.current()
	if tok.kind != kind {
		return token{}, p.errorf(tok, "%s", msg)
	}
	p.pos++
	return tok, nil
}

func (p *textParser) errorf(tok token, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("line %d:%d: %s", tok.line, tok.col, msg)
}

func marshalExpr(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encode expression: %w", err)
	}
	return json.RawMessage(b), nil
}

func pathExpr(path string) map[string]any {
	return map[string]any{"$": path}
}
