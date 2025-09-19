package license

import (
	"errors"

	"github.com/mr-tron/base58"
)


func ReadCert(str string) (License, error) {
	bytes, err := base58.Decode(str)
	if err != nil {
		return nil, err
	}
	certType := bytes[0]
	var cert License
	switch Type(certType) {
	case TypeDomain:
		cert = &LicenseDomain{}
	default:
		return nil, errors.New("unsupported cert type")
	}
	err = cert.Decode(bytes)
	return cert, err
}
