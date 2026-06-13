package xqrcode

import (
    "fmt"
    "slices"
    "strconv"
    "strings"
)

type encodeMode int

const (
    nullEncode       encodeMode = 0
    numericMode      encodeMode = 1
    alphaNumericMode encodeMode = 2
    byteMode         encodeMode = 3
    kanjiMode        encodeMode = 4
    eciMode          encodeMode = 5
    fnc1Mode1        encodeMode = 6
    fnc1Mode2        encodeMode = 7
    structAppendMode encodeMode = 8
)

func modeName(mode encodeMode) string {
    names := []string{"", "Numeric", "AlphaNumeric", "Byte", "Kanji", "ECI", "Fnc1 1st", "Fnc1 2nd", "Struct Append"}
    return names[mode]
}

type encodedSegment struct {
    mode  encodeMode
    count int
    data  []uint32
}

func (seg *encodedSegment) append(value uint32, count uint32) {
    seg.data = append(seg.data, (count<<16)|value)
}

type segmentIndicator struct {
    modeBits  byte
    modeData  byte
    countBits byte
    countData uint16
}

func getSegmentIndicator(kind QRCodeKind, version QRCodeVersion, mode encodeMode,
    count uint16) (*segmentIndicator, error, bool) {
    if mode == kanjiMode {
        count = count / 2
    }
    switch kind {
    case RMQRCode:
        if version >= 1 && version <= 32 {
            countBits := []map[encodeMode]byte{
                {numericMode: 4, alphaNumericMode: 3, byteMode: 3, kanjiMode: 2},
                {numericMode: 5, alphaNumericMode: 5, byteMode: 4, kanjiMode: 3},
                {numericMode: 6, alphaNumericMode: 5, byteMode: 5, kanjiMode: 4},
                {numericMode: 7, alphaNumericMode: 6, byteMode: 5, kanjiMode: 5},
                {numericMode: 7, alphaNumericMode: 6, byteMode: 6, kanjiMode: 5},
                {numericMode: 5, alphaNumericMode: 5, byteMode: 4, kanjiMode: 3},
                {numericMode: 6, alphaNumericMode: 5, byteMode: 5, kanjiMode: 4},
                {numericMode: 7, alphaNumericMode: 6, byteMode: 5, kanjiMode: 5},
                {numericMode: 7, alphaNumericMode: 6, byteMode: 6, kanjiMode: 5},
                {numericMode: 8, alphaNumericMode: 7, byteMode: 6, kanjiMode: 6},
                {numericMode: 4, alphaNumericMode: 4, byteMode: 3, kanjiMode: 2},
                {numericMode: 6, alphaNumericMode: 5, byteMode: 5, kanjiMode: 4},
                {numericMode: 7, alphaNumericMode: 6, byteMode: 5, kanjiMode: 5},
                {numericMode: 7, alphaNumericMode: 6, byteMode: 6, kanjiMode: 5},
                {numericMode: 8, alphaNumericMode: 7, byteMode: 6, kanjiMode: 6},
                {numericMode: 8, alphaNumericMode: 7, byteMode: 7, kanjiMode: 6},
                {numericMode: 5, alphaNumericMode: 5, byteMode: 4, kanjiMode: 3},
                {numericMode: 6, alphaNumericMode: 6, byteMode: 5, kanjiMode: 5},
                {numericMode: 7, alphaNumericMode: 6, byteMode: 6, kanjiMode: 5},
                {numericMode: 7, alphaNumericMode: 7, byteMode: 6, kanjiMode: 6},
                {numericMode: 8, alphaNumericMode: 7, byteMode: 7, kanjiMode: 6},
                {numericMode: 8, alphaNumericMode: 8, byteMode: 7, kanjiMode: 7},
                {numericMode: 7, alphaNumericMode: 6, byteMode: 6, kanjiMode: 5},
                {numericMode: 7, alphaNumericMode: 7, byteMode: 6, kanjiMode: 5},
                {numericMode: 8, alphaNumericMode: 7, byteMode: 7, kanjiMode: 6},
                {numericMode: 8, alphaNumericMode: 7, byteMode: 7, kanjiMode: 6},
                {numericMode: 9, alphaNumericMode: 8, byteMode: 7, kanjiMode: 7},
                {numericMode: 7, alphaNumericMode: 6, byteMode: 6, kanjiMode: 5},
                {numericMode: 8, alphaNumericMode: 7, byteMode: 6, kanjiMode: 6},
                {numericMode: 8, alphaNumericMode: 7, byteMode: 7, kanjiMode: 6},
                {numericMode: 8, alphaNumericMode: 8, byteMode: 7, kanjiMode: 6},
                {numericMode: 9, alphaNumericMode: 8, byteMode: 8, kanjiMode: 7},
            }
            cb := byte(0)
            if c, ok := countBits[version-1][mode]; ok {
                cb = c
            }
            modeData := map[encodeMode]byte{
                numericMode: 1, alphaNumericMode: 2, byteMode: 3, kanjiMode: 4, fnc1Mode1: 5, fnc1Mode2: 6, eciMode: 7,
            }
            if m, ok := modeData[mode]; ok {
                return &segmentIndicator{modeBits: 3, modeData: m, countBits: cb, countData: count}, nil, false
            } else {
                return nil, fmt.Errorf("mode '%s' is not supported", modeName(mode)), false
            }
        }
    case QRCode:
        if version >= 1 && version <= 40 {
            countSeg := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2,
                2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
            seg := countSeg[version-1]
            countBits := []map[encodeMode]byte{
                {numericMode: 10, alphaNumericMode: 9, byteMode: 8, kanjiMode: 8},
                {numericMode: 12, alphaNumericMode: 11, byteMode: 16, kanjiMode: 10},
                {numericMode: 14, alphaNumericMode: 13, byteMode: 16, kanjiMode: 12},
            }
            cb := byte(0)
            if c, ok := countBits[seg][mode]; ok {
                cb = c
            }
            modeData := map[encodeMode]byte{numericMode: 1, alphaNumericMode: 2, byteMode: 4, kanjiMode: 8, eciMode: 7,
                fnc1Mode1: 5, fnc1Mode2: 9, structAppendMode: 3}
            if m, ok := modeData[mode]; ok {
                return &segmentIndicator{modeBits: 4, modeData: m, countBits: cb, countData: count}, nil, false
            } else {
                return nil, fmt.Errorf("mode '%s' is not supported", modeName(mode)), false
            }
        }
    case MicroQRCode:
        if version >= 1 && version <= 4 {
            countBits := []map[encodeMode]byte{
                {numericMode: 3, alphaNumericMode: 0, byteMode: 0, kanjiMode: 0},
                {numericMode: 4, alphaNumericMode: 3, byteMode: 0, kanjiMode: 0},
                {numericMode: 5, alphaNumericMode: 4, byteMode: 4, kanjiMode: 3},
                {numericMode: 6, alphaNumericMode: 5, byteMode: 5, kanjiMode: 4},
            }
            cb := byte(0)
            if c, ok := countBits[version-1][mode]; ok {
                cb = c
                if cb == 0 {
                    return nil, fmt.Errorf("mode '%s' is not supported by %s version '%d'", modeName(mode), kind,
                        version), true
                }
            } else {
                return nil, fmt.Errorf("mode '%s' is not supported", modeName(mode)), true
            }
            modeData := map[encodeMode]byte{numericMode: 0, alphaNumericMode: 1, byteMode: 2, kanjiMode: 3}
            if m, ok := modeData[mode]; ok {
                return &segmentIndicator{modeBits: byte(version - 1), modeData: m, countBits: cb,
                    countData: count}, nil, false
            } else {
                return nil, fmt.Errorf("mode '%s' is not supported", modeName(mode)), false
            }
        }
    default:
        return nil, fmt.Errorf("invalid barcode kind: '%s'", kind), false
    }
    return nil, fmt.Errorf("version '%d' is not supported", version), false
}

type segmentDivision struct {
    mode encodeMode
    data []byte
    head int
}

func appendSegment(segments *[]segmentDivision, segment segmentDivision, disableMerge bool) {
    if segment.mode != nullEncode {
        *segments = append(*segments, segment)
    }
    index := len(*segments) - 1
    for index > 0 {
        curMode := (*segments)[index].mode
        curHead := (*segments)[index].head
        curData := len((*segments)[index].data)
        curSize := curHead + encodedLen(curData, curMode)
        preMode := (*segments)[index-1].mode
        preHead := (*segments)[index-1].head
        preData := len((*segments)[index-1].data)
        preSize := preHead + encodedLen(preData, preMode)
        pprMode := nullEncode
        pprHead := 0
        pprData := 0
        pprSize := 0
        if index > 1 {
            pprMode = (*segments)[index-2].mode
            pprHead = (*segments)[index-2].head
            pprData = len((*segments)[index-2].data)
            pprSize = pprHead + encodedLen(pprData, pprMode)
        }
        if curMode == preMode {
            (*segments)[index-1].data = append((*segments)[index-1].data, (*segments)[index].data...)
            *segments = (*segments)[0 : len(*segments)-1]
            if preMode != byteMode {
                index--
                continue
            }
        } else {
            switch curMode {
            case numericMode:
                switch preMode {
                case alphaNumericMode, byteMode:
                    if !disableMerge && ((preHead + encodedLen(preData+curData, preMode)) <= (preSize + curSize)) {
                        (*segments)[index-1].data = append((*segments)[index-1].data, (*segments)[index].data...)
                        *segments = (*segments)[0 : len(*segments)-1]
                        if preMode == alphaNumericMode {
                            index--
                            continue
                        }
                    }
                }
            case alphaNumericMode:
                switch preMode {
                case numericMode:
                    if pprMode == curMode {
                        if (pprHead + encodedLen(pprData+preData+curData, pprMode)) < (pprSize + preSize + curSize) {
                            (*segments)[index-2].data = append((*segments)[index-2].data, (*segments)[index-1].data...)
                            (*segments)[index-2].data = append((*segments)[index-2].data, (*segments)[index].data...)
                            *segments = (*segments)[0 : len(*segments)-2]
                            index--
                            index--
                            continue
                        }
                    } else {
                        if (curHead + encodedLen(preData+curData, curMode)) < (preSize + curSize) {
                            (*segments)[index-1].mode = curMode
                            (*segments)[index-1].head = curHead
                            (*segments)[index-1].data = append((*segments)[index-1].data, (*segments)[index].data...)
                            *segments = (*segments)[0 : len(*segments)-1]
                            index--
                            continue
                        }
                    }
                case byteMode:
                    if !disableMerge && ((preHead + encodedLen(preData+curData, preMode)) <= (preSize + curSize)) {
                        (*segments)[index-1].data = append((*segments)[index-1].data, (*segments)[index].data...)
                        *segments = (*segments)[0 : len(*segments)-1]
                    }
                }
            case byteMode:
                switch preMode {
                case numericMode, alphaNumericMode:
                    if pprMode == curMode {
                        if (pprHead + encodedLen(pprData+preData+curData, pprMode)) < (pprSize + preSize + curSize) {
                            (*segments)[index-2].data = append((*segments)[index-2].data, (*segments)[index-1].data...)
                            (*segments)[index-2].data = append((*segments)[index-2].data, (*segments)[index].data...)
                            *segments = (*segments)[0 : len(*segments)-2]
                            index--
                            index--
                            continue
                        }
                    } else {
                        if curHead+encodedLen(preData+curData, curMode) <= (preSize + curSize) {
                            (*segments)[index-1].mode = curMode
                            (*segments)[index-1].head = curHead
                            (*segments)[index-1].data = append((*segments)[index-1].data, (*segments)[index].data...)
                            *segments = (*segments)[0 : len(*segments)-1]
                            index--
                            continue
                        }
                    }
                }
            case kanjiMode:
                switch preMode {
                case byteMode:
                    if !disableMerge && ((preHead + encodedLen(preData+curData, preMode)) <= (preSize + curSize)) {
                        (*segments)[index-1].data = append((*segments)[index-1].data, (*segments)[index].data...)
                        *segments = (*segments)[0 : len(*segments)-1]
                    }
                }
            }
        }
        return
    }
}

func headerBits(kind QRCodeKind, version QRCodeVersion, mode encodeMode) int {
    seg, _, _ := getSegmentIndicator(kind, version, mode, 0)
    return int(seg.modeBits + seg.countBits)
}

func countModeChars(data []byte, mode encodeMode, index int) int {
    result := 0
    if index >= len(data) {
        return 0
    }
    kanji := false
    for i := index; i < len(data); i++ {
        if kanji {
            result++
            kanji = false
            continue
        }
        ch := data[i]
        switch mode {
        case numericMode:
            if ch >= '0' && ch <= '9' {
                result++
            } else {
                return result
            }
        case alphaNumericMode:
            if (ch >= 'A' && ch <= 'Z') || ch == ' ' || ch == '$' || ch == '%' || ch == '*' || ch == '+' ||
                ch == '-' || ch == '.' || ch == '/' || ch == ':' {
                result++
            } else {
                return result
            }
        case byteMode:
            if ch <= 31 || ch >= 33 && ch <= 35 || ch >= 38 && ch <= 41 || ch == 44 || ch >= 59 && ch <= 64 ||
                ch >= 91 && ch <= 128 {
                result++
            } else if ch > 128 {
                if QRCodeEnableKanji && (i < len(data)-1) {
                    wd := (uint16(ch) << 8) | uint16(data[i+1])
                    if (wd >= 0x8140 && wd <= 0x9ffc) || (wd >= 0xe040 && wd <= 0xebbf) {
                        return result
                    }
                }
                result++
            } else {
                return result
            }
        case kanjiMode:
            if i < len(data)-1 {
                wd := (uint16(ch) << 8) | uint16(data[i+1])
                if (wd >= 0x8140 && wd <= 0x9ffc) || (wd >= 0xe040 && wd <= 0xebbf) {
                    result++
                    kanji = true
                } else {
                    return result
                }
            } else {
                return result
            }
        }
    }
    return result
}

func detectCharMode(data []byte, index int) encodeMode {
    if index >= len(data) {
        return nullEncode
    }
    ch := data[index]
    if ch >= '0' && ch <= '9' {
        return numericMode
    } else if (ch >= 'A' && ch <= 'Z') || ch == ' ' || ch == '$' || ch == '%' || ch == '*' || ch == '+' ||
        ch == '-' || ch == '.' || ch == '/' || ch == ':' {
        return alphaNumericMode
    } else if ch <= 31 || ch <= 35 || ch <= 41 || (ch == 44) || ch <= 64 || ch <= 128 {
        return byteMode
    } else {
        if QRCodeEnableKanji && (index < len(data)-1) {
            wd := (uint16(ch) << 8) | uint16(data[index+1])
            if (wd >= 0x8140 && wd <= 0x9ffc) || (wd >= 0xe040 && wd <= 0xebbf) {
                return kanjiMode
            }
        }
        return byteMode
    }
}

func encodedLen(count int, mode encodeMode) int {
    switch mode {
    case numericMode:
        return 10*(count/3) + []int{0, 4, 7}[count%3]
    case alphaNumericMode:
        return 11*(count/2) + 6*(count%2)
    case byteMode:
        return count * 8
    case kanjiMode:
        return (count / 2) * 13
    }
    return 0
}

func optimizeEncode(data []byte, kind QRCodeKind, version QRCodeVersion) ([]segmentDivision, error, bool) {
    if len(data) == 0 {
        return []segmentDivision{{mode: numericMode, head: headerBits(kind, version, numericMode)}}, nil, false
    } else {
        result := make([]segmentDivision, 0)
        curMode := nullEncode
        index := 0
        for index < len(data) {
            nxtMode := detectCharMode(data, index)
            nxtCount := countModeChars(data, nxtMode, index)
            nxtSize := headerBits(kind, version, nxtMode) + encodedLen(nxtCount, nxtMode)
            sucMode := nullEncode
            if index+nxtCount < len(data) {
                sucMode = detectCharMode(data, index+nxtCount)
            }
            newMode := nxtMode
            disableMerge := false
            switch curMode {
            case nullEncode, kanjiMode:
                switch nxtMode {
                case numericMode:
                    if sucMode == alphaNumericMode || sucMode == byteMode {
                        if nxtSize >= encodedLen(nxtCount, sucMode) {
                            newMode = sucMode
                        }
                    }
                case alphaNumericMode, kanjiMode:
                    if sucMode == byteMode {
                        if nxtSize >= encodedLen(nxtCount, sucMode) {
                            newMode = sucMode
                        }
                    }
                }
            case numericMode:
                if (nxtMode == alphaNumericMode || nxtMode == byteMode) && (sucMode == byteMode) {
                    if nxtSize >= encodedLen(nxtCount, sucMode) {
                        newMode = sucMode
                    }
                }
            case alphaNumericMode:
                switch nxtMode {
                case numericMode:
                    if sucMode == alphaNumericMode {
                        if nxtSize+headerBits(kind, version, sucMode) >= encodedLen(nxtCount, sucMode) {
                            newMode = sucMode
                        }
                    } else {
                        if nxtSize >= encodedLen(nxtCount, alphaNumericMode) {
                            newMode = alphaNumericMode
                        }
                    }
                }
            case byteMode:
                switch nxtMode {
                case numericMode:
                    switch sucMode {
                    case alphaNumericMode:
                        if nxtSize >= encodedLen(nxtCount, byteMode) {
                            newMode = byteMode
                        }
                        disableMerge = true
                    case byteMode:
                        if (nxtSize + headerBits(kind, version, byteMode)) >= encodedLen(nxtCount, byteMode) {
                            newMode = byteMode
                        }
                    default:
                        if nxtSize >= encodedLen(nxtCount, byteMode) {
                            newMode = byteMode
                        }
                    }
                case alphaNumericMode:
                    switch sucMode {
                    case numericMode:
                        if nxtSize >= encodedLen(nxtCount, byteMode) {
                            newMode = byteMode
                        }
                        disableMerge = true
                    case byteMode:
                        if (nxtSize + headerBits(kind, version, byteMode)) >= encodedLen(nxtCount, byteMode) {
                            newMode = byteMode
                        }
                    default:
                        if nxtSize >= encodedLen(nxtCount, byteMode) {
                            newMode = byteMode
                        }
                    }
                case kanjiMode:
                    switch sucMode {
                    case byteMode:
                        if nxtSize+headerBits(kind, version, byteMode) >= encodedLen(nxtCount, byteMode) {
                            newMode = byteMode
                        }
                    default:
                        if nxtSize >= encodedLen(nxtCount, byteMode) {
                            newMode = byteMode
                        }
                    }
                }
            }
            appendSegment(&result, segmentDivision{mode: newMode, head: headerBits(kind, version, newMode),
                data: data[index : index+nxtCount]}, disableMerge)
            index += nxtCount
            curMode = newMode
        }
        appendSegment(&result, segmentDivision{mode: nullEncode}, false)
        return result, nil, false
    }
}

func splitBarcode(data []byte, kind QRCodeKind, version QRCodeVersion, escape bool) ([]segmentDivision, error, bool) {
    for i, ch := range data {
        if kind == MicroQRCode {
            valid := false
            switch version {
            case 1:
                valid = ch >= 48 && ch <= 57
            case 2:
                valid = (ch >= 48 && ch <= 57) || (ch >= 65 && ch <= 90) || ch == 32 || ch == 36 || ch == 37 ||
                    ch == 42 || ch == 43 || ch == 45 || ch == 46 || ch == 47 || ch == 58
            default:
                valid = true
            }
            if !valid {
                return nil, fmt.Errorf("character is not supported by MicroQRCode version %d: %d: %s", version, i+1,
                    string(ch)), true
            }
        }
    }
    if !escape {
        return optimizeEncode(data, kind, version)
    }
    result := make([]segmentDivision, 0)
    mode := ""
    eci := 0
    saIndex := 0
    saTotal := 0
    saParity := 0
    var cur []byte
    for i := 0; i < len(data); i++ {
        switch mode {
        default:
            if data[i] == '\\' {
                mode = "wait"
            } else {
                cur = append(cur, data[i])
            }
        case "wait":
            switch data[i] {
            case '\\':
                cur = append(cur, data[i])
                mode = ""
            case 'g':
                cur = append(cur, 29)
                mode = ""
            case 'f':
                if len(cur) != 0 {
                    if s, e, u := optimizeEncode(cur, kind, version); e != nil {
                        return nil, e, u
                    } else {
                        result = append(result, s...)
                    }
                }
                result = append(result, segmentDivision{mode: fnc1Mode1})
                cur = []byte{}
                mode = ""
            case 'd':
                mode = "fnc0"
            case 'e':
                mode = "eci0"
                eci = 0
            case 's':
                if i != 1 {
                    return nil, fmt.Errorf(
                        "structured append must be located at the beginning of the barcode text: %d: %s",
                        i+1, string(data[i])), false
                }
                mode = "sa0"
                saIndex = -1
                saTotal = -1
                saParity = -1
            default:
                return nil, fmt.Errorf("invalid charactre in escape sequence: %d: %s", i+1, string(data[i])), false
            }
        case "fnc0":
            if (data[i] >= 'A' && data[i] <= 'Z') || (data[i] >= 'a' && data[i] <= 'z') {
                if len(cur) != 0 {
                    if s, e, u := optimizeEncode(cur, kind, version); e != nil {
                        return nil, e, u
                    } else {
                        result = append(result, s...)
                    }
                }
                result = append(result, segmentDivision{mode: fnc1Mode2, data: data[i : i+1]})
                cur = []byte{}
                mode = ""
            } else if data[i] >= '0' && data[i] <= '9' {
                mode = "fnc1"
            } else {
                return nil, fmt.Errorf("invalid character in GS1 AI: %d: %s", i+1, string(data[i])), false
            }
        case "fnc1":
            if data[i] >= '0' && data[i] <= '9' {
                if len(cur) != 0 {
                    if s, e, u := optimizeEncode(cur, kind, version); e != nil {
                        return nil, e, u
                    } else {
                        result = append(result, s...)
                    }
                }
                result = append(result, segmentDivision{mode: fnc1Mode2, data: data[i-1 : i+1]})
                cur = []byte{}
                mode = ""
            } else {
                return nil, fmt.Errorf("invalid character in GS1 AI: %d: %s", i+1, string(data[i])), false
            }
        case "eci0":
            if data[i] == '[' {
                eci = -1
                mode = "eci1"
            } else {
                return nil, fmt.Errorf("invalid character in ECI parameter: %d: %s", i+1, string(data[i])), false
            }
        case "eci1":
            if data[i] == ']' {
                if eci == -1 {
                    return nil, fmt.Errorf("missing ECI value: %d: %s", i+1, string(data[i])), false
                }
                if len(cur) != 0 {
                    if s, e, u := optimizeEncode(cur, kind, version); e != nil {
                        return nil, e, u
                    } else {
                        result = append(result, s...)
                    }
                }
                result = append(result, segmentDivision{mode: eciMode, data: []byte(strconv.Itoa(eci))})
                cur = []byte{}
                mode = ""
            } else if data[i] >= '0' && data[i] <= '9' {
                if eci == -1 {
                    eci = int(data[i]) - 48
                } else {
                    eci = eci*10 + int(data[i]) - 48
                }
                if eci > 999999 {
                    return nil, fmt.Errorf("invalid ECI value: %d: %d: %s", eci, i+1, string(data[i])), false
                }
            } else {
                return nil, fmt.Errorf("invalid character in ECI value: %d: %s", i+1, string(data[i])), false
            }
        case "sa0":
            if data[i] == '[' {
                mode = "sa1"
            } else {
                return nil, fmt.Errorf("invalid character in Structured Append parameter: %d: %s", i+1,
                    string(data[i])), false
            }
        case "sa1":
            if data[i] == ',' {
                if saIndex == -1 {
                    return nil, fmt.Errorf("missing index value of Structured Append: %d: %s", i+1, string(data[i])),
                        false
                }
                mode = "sa2"
            } else if data[i] >= '0' && data[i] <= '9' {
                if saIndex == -1 {
                    saIndex = int(data[i]) - 48
                } else {
                    saIndex = saIndex*10 + int(data[i]) - 48
                }
                if saIndex > 16 || saIndex < 1 {
                    return nil, fmt.Errorf(
                        "the index value %d of Structured Append must be between 1 and 16 (inclusive): %d: %s", saIndex,
                        i+1, string(data[i])), false
                }
            } else {
                return nil, fmt.Errorf("invalid character in index value of Structured Append: %d: %s", i+1,
                    string(data[i])), false
            }
        case "sa2":
            if data[i] == ',' {
                if saTotal == -1 {
                    return nil, fmt.Errorf("missing total count of Structured Append: %d: %s", i+1, string(data[i])),
                        false
                }
                mode = "sa3"
            } else if data[i] >= '0' && data[i] <= '9' {
                if saTotal == -1 {
                    saTotal = int(data[i]) - 48
                } else {
                    saTotal = saTotal*10 + int(data[i]) - 48
                }
                if saTotal > 16 || saTotal < 1 {
                    return nil, fmt.Errorf(
                        "the total count %d of Structured Append must be between 1 and 16 (inclusive): %d: %s",
                        saTotal, i+1, string(data[i])), false
                } else if saIndex > saTotal {
                    return nil, fmt.Errorf("the index value %d of Structured Append exceeds its total count %d: %d: %s",
                        saIndex, saTotal, i+1, string(data[i])), false
                }
            } else {
                return nil, fmt.Errorf("invalid character in total count of Structured Append: %d: %s", i+1,
                    string(data[i])), false
            }
        case "sa3":
            if data[i] == ']' {
                if saParity == -1 {
                    return nil, fmt.Errorf("missing parity value of Structured Append: %d: %s", i+1, string(data[i])),
                        false
                }
                if len(cur) != 0 {
                    if s, e, u := optimizeEncode(cur, kind, version); e != nil {
                        return nil, e, u
                    } else {
                        result = append(result, s...)
                    }
                }
                result = append(result, segmentDivision{mode: eciMode, data: []byte(fmt.Sprintf("%d,%d,%d", saIndex,
                    saParity, saParity))})
                cur = []byte{}
                mode = ""
            } else if data[i] >= '0' && data[i] <= '9' {
                if saParity == -1 {
                    saParity = int(data[i]) - 48
                } else {
                    saParity = saParity*10 + int(data[i]) - 48
                }
                if saParity < 0 || saParity > 255 {
                    return nil, fmt.Errorf(
                        "the parity value %d of Structured Append must be between 0 and 255 (inclusive): %d: %s",
                        saParity, i+1, string(data[i])), false
                }
            } else {
                return nil, fmt.Errorf("invalid character in parity value of Structured Append: %d: %s", i+1,
                    string(data[i])), false
            }
        }
    }
    if len(cur) != 0 {
        if s, e, u := optimizeEncode(cur, kind, version); e != nil {
            return nil, e, u
        } else {
            result = append(result, s...)
        }
    }
    return result, nil, false
}

func encodeNumeric(data []byte) encodedSegment {
    count := len(data)
    result := encodedSegment{mode: numericMode, count: count, data: make([]uint32, 0)}
    for i := 0; i < count/3; i++ {
        result.append(uint32(data[i*3]-48)*100+uint32(data[i*3+1]-48)*10+uint32(data[i*3+2]-48), 10)
    }
    last := count - 1
    switch count % 3 {
    case 1:
        result.append(uint32(data[last]-48), 4)
    case 2:
        result.append(uint32(data[last-1]-48)*10+uint32(data[last]-48), 7)
    }
    return result
}

func alphaNumericValue(ch byte) uint8 {
    switch {
    case ch >= '0' && ch <= '9':
        return ch - '0'
    case ch >= 'A' && ch <= 'Z':
        return ch - 'A' + 10
    case ch == ' ':
        return 36
    case ch == '$':
        return 37
    case ch == '%':
        return 38
    case ch == '*':
        return 39
    case ch == '+':
        return 40
    case ch == '-':
        return 41
    case ch == '.':
        return 42
    case ch == '/':
        return 43
    case ch == ':':
        return 44
    }
    return 0
}

func encodeAlphaNumeric(data []byte) encodedSegment {
    count := len(data)
    result := encodedSegment{mode: alphaNumericMode, count: count, data: make([]uint32, 0)}
    for i := 0; i < count/2; i++ {
        result.append(uint32(alphaNumericValue(data[i*2]))*45+uint32(alphaNumericValue(data[i*2+1])), 11)
    }
    if count%2 > 0 {
        result.append(uint32(alphaNumericValue(data[count-1])), 6)
    }
    return result
}

func encodeByte(data []byte) encodedSegment {
    count := len(data)
    result := encodedSegment{mode: byteMode, count: count, data: make([]uint32, 0)}
    for i := 0; i < count; i++ {
        result.append(uint32(data[i]), 8)
    }
    return result
}

func encodeKanji(data []byte) encodedSegment {
    count := len(data)
    result := encodedSegment{mode: kanjiMode, count: count, data: make([]uint32, 0)}
    for i := 0; i < count/2; i++ {
        kanji := (uint32(data[i*2]) << 8) | uint32(data[i*2+1])
        if kanji >= 0x8140 && kanji <= 0x9ffc {
            kanji -= 0x8140
        } else if kanji >= 0xe040 && kanji <= 0xebbf {
            kanji -= 0xc140
        }
        result.append((kanji>>8)*0xc0+(kanji&0xff), 13)
    }
    return result
}

func encodeFNC1Mode1() encodedSegment {
    return encodedSegment{mode: fnc1Mode1, count: 0, data: []uint32{}}
}

func encodeFNC1Mode2(data []byte) encodedSegment {
    result := encodedSegment{mode: fnc1Mode2, count: 0, data: make([]uint32, 0)}
    if len(data) == 1 {
        result.append(uint32(data[0]+100), 8)
    } else {
        ai, _ := strconv.Atoi(string(data))
        result.append(uint32(ai), 8)
    }
    return result
}

func encodeECI(data []byte) encodedSegment {
    result := encodedSegment{mode: eciMode, count: 0, data: make([]uint32, 0)}
    eci, _ := strconv.ParseInt(string(data), 10, 32)
    if eci <= 127 {
        result.append(uint32(eci), 8)
    } else if eci <= 16383 {
        result.append(uint32(eci>>8)|128, 8)
        result.append(uint32(eci&0xff), 8)
    } else if eci <= 999999 {
        result.append(uint32(eci>>16)|192, 8)
        result.append(uint32(eci>>8)&0xff, 8)
        result.append(uint32(eci&0xff), 8)
    }
    return result
}

func encodeStructAppend(data []byte) encodedSegment {
    result := encodedSegment{mode: structAppendMode, count: 0, data: make([]uint32, 0)}
    str := string(data)
    items := strings.SplitN(str[3:len(str)-1], ",", 3)
    index, _ := strconv.Atoi(items[0])
    total, _ := strconv.Atoi(items[0])
    parity, _ := strconv.Atoi(items[0])
    result.append(uint32((index-1)<<4)|uint32(total-1), 8)
    result.append(uint32(parity), 8)
    return result
}

func encodeData(data []byte, kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel,
    escape bool) ([]byte, error, bool, QRCodeEccLevel) {
    if divisions, err, up := splitBarcode(data, kind, version, escape); err != nil {
        return nil, err, up, ecc
    } else {
        var bitGroups []uint32
        for _, segDiv := range divisions {
            var segment encodedSegment
            switch segDiv.mode {
            case numericMode:
                segment = encodeNumeric(segDiv.data)
            case alphaNumericMode:
                segment = encodeAlphaNumeric(segDiv.data)
            case kanjiMode:
                segment = encodeKanji(segDiv.data)
            case byteMode:
                segment = encodeByte(segDiv.data)
            case eciMode:
                segment = encodeECI(segDiv.data)
            case fnc1Mode1:
                segment = encodeFNC1Mode1()
            case fnc1Mode2:
                segment = encodeFNC1Mode2(segDiv.data)
            case structAppendMode:
                segment = encodeStructAppend(segDiv.data)
            }
            if header, err, up := getSegmentIndicator(kind, version, segment.mode, uint16(segment.count)); err != nil {
                return nil, err, up, ecc
            } else {
                bitGroups = append(bitGroups, (uint32(header.modeBits)<<16)|uint32(header.modeData))
                if header.countData > 0 {
                    bitGroups = append(bitGroups, (uint32(header.countBits)<<16)|uint32(header.countData))
                }
                bitGroups = append(bitGroups, segment.data...)
            }
        }
        currBits := 0
        for _, code := range bitGroups {
            currBits += int(code>>16) & 0xff
        }
        switch kind {
        case QRCode:
            for _, ec := range []QRCodeEccLevel{QREccHighest, QREccQuality, QREccMedium, QREccLowest} {
                if ec == ecc {
                    break
                } else if totalDataBits(kind, version, ec) >= currBits {
                    ecc = ec
                    break
                }
            }
        case MicroQRCode:
            switch version {
            case 2, 3:
                if ecc == QREccLowest && totalDataBits(kind, version, QREccMedium) >= currBits {
                    ecc = QREccMedium
                }
            case 4:
                for _, ec := range []QRCodeEccLevel{QREccQuality, QREccMedium, QREccLowest} {
                    if ec == ecc {
                        break
                    } else if totalDataBits(kind, version, ec) >= currBits {
                        ecc = ec
                        break
                    }
                }
            }
        case RMQRCode:
            if ecc == QREccMedium && totalDataBits(kind, version, QREccHighest) >= currBits {
                ecc = QREccHighest
            }
        }
        dataBits := totalDataBits(kind, version, ecc)
        if currBits > dataBits {
            return nil, fmt.Errorf("the barcode data is too large to be contained within the symbol"), true, ecc
        }
        termBits := min(dataBits-currBits, terminateBits(kind, version))
        if termBits > 0 {
            bitGroups = append(bitGroups, uint32(termBits<<16))
        }
        currBits += termBits
        codes := make([]byte, totalDataWords(kind, version, ecc))
        rBits := 8
        index := 0
        for _, code := range bitGroups {
            cData := code & 0xffff
            cBits := int(code >> 16)
            for cBits > 0 {
                if rBits > cBits {
                    codes[index] = (codes[index] << cBits) | byte(cData)
                    rBits -= cBits
                    cBits = 0
                } else {
                    cBits -= rBits
                    codes[index] = (codes[index] << rBits) | byte(cData>>cBits)
                    if cBits > 0 {
                        cData = cData & ((1 << cBits) - 1)
                    }
                    index++
                    rBits = 8
                }
            }
        }
        if rBits != 8 {
            codes[index] = codes[index] << rBits
        }
        padding := byte(0xec)
        for i := index + 1; i < len(codes); i++ {
            codes[i] = padding
            padding = padding ^ 0xfd
        }
        var dataBlocks [][]byte
        define := eccDefines[kind][version-1][ecc]
        src := 0
        for i := 0; i < define.count; i++ {
            dataBlock := codes[src : src+define.data]
            src += define.data
            dataBlocks = append(dataBlocks, dataBlock)
        }
        for i := 0; i < define.next; i++ {
            dataBlock := codes[src : src+define.data+1]
            src += define.data + 1
            dataBlocks = append(dataBlocks, dataBlock)
        }
        var eccBlocks [][]byte
        for i := 0; i < define.count+define.next; i++ {
            eccBlocks = append(eccBlocks, rsEncode(dataBlocks[i], define.total-define.data))
        }
        result := make([]byte, totalWords(kind, version))
        index = 0
        for i := 0; i < define.data; i++ {
            for j := 0; j < define.count+define.next; j++ {
                result[index] = dataBlocks[j][i]
                index++
            }
        }
        for j := 0; j < define.next; j++ {
            result[index] = dataBlocks[define.count+j][define.data]
            index++
        }
        for i := 0; i < define.total-define.data; i++ {
            for j := 0; j < define.count+define.next; j++ {
                result[index] = eccBlocks[j][i]
                index++
            }
        }
        return result, nil, false, ecc
    }
}

// EncodeGS1 : Encode a GS1 barcode string in order to generate GS1 barcode (to be used as the barcode data for QRCode
// or rMQRCode).
//
// The `escape` parameter must be set to `true` when using this function.
//
// Note: When inserting `"("` into a GS1 AI element value, please use `"(("`.
func EncodeGS1(barcode string) ([]byte, error) {
    result := "\\f"
    inAi := false
    ignore := false
    ai := ""
    value := ""
    for i, c := range barcode {
        if inAi {
            switch c {
            case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
                ai += string(c)
            case ')':
                inAi = false
                value = ""
            default:
                return nil, fmt.Errorf("invalid GS1 barcode character: %d: %s", i+1, string(c))
            }
        } else {
            switch c {
            case '(':
                if ignore {
                    ignore = false
                    continue
                }
                if i < len(barcode)-1 {
                    if barcode[i+1] == '(' {
                        value += string(c)
                        ignore = true
                    } else {
                        if ai != "" {
                            result += ai + value
                            if !(strings.HasPrefix(ai, "31") || strings.HasPrefix(ai, "32") ||
                                strings.HasPrefix(ai, "33") || strings.HasPrefix(ai, "34") ||
                                strings.HasPrefix(ai, "35") || strings.HasPrefix(ai, "36") ||
                                strings.HasPrefix(ai, "41") ||
                                slices.Contains([]string{"00", "01", "02", "03", "11", "12", "13", "15", "16", "17",
                                    "20"}, ai)) {
                                result += "\\g"
                            }
                            ai = ""
                            value = ""
                        }
                        inAi = true
                    }
                } else {
                    return nil, fmt.Errorf("invalid GS1 barcode character: %d: %s", i+1, string(c))
                }
            default:
                if ai == "" {
                    return nil, fmt.Errorf("invalid GS1 barcode character: %d: %s", i+1, string(c))
                } else if c == '\\' {
                    value += "\\\\"
                } else {
                    value += string(c)
                }
            }
        }
    }
    if ai != "" {
        result += ai + value
    } else if strings.HasSuffix(result, "\\g") {
        result = result[:len(result)-2]
    }
    return []byte(result), nil
}

// GetParity : Calculate the parity for the entire barcode to create a series of structured append barcode symbols.
func GetParity(data []byte, escape bool) (result byte, err error) {
    if escape {
        mode := ""
        eci := 0
        for i := 0; i < len(data); i++ {
            switch mode {
            default:
                if data[i] == '\\' {
                    mode = "wait"
                } else {
                    result ^= data[i]
                }
            case "wait":
                switch data[i] {
                case '\\':
                    result ^= '\\'
                    mode = ""
                case 'g':
                    result ^= 29
                    mode = ""
                case 'f':
                    mode = ""
                case 'd':
                    mode = "fnc0"
                case 'e':
                    mode = "eci0"
                    eci = 0
                default:
                    err = fmt.Errorf("invalid character in escape sequence: %d: %s", i+1, string(data[i]))
                    return
                }
            case "fnc0":
                if (data[i] >= 'A' && data[i] <= 'Z') || (data[i] >= 'a' && data[i] <= 'z') {
                    mode = ""
                } else if data[i] >= '0' && data[i] <= '9' {
                    mode = "fnc1"
                } else {
                    err = fmt.Errorf("invalid character in GS1 AI: %d: %s", i+1, string(data[i]))
                    return
                }
            case "fnc1":
                if data[i] >= '0' && data[i] <= '9' {
                    mode = ""
                } else {
                    err = fmt.Errorf("invalid character in GS1 AI: %d: %s", i+1, string(data[i]))
                    return
                }
            case "eci0":
                if data[i] == '[' {
                    eci = -1
                    mode = "eci1"
                } else {
                    err = fmt.Errorf("invalid character in ECI parameter %d: %s", i+1, string(data[i]))
                    return
                }
            case "eci1":
                if data[i] == ']' {
                    mode = ""
                    if eci == -1 {
                        err = fmt.Errorf("missing ECI value: %d: %s", i+1, string(data[i]))
                        return
                    }
                } else if data[i] >= '0' && data[i] <= '9' {
                    if eci == -1 {
                        eci = int(data[i]) - 48
                    } else {
                        eci = eci*10 + int(data[i]) - 48
                    }
                    if eci > 999999 {
                        err = fmt.Errorf("invalid ECI value %d: %d: %s", eci, i+1, string(data[i]))
                        return
                    }
                } else {
                    err = fmt.Errorf("invalid character in ECI value: %d: %s", i+1, string(data[i]))
                    return
                }
            }
        }
    } else {
        for i := 0; i < len(data); i++ {
            result ^= data[i]
        }
    }
    return
}
