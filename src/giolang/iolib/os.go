package iolib

import "fmt"
import "io"
import "os"
import "os/exec"
import . "giolang"

type Os struct{}
func (i Os) Wait() Value {
	return i
}
func (i Os) CodeBlock() *CodeBlock {
	return nil
}
func (i Os) Bool() bool {
	return true
}
func (i Os) String() string {
	return "Package_os"
}
func (i Os) Bytes() []byte {
	return []byte(i.String())
}
func (i Os) Send(m MsgObject) Future {
	switch m.Method {
		case "open":
			if len(m.Params)==1 {
				f,e := os.Open(m.Params[0].Wait().String())
				if e!=nil { return Nil{} }
				return CreateInputSrc(f,true)
			}
		case "create":
			if len(m.Params)==1 {
				f,e := os.Create(m.Params[0].Wait().String())
				if e!=nil { return Nil{} }
				return CreateOutputDest(f,true)
			}
		case "execute":
			if len(m.Params)>0 {
				s := make([]string,len(m.Params))
				for i,p := range m.Params {
					s[i] = p.Wait().String()
				}
				cmd := exec.Command(s[0],s[1:]...)
				pout,e := cmd.StdoutPipe()
				if e!=nil { return Nil{} }
				pin,e := cmd.StdinPipe()
				if e!=nil { return Nil{} }
				return sioRaW(pout,pin)
			}
		case "mkfifo":
			{
				r,w := io.Pipe()
				return RaW(r,w)
			}
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}

func sioRaW(r io.Reader, w io.Writer) Value {
	obj := NewObjectStruct()
	*(obj.CreateField("in")) = CreateInputSrc(r,true)
	*(obj.CreateField("out")) = CreateOutputDest(w,true)
	return obj
}

