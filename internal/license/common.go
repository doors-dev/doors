package license

import (
	"crypto/ed25519"
	"fmt"

	"github.com/mr-tron/base58"
)

type License interface {
	GetType() Type
	GetCert() string
	GetId() string
	GetDomain() string
	GetTier() Tier
	GetIssuer() string
	Decode([]byte) error
	Print()
}

type Type uint8

const (
	TypeDomain Type = 1
)

func (t Type) Value() uint8 {
	return uint8(t)
}

func (t Type) Format() string {
	str := t.String()
	if str == "" {
		str = "unknown"
	}
	return fmt.Sprintf("%s (%d)", str, t)
}

func (t Type) String() string {
	switch t {
	case TypeDomain:
		return "domain"
	}
	return ""
}

type Tier uint8

const (
	TierStartup  Tier = 1
	TierBusiness Tier = 2
)

func (t Tier) Value() uint8 {
	return uint8(t)
}

func (t Tier) String() string {
	switch t {
	case TierStartup:
		return "Startup"
	case TierBusiness:
		return "Business"
	default:
		return ""
	}
}

func (t Tier) Format() string {
	str := t.String()
	if str == "" {
		str = "unknown"
	}
	return fmt.Sprintf("%s (%d)", str, t)
}

func DecodePrivateKey(secret string) (ed25519.PrivateKey, error) {
	secretBytes, err := base58.Decode(secret)
	if err != nil {
		return nil, err
	}
	return ed25519.NewKeyFromSeed(secretBytes), nil
}
