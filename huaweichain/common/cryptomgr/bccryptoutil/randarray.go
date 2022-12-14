/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

// Package bccryptoutil is the util package for crypto manager
package bccryptoutil

import (
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"math/big"
	"sync"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/logger"
)

var log = logger.GetModuleLogger("cryptomgr", "bccryptoutil")

// RandomArrays random arrays for batch verify
type RandomArrays struct {
	CoeffiArrFinal []*big.Int
	Part           []int
	IArr           []int
	RandRxPosition []int
}

type ecdsaInt struct {
	INT *big.Int
}

// RandMap random array map for different batch size
var RandMap = &sync.Map{}

// InitRandom init the RandMap
func InitRandom() {
	cv := elliptic.P256()
	var minBatchSize1 = 31
	var maxBatchSize1 = 100
	GenRandom(cv, minBatchSize1, maxBatchSize1, RandMap)

	var minBatchSize2 = 499
	var maxBatchSize2 = 501
	GenRandom(cv, minBatchSize2, maxBatchSize2, RandMap)
	log.Infof("Init random array success.")
}

// GenRandom get random2
func GenRandom(curve elliptic.Curve, minBatchSize int, maxBatchSize int, p *sync.Map) {
	var wg sync.WaitGroup
	for i := minBatchSize; i <= maxBatchSize; i++ {
		wg.Add(1)
		go func(p *sync.Map, cv elliptic.Curve, n int) {
			randomArray, err := genRandomArrays(cv, n)
			if err != nil {
				log.Infof("generate random failed for %d", n)
				wg.Done()
				return
			}
			p.Store(n, randomArray)
			wg.Done()
		}(p, curve, i)
	}
	wg.Wait()
}

// genRandomArrays get random arrays
func genRandomArrays(curve elliptic.Curve, batchSize int) (*RandomArrays, error) {
	result := RandomArrays{}
	result.IArr = getRandomBinArr(batchSize)
	result.RandRxPosition = getIntRandArray(batchSize)

	m := 10
	n := 30
	rowsNum := getIntRandInRange(m, n)
	result.Part = getIntRandArrayPart(batchSize, rowsNum)
	coeffiArr, err := getCoeffiArrBigInt(curve, result.IArr, result.RandRxPosition)
	if err != nil {
		return nil, err
	}
	result.CoeffiArrFinal, err = getPartSumOfCoeffiArrBigInt(curve, coeffiArr, result.Part)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Calculates the position of the randomly selected row in the coefficient matrix.
func getIntRandArrayPart(n int, p int) []int {
	var result, tmpArr []int
	for i := 0; i < n; i++ {
		tmpArr = append(tmpArr, i)
	}
	for i := 0; i < p; i++ {
		tmpArrLen := len(tmpArr)
		num := getIntRand(tmpArrLen)
		if tmpArrLen <= num {
			return nil
		}
		tmp := tmpArr[num]
		result = append(result, tmp)
		tmpArr = append(tmpArr[:num], tmpArr[num+1:]...)
	}
	return result
}

func calcBigInt(z1Str1 []byte, n int) ([]*big.Int, error) {
	z1Len := len(z1Str1)
	const certainPos = 4
	if z1Len <= certainPos {
		return nil, errors.Errorf("z1 string len is <= %d", certainPos)
	}
	tmp := []*big.Int{}
	for j := 0; j < n; j++ {
		z0Int := &ecdsaInt{}
		z1Str1[certainPos] = '\x00'
		if _, err := asn1.Unmarshal(z1Str1, z0Int); err != nil {
			return nil, errors.New("z0Int asn1.Unmarshal signature failed")
		}
		tmp = append(tmp, z0Int.INT)
	}
	return tmp, nil
}

func calcResult(i int, result [][]*big.Int, iArr []int, tmpN *big.Int) {
	n := len(iArr)
	if i > 0 {
		for j := 0; j < n; j++ {
			if iArr[i] == 1 {
				result[i][j].Add(result[i-1][j], result[i-1][j])
				result[i][j].Mod(result[i][j], tmpN)
			} else {
				result[i][j].Mod(result[i-1][j], tmpN)
			}
		}
	}
}

// Calculate the coefficient matrix to the right of the equal sign
func getCoeffiArrBigInt(curve elliptic.Curve, iArr []int, randomRxposition []int) ([][]*big.Int, error) {
	var result [][]*big.Int
	c := curve
	N := c.Params().N
	n := len(iArr)
	var z1Str1 []byte
	z1Str1 = append(z1Str1, '\x30', '\x03', '\x02', '\x01', '\x00')

	for i := 0; i < n; i++ {
		tmp, err := calcBigInt(z1Str1, n)
		if err != nil {
			return nil, err
		}

		result = append(result, tmp)
		calcResult(i, result, iArr, N)

		z1Int := &ecdsaInt{}
		z1Str1[4] = '\x01'
		if _, err := asn1.Unmarshal(z1Str1, z1Int); err != nil {
			return nil, errors.New("z1Int asn1.Unmarshal signature failed")
		}
		if i >= len(result) || i >= len(result[i]) {
			return nil, errors.New("the result length is not correct")
		}
		result[i][i].Add(result[i][i], z1Int.INT)
		result[i][i].Mod(result[i][i], N)
		z1Inttmp := &ecdsaInt{}
		if _, err := asn1.Unmarshal(z1Str1, z1Inttmp); err != nil {
			return nil, errors.New("z1Inttmp asn1.Unmarshal signature failed")
		}
		randomPosValue := randomRxposition[i]
		if randomPosValue >= len(result[i]) {
			return nil, errors.New("randomRxposition[i] is bigger than the length of result[i]")
		}
		result[i][randomPosValue].Add(result[i][randomPosValue], z1Inttmp.INT)
		result[i][randomPosValue].Mod(result[i][randomPosValue], N)
	}
	return result, nil
}

// Calculate the random number used by the coefficient matrix
func getPartSumOfCoeffiArrBigInt(curve elliptic.Curve, coeffiArr [][]*big.Int, part []int) ([]*big.Int, error) {
	var result []*big.Int
	n := len(coeffiArr[0])
	c := curve
	N := c.Params().N
	var z1Str1 []byte
	z1Str1 = append(z1Str1, '\x30', '\x03', '\x02', '\x01', '\x00')

	for i := 0; i < n; i++ {
		z0Int := &ecdsaInt{}
		if _, err := asn1.Unmarshal(z1Str1, z0Int); err != nil {
			return nil, errors.New("z0Int asn1.Unmarshal signature failed")
		}
		for j := 0; j < len(part); j++ {
			valuePart := part[j]
			if valuePart >= len(coeffiArr) {
				return nil, errors.New("length of coeffiArr is not correct")
			}
			valueCoeffiArr := coeffiArr[valuePart]
			if i >= len(valueCoeffiArr) {
				return nil, errors.New("length of coeffiArr[part[j]] is not correct")
			}
			z0Int.INT.Add(z0Int.INT, valueCoeffiArr[i])
			z0Int.INT.Mod(z0Int.INT, N)
		}
		result = append(result, z0Int.INT)
	}
	for i := 0; i < n; i++ {
		if n-1 >= len(coeffiArr) {
			return nil, errors.New(" n-1 >= len(coeffiArr)")
		}
		valueCoeffiArr := coeffiArr[n-1]
		if i >= len(valueCoeffiArr) {
			return nil, errors.New("valueCoeffiArr :=  coeffiArr[n-1]")
		}
		if i >= len(result) {
			return nil, errors.New("i >= len(result)")
		}
		result[i].Add(result[i], valueCoeffiArr[i])
		result[i].Mod(result[i], N)
	}
	return result, nil
}

// Randomly select a number in the range [0, n-1]
func getIntRand(n int) int {
	b := new(big.Int).SetInt64(int64(n))
	b, err := rand.Int(rand.Reader, b)
	if err != nil {
		// todo add err return
		return 0
	}
	return int(b.Int64())
}

// Randomly select a number in the range [m, n].
func getIntRandInRange(m int, n int) int {
	return getIntRand(n-m+1) + m
}

func getRandomBinArr(n int) []int {
	var result []int

	var maxRandNum = 2
	for i := 0; i < n; i++ {
		intRand := getIntRand(maxRandNum)
		result = append(result, intRand)
	}
	return result
}

// Random sorting results of 0, 1, 2, and n-1 are generated
func getIntRandArray(n int) []int {
	var result, tmpArr []int
	for i := 0; i < n; i++ {
		tmpArr = append(tmpArr, i)
	}
	for i := 0; i < n; i++ {
		num := getIntRand(len(tmpArr))
		if tmpArr == nil {
			// todo print err log
			return result
		}
		if num >= len(tmpArr) {
			return result
		}
		tmp := tmpArr[num]
		result = append(result, tmp)
		tmpArr = append(tmpArr[:num], tmpArr[num+1:]...)
	}
	return result
}
