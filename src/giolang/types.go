package giolang

import "fmt"
import "strconv"
import "math"

type FutureImpl chan Value
func NewFutureImpl() FutureImpl{
	return make(FutureImpl,1)
}
func (fi FutureImpl) Wait() Value {
	v := <- fi
	fi <- v
	return v
}
func FutureOfFuture() (Future,chan <- Future) {
	fi := NewFutureImpl()
	ch := make(chan Future)
	go func(){
		fi <- (<- ch).Wait()
	}()
	return fi,ch
}

type Nil struct{}
func (i Nil) Wait() Value {
	return i
}
func (i Nil) CodeBlock() *CodeBlock {
	return nil
}
func (i Nil) Bool() bool {
	return false
}
func (i Nil) String() string {
	return "nil"
}
func (i Nil) Bytes() []byte {
	return []byte(i.String())
}
func (i Nil) Send(m MsgObject) Future {
	panic(fmt.Sprint("nil-pointer:",m.Method,"/",len(m.Params)))
}

type Integer int64
func (i Integer) Wait() Value {
	return i
}
func (i Integer) CodeBlock() *CodeBlock {
	return nil
}
func (i Integer) Bool() bool {
	return true
}
func (i Integer) String() string {
	return fmt.Sprint(int64(i))
}
func (i Integer) Bytes() []byte {
	return []byte(i.String())
}
func (i Integer) Send(m MsgObject) Future {
	switch len(m.Params) {
		case 0:
		switch m.Method{
			case "asInteger","int": return i
			case "asFloat","float": return Float(float64(int64(i)))
			case "asBool","bool": return Boolean(int64(i)!=0)
			case "negative","neg": return -i
		}
		case 1:
		switch m.Method{
			case "+": return i+(m.Params[0].Wait().(Integer))
			case "-": return i-(m.Params[0].Wait().(Integer))
			case "*": return i*(m.Params[0].Wait().(Integer))
			case "/": return i/(m.Params[0].Wait().(Integer))
			case "%": return i%(m.Params[0].Wait().(Integer))
			case "^": return i^(m.Params[0].Wait().(Integer))
			case "&": return i&(m.Params[0].Wait().(Integer))
			case "|": return i|(m.Params[0].Wait().(Integer))
			case "<": return Boolean(i<(m.Params[0].Wait().(Integer)))
			case ">": return Boolean(i>(m.Params[0].Wait().(Integer)))
			case "lshift": return i<<uint(m.Params[0].Wait().(Integer))
			case "rshift": return i>>uint(m.Params[0].Wait().(Integer))
		}
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}

type Boolean bool
func (i Boolean) Wait() Value {
	return i
}
func (i Boolean) CodeBlock() *CodeBlock {
	return nil
}
func (i Boolean) Bool() bool {
	return bool(i)
}
func (i Boolean) String() string {
	return fmt.Sprint(bool(i))
}
func (i Boolean) Bytes() []byte {
	return []byte(i.String())
}
func (b Boolean) Send(m MsgObject) Future {
	switch len(m.Params) {
		case 0:
		switch m.Method{
			case "asBool","bool": return b
		}
		case 1:
		o := (m.Params[0].Wait().(Boolean))
		switch m.Method{
			case "&","and": return b&&o
			case "|","or": return b||o
		}
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}

type Float float64
func (i Float) Wait() Value {
	return i
}
func (i Float) CodeBlock() *CodeBlock {
	return nil
}
func (i Float) Bool() bool {
	return true
}
func (i Float) String() string {
	return fmt.Sprint(float64(i))
}
func (i Float) Bytes() []byte {
	return []byte(i.String())
}
func (i Float) Send(m MsgObject) Future {
	switch len(m.Params) {
		case 0:
		switch m.Method{
			case "asInteger","int": return Integer(int64(float64(i)))
			case "asFloat","float": return i
			case "asBool","bool": return Integer(int64(float64(i))).Send(m)
			case "isNaN": return Boolean(math.IsNaN(float64(i)))
			case "isInf": return Boolean(math.IsInf(float64(i),0))
			case "sin": return Float(math.Sin(float64(i)))
			case "cos": return Float(math.Cos(float64(i)))
			case "tan": return Float(math.Tan(float64(i)))
			case "floor": return Float(math.Floor(float64(i)))
			case "ceil": return Float(math.Ceil(float64(i)))
			case "asin": return Float(math.Asin(float64(i)))
			case "acos": return Float(math.Acos(float64(i)))
			case "atan": return Float(math.Atan(float64(i)))
			case "sqrt": return Float(math.Sqrt(float64(i)))
			case "cbrt": return Float(math.Cbrt(float64(i)))
			case "sqpow": return Float(math.Pow(float64(i),2.0))
			case "cbpow": return Float(math.Pow(float64(i),3.0))
			case "log": return Float(math.Log(float64(i)))
			case "exp": return Float(math.Exp(float64(i)))
			case "negative","neg": return -i
		}
		case 1:
		switch m.Method{
			case "+": return i+(m.Params[0].Wait().(Float))
			case "-": return i-(m.Params[0].Wait().(Float))
			case "*": return i*(m.Params[0].Wait().(Float))
			case "/": return i/(m.Params[0].Wait().(Float))
			case "atan","atan2","%": return Float(math.Atan2(float64(i),float64(m.Params[0].Wait().(Float))))
			case "<": return Boolean(i<(m.Params[0].Wait().(Float)))
			case ">": return Boolean(i>(m.Params[0].Wait().(Float)))
		}
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}
type String string
func (i String) Wait() Value {
	return i
}
func (i String) CodeBlock() *CodeBlock {
	return nil
}
func (i String) Bool() bool {
	return true
}
func (i String) String() string {
	return string(i)
}
func (i String) Bytes() []byte {
	return []byte(i.String())
}
func (i String) Send(m MsgObject) Future {
	switch len(m.Params) {
		case 0:
		switch m.Method{
			case "asInteger","int":
				pi,_ := strconv.ParseInt(string(i),0,64)
				return Integer(pi)
			case "asFloat","float":
				pf,_ := strconv.ParseFloat(string(i),64)
				return Float(pf)
			case "asBool","bool": return Boolean(string(i)!="")
		}
		case 1:
		switch m.Method{
			case "+": return i+(m.Params[0].Wait().(String))
			case "<": return Boolean(i<(m.Params[0].Wait().(String)))
			case ">": return Boolean(i>(m.Params[0].Wait().(String)))
		}
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}
