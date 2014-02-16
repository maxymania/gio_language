package iolib

import "io"
import "bufio"
import "fmt"
import . "giolang"

type EClose struct{}
func (e EClose) Close() error { return nil }

type InputSrc struct{
	Buffer *bufio.Reader
	Closer io.Closer
	DefaultBufSize int
	Error error
}
func CreateInputSrc(r io.Reader,doclose bool) *InputSrc {
	return CreateInputSrc2(r,1024,doclose)
}
func CreateInputSrc2(r io.Reader, dbs int, doclose bool) *InputSrc {
	var c io.Closer
	if doclose {
		if c2,ok := r.(io.Closer) ; ok {
			c=c2
		} else {
			c = EClose{}
		}
	} else {
		c = EClose{}
	}
	b := bufio.NewReader(r)
	return &InputSrc{Buffer:b, Closer:c, DefaultBufSize:dbs }
}

func (i *InputSrc) Wait() Value {
	return i
}
func (i *InputSrc) CodeBlock() *CodeBlock {
	return nil
}
func (i *InputSrc) Bool() bool {
	return true
}
func (i *InputSrc) String() string {
	return fmt.Sprintf("Reader_%p",i)
}
func (i *InputSrc) Bytes() []byte {
	return []byte(i.String())
}
func (i *InputSrc) Send(m MsgObject) Future {
	switch m.Method {
		case "read": {
			nread := i.DefaultBufSize
			if len(m.Params)>0 {
				mp0i,ok := m.Params[0].Wait().(Integer)
				if ok {
					nread = int(mp0i)
				}
			}
			buf := make([]byte,nread)
			nread,i.Error = i.Buffer.Read(buf)
			if nread>0 { return String(string(buf[:nread])) }
			return String("")
		}
		case "error": {
			e := i.Error
			if e==nil { return Nil{} }
			return String(fmt.Sprint(i.Error))
		}
		case "line","readLine","readline": {
			line,pre,e := i.Buffer.ReadLine()
			res := string(line)
			for (pre && (e==nil)) {
				line,pre,e = i.Buffer.ReadLine()
				res += string(line)
			}
			i.Error = e
			return String(res)
		}
		case "eof":
			return Boolean(i.Error == io.EOF)
		case "close":
			i.Error = i.Closer.Close()
			return Nil{}
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}

type OutputDest struct{
	Buffer *bufio.Writer
	Closer io.Closer
	Error error
}
func CreateOutputDest(w io.Writer,doclose bool) *OutputDest {
	var c io.Closer
	if doclose {
		if c2,ok := w.(io.Closer) ; ok {
			c=c2
		} else {
			c = EClose{}
		}
	} else {
		c = EClose{}
	}
	b := bufio.NewWriter(w)
	return &OutputDest{Buffer:b, Closer:c }
}

func (i *OutputDest) Wait() Value {
	return i
}
func (i *OutputDest) CodeBlock() *CodeBlock {
	return nil
}
func (i *OutputDest) Bool() bool {
	return true
}
func (i *OutputDest) String() string {
	return fmt.Sprintf("Writer_%p",i)
}
func (i *OutputDest) Bytes() []byte {
	return []byte(i.String())
}
func (i *OutputDest) Send(m MsgObject) Future {
	switch m.Method {
		case "write": {
			n := 0
			for _,p := range m.Params {
				w,e := i.Buffer.WriteString(p.Wait().String())
				if w>0 { n+=w }
				if e!=nil { break }
			}
			i.Error = i.Buffer.Flush()
			return Integer(n)
		}
		case "error": {
			e := i.Error
			if e==nil { return Nil{} }
			return String(fmt.Sprint(i.Error))
		}
		case "close":
			i.Error = i.Closer.Close()
			return Nil{}
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}

func RW(rw io.ReadWriter) Value {
	obj := NewObjectStruct()
	*(obj.CreateField("in")) = CreateInputSrc(rw,false)
	*(obj.CreateField("out")) = CreateOutputDest(rw,true)
	return obj
}
func RaW(r io.Reader, w io.Writer) Value {
	obj := NewObjectStruct()
	*(obj.CreateField("in")) = CreateInputSrc(r,false)
	*(obj.CreateField("out")) = CreateOutputDest(w,true)
	return obj
}


