# Go interface for BIRD2

```go
package main

import (
	"log"

	"github.com/natesales/go-bird"
)

func main() {
	d, err := bird.New("/var/run/bird/bird.ctl")
	if err != nil {
		log.Fatal(err)
	}

	d.Read(nil)
	d.Write("show status")
	out, err := d.ReadString()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(out)

	protos, err := d.Protocols()
	if err != nil {
		log.Fatal(err)
	}
	for _, proto := range protos {
		log.Printf("%+v\n", proto)
	}
}
```

Based on https://github.com/xddxdd/bird-lg-go/
