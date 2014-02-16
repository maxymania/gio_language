gio_language
============

An interpreted programming language inspired from IO and implemented in Go

Example of gio
==============

```
obj := new_object;
obj handle = method("conn",{
	conn out write("Your name?\r\n");
	l := conn in readLine;
	conn out write("Thank you ",l,"!\r\n");
	conn in readLine;
	conn out close;
	nil
});
obj std := std;
listener := net listen("tcp",":9999");
c := nil;
for({
	c = listener accept;
	if(c,{
		obj clone(true) handle(c)
	},{
		nil
	})
});
nil
```

How To use the interpreter
==========================

```go
package main

import "fmt"
import "os"
import "strings"

// we need the basic giolang package and the giolang parser
import "giolang"
import "giolang/parser"

// we will also use the Input-Output-Lib
import "giolang/iolib"

const src = `
obj := new_object;
obj handle = method("conn",{
	conn out write("Your name?\r\n");
	l := conn in readLine;
	conn out write("Thank you ",l,"!\r\n");
	conn in readLine;
	conn out close;
	nil
});
obj std := std;
listener := net listen("tcp",":9999");
c := nil;
for({
	c = listener accept;
	if(c,{
		obj clone(true) handle(c)
	},{
		nil
	})
});
nil
`

func main() {
	// create an object (of globals) where the scope will be based on
	obj := giolang.NewObjectStruct()
	
	// initialize some global things ...
	*(obj.CreateField("std")) = iolib.RaW(os.Stdin,os.Stdout) // stdin and stdout
	*(obj.CreateField("os")) = iolib.Os{} // os library for opening files
	*(obj.CreateField("net")) = iolib.Net{} // network library
	
	//create the scope. This is mandatory.
	scope := giolang.NewScope(obj,obj)
	
	// parse the source code
	r := parser.ParseSrc(strings.NewReader(src))
	
	// run the parsed code
	res := r.Execute(scope)
	res.Wait()
	fmt.Println()
	fmt.Println("EOP")
}
```

