package controller

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math/big"
	"crypto/rand"
	"os"
	"time"
)

func Create(c *gin.Context)  {
	uuid := c.Param("uuid")

	rootCaPath := "../ca/meross_demo_ca.cert"
	caFile, err := ioutil.ReadFile(rootCaPath)
	if err != nil {
		c.JSON(200, gin.H{"code":1001, "message":"io error"})
		return
	}
	caBlock, _ := pem.Decode(caFile)
	if caBlock == nil {
		c.JSON(200, gin.H{"code":1002, "message":"ca file is wrong pem format"})
		return
	}
	rootCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		c.JSON(200, gin.H{"code":1003, "message":"fail to parse caBlock"})
		return
	}

	rootKeyPath := "../ca/meross_demo_ca.key"
	keyFile, err := ioutil.ReadFile(rootKeyPath)
	if err != nil {
		c.JSON(200, gin.H{"code":1001, "message":"io error"})
		return
	}
	keyBlock, _ := pem.Decode(keyFile)
	if keyBlock == nil {
		c.JSON(200, gin.H{"code":1004, "message":"ca key file is wrong pem format"})
		return
	}
	rootKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		c.JSON(200, gin.H{"code":1005, "message":"fail to parse ca key"})
		return
	}

	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNum, _ := rand.Int(rand.Reader, max)
	certTemplate := &x509.Certificate{
		SerialNumber: serialNum,
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"Chengdu Meross Technology Co., Ltd."},
			OrganizationalUnit: []string{"Iot Rd"},
			CommonName:         uuid,
		},
		NotBefore: time.Now(),
		NotAfter: time.Now().AddDate(30, 0, 0),
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA: false,
	}

	devicePrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		c.JSON(200, gin.H{"code":1006, "message":"fail to generate private key"})
		return
	}
	deviceCert, err := x509.CreateCertificate(rand.Reader,certTemplate,rootCert,&devicePrivKey.PublicKey,rootKey)
	if err != nil {
		c.JSON(200, gin.H{"code":1007, "message":"fail to create device certificate"})
		return
	}
	pemCert := &pem.Block{
		Type:    "CERTIFICATE",
		Bytes:   deviceCert,
	}
	deviceCert = pem.EncodeToMemory(pemCert)
	certOut, err := os.Create("../ca/" + uuid + ".cert")
	pem.Encode(certOut, pemCert)

	deviceKeyBuf := x509.MarshalPKCS1PrivateKey(devicePrivKey)
	pemDeviceKey := &pem.Block{
		Type:    "RSA PRIVATE KEY",
		Bytes:   deviceKeyBuf,
	}
	deviceKey := pem.EncodeToMemory(pemDeviceKey)
	keyOut, _ := os.Create("../ca/" + uuid + ".key")
	pem.Encode(keyOut, pemDeviceKey)
	fmt.Println("device certificate: ")
	fmt.Printf("%s ", deviceCert)
	fmt.Println("device private key: ")
	fmt.Printf("%s ", deviceKey)
}



