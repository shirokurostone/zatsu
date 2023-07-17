package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"strings"
)

type Part struct {
	Header map[string][]string `json:"header"`
	Body   string              `json:"body,omitempty"`
	Parts  []Part              `json:"parts,omitempty"`
}

type Option struct {
	decodeTransferEncoding bool
}

func main() {
	option := Option{
		decodeTransferEncoding: false,
	}

	flag.BoolVar(&option.decodeTransferEncoding, "decode", false, "decode contents")
	flag.BoolVar(&option.decodeTransferEncoding, "d", false, "decode contents")
	flag.Parse()

	r := os.Stdin

	msg, err := ReadMail(r, option)
	if err != nil {
		log.Fatal(err)
	}

	v, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(v))
}

func ReadMail(r io.Reader, option Option) (*Part, error) {
	var part Part

	msg, err := mail.ReadMessage(r)
	if err != nil {
		return nil, err
	}

	part.Header = msg.Header
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if !strings.HasPrefix(mediaType, "multipart/") {
		encoding := msg.Header.Get("Content-Transfer-Encoding")
		body, err := ReadBody(msg.Body, mediaType, encoding, option)
		if err != nil {
			return nil, err
		}
		part.Body = body
		return &part, nil
	}

	var boundary string
	var ok bool

	if boundary, ok = params["boundary"]; !ok {
		return nil, fmt.Errorf("boundary not found")
	}

	if part.Parts, err = ReadMultiPart(msg.Body, boundary, option); err != nil {
		return nil, err
	}
	return &part, nil
}

func ReadMultiPart(r io.Reader, boundary string, option Option) ([]Part, error) {
	var ret []Part

	reader := multipart.NewReader(r, boundary)
	for {
		var part Part
		rawPart, err := reader.NextRawPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		part.Header = rawPart.Header

		mediaType, params, err := mime.ParseMediaType(rawPart.Header.Get("Content-Type"))
		if !strings.HasPrefix(mediaType, "multipart/") {

			encoding := rawPart.Header.Get("Content-Transfer-Encoding")
			body, err := ReadBody(rawPart, mediaType, encoding, option)
			if err != nil {
				return nil, err
			}
			part.Body = body
			ret = append(ret, part)
			continue
		}

		var subboundary string
		var ok bool

		if subboundary, ok = params["boundary"]; !ok {
			return nil, fmt.Errorf("boundary not found")
		}

		part.Parts, err = ReadMultiPart(rawPart, subboundary, option)
		if err != nil {
			return nil, err
		}
		ret = append(ret, part)
	}

	return ret, nil
}

func ReadBody(r io.Reader, mediaType string, contentTransferEncoding string, option Option) (string, error) {

	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	if !option.decodeTransferEncoding || !strings.HasPrefix(mediaType, "text/") {
		return string(b), nil
	}

	if contentTransferEncoding == "base64" {
		dst := make([]byte, base64.StdEncoding.DecodedLen(len(b)))
		n, err := base64.StdEncoding.Decode(dst, b)
		if err != nil {
			return "", err
		}
		dst = dst[:n]
		return string(dst), nil
	} else if contentTransferEncoding == "quoted-printable" {
		qr := quotedprintable.NewReader(bytes.NewReader(b))
		q, err := io.ReadAll(qr)
		if err != nil {
			return "", err
		}
		return string(q), nil
	} else {
		return string(b), nil
	}
}
