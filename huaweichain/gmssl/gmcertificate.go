/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// gmssl: Before compile the wnode, copy this file to the open source gmssl folder as a patch!
package gmssl

/*
#cgo CFLAGS: -I/usr/local/include/openssl
#cgo LDFLAGS: -L/usr/local/include/openssl  -lgmcryp

#include <stdlib.h>
#include "gmcert.h"
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"
)

var monthMap = map[string]time.Month{
	"Jan": 1,
	"Feb": 2,
	"Mar": 3,
	"Apr": 4,
	"May": 5,
	"Jun": 6,
	"Jul": 7,
	"Aug": 8,
	"Sep": 9,
	"Oct": 10,
	"Nov": 11,
	"Dec": 12}

// GetExpireTime get the certificate expire time
func (cert *Certificate) GetExpireTime() time.Time {
	var bufLen C.uint = 128
	var timeBuf [128]byte
	res := C.GetExipreTime(cert.x509, unsafe.Pointer(&timeBuf), bufLen)
	if res == 0 {
		return time.Now()
	}

	var month string
	var day int
	var hour int
	var min int
	var sec int
	var year int
	var timeZone string

	sscanf, err := fmt.Sscanf(string(timeBuf[:]), "%s %d %d:%d:%d %d %s", &month, &day, &hour, &min, &sec, &year, &timeZone)
	if err != nil {
		fmt.Printf("the sscanf is %d, err is %s\n", sscanf, err.Error())
		return time.Now()
	}
	mon, ok := monthMap[month]
	if !ok {
		fmt.Printf("get month failed\n")
		return time.Now()
	}

	location := time.FixedZone(timeZone, 0)
	return time.Date(year, mon, day, hour, min, sec, sec, location)
}

// GetFingerPrint get finger print of certificate
func (cert *Certificate) GetFingerPrint() string {
	var fingerPrintBuf [int(C.EVP_MAX_MD_SIZE)]byte
	var bufLen C.uint = C.EVP_MAX_MD_SIZE
	var fingerPrint string

	C.GetFingerPrintf(cert.x509, unsafe.Pointer(&fingerPrintBuf), (*C.uint)(&bufLen))

	length := int(bufLen)
	for i := 0; i < length-1; i++ {
		item := fmt.Sprintf("%02x", fingerPrintBuf[i])
		fingerPrint = fingerPrint + item
	}
	fingerPrint = fingerPrint + fmt.Sprintf("%02x", fingerPrintBuf[length-1])
	return fingerPrint
}

// VerifyCert verify cert with CA cert
func (cert *Certificate) VerifyCert(caCert *Certificate) bool {
	var res C.int
	res = C.VerifyCert(cert.x509, caCert.x509)

	if res == 1 {
		return true
	}

	return false
}

func freeCString(cStrs ...*C.char) {
	for _, cStr := range cStrs {
		C.free(unsafe.Pointer(cStr))
	}
}

// GenerateMiddleCACert ge middle ca cert
func GenerateMiddleCACert(country, province, locality, org, ou,
	commonName string, years uint, pubKey *PublicKey, caPriKey *PrivateKey, caCert *Certificate) (*Certificate, error) {
	cCountry := C.CString(country)
	cProvince := C.CString(province)
	cLocality := C.CString(locality)
	cOrg := C.CString(org)
	cOu := C.CString(ou)
	cCommonName := C.CString(commonName)

	defer freeCString(cCountry, cProvince, cLocality, cOrg, cOu, cCommonName)

	x509 := C.GenerateMiddleCACert(unsafe.Pointer(cCountry), unsafe.Pointer(cProvince), unsafe.Pointer(cLocality), unsafe.Pointer(cOrg), unsafe.Pointer(cOu), unsafe.Pointer(cCommonName), C.uint(years), pubKey.pkey, caPriKey.pkey, caCert.x509) ///nolint
	cert := &Certificate{}
	cert.x509 = x509
	return cert, nil
}

// GenerateCert generate self sign certificate
func GenerateCert(country, province, locality, org, ou,
	commonName string, years uint, pubKey *PublicKey, caPriKey *PrivateKey, caCert *Certificate) (*Certificate, error) {
	cCountry := C.CString(country)
	cProvince := C.CString(province)
	cLocality := C.CString(locality)
	cOrg := C.CString(org)
	cOu := C.CString(ou)
	cCommonName := C.CString(commonName)

	defer freeCString(cCountry, cProvince, cLocality, cOrg, cOu, cCommonName)

	x509 := C.GenerateCert(unsafe.Pointer(cCountry), unsafe.Pointer(cProvince), unsafe.Pointer(cLocality), unsafe.Pointer(cOrg), unsafe.Pointer(cOu), unsafe.Pointer(cCommonName), C.uint(years), pubKey.pkey, caPriKey.pkey, caCert.x509) ///nolint
	cert := &Certificate{}
	cert.x509 = x509
	return cert, nil
}

// GenerateSelfSignCert generate self sign certificate
func GenerateSelfSignCert(country, province, locality, org, ou,
	commonName string, years uint, key *PrivateKey) (*Certificate, error) {
	cCountry := C.CString(country)
	cProvince := C.CString(province)
	cLocality := C.CString(locality)
	cOrg := C.CString(org)
	cOu := C.CString(ou)
	cCommonName := C.CString(commonName)

	defer freeCString(cCountry, cProvince, cLocality, cOrg, cOu, cCommonName)

	sm2Key := key
	x509 := C.GenerateSelfSignCert(unsafe.Pointer(cCountry), unsafe.Pointer(cProvince), unsafe.Pointer(cLocality), unsafe.Pointer(cOrg), unsafe.Pointer(cOu), unsafe.Pointer(cCommonName), C.uint(years), sm2Key.pkey) ///nolint

	cert := &Certificate{
		x509: x509,
	}

	return cert, nil
}

// GetPemFromCert get pem from cert
func GetPemFromCert(certificate *Certificate) []byte {
	var pemBuf [1024]byte
	var pemBufLen C.uint = 1024
	C.X509_to_PEM(certificate.x509, unsafe.Pointer(&pemBuf), (*C.uint)(&pemBufLen))

	pemCert := pemBuf[:pemBufLen]
	return pemCert
}

// GetPemWithoutEnc get pem private key without encryption
func (sk *PrivateKey) GetPemWithoutEnc() (string, error) {
	bio := C.BIO_new(C.BIO_s_mem())
	if bio == nil {
		return "", GetErrors()
	}
	defer C.BIO_free(bio)

	if 1 != C.PEM_write_bio_PKCS8PrivateKey(bio, sk.pkey, nil,
		nil, C.int(0), nil, nil) {
		return "", GetErrors()
	}

	var p *C.char
	len := C.BIO_get_mem_data_func(bio, &p)
	if len <= 0 {
		return "", GetErrors()
	}

	return C.GoString(p)[:len], nil
}
