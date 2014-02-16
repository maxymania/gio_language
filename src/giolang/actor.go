package giolang

import "fmt"
import "runtime"

type Logger interface{
	Error(s string)
}
type NoLogger struct{}
func (n NoLogger) Error(s string) {}

var ActorLogger Logger = NoLogger{}
var ActorDefaultQueueSize int = 1000

type actorMsg struct{
	Free bool
	Sender *Actor
	Msg MsgObject
	Answer chan <- Future
}
type Actor struct{
	base Value
	thread chan actorMsg
}
func MakeActorDefault(b Value) *Actor {
	return MakeActor(b,ActorDefaultQueueSize)
}
func MakeActor(b Value,n int) *Actor {
	am := make(chan actorMsg,n)
	a := new(Actor)
	a.base = b
	a.thread = am
	go actorLoop(am)
	runtime.SetFinalizer(a,FreeActor)
	return a
}
func actorLoop(am chan actorMsg) {
	for {
		m := <- am
		if m.Free { return }
		m.Answer<- m.Sender.base.Send(m.Msg)
		m.Sender = nil
	}
}
func FreeActor(a *Actor) {
	go freeActorLoop(a.thread)
}
func freeActorLoop(am chan actorMsg){
	am <- actorMsg{Free:true}
}
func (a *Actor) Wait() Value {
	return a
}
func (a *Actor) CodeBlock() *CodeBlock {
	return nil
}
func (a *Actor) Bool() bool {
	return true
}
func (a *Actor) String() string {
	return fmt.Sprintf("Actor_%p",a)
}
func (a *Actor) Bytes() []byte {
	return []byte(a.String())
}
func (a *Actor) Send(m MsgObject) Future {
	fu,res := FutureOfFuture()
	a.thread <- actorMsg{
		Free:false,
		Sender:a,
		Msg:m,
		Answer:res,
	}
	return fu
}

