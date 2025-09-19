package license

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/mr-tron/base58"
)

type LicenseDomain struct {
	Tier   Tier
	Domain string
	TS     time.Time
	Id     []byte
	Sign   []byte
	Issuer []byte
	Cert   []byte
}

func (c *LicenseDomain) GetCert() string {
	return base58.Encode(c.Cert)
}
func (c *LicenseDomain) GetType() Type {
	return TypeDomain
}

func (c *LicenseDomain) GetId() string {
	return base58.Encode(c.Id)
}

func (c *LicenseDomain) GetDomain() string {
	return c.Domain
}

func (c *LicenseDomain) Print() {
	fmt.Printf(
		"Type:\t%s\nTS:\t%s\nId:\t%s\nTier:\t%s\nDomain:\t%s\nSign:\t%s\nIssuer:\t%s\nCert:\t%s\n",
		c.GetType().Format(),
		c.TS.Format(time.RFC3339),
		c.GetId(),
		c.Tier.Format(),
		c.Domain,
		base58.Encode(c.Sign),
		c.GetIssuer(),
		c.GetCert(),
	)
}

func (c *LicenseDomain) GetIssuer() string {
	return base58.Encode(c.Issuer)
}

func (c *LicenseDomain) GetTier() Tier {
	return c.Tier
}

func (c *LicenseDomain) Decode(cert []byte) error {
	r := bytes.NewReader(cert)
	var certType uint8
	if err := binary.Read(r, binary.BigEndian, &certType); err != nil {
		return errors.Join(errors.New("failed to read type"), err)
	}
	if certType != TypeDomain.Value() {
		return fmt.Errorf("wrong cert type: expected %02X got %02X", TypeDomain, certType)
	}
	var ts int64
	if err := binary.Read(r, binary.BigEndian, &ts); err != nil {
		return errors.Join(errors.New("failed to read timestamp"), err)
	}
	c.TS = time.UnixMilli(ts)
	c.Id = make([]byte, 16)
	if _, err := io.ReadFull(r, c.Id); err != nil {
		return errors.Join(errors.New("failed to read kicense key"), err)
	}
	if err := binary.Read(r, binary.BigEndian, &c.Tier); err != nil {
		return errors.Join(errors.New("failed to read tier"), err)
	}
	if c.Tier.String() == "" {
		return fmt.Errorf("unknown tier: %d", c.Tier.Value())
	}
	var l uint8
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return errors.Join(errors.New("failed to read domain length"), err)
	}
	domain := make([]byte, l)
	if _, err := io.ReadFull(r, domain); err != nil {
		return errors.Join(errors.New("failed to read domain"), err)
	}
	c.Domain = string(domain)
	c.Issuer = make([]byte, 32)
	if _, err := io.ReadFull(r, c.Issuer); err != nil {
		return errors.Join(errors.New("failed to read public key"), err)
	}
	c.Sign = make([]byte, 64)
	if _, err := io.ReadFull(r, c.Sign); err != nil {
		return errors.Join(errors.New("failed to read signature"), err)
	}
	if r.Len() != 0 {
		return fmt.Errorf("unexpected %d trailing bytes", r.Len())
	}
	payloadLen := 1 + 1 + 8 + 16 + 1 + int(l)
	body := cert[:payloadLen]
	if !ed25519.Verify(c.Issuer, body, c.Sign) {
		return errors.New("signature verification failed")
	}
	c.Cert = cert
	return nil
}

func (c *LicenseDomain) Encode(secret ed25519.PrivateKey) error {
	if c.Tier.String() == "" {
		return fmt.Errorf("unknown tier: %d", c.Tier.Value())
	}
	if c.TS.IsZero() {
		c.TS = time.Now()
	}
	if len(c.Id) == 0 {
		c.Id = make([]byte, 16)
		_, err := rand.Read(c.Id)
		if err != nil {
			log.Fatalf("failed to generate random bytes: %v", err)
		}
	}
	buf := &bytes.Buffer{}
	err := c.body(buf)
	if err != nil {
		return err
	}
	c.Sign = ed25519.Sign(secret, buf.Bytes())
	c.Issuer = secret.Public().(ed25519.PublicKey)
	_, err = buf.Write(c.Issuer)
	if err != nil {
		return err
	}
	_, err = buf.Write(c.Sign)
	if err != nil {
		return err
	}
	c.Cert = buf.Bytes()
	return nil
}

func (c *LicenseDomain) body(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, TypeDomain)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, c.TS.UnixMilli())
	if err != nil {
		return err
	}
	_, err = w.Write(c.Id)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, c.Tier)
	if err != nil {
		return err
	}
	length := len(c.Domain)
	if length > 255 {
		return errors.New("domain is too long")
	}
	err = binary.Write(w, binary.BigEndian, uint8(length))
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, c.Domain)
	return err
}
