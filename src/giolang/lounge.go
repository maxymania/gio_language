package giolang

import "fmt"
import "sync"

var GlobalObject Value = Nil{}

func RTFunc(f Future,m MsgObject) (Future,bool) {
	switch m.Method{
		case "if":
			switch len(m.Params){
			case 1: return Nil{},true
			case 2:
				if m.Params[0].Wait().Bool() {
					return m.Params[1].Wait().CodeBlock().Execute(f),true
				} else {
					return Nil{},true
				}
			case 3:
				if m.Params[0].Wait().Bool() {
					return m.Params[1].Wait().CodeBlock().Execute(f),true
				} else {
					return m.Params[2].Wait().CodeBlock().Execute(f),true
				}
			}
		case "for":
			switch len(m.Params){
			case 1: // for(;;) $0
				for {
					m.Params[0].Wait().CodeBlock().Execute(f)
				}
			case 2: // for(;$0;)$1
				for m.Params[0].Wait().CodeBlock().Execute(f).Wait().Bool(){
					m.Params[1].Wait().CodeBlock().Execute(f)
				}
			case 3: // for(;$0;$1)$2
				for m.Params[0].Wait().CodeBlock().Execute(f).Wait().Bool() {
					m.Params[2].Wait().CodeBlock().Execute(f)
					m.Params[1].Wait().CodeBlock().Execute(f)
				}
			case 4:  // for($0;$1;$2)$3
				m.Params[0].Wait().CodeBlock().Execute(f)
				for m.Params[1].Wait().CodeBlock().Execute(f).Wait().Bool() {
					m.Params[3].Wait().CodeBlock().Execute(f)
					m.Params[2].Wait().CodeBlock().Execute(f)
				}
			}
			return Nil{},true
		case "true": return Boolean(true),true
		case "false": return Boolean(false),true
		case "nil": return Nil{},true
		case "global": return GlobalObject,true
		case "method":
			return MakeDynMethod(m.Params,true),true
		// case "void","voidmethod":
		//	return MakeDynMethod(m.Params,false),true
		case "new_object":
			return NewObjectStruct(),true
		case "wait":
			if len(m.Params)==1{
				return m.Params[0].Wait(),true
			}
		case "eval":
			if len(m.Params)==1{
				return m.Params[0].Wait().CodeBlock().Execute(f),true
			}
		case "try_error":
			if len(m.Params)==1{
				cb := m.Params[0].Wait().CodeBlock()
				err := func()(err interface{}){
					defer func(){
						err = recover()
					}()
					err = nil
					cb.Execute(f)
					return
				}()
				if err==nil { return Nil{},true }
				return String(fmt.Sprint(err)),true
			}
		case "string":
			if len(m.Params)==1{
				return String(m.Params[0].Wait().String()),true
			}
	}
	return nil,false
}

type DynMethod struct{
	Args []string
	Returns bool
	Body *CodeBlock
}
func MakeDynMethod(v []Future,ret bool) *DynMethod {
	if len(v)==0 { panic("function/procedure without content is not allowed") }
	n := len(v)-1
	cb := v[n].Wait().CodeBlock()
	if cb==nil { panic("last parameter must be a CodeBlock") }
	args := make([]string,n)
	for i:=0;i<n;i++{
		args[i]=v[i].Wait().String()
	}
	return &DynMethod{
		Args: args,
		Returns: ret,
		Body: cb,
	}
}
func (d *DynMethod) Wait() Value {
	return d
}
func (d *DynMethod) CodeBlock() *CodeBlock {
	return nil
}
func (d *DynMethod) Bool() bool {
	return true
}
func (d *DynMethod) String() string {
	return fmt.Sprint("method",d.Args)
}
func (d *DynMethod) Bytes() []byte {
	return []byte(d.String())
}
func (d *DynMethod) Send(m MsgObject) Future {
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}

type ObjectStruct struct{
	Fields map[string]*Future
	Methods map[string]*DynMethod
	This Value
	Lock sync.Mutex
}
func NewObjectStruct() *ObjectStruct{
	o := new(ObjectStruct)
	o.This = o
	o.Fields = make(map[string]*Future)
	o.Methods = make(map[string]*DynMethod)
	return o
}
func (o *ObjectStruct) Clone() *ObjectStruct {
	d := NewObjectStruct()
	for n,m := range o.Methods {
		d.Methods[n] = m
	}
	for n,fp := range o.Fields {
		nfp := new(Future)
		*nfp = *fp
		d.Fields[n] = nfp
	}
	return d
}
func (o *ObjectStruct) CreateField(n string) *Future{
	o.Lock.Lock(); defer o.Lock.Unlock()
	fp,ok := o.Fields[n]
	if !ok {
		fp = new(Future)
		*fp = Nil{}
		o.Fields[n] = fp
	}
	return fp
}
func (o *ObjectStruct) SetMethod(n string,m *DynMethod) {
	o.Lock.Lock(); defer o.Lock.Unlock()
	o.Methods[n] = m
}
func (o *ObjectStruct) GetField(n string) (*Future,bool){
	o.Lock.Lock(); defer o.Lock.Unlock()
	fp,ok := o.Fields[n]
	return fp,ok
}
func (o *ObjectStruct) GetMethod(n string) (*DynMethod,bool){
	o.Lock.Lock(); defer o.Lock.Unlock()
	meth,ok := o.Methods[n]
	return meth,ok
}
func (o *ObjectStruct) Wait() Value {
	return o
}
func (o *ObjectStruct) CodeBlock() *CodeBlock {
	return nil
}
func (o *ObjectStruct) Bool() bool {
	return true
}
func (o *ObjectStruct) String() string {
	return fmt.Sprintf("Object_%p",o)
}
func (o *ObjectStruct) Bytes() []byte {
	return []byte(o.String())
}
func (o *ObjectStruct) Send(m MsgObject) Future {
	switch m.Method {
		case "clone":
			switch len(m.Params) {
			case 0: {
				aCopy := o.Clone()
				if _,ok := o.This.(*Actor) ; ok {
					actor := MakeActorDefault(aCopy)
					aCopy.This = actor
					return actor
				} else {
					return aCopy
				}
			}
			case 1: { // clone(true) -> actor, clone(false) -> no actor
				aCopy := o.Clone()
				if m.Params[0].Wait().Bool() {
					actor := MakeActorDefault(aCopy)
					aCopy.This = actor
					return actor
				} else {
					return aCopy
				}
			}
			}
		case "updateSlot","createSlot":
			if len(m.Params)==2 {
				n := m.Params[0].Wait().String()
				v := m.Params[1].Wait()
				asMeth,isMeth := v.(*DynMethod)
				if isMeth {
					o.SetMethod(n,asMeth)
				} else {
					fp := o.CreateField(n)
					*fp = v
				}
				return v
			}
	}
	meth,_ := o.GetMethod(m.Method)
	if meth!=nil {
		s := NewScope(o,o.This)
		n1,n2 := len(meth.Args),len(m.Params)
		for i:=0 ; i<n1 ; i++ {
			fp := new(Future)
			s.Local[meth.Args[i]] = fp
			if i<n2 {
				*fp = m.Params[i]
			} else {
				*fp = Nil{}
			}
		}
		return meth.Body.Execute(s)
	}
	if len(m.Params)==0 {
		fp,_ := o.GetField(m.Method)
		if fp!=nil {
			return *fp
		}
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}

type Scope struct{
	Base Value
	This Value
	Local map[string]*Future
}
func NewScope(b,t Value) *Scope {
	return &Scope{b,t,make(map[string]*Future)}
}
func (s *Scope) Wait() Value {
	return s
}
func (s *Scope) CodeBlock() *CodeBlock {
	return nil
}
func (s *Scope) Bool() bool {
	return true
}
func (s *Scope) String() string {
	return "L()"+s.Base.String()
}
func (s *Scope) Bytes() []byte {
	return []byte(s.String())
}
func (s *Scope) Send(m MsgObject) Future {
	rtr,ok := RTFunc(s,m)
	if ok { return rtr }
	switch len(m.Params) {
		case 0:
		if m.Method=="this" {
			return s.This
		}
		fp,_ := s.Local[m.Method]
		if fp!=nil {
			return *fp
		}
		case 2:
		switch m.Method{
			case "updateSlot": {
				fp,_ := s.Local[m.Params[0].Wait().String()]
				if fp!=nil {
					*fp = m.Params[1]
					return *fp
				}
			}
			case "createSlot": {
				fp := new(Future)
				s.Local[m.Params[0].Wait().String()] = fp
				*fp = m.Params[1]
				return *fp
			}
		}
	}
	return s.Base.Send(m)
}

