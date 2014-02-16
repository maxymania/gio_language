package iolib

import "fmt"
import "net"
import . "giolang"

type Net struct{}
func (i Net) Wait() Value {
	return i
}
func (i Net) CodeBlock() *CodeBlock {
	return nil
}
func (i Net) Bool() bool {
	return true
}
func (i Net) String() string {
	return "Package_net"
}
func (i Net) Bytes() []byte {
	return []byte(i.String())
}
func (i Net) Send(m MsgObject) Future {
	switch m.Method {
		case "listen":
			if len(m.Params)==2 {
				n := m.Params[0].Wait().String()
				a := m.Params[1].Wait().String()
				l,e := net.Listen(n,a)
				if e!=nil { return Nil{} }
				return &NetListener{Listener:l}
			}
		case "dial":
			if len(m.Params)==2 {
				n := m.Params[0].Wait().String()
				a := m.Params[1].Wait().String()
				c,e := net.Dial(n,a)
				if e!=nil { return Nil{} }
				return RW(c)
			}
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}

type NetListener struct{
	Listener net.Listener
	Error error
}
func (i *NetListener) Wait() Value {
	return i
}
func (i *NetListener) CodeBlock() *CodeBlock {
	return nil
}
func (i *NetListener) Bool() bool {
	return true
}
func (i *NetListener) String() string {
	return fmt.Sprintf("Listener_%p",i)
}
func (i *NetListener) Bytes() []byte {
	return []byte(i.String())
}
func (i *NetListener) Send(m MsgObject) Future {
	switch m.Method {
		case "accept":{
			c,e := i.Listener.Accept()
			i.Error = e
			if e!=nil {
				return Nil{}
			}
			return RW(c)
		}
		case "close":{
			i.Error = i.Listener.Close()
			return Nil{}
		}
		case "addr","address":
			return String(fmt.Sprint(i.Listener.Addr()))
	}
	panic(fmt.Sprint("unsupported method:",m.Method,"/",len(m.Params)))
}


