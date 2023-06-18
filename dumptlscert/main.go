package main

import (
	"crypto/tls"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
)

func main() {
	var noVerify bool
	var subject bool
	flag.BoolVar(&noVerify, "noverify", false, "not verify server certificates")
	flag.BoolVar(&subject, "subject", false, "print subject")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s host:port\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(0)
	}

	conn, err := tls.Dial("tcp", flag.Arg(0), &tls.Config{
		InsecureSkipVerify: noVerify,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	for _, c := range certs {
		if subject {
			fmt.Fprintf(os.Stdout, "%s\n", c.Subject)
		} else {
			pem.Encode(os.Stdout, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: c.Raw,
			})
		}
	}
}
