# go-mailparser

Go lib for parsing email in simple way.

## Usage Examle

```go
package main

import (
	"net/mail"

	"github.com/knadh/go-pop3"
  "github.com/windvalley/go-mailparser"
)

func main() {
	p := pop3.New(pop3.Opt{
		Host:       "mail.xxx.com",
		Port:       110,
		TLSEnabled: false,
	})

	c, err := p.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := c.Quit(); err != nil {
			log.Print(err)
		}
	}()

	if err := c.Auth("xxx@xxx.com", "yourpassword"); err != nil {
    log.Fatal(err)
	}

	msgs, err := c.List(0)
	if err != nil {
		return
	}

	for _, v := range msgs {
		b, err := c.Cmd("RETR", true, v.ID)
		if err != nil {
			fmt.Println(err)
			continue
		}

		m, err := mail.ReadMessage(b)
		if err != nil {
			fmt.Println(err)
			continue
		}

		mailMessage, err := mailparser.Parse(m)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("%+v\n", mailMessage)
	}
}
```
