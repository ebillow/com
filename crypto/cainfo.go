package crypto

import (
	"crypto/x509"
	"io/ioutil"
	"server/com/log"
)

func LoadCA() *x509.CertPool {
	pool := x509.NewCertPool()
	if ca, err := ioutil.ReadFile("cacert.pem"); err != nil {
		log.Warnf("Read File cacert.pem error %v", err)
	} else {
		pool.AppendCertsFromPEM(ca)
	}
	return pool
}
