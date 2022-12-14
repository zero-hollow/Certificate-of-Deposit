/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package ecdsaalg

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha512"
	"encoding/asn1"
	"io"
	"math/big"
	"sync"

	"git.huawei.com/huaweichain/common/cryptomgr/bccryptoutil"

	"github.com/pkg/errors"
)

type zr struct {
	io.Reader
}

type ecdsaSignatureRich struct {
	R, S, Rx, Ry *big.Int
}
type ecdsaSigParaArr struct {
	arrS, arrR, arrRx, arrRy []*big.Int
}

type invertible interface {
	// Inverse returns the inverse of k in GF(P)
	Inverse(k *big.Int) *big.Int
}
type combinedMult interface {
	CombinedMult(bigX, bigY *big.Int, baseScalar, scalar []byte) (x, y *big.Int)
}

var (
	closedChanOnce sync.Once
	closedChan     chan struct{}
)

const (
	aesIV = "IV for ECDSA CTR"
)

var one = new(big.Int).SetInt64(1)
var zeroReader = &zr{}

var errZeroParam = errors.New("zero parameter")

func signBatchMarshal(priv *ecdsa.PrivateKey, rand io.Reader, digest []byte) ([]byte, error) {
	sigRes, err := signBatch(rand, priv, digest)
	if err != nil {
		return nil, err
	}

	asn1Res, err := asn1.Marshal(*sigRes)
	if err != nil {
		return nil, err
	}

	return asn1Res, nil
}

func getCsprng(rand io.Reader, priv *ecdsa.PrivateKey, hash []byte) (*cipher.StreamReader, error) {
	maybeReadByte(rand)

	// Get min(log2(q) / 2, 256) bits of entropy from rand.
	const m = 7
	const n = 16
	const maxLen = 32
	entropyLen := (priv.Curve.Params().BitSize + m) / n
	if entropyLen > maxLen {
		entropyLen = maxLen
	}
	if entropyLen <= 0 {
		return nil, errors.Errorf("the entropyLen:%d is not correct", entropyLen)
	}
	entropy := make([]byte, entropyLen)
	_, err := io.ReadFull(rand, entropy)
	if err != nil {
		return nil, errors.New("read full failed")
	}

	// Initialize an SHA-512 hash context; digest ...
	md := sha512.New()

	// the private key,
	_, err = md.Write(priv.D.Bytes())
	if err != nil {
		return nil, err
	}

	// the entropy,
	_, err = md.Write(entropy)
	if err != nil {
		return nil, err
	}
	// and the input hash;
	_, err = md.Write(hash)
	if err != nil {
		return nil, err
	}
	key := md.Sum(nil)[:maxLen]

	// Create an AES-CTR instance to use as a CSPRNG.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a CSPRNG that xors a stream of zeros with
	// the output of the AES-CTR instance.
	return &cipher.StreamReader{
		R: zeroReader,
		S: cipher.NewCTR(block, []byte(aesIV)),
	}, nil
}

// sign for generating ecdsaSignatureRich struct for batch verify
func signBatch(rand io.Reader, priv *ecdsa.PrivateKey, hash []byte) (*ecdsaSignatureRich, error) {
	var r, s, Rx, Ry *big.Int
	csprng, err := getCsprng(rand, priv, hash)
	if err != nil {
		return nil, err
	}

	// See [NSA] 3.4.1
	c := priv.PublicKey.Curve
	N := c.Params().N
	if N.Sign() == 0 {
		return nil, errZeroParam
	}
	var k, kInv *big.Int
	for {
		for {
			k, err = randFieldElement(c, csprng)
			if err != nil {
				return nil, err
			}

			if in, ok := priv.Curve.(invertible); ok {
				kInv = in.Inverse(k)
			} else {
				kInv = fermatInverse(k, N) // N != 0
			}
			Rx, Ry = priv.Curve.ScalarBaseMult(k.Bytes())
			r = Rx
			r.Mod(r, N)
			if r.Sign() != 0 {
				break
			}
		}

		e := hashToInt(hash, c)
		s = new(big.Int).Mul(priv.D, r)
		s.Add(s, e)
		s.Mul(s, kInv)
		// N != 0
		s.Mod(s, N)
		if s.Sign() != 0 {
			break
		}
	}

	return &ecdsaSignatureRich{
		R:  r,
		S:  s,
		Rx: Rx,
		Ry: Ry,
	}, nil
}

func randFieldElement(c elliptic.Curve, rand io.Reader) (k *big.Int, err error) {
	const byteSize = 8
	params := c.Params()
	b := make([]byte, params.BitSize/byteSize+byteSize)
	_, err = io.ReadFull(rand, b)
	if err != nil {
		return
	}

	k = new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(params.N, one)
	k.Mod(k, n)
	k.Add(k, one)
	return
}

func maybeReadByte(r io.Reader) {
	closedChanOnce.Do(func() {
		closedChan = make(chan struct{})
		close(closedChan)
	})

	if closedChan == nil {
		return
	}
	select {
	case _, ok := <-closedChan:
		if !ok {
			return
		}
	case _, ok := <-closedChan:
		if !ok {
			var buf [1]byte
			_, err := r.Read(buf[:])
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
}

// Read replaces the contents of dst with zeros.
func (z *zr) Read(dst []byte) (n int, err error) {
	for i := range dst {
		dst[i] = 0
	}
	return len(dst), nil
}
func fermatInverse(k, bigN *big.Int) *big.Int {
	var x int64 = 2
	two := big.NewInt(x)
	nMinus2 := new(big.Int).Sub(bigN, two)
	return new(big.Int).Exp(k, nMinus2, bigN)
}

// Single-PubKey Batch Verify
func verifyBatchWithRandArray(pub *ecdsa.PublicKey, hash [][]byte, sigs [][]byte,
	randArr *bccryptoutil.RandomArrays) bool {
	var batchSize = len(sigs)

	paraArr := &ecdsaSigParaArr{}
	paraArr.arrS = make([]*big.Int, 0, batchSize)
	paraArr.arrR = make([]*big.Int, 0, batchSize)
	paraArr.arrRx = make([]*big.Int, 0, batchSize)
	paraArr.arrRy = make([]*big.Int, 0, batchSize)

	sigInfo := &ecdsaSignatureRich{
		R:  nil,
		S:  nil,
		Rx: nil,
		Ry: nil,
	}
	for i := 0; i < batchSize; i++ {
		_, err := asn1.Unmarshal(sigs[i], sigInfo)
		if err != nil {
			return false
		}

		paraArr.arrR = append(paraArr.arrR, sigInfo.R)
		paraArr.arrS = append(paraArr.arrS, sigInfo.S)

		paraArr.arrRx = append(paraArr.arrRx, sigInfo.Rx)
		paraArr.arrRy = append(paraArr.arrRy, sigInfo.Ry)
	}

	return verifyBatchCore(pub, hash, paraArr, randArr)
}

func hashToInt(hash []byte, c elliptic.Curve) *big.Int {
	hashTmp := hash
	const m = 7
	const n = 8
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + m) / n
	if len(hash) > orderBytes {
		hashTmp = hash[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hashTmp)
	excess := len(hashTmp)*n - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}

// Calculate the Rx and Ry vectors based on iArr and the random position array randRxposition
func getRPointArr(pub *ecdsa.PublicKey, iArr []int, randRxposition []int, arrRx []*big.Int,
	arrRy []*big.Int) ([]*big.Int, []*big.Int) {
	var arrRx1, arrRy1 []*big.Int
	c := pub.Curve
	posLen := len(randRxposition)
	rxLen := len(arrRx)
	ryLen := len(arrRy)
	if ryLen < rxLen || posLen < rxLen {
		return nil, nil
	}

	for i := 0; i < rxLen; i++ {
		posValue := randRxposition[i]
		if posValue >= rxLen {
			return nil, nil
		}
		if i == 0 {
			tmpx, tmpy := c.Add(arrRx[0], arrRy[0], arrRx[posValue], arrRy[posValue])
			arrRx1 = append(arrRx1, tmpx)
			arrRy1 = append(arrRy1, tmpy)
		} else {
			if len(arrRx1) <= i-1 || len(arrRy1) <= i-1 {
				return nil, nil
			}
			if iArr[i] == 1 {
				tmpx1, tmpy1 := c.Double(arrRx1[i-1], arrRy1[i-1])
				tmpx2, tmpy2 := c.Add(tmpx1, tmpy1, arrRx[i], arrRy[i])

				tmpx3, tmpy3 := c.Add(tmpx2, tmpy2, arrRx[posValue], arrRy[posValue])
				arrRx1 = append(arrRx1, tmpx3)
				arrRy1 = append(arrRy1, tmpy3)
			} else {
				tmpx2, tmpy2 := c.Add(arrRx1[i-1], arrRy1[i-1], arrRx[i], arrRy[i])
				tmpx3, tmpy3 := c.Add(tmpx2, tmpy2, arrRx[posValue], arrRy[posValue])
				arrRx1 = append(arrRx1, tmpx3)
				arrRy1 = append(arrRy1, tmpy3)
			}
		}
	}
	return arrRx1, arrRy1
}

func getPartSumOfRPointArr(pub *ecdsa.PublicKey, arrRx []*big.Int, arrRy []*big.Int,
	part []int) (arrRx1 *big.Int, arrRy1 *big.Int) {
	c := pub.Curve
	for i := 0; i < len(part); i++ {
		if i == 0 {
			arrRx1 = arrRx[part[i]]
			arrRy1 = arrRy[part[i]]
		} else {
			arrRx1, arrRy1 = c.Add(arrRx1, arrRy1, arrRx[part[i]], arrRy[part[i]])
		}
	}
	arrRx1, arrRy1 = c.Add(arrRx1, arrRy1, arrRx[len(arrRx)-1], arrRy[len(arrRx)-1])
	return
}

// Core code for batch signature verification
func verifyBatchCore(pub *ecdsa.PublicKey, hash [][]byte, paraArr *ecdsaSigParaArr,
	randArr *bccryptoutil.RandomArrays) bool {
	arrS := paraArr.arrS
	arrR := paraArr.arrR
	arrRx := paraArr.arrRx
	arrRy := paraArr.arrRy
	c := pub.Curve
	N := c.Params().N

	u := &big.Int{}
	v := &big.Int{}
	tmpRx1, tmpRy1 := getRPointArr(pub, randArr.IArr, randArr.RandRxPosition, arrRx, arrRy)
	tmpRxFinal, _ := getPartSumOfRPointArr(pub, tmpRx1, tmpRy1, randArr.Part)

	var hashLen = len(hash)
	for i := 0; i < hashLen; i++ {
		if arrR[i].Sign() <= 0 || arrS[i].Sign() <= 0 || arrRx[i].Sign() <= 0 || arrRy[i].Sign() <= 0 {
			return false
		}
		if arrR[i].Cmp(N) >= 0 || arrS[i].Cmp(N) >= 0 || arrRx[i].Cmp(N) >= 0 || arrRy[i].Cmp(N) >= 0 {
			return false
		}
		e := hashToInt(hash[i], c)
		var w *big.Int
		if in, ok := c.(invertible); ok {
			w = in.Inverse(arrS[i])
		} else {
			w = new(big.Int).ModInverse(arrS[i], N)
		}
		if w == nil {
			return false
		}
		// sum U,V
		var tmpU, tmpV *big.Int
		tmpU = e.Mul(e, w)
		tmpU.Mod(tmpU, N)
		tmpU = tmpU.Mul(tmpU, randArr.CoeffiArrFinal[i])
		tmpU.Mod(tmpU, N)
		tmpV = w.Mul(arrR[i], w)
		tmpV.Mod(tmpV, N)
		tmpV = tmpV.Mul(tmpV, randArr.CoeffiArrFinal[i])
		tmpV.Mod(tmpV, N)

		if tmpU.Sign() == 0 {
			u = tmpU
			v = tmpV
		} else {
			u.Add(u, tmpU)
			u.Mod(u, N)
			v.Add(v, tmpV)
			v.Mod(v, N)
		}
	}

	return checkXY(pub, u, v, N, tmpRxFinal)
}

// Check if implements S1*g + S2*p
func checkXY(pub *ecdsa.PublicKey, u, v *big.Int, bigN *big.Int, rxFinal *big.Int) bool {
	c := pub.Curve
	var x, y *big.Int
	if opt, ok := c.(combinedMult); ok {
		x, y = opt.CombinedMult(pub.X, pub.Y, u.Bytes(), v.Bytes())
	} else {
		x1, y1 := c.ScalarBaseMult(u.Bytes())
		x2, y2 := c.ScalarMult(pub.X, pub.Y, v.Bytes())
		x, y = c.Add(x1, y1, x2, y2)
	}
	if x == nil || y == nil {
		return false
	}

	if x.Sign() == 0 && y.Sign() == 0 {
		return false
	}
	x.Mod(x, bigN)
	return x.Cmp(rxFinal) == 0
}
