#ifndef GM_CERTIFICATE_H
#define GM_CERTIFICATE_H

//the included header files
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/uio.h>
#include <unistd.h>
#include <openssl/bio.h>
#include <openssl/evp.h>
#include <openssl/pem.h>
#include <openssl/x509v3.h>
#include <openssl/ossl_typ.h>
#include <openssl/err.h>

long BIO_get_mem_data_func(BIO *b, char **pp);
void GetFingerPrintf(X509* x, void* fingerPrintBuf, uint* bufLen);
int VerifyCert(X509 *cert, X509 *cacert);
X509 * GenerateCertBasic(void* country, void* province, void* locality, void* org, void* ou, void* commonName, uint years, EVP_PKEY *pubKey);
X509 * GenerateMiddleCACert(void* country, void* province, void* locality, void* org, void* ou, void* commonName, uint years, EVP_PKEY *pubKey, EVP_PKEY *caPriKey, X509 * caCert);
X509 * GenerateCert(void* country, void* province, void* locality, void* org, void* ou, void* commonName, uint years, EVP_PKEY *pubKey, EVP_PKEY *caPriKey, X509 * caCert);
X509 * GenerateSelfSignCert(void* country, void* province, void* locality, void* org, void* ou, void* commonName, uint years, EVP_PKEY *pkey);
void X509_to_PEM(X509 *cert, void* pemBuf, uint* pemBufLen);
int GetExipreTime(X509* x, void* timeBuf, uint bufLen);


#endif
