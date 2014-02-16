package giolang

import "fmt"

type Message struct{
	Method string
	Params []Expression
}
func (m Message)String() string{
	return fmt.Sprint(m.Method,m.Params)
}

type Expression interface{
	Execute(f Future) Future
}
type ExpressionL struct{
	V Value
	Trailer ExpressionC
}
func (e ExpressionL)Execute(f Future) Future{
	return e.Trailer.Execute(e.V)
}
func (e ExpressionL)String() string{
	return fmt.Sprint(e.V,e.Trailer)
}

type ExpressionC []Message
func (e ExpressionC)Execute(f Future) Future{
	a := f
	for _,m := range e {
		var mo MsgObject
		mo.Method = m.Method
		mo.Params = make([]Future,len(m.Params))
		for i,p := range m.Params {
			mo.Params[i]=p.Execute(f)
		}
		a = a.Wait().Send(mo)
	}
	return a
}

type CodeBlock struct{
	Code []Expression
}
func (b *CodeBlock)Execute(f Future) Future{
	a := f
	for _,e := range b.Code {
		a = e.Execute(f)
	}
	return a
}
func (b *CodeBlock)Wait() Value{
	return b
}
func (b *CodeBlock)CodeBlock() *CodeBlock{
	return b
}
func (b *CodeBlock)Bool() bool{
	return true
}
func (b *CodeBlock)String() string{
	return fmt.Sprint(b.Code)
}
func (b *CodeBlock)Bytes() []byte{
	return []byte(b.String())
}
func (b *CodeBlock)Send(m MsgObject) Future{
	panic("unsupported message")
}

type MsgObject struct{
	Method string
	Params []Future
}

type Future interface{
	Wait() Value
}
type Value interface{
	Future
	CodeBlock() *CodeBlock
	Bool() bool
	String() string
	Bytes() []byte
	Send(m MsgObject) Future
}
