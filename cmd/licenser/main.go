package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/doors-dev/doors/internal/license"
	"github.com/mr-tron/base58"
)

const (
	commandKeyPair = "keypair"
	commanCreate   = "create"
	commandVerify  = "verify"
)

func keypair() {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	encodedPub := base58.Encode(pub)
	encodedPriv := base58.Encode(priv.Seed())
	fmt.Printf("PK:\t%s\nSK:\t%s\n", encodedPub, encodedPriv)
}

func verify() {
	fs := flag.NewFlagSet(commandVerify, flag.ExitOnError)
	certArg := fs.String("cert", "", "customer's cert")
	publicArg := fs.String("pk", "", "secret")
	fs.Parse(os.Args[2:])
	cert, err := license.ReadCert(*certArg)
	if err != nil {
		println("❌ FAILED READING")
		cert.Print()
		log.Fatal(err.Error())
		return
	}
	if *publicArg == "" {
		println("⚠️ Public key is not verified, provide -pk arg")
	} else {
		if *publicArg == cert.GetIssuer() {
			println("✅ OK")
		} else {
			println("☠️ FAILED VERIFICATION")
		}
	}
	println()
	cert.Print()
}

func create() {
	fs := flag.NewFlagSet(commanCreate, flag.ExitOnError)
	typeArg := fs.String("type", "", "cert type")
	domainArg := fs.String("domain", "", "customer's domain")
	tierArg := fs.String("tier", "", "customer's domain")
	secretArg := fs.String("sk", "", "secret")
	fs.Parse(os.Args[2:])
	if *typeArg == "" {
		log.Fatalf("must provide: -type")
	}
	certType, err := strconv.ParseUint(*typeArg, 10, 8)
	if err != nil {
		panic(errors.Join(errors.New("cert type parsing error"), err))
	}
	switch license.Type(uint8(certType)) {
	case license.TypeDomain:
		if *domainArg == "" || *secretArg == "" || *tierArg == "" {
			log.Fatalf("must provide for the domain cert: -domain, -tier and -sk")
		}
		tier, err := strconv.ParseUint(*tierArg, 10, 8)
		if err != nil {
			panic(errors.Join(errors.New("cert tier parsing error"), err))
		}
		certDomain := license.LicenseDomain{
			Domain: *domainArg,
			Tier:   license.Tier(uint8(tier)),
		}
		privateKey, err := license.DecodePrivateKey(*secretArg)
		if err != nil {
			panic(err)
		}
		err = certDomain.Encode(privateKey)
		if err != nil {
			panic(err)
		}
		certDomain.Print()
	default:
		log.Fatalf("uknown cert type %d", certType)
	}

}

func main() {
	println()
	defer println()
	if len(os.Args) < 2 {
		log.Fatalf("No command provided")
		return
	}
	cmd := os.Args[1]

	switch cmd {
	case commandKeyPair:
		keypair()
	case commanCreate:
		create()
	case commandVerify:
		verify()
	default:
		log.Fatalf("Usupported command")
	}
}
