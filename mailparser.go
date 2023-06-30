package mailparser

import (
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Header struct {
	Date      string `json:"Date"`       // 邮件的日期和时间
	From      string `json:"From"`       // 发件人的电子邮件地址
	To        string `json:"To"`         // 收件人的电子邮件地址
	Cc        string `json:"Cc"`         // 抄送(Carbon Copy)的收件人的电子邮件地址
	Bcc       string `json:"Bcc"`        // 密送(Blind Carbon Copy)的收件人的电子邮件地址
	Subject   string `json:"Subject"`    // 邮件的主题
	MessageID string `json:"Message-ID"` // 邮件的唯一标识符
}

type MailMessage struct {
	Header *Header

	Body string `json:"Body"`
}

func Parse(m *mail.Message) (*MailMessage, error) {
	header, err := ParseHeader(m)
	if err != nil {
		return nil, err
	}

	body, err := ParseBody(m)
	if err != nil {
		return nil, err
	}

	return &MailMessage{
		Header: header,
		Body:   body,
	}, nil
}

func ParseHeader(m *mail.Message) (*Header, error) {
	dec := new(mime.WordDecoder)
	dec.CharsetReader = charsetReader

	date, err := dec.DecodeHeader(m.Header.Get("Date"))
	if err != nil {
		return nil, err
	}

	from, err := dec.DecodeHeader(m.Header.Get("From"))
	if err != nil {
		return nil, err
	}

	to, err := dec.DecodeHeader(m.Header.Get("To"))
	if err != nil {
		return nil, err
	}

	cc, err := dec.DecodeHeader(m.Header.Get("Cc"))
	if err != nil {
		return nil, err
	}

	bcc, err := dec.DecodeHeader(m.Header.Get("Bcc"))
	if err != nil {
		return nil, err
	}

	subject, err := dec.DecodeHeader(m.Header.Get("Subject"))
	if err != nil {
		return nil, err
	}

	messageID, err := dec.DecodeHeader(m.Header.Get("Message-ID"))
	if err != nil {
		return nil, err
	}

	header := &Header{
		Date:      date,
		From:      from,
		To:        to,
		Cc:        cc,
		Bcc:       bcc,
		Subject:   subject,
		MessageID: messageID,
	}

	return header, nil
}

func ParseBody(m *mail.Message) (string, error) {
	mediaType, params, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	body := ""

	if strings.HasPrefix(mediaType, "multipart/") {
		content, err := parseMultipartBody(m.Body, params["boundary"])
		if err != nil {
			return "", err
		}

		body = strings.Join(content, "\n")
	} else if strings.HasPrefix(mediaType, "text/plain") {
		textBody, err := parseTextBody(m)
		if err != nil {
			return "", err
		}

		body = textBody
	}

	return body, nil
}

func parseTextBody(m *mail.Message) (string, error) {
	contentType := m.Header.Get("Content-Type")
	contentTransferEncoding := m.Header.Get("Content-Transfer-Encoding")

	bodyBytes, err := io.ReadAll(m.Body)
	if err != nil {
		return "", err
	}

	decodedBody, err := deTransferEncoding(contentTransferEncoding, bodyBytes)
	if err != nil {
		return "", err
	}

	body, err := decodeContent(contentType, decodedBody)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseMultipartBody(body io.Reader, boundary string) ([]string, error) {
	var content []string

	mr := multipart.NewReader(body, boundary)

	dec := new(mime.WordDecoder)
	dec.CharsetReader = charsetReader

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		contentType := part.Header.Get("Content-Type")
		contentTransferEncoding := part.Header.Get("Content-Transfer-Encoding")

		bodyPart, err := io.ReadAll(part)
		if err != nil {
			return nil, err
		}

		deTransferedBody, err := deTransferEncoding(contentTransferEncoding, bodyPart)
		if err != nil {
			return nil, err
		}

		decodedBody, err := decodeContent(contentType, deTransferedBody)
		if err != nil {
			return nil, err
		}

		content = append(content, string(decodedBody))

		if part.Header.Get("Content-Type") == "multipart/alternative" {
			break
		}
	}

	return content, nil
}

func deTransferEncoding(contentTransferEncoding string, body []byte) ([]byte, error) {
	switch contentTransferEncoding {
	case "base64":
		decodedBody, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			return nil, err
		}
		return decodedBody, nil
	case "quoted-printable":
		decodedBody, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(string(body))))
		if err != nil {
			return nil, err
		}
		return decodedBody, nil
	default:
		return body, nil
	}
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch charset {
	case "gb2312", "gb18030":
		decoder := simplifiedchinese.GB18030.NewDecoder()
		reader := transform.NewReader(input, decoder)
		return reader, nil
	default:
		return input, nil
	}
}

func decodeContent(contentType string, body []byte) ([]byte, error) {
	charset, err := getContentCharset(contentType)
	if err != nil {
		return nil, err
	}

	switch charset {
	case "gb2312", "gb18030":
		decoder := simplifiedchinese.GB18030.NewDecoder()
		decodedBody, err := io.ReadAll(transform.NewReader(strings.NewReader(string(body)), decoder))
		if err != nil {
			return nil, err
		}
		return decodedBody, nil
	default:
		return body, nil
	}
}

func getContentCharset(contentType string) (string, error) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}

	charset := params["charset"]
	if charset == "" {
		charset = "utf-8"
	}

	return charset, nil
}
