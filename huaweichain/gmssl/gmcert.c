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
#include "gmcert.h"

extern long _BIO_get_mem_data(BIO *b, char **pp);
extern void _OPENSSL_free(void *addr);

int GetExipreTime(X509* x, void* timeBuf, uint bufLen){
    char* buf = (char*)timeBuf;
    ASN1_TIME* time = X509_get_notAfter(x);
    BIO *mem = BIO_new(BIO_s_mem());
    ASN1_TIME_print(mem, time);
    BUF_MEM *bptr;
    BIO_get_mem_ptr(mem, &bptr);
    if (bufLen < bptr->length) {
        return 0;
    }
    memcpy(timeBuf, bptr->data, bptr->length-1);
    buf[bptr->length-1] = 0;
    BIO_free(mem);
    return 1;
}

long BIO_get_mem_data_func(BIO *b, char **pp) {
	return BIO_get_mem_data(b, pp);
}


void GetFingerPrintf(X509* x, void* fingerPrintBuf, uint* bufLen)
{
    unsigned char         md[EVP_MAX_MD_SIZE] = {0};
    unsigned int          n = 0;
    const EVP_MD          *digest;
    int                   pos;

    digest = EVP_get_digestbyname("sha256");
    X509_digest(x, digest, md, &n);

    if (*bufLen < n) {
        printf("the digest length:%d is not correct\n",n);
        return;
    }
    memcpy(fingerPrintBuf, md, n);

    *bufLen = n;
}

int VerifyCert(X509 *cert, X509 *cacert)
{
    int ret;
    int error = 0;
    X509_STORE *store;
    X509_STORE_CTX *ctx;

    store = X509_STORE_new();
    X509_STORE_add_cert(store, cacert);

    ctx = X509_STORE_CTX_new();
    X509_STORE_CTX_init(ctx, store, cert, NULL);

    ret = X509_verify_cert(ctx);

    X509_STORE_CTX_free(ctx);
    X509_STORE_free(store);

    return ret;
}


static int set_extension(X509 * issuer, X509 * cert, int nid, char * value)
{
    X509_EXTENSION * ext ;
    X509V3_CTX ctx ;
    unsigned long err;

    X509V3_set_ctx(&ctx, issuer, cert, NULL, NULL, 0);
    ext = X509V3_EXT_conf_nid(NULL, &ctx, nid, value);
    if (!ext) {
        err = ERR_get_error();
        printf("X509V3_EXT_conf_nid set extension failed, err: 0x%lx\n",err);
        return -1;
    }

    X509_add_ext(cert, ext, -1);
    X509_EXTENSION_free(ext);
    return 0 ;
}


X509 * GenerateCertBasic(void* country, void* province, void* locality, void* org, void* ou, void* commonName, uint years, EVP_PKEY *pubKey){
    X509 * x509;
    BIGNUM *btmp = NULL;
	ASN1_INTEGER *sno;
    unsigned long err;
    int ret = 0;

	x509 = X509_new();
    if (x509 == NULL) {
        err = ERR_get_error();
        printf("X509_new failed, err: 0x%lx\n",err);
        return NULL;
    }

    int svRet = X509_set_version(x509, 2);
    if (svRet != 1) {
        err = ERR_get_error();
        printf("X509_set_version failed, err: 0x%lx\n",err);
        ret = -1;
        goto END;
    }

    sno = ASN1_INTEGER_new();
	if (!sno) {
        ret = -1;
		goto END;
	}
	btmp = BN_new();
	if (!btmp) {
        ret = -1;
		goto END;
	}
	if (!BN_rand(btmp, 128, 0, 0)) {
        ret = -1;
		goto END;
	}
	if (!BN_to_ASN1_INTEGER(btmp, sno)) {
        ret = -1;
		goto END;
	}
	if (!X509_set_serialNumber(x509, sno)) {
        ret = -1;
		goto END;
	}

	X509_gmtime_adj(X509_get_notBefore(x509), 0);
	X509_gmtime_adj(X509_get_notAfter(x509), 31536000*years);

    X509_set_pubkey(x509, pubKey);

	X509_NAME * name;
	name = X509_get_subject_name(x509);

	X509_NAME_add_entry_by_txt(name, "C",  MBSTRING_ASC,
						   (unsigned char *)country, -1, -1, 0);
	X509_NAME_add_entry_by_txt(name, "ST", MBSTRING_ASC,
						   (unsigned char *)province, -1, -1, 0);
	X509_NAME_add_entry_by_txt(name, "L", MBSTRING_ASC,
						   (unsigned char *)locality, -1, -1, 0);
	X509_NAME_add_entry_by_txt(name, "O",  MBSTRING_ASC,
						   (unsigned char *)org, -1, -1, 0);
	X509_NAME_add_entry_by_txt(name, "OU",  MBSTRING_ASC,
						   (unsigned char *)ou, -1, -1, 0);
	X509_NAME_add_entry_by_txt(name, "CN", MBSTRING_ASC,
						   (unsigned char *)commonName, -1, -1, 0);

END:
    if (sno != NULL){
        ASN1_INTEGER_free(sno);
    }
    if (btmp != NULL){
        BN_free(btmp);
    }
    if (ret != 0){
		if (x509 != NULL){
			X509_free(x509);
			x509 = NULL;
		}
    }
    return x509;
}

X509 * GenerateMiddleCACert(void* country, void* province, void* locality, void* org, void* ou, void* commonName, uint years, EVP_PKEY *pubKey, EVP_PKEY *caPriKey, X509 * caCert){
    int ret = 0;
    X509* x509 = GenerateCertBasic(country, province, locality, org, ou, commonName, years, pubKey);
    if (x509 == NULL){
        return NULL;
    }

    X509_NAME * caName;
    caName = X509_get_subject_name(caCert);
    X509_set_issuer_name(x509, caName);

    // todo judge return value err process
    ret = set_extension(caCert,x509,NID_ext_key_usage,"clientAuth, serverAuth");
    if (ret != 0){
        goto END;
    }
    ret = set_extension(caCert,x509,NID_subject_key_identifier,"hash");
    if (ret != 0){
       goto END;
    }
    ret = set_extension(caCert,x509,NID_authority_key_identifier,"keyid:always");
    if (ret != 0){
        goto END;
    }
    ret = set_extension(caCert,x509,NID_key_usage, "digitalSignature, keyEncipherment, keyCertSign, cRLSign");
    if (ret != 0){
        goto END;
    }
    ret = set_extension(caCert,x509,NID_basic_constraints, "CA:true");
    if (ret != 0){
        goto END;
    }

	X509_sign(x509, caPriKey, EVP_sm3());
END:
    if (ret != 0){
		if (x509 != NULL){
			X509_free(x509);
			x509 = NULL;
		}
    }
    return x509;
}

// 传入私钥（内含公钥）
X509 * GenerateCert(void* country, void* province, void* locality, void* org, void* ou, void* commonName, uint years, EVP_PKEY *pubKey, EVP_PKEY *caPriKey, X509 * caCert){
    int ret = 0;
    X509* x509 = GenerateCertBasic(country, province, locality, org, ou, commonName, years, pubKey);
    if (x509 == NULL){
        return NULL;
    }

    X509_NAME * caName;
    caName = X509_get_subject_name(caCert);
    X509_set_issuer_name(x509, caName);

    // todo judge return value err process
    ret = set_extension(caCert, x509, NID_authority_key_identifier, "keyid:always");
    if (ret != 0){
        goto END;
    }
    ret = set_extension(caCert,x509,NID_key_usage, "digitalSignature");
    if (ret != 0){
        goto END;
    }
    ret = set_extension(caCert,x509,NID_basic_constraints, "CA:false");
    if (ret != 0){
        goto END;
    }

	X509_sign(x509, caPriKey, EVP_sm3());


END:
    if (ret != 0){
		if (x509 != NULL){
			X509_free(x509);
			x509 = NULL;
		}
    }

	return x509;
}


// 传入私钥（内含公钥）
X509 * GenerateSelfSignCert(void* country, void* province, void* locality, void* org, void* ou, void* commonName, uint years, EVP_PKEY *pkey){
    int ret = 0;
    X509* x509 = GenerateCertBasic(country, province, locality, org, ou, commonName, years, pkey);
    if (x509 == NULL){
        return NULL;
    }
    X509_NAME * name;
    name = X509_get_subject_name(x509);
	X509_set_issuer_name(x509, name);


    ret = set_extension(x509,x509,NID_ext_key_usage,"clientAuth, serverAuth");
    if (ret != 0){
        goto END;
    }
    // todo hash is the default
    ret = set_extension(x509,x509,NID_subject_key_identifier,"hash");
    if (ret != 0){
        goto END;
    }
    ret = set_extension(x509,x509,NID_authority_key_identifier,"keyid:always");
    if (ret != 0){
        goto END;
    }
    ret = set_extension(x509,x509,NID_key_usage, "digitalSignature, keyEncipherment, keyCertSign, cRLSign");
    if (ret != 0){
        goto END;
    }
    ret = set_extension(x509,x509,NID_basic_constraints, "CA:true");
    if (ret != 0){
        goto END;
    }

	// for EVP_sm3(), the signature algorithm is : 1.2.156.10197.1.501
    // for EVP_sha256(), the signature algorith is :sha256ECDSA
    X509_sign(x509, pkey, EVP_sm3());

END:
    if (ret != 0){
		if (x509 != NULL){
			X509_free(x509);
			x509 = NULL;
		}
    }
	return x509;
}


void X509_to_PEM(X509 *cert, void* pemBuf, uint* pemBufLen) {
    BIO *bio = NULL;
    char *pem = NULL;

    if (NULL == cert) {
        return ;
    }

    bio = BIO_new(BIO_s_mem());
    if (NULL == bio) {
        return ;
    }

    if (0 == PEM_write_bio_X509(bio, cert)) {
        BIO_free(bio);
        return ;
    }

    BUF_MEM *mem = NULL;
    BIO_get_mem_ptr(bio, &mem);

    if (*pemBufLen <= mem->length){
       return;
    }
    memcpy(pemBuf, mem->data, mem->length);

    *pemBufLen = mem->length;
    BIO_set_close(bio, BIO_NOCLOSE);
    BIO_free(bio);
    return ;
}