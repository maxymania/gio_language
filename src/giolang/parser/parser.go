package parser

import "giolang"
import "text/scanner"
import "strconv"
import "io"

const (
	PO = '('
	PC = ')'
	BO = '{'
	BC = '}'
)

type Parser struct{
	sc scanner.Scanner
	ispecial bool
	tok rune
	text string
	f float64
	i int64
}
func (p *Parser) next(){
	p.tok=p.sc.Scan()
	p.text=p.sc.TokenText()
	p.ispecial = false
	switch(p.tok){
	case '+','-','!','$','%','&','*','/','|','#','~','<','>','^':
		p.ispecial = true
		p.tok = scanner.Ident
	case scanner.RawString,scanner.String:
		p.tok = scanner.String
		p.text,_ = strconv.Unquote(p.text)
	case scanner.Float:
		p.f,_ = strconv.ParseFloat(p.text,64)
	case scanner.Int:
		p.i,_ = strconv.ParseInt(p.text,0,64)
	}
}
func (p *Parser) ParamList() []giolang.Expression{
	l := []giolang.Expression{}
	if p.tok==PO {
		p.next()
		if p.tok==PC { p.next(); return l }
		l=append(l,p.Expression(true))
		for{
			if p.tok==PC { p.next(); return l }
			if p.tok!=',' { panic("required , got "+p.text) }
			p.next()
			l=append(l,p.Expression(true))
		}
	}
	return l
}
func (p *Parser) Expression(trailer bool) giolang.Expression{
	switch p.tok{
		case scanner.Ident:
			return p.ExpressionC()
		case PO:
			p.next()
			e := p.Expression(true)
			if p.tok!=PC { panic("required PC got "+p.text) }
			p.next()
			return e
		case scanner.String,scanner.Float,scanner.Int,BO:
		{
			l := new(giolang.ExpressionL)
			switch p.tok{
			case BO:
				l.V = &giolang.CodeBlock{p.Block()}
			case scanner.String:
				l.V = giolang.String(p.text)
				p.next()
			case scanner.Float:
				l.V = giolang.Float(p.f)
				p.next()
			case scanner.Int:
				l.V = giolang.Integer(p.i)
				p.next()
			}
			if trailer {
				l.Trailer = p.ExpressionC()
			} else {
				l.Trailer = giolang.ExpressionC{}
			}
			return l
		}
	}
	panic("required ident, PO or BO got "+p.text)
}
func (p *Parser) ExpressionC() giolang.ExpressionC{
	ec := giolang.ExpressionC{}
	for p.tok == scanner.Ident {
		var m giolang.Message
		m.Method = p.text
		spec := p.ispecial
		p.next()
		if spec && (p.tok!=PO) {
			m.Params = []giolang.Expression{p.Expression(false)}
		} else {
			switch p.tok {
				case ':': {
					p.next()
					if p.tok!='=' { panic("required = got "+p.text) }
					p.next()
					s := m.Method
					m.Method = "createSlot"
					m.Params = []giolang.Expression{
						giolang.ExpressionL{giolang.String(s),giolang.ExpressionC{}},
						p.Expression(true)}
				}
				case '=': {
					p.next()
					s := m.Method
					m.Method = "updateSlot"
					m.Params = []giolang.Expression{
						giolang.ExpressionL{giolang.String(s),giolang.ExpressionC{}},
						p.Expression(true)}
				}
				default:
				m.Params = p.ParamList()
			}
		}
		ec=append(ec,m)
	}
	return ec
}
func (p *Parser) Block() []giolang.Expression{
	l := []giolang.Expression{}
	if p.tok==BO {
		p.next()
		if p.tok==BC { p.next(); return l }
		l=append(l,p.Expression(true))
		for{
			if p.tok==BC { p.next(); return l }
			if p.tok!=';' { panic("required ; got "+p.text) }
			p.next()
			l=append(l,p.Expression(true))
		}
	}
	return l
}
func (p *Parser) Src() []giolang.Expression{
	l := []giolang.Expression{}
	if p.tok!=scanner.EOF {
		l=append(l,p.Expression(true))
		for{
			if p.tok==scanner.EOF { return l }
			if p.tok!=';' { panic("required ; got "+p.text) }
			p.next()
			if p.tok==scanner.EOF { return l }
			l=append(l,p.Expression(true))
		}
	}
	return l
}
func (p *Parser) ParseSrc(r io.Reader) *giolang.CodeBlock{
	p.sc.Init(r)
	p.next()
	return &giolang.CodeBlock{p.Src()}
}
func ParseSrc(r io.Reader) *giolang.CodeBlock{
	var p Parser
	return p.ParseSrc(r)
}

