package xqrcode

import (
    "bytes"
    "fmt"
    "image"
    "image/color"
    "image/png"
    "math"
    "net/http"
    "os"
    "slices"
    "strconv"
    "strings"
)

func init() {
    rsInitTables()
}

func bchDigit(data int) (result int) {
    for data != 0 {
        result++
        data >>= 1
    }
    return
}

func bch(data, eccLen, genPoly int) (result int) {
    result = data << eccLen
    genBits := bchDigit(genPoly)
    for bchDigit(result)-genBits >= 0 {
        result ^= genPoly << (bchDigit(result) - genBits)
    }
    return
}

var rsExpTable [256]byte

var rsLogTable [256]byte

func rsInitTables() {
    for i := 0; i < 8; i++ {
        rsExpTable[i] = byte(1 << i)
    }
    for i := 8; i < 256; i++ {
        rsExpTable[i] = rsExpTable[i-4] ^ rsExpTable[i-5] ^ rsExpTable[i-6] ^ rsExpTable[i-8]
    }
    for i := 0; i < 256; i++ {
        rsLogTable[rsExpTable[i]] = byte(i)
    }
}

func rsExp(e int) byte {
    for e < 0 {
        e += 255
    }
    for e > 255 {
        e -= 255
    }
    return rsExpTable[e]
}

func rsPolyMul(a, b []byte, count int) []byte {
    bLen := len(b)
    tmp := make([]byte, count+bLen+1)
    for i := 0; i < count; i++ {
        for j := 0; j < bLen; j++ {
            tmp[i+j] = tmp[i+j] ^ rsExp(int(rsLogTable[a[i]])+int(rsLogTable[b[j]]))
        }
    }
    return tmp[:count+bLen-1]
}

func rsGenPoly(poly *[]byte, count int) {
    tmp := make([]byte, 2)
    tmp[0] = 1
    (*poly)[0] = 1
    for i := 1; i < count; i++ {
        tmp[1] = rsExp(i)
        mul := rsPolyMul(*poly, tmp, i+1)
        _ = copy(*poly, mul)
    }
}

func rsEncode(data []byte, eccLen int) []byte {
    gen := make([]byte, eccLen+1)
    rsGenPoly(&gen, eccLen)
    dataLen := len(data)
    tmp := make([]byte, dataLen+eccLen)
    copy(tmp, data)
    for i := 0; i < dataLen; i++ {
        if tmp[i] != 0 {
            r := int(rsLogTable[tmp[i]]) - int(rsLogTable[gen[0]])
            for j := 0; j <= eccLen; j++ {
                tmp[i+j] = tmp[i+j] ^ rsExp(int(rsLogTable[gen[j]])+r)
            }
        }
    }
    return tmp[dataLen:]
}
