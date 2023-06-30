# go-mailparser

Go lib for parsing email in simple way.

## Features

- Support parsing emails with content types of `text/*` and `multipart/*`.
- Support parsing Chinese content, such as Chinese characters in email address aliases, email subject, and email content.
- Support parsing email attachments.
- Support parsing emails with content encoded in base64.
- Support parsing email headers and email content separately, or parse them all at once.

## Examle

```go
package main

import (
	"fmt"
	"log"
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
			fmt.Println(err)
		}
	}()

	if err := c.Auth("xxx@xxx.com", "yourpassword"); err != nil {
		log.Fatal(err)
	}

	msgs, err := c.List(0)
	if err != nil {
		log.Fatal(err)
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

		res, err := mailparser.Parse(m)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// check MailMessage
		fmt.Printf("result: %+v\n", res)

		// check attachments
		for _, v := range res.Attachments {
			// You can handle the file data (v.Data) appropriately based on the content type.
			fmt.Printf("filename: %s, content-type: %s\n", v.Filename, v.ContentType)
		}
	}
}
```

## License

This project is under the MIT License.
See the [LICENSE](LICENSE) file for the full license text.
