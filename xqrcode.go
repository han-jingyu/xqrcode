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

// QRCodeFormat : QRCode output format.
type QRCodeFormat string

// QRCode QRCode output format values.
const (
    QRCodeFormatSvg QRCodeFormat = "svg"
    QRCodeFormatPng QRCodeFormat = "png"
)

// QRCodeKind : QRCode kind.
type QRCodeKind string

// QRCode types.
const (
    MicroQRCode QRCodeKind = "MicroQRCode"
    RMQRCode    QRCodeKind = "rMQRCode"
    QRCode      QRCodeKind = "QRCode"
)

// QRCodeVersion : Version (MicroQRCode: 1-4; rMQRCode: 1-32; QRCode: 1-40).
type QRCodeVersion = int

// QRCodeEccLevel : Error correction level (MicroQRCode: L, M, Q; rMQRCode: M, H; QRCode: L, M, Q, H).
type QRCodeEccLevel string

// QRCode error correction levels.
const (
    QREccLowest  QRCodeEccLevel = "L"
    QREccMedium  QRCodeEccLevel = "M"
    QREccQuality QRCodeEccLevel = "Q"
    QREccHighest QRCodeEccLevel = "H"
)

// QRCodeEnableKanji : Determines whether the Kanji encoding mode is allowed. Enabled by default.
var QRCodeEnableKanji = true

// DrawQRCode1 : Draw barcode by drawing rectangles within images (such as PNG or SVG).
//
// The `left` parameter specifies the left margin for the left edge of the barcode, or the left margin for the right
// edge when `rightAlign` is true. This includes the quiet zone.
//
// The `top` parameter specifies the top margin for the top edge of the barcode, or the top margin for the bottom edge
// when `bottomAlign` is true. This includes the quiet zone.
//
// The color values for the `background` and `foreground` parameters use the `#RRGGBB` or `#RRGGBBAA` format.
//
// When `escape` is true, the following escape sequences can be inserted:
//
//   - \\: A backslash character
//   - \f: FNC1 in first position
//   - \dN: FNC1 in second position, where N is A-Z, a-z, or 00-99
//   - \e[N]: Extended Channel Interpretation (ECI) designator, where N is a number ranging from 0 to 999999
//   - \g: A GS character (ASCII 29)
//   - \s[INDEX,TOTAL,PARITY]: Structured Append, Where INDEX and TOTAL is a number ranging from 1 to 16, PARITY is a
//     number ranging from 1 to 255, use function `GetParity` to obtain the parity value of the barcode data.
//
// The `drawRect` parameter specifies a function used to draw rectangles within the image; when the overall background
// is being drawn, its `background` parameter is set to true.
//
// This function returns the width and height of the barcode symbol. When an error occurs, the final return value
// indicates whether the error can be resolved by upgrading the barcode version.
func DrawQRCode1(data []byte, kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel, quietZone int,
    escape, mirror, invert bool, background, foreground string, module, left, top float64, rightAlign, bottomAlign bool,
    drawRect func(left, top, width, height float64, r, g, b, a uint8, background bool) error) (float64, float64, error,
    bool) {
    if module <= 0 {
        return 0, 0, fmt.Errorf("invalid module value: %f", module), false
    }
    if quietZone < 0 {
        return 0, 0, fmt.Errorf("invalid quiet zone value: %d", quietZone), false
    }
    if matrix, err, up := buildMatrix(data, kind, version, ecc, escape, mirror); err != nil {
        return 0, 0, err, up
    } else {
        hMargin := float64(quietZone) * module
        vMargin := hMargin
        height := float64(quietZone*2+len(*matrix)) * module
        width := float64(quietZone*2+len((*matrix)[0])) * module
        if bottomAlign {
            top -= height
        }
        if rightAlign {
            left -= width
        }
        if invert {
            tmpColor := background
            background = foreground
            foreground = tmpColor
        }
        rgba := ColorHex2RGBA(background, color.RGBA{R: 255, G: 255, B: 255, A: 255})
        if e := drawRect(left, top, width, height, rgba.R, rgba.G, rgba.B, rgba.A, true); e != nil {
            return 0, 0, e, false
        }
        rgba = ColorHex2RGBA(foreground, color.RGBA{R: 0, G: 0, B: 0, A: 255})
        top += vMargin
        left += hMargin
        rows := len(*matrix)
        cols := len((*matrix)[0])
        fromY := top
        for i := 0; i < rows; i++ {
            fromX := left
            count := 0
            state := false
            for j := 0; j < cols; j++ {
                if ((*matrix)[i][j] & 1) > 0 {
                    if state {
                        count++
                    } else {
                        if count > 0 {
                            fromX += float64(count) * module
                        }
                        state = true
                        count = 1
                    }
                } else {
                    if state {
                        if count > 0 {
                            dX := float64(count) * module
                            if e := drawRect(fromX, fromY, dX, module, rgba.R, rgba.G, rgba.B, rgba.A,
                                false); e != nil {
                                return 0, 0, e, false
                            }
                            fromX += dX
                        }
                        state = false
                        count = 1
                    } else {
                        count++
                    }
                }
            }
            if count > 0 && state {
                dX := float64(count) * module
                if e := drawRect(fromX, fromY, dX, module, rgba.R, rgba.G, rgba.B, rgba.A, false); e != nil {
                    return 0, 0, e, false
                }
            }
            fromY += module
        }
        return width, height, nil, false
    }
}

// DrawQRCode2 : Draw barcodes by setting individual pixels in raster images, such as PNGs.
//
// The `left` parameter specifies the left margin for the left edge of the barcode, or the left margin for the right
// edge when `rightAlign` is true. This includes the quiet zone.
//
// The `top` parameter specifies the top margin for the top edge of the barcode, or the top margin for the bottom edge
// when `bottomAlign` is true. This includes the quiet zone.
//
// The color values for the `background` and `foreground` parameters use the `#RRGGBB` or `#RRGGBBAA` format.
//
// When `escape` is true, the following escape sequences can be inserted:
//
//   - \\: A backslash character
//   - \f: FNC1 in first position
//   - \dN: FNC1 in second position, where N is A-Z, a-z, or 00-99
//   - \e[N]: Extended Channel Interpretation (ECI) designator, where N is a number ranging from 0 to 999999
//   - \g: A GS character (ASCII 29)
//   - \s[INDEX,TOTAL,PARITY]: Structured Append, Where INDEX and TOTAL is a number ranging from 1 to 16, PARITY is a
//     number ranging from 1 to 255, use function `GetParity` to obtain the parity value of the barcode data.
//
// The `drawDot` parameter specifies a function used to set a pixel within the image; when the overall background is
// being drawn, its `background` parameter is set to true.
//
// This function returns the width and height of the barcode symbol. When an error occurs, the final return value
// indicates whether the error can be resolved by upgrading the barcode version.
func DrawQRCode2(data []byte, left, top int, kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel,
    module, quietZone int, background, foreground string, escape, mirror, invert, rightAlign, bottomAlign bool,
    drawDot func(left, top int, r, g, b, a uint8, background bool) error) (int, int, error, bool) {
    if module <= 0 {
        return 0, 0, fmt.Errorf("invalid module value: %d", module), false
    }
    if quietZone < 0 {
        return 0, 0, fmt.Errorf("invalid quiet zone value: %d", quietZone), false
    }
    if matrix, err, up := buildMatrix(data, kind, version, ecc, escape, mirror); err != nil {
        return 0, 0, err, up
    } else {
        height := (len(*matrix) + quietZone*2) * module
        width := (len((*matrix)[0]) + quietZone*2) * module
        if bottomAlign {
            top -= height
        }
        if rightAlign {
            left -= width
        }
        if invert {
            tmpColor := foreground
            foreground = background
            background = tmpColor
        }
        if e := drawQRDots(matrix, left, top, module, quietZone, background, foreground, drawDot); e != nil {
            return width, height, e, false
        }
        return width, height, nil, false
    }
}

// QRCodeToFile : Generate a barcode image and write it to a file.
//
// The color values for the `background` and `foreground` parameters use the `#RRGGBB` or `#RRGGBBAA` format.
//
// When `escape` is true, the following escape sequences can be inserted:
//
//   - \\: A backslash character
//   - \f: FNC1 in first position
//   - \dN: FNC1 in second position, where N is A-Z, a-z, or 00-99
//   - \e[N]: Extended Channel Interpretation (ECI) designator, where N is a number ranging from 0 to 999999
//   - \g: A GS character (ASCII 29)
//   - \s[INDEX,TOTAL,PARITY]: Structured Append, Where INDEX and TOTAL is a number ranging from 1 to 16, PARITY is a
//     number ranging from 1 to 255, use function `GetParity` to obtain the parity value of the barcode data.
//
// The `output` parameter specifies the target file path.
//
// When an error occurs, the final return value indicates whether the error can be resolved by upgrading the barcode
// version.
func QRCodeToFile(data []byte, kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel, module, quietZone int,
    background, foreground string, escape, mirror, invert bool, format QRCodeFormat, output string) (error, bool) {
    if img, err, up := qrCodeToImage(data, kind, version, ecc, module, quietZone, background, foreground, escape,
        mirror, invert, format); err != nil {
        return err, up
    } else {
        return os.WriteFile(output, img, 0644), false
    }
}

// QRCodeToHTTP : Generate a barcode images and respond to HTTP requests.
//
// The color values for the `background` and `foreground` parameters use the `#RRGGBB` or `#RRGGBBAA` format.
//
// When `escape` is true, the following escape sequences can be inserted:
//
//   - \\: A backslash character
//   - \f: FNC1 in first position
//   - \dN: FNC1 in second position, where N is A-Z, a-z, or 00-99
//   - \e[N]: Extended Channel Interpretation (ECI) designator, where N is a number ranging from 0 to 999999
//   - \g: A GS character (ASCII 29)
//   - \s[INDEX,TOTAL,PARITY]: Structured Append, Where INDEX and TOTAL is a number ranging from 1 to 16, PARITY is a
//     number ranging from 1 to 255, use function `GetParity` to obtain the parity value of the barcode data.
//
// This function returns the number of response bytes. When an error occurs, the final return value indicates whether
// the error can be resolved by upgrading the barcode version.
func QRCodeToHTTP(data []byte, kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel, module, quietZone int,
    background, foreground string, escape, mirror, invert bool, format QRCodeFormat, w http.ResponseWriter) (int64,
    error, bool) {
    if img, err, up := qrCodeToImage(data, kind, version, ecc, module, quietZone, background, foreground, escape,
        mirror, invert, format); err != nil {
        return 0, err, up
    } else {
        size := len(img)
        if format == "png" {
            w.Header().Set("Content-Type", "image/png")
        } else if format == "svg" {
            w.Header().Set("Content-Type", "image/svg+xml")
        }
        w.Header().Set("Content-Length", strconv.Itoa(size))
        w.Header().Set("Cache-Control", "no-store")
        w.WriteHeader(http.StatusOK)
        if n, err := w.Write(img); err != nil {
            return int64(n), err, false
        } else {
            return int64(n), nil, false
        }
    }
}

// SprintQRCode : Output a barcode to the command-line terminal.
//
// The `color` parameter can be used to specify the foreground and(or) background colors of the barcode, for example,
// "\033[48;5;190m\033[38;5;19m"
//
// When `escape` is true, the following escape sequences can be inserted:
//
//   - \\: A backslash character
//   - \f: FNC1 in first position
//   - \dN: FNC1 in second position, where N is A-Z, a-z, or 00-99
//   - \e[N]: Extended Channel Interpretation (ECI) designator, where N is a number ranging from 0 to 999999
//   - \g: A GS character (ASCII 29)
//   - \s[INDEX,TOTAL,PARITY]: Structured Append, Where INDEX and TOTAL is a number ranging from 1 to 16, PARITY is a
//     number ranging from 1 to 255, use function `GetParity` to obtain the parity value of the barcode data.
//
//  - This function returns a string for output to the command-line terminal.  When an error occurs, the final return
//    value indicates whether the error can be resolved by upgrading the barcode version.
func SprintQRCode(data []byte, kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel, quietZone int, color string,
    escape, mirror bool) (string, error, bool) {
    if quietZone < 0 {
        return "", fmt.Errorf("invalid quiet zone value: %d", quietZone), false
    }
    if matrix, err, up := buildMatrix(data, kind, version, ecc, escape, mirror); err != nil {
        return "", err, up
    } else {
        clean := ""
        if color != "" {
            clean = "\033[0m"
        }
        height := len(*matrix)
        width := len((*matrix)[0])
        output := ""
        for i := 0; i < (quietZone+1)/2; i++ {
            output += color + strings.Repeat(" ", quietZone*2+width) + clean + "\n"
        }
        row := 0
        for {
            output += color + strings.Repeat(" ", quietZone)
            for col := 0; col < width; col++ {
                idx := ((*matrix)[row][col] & 1) << 1
                if row+1 < height {
                    idx |= (*matrix)[row+1][col] & 1
                }
                output += []string{" ", "▄", "▀", "█"}[idx]
            }
            row += 2
            output += strings.Repeat(" ", quietZone) + clean + "\n"
            if row >= height {
                break
            }
        }
        for i := 0; i < quietZone/2; i++ {
            output += color + strings.Repeat(" ", quietZone*2+width) + clean + "\n"
        }
        return output, nil, false
    }
}

func qrCodeToImage(data []byte, kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel, module, quietZone int,
    spcColor, dotColor string, escape, mirror, invert bool, format QRCodeFormat) ([]byte, error, bool) {
    if format != QRCodeFormatSvg && format != QRCodeFormatPng {
        return nil, fmt.Errorf("unknow QRCode format (svg, png): %s", format), false
    }
    if module <= 0 {
        return nil, fmt.Errorf("invalid module value: %d", module), false
    }
    if quietZone < 0 {
        return nil, fmt.Errorf("invalid quiet zone value: %d", quietZone), false
    }
    if matrix, err, up := buildMatrix(data, kind, version, ecc, escape, mirror); err != nil {
        return nil, err, up
    } else {
        if invert {
            tmpColor := dotColor
            dotColor = spcColor
            spcColor = tmpColor
        }
        height := len(*matrix)
        width := len((*matrix)[0])
        if format == QRCodeFormatSvg {
            svg := fmt.Sprintf(svgQRCodeHeader, (width+quietZone*2)*module, (height+quietZone*2)*module,
                width+quietZone*2, height+quietZone*2, ColorHex2CSS(spcColor, "#ffffff"))
            svg += "<path d=\""
            for i := 0; i < height; i++ {
                svg += fmt.Sprintf("M%d %d.5", quietZone, quietZone+i)
                count := 0
                state := false
                for j := 0; j < width; j++ {
                    if ((*matrix)[i][j] & 1) > 0 {
                        if state {
                            count++
                        } else {
                            if count > 0 {
                                svg += fmt.Sprintf("m%d 0", count)
                            }
                            state = true
                            count = 1
                        }
                    } else {
                        if state {
                            if count > 0 {
                                svg += fmt.Sprintf("l%d 0", count)
                            }
                            state = false
                            count = 1
                        } else {
                            count++
                        }
                    }
                }
                if count > 0 && state {
                    svg += fmt.Sprintf("l%d 0", count)
                }
            }
            svg += fmt.Sprintf(`" style="stroke-width: 1; stroke-linecap: butt; stroke: %s"/></svg>`,
                ColorHex2CSS(dotColor, "#000000"))
            return []byte(svg), nil, false
        } else {
            iPng := image.NewRGBA(image.Rect(0, 0, (width+quietZone*2)*module, (height+quietZone*2)*module))
            if e := drawQRDots(matrix, 0, 0, module, quietZone, spcColor, dotColor,
                func(left, top int, r, g, b, a uint8, background bool) error {
                    iPng.Set(left, top, color.RGBA{R: r, G: g, B: b, A: a})
                    return nil
                }); e != nil {
                return nil, e, false
            }
            if strings.HasSuffix(dotColor, "#") {
                colorPng(iPng, matrix, module, quietZone, invert)
            }
            buffer := new(bytes.Buffer)
            if err := png.Encode(buffer, iPng); err != nil {
                return nil, err, false
            } else {
                return buffer.Bytes(), nil, false
            }
        }
    }
}

func ColorHex2RGBA(hex string, def color.RGBA) color.RGBA {
    hex = strings.Trim(hex, "# \n\t\v")
    if values, err := strconv.ParseUint(hex, 16, 32); err != nil {
        return def
    } else {
        if len(hex) <= 6 {
            return color.RGBA{
                R: uint8(values>>16) & 0xff,
                G: uint8((values >> 8) & 0xff),
                B: uint8(values & 0xff),
                A: 0xff,
            }
        } else {
            return color.RGBA{
                R: uint8(values>>24) & 0xff,
                G: uint8((values >> 16) & 0xff),
                B: uint8((values >> 8) & 0xff),
                A: uint8(values & 0xff),
            }
        }
    }
}

func ColorHex2CSS(hex string, def string) string {
    hex = strings.Trim(hex, "# \n\t\v")
    if values, err := strconv.ParseUint(hex, 16, 32); err != nil {
        return def
    } else {
        if len(hex) <= 6 {
            return fmt.Sprintf("#%02x%02x%02x", uint8(values>>16)&0xff, uint8((values>>8)&0xff),
                uint8(values&0xff))
        } else {
            return fmt.Sprintf("#%02x%02x%02x%02x", uint8(values>>24)&0xff, uint8((values>>16)&0xff),
                uint8((values>>8)&0xff), uint8(values&0xff))
        }
    }
}

const svgQRCodeHeader = `<?xml version="1.0"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg width="%[1]dpx" height="%[2]dpx" viewBox="0 0 %[3]d %[4]d" xmlns="http://www.w3.org/2000/svg">
<rect x="0" y="0" width="%[3]d" height="%[4]d" style="fill: %[5]s;"/>
`

func qrCodeSize(kind QRCodeKind, ver QRCodeVersion) (width, height int) {
    switch kind {
    case MicroQRCode:
        width = int(ver)*2 + 9
        height = width
    case RMQRCode:
        var colIdx int
        if ver < 11 {
            colIdx = (ver-1)%5 + 1
            if ver < 6 {
                height = 7
            } else {
                height = 9
            }
        } else if ver < 23 {
            colIdx = (ver - 11) % 6
            if ver < 17 {
                height = 11
            } else {
                height = 13
            }
        } else {
            colIdx = (ver-23)%5 + 1
            if ver < 28 {
                height = 15
            } else {
                height = 17
            }
        }
        width = []int{27, 43, 59, 77, 99, 139}[colIdx]
    case QRCode:
        width = int(ver)*4 + 17
        height = width
    }
    return
}

func qrCodeNewMatrix(kind QRCodeKind, ver QRCodeVersion) *[][]uint8 {
    width, height := qrCodeSize(kind, ver)
    matrix := make([][]uint8, height)
    for i := 0; i < height; i++ {
        matrix[i] = make([]uint8, width)
    }
    return &matrix
}

var qrCodeAlignments = [][]int{
    {}, {6, 18}, {6, 22}, {6, 26}, {6, 30}, {6, 34}, {6, 22, 38}, {6, 24, 42}, {6, 26, 46}, {6, 28, 50},
    {6, 30, 54}, {6, 32, 58}, {6, 34, 62}, {6, 26, 46, 66}, {6, 26, 48, 70}, {6, 26, 50, 74}, {6, 30, 54, 78},
    {6, 30, 56, 82}, {6, 30, 58, 86}, {6, 34, 62, 90}, {6, 28, 50, 72, 94}, {6, 26, 50, 74, 98},
    {6, 30, 54, 78, 102}, {6, 28, 54, 80, 106}, {6, 32, 58, 84, 110}, {6, 30, 58, 86, 114},
    {6, 34, 62, 90, 118}, {6, 26, 50, 74, 98, 122}, {6, 30, 54, 78, 102, 126}, {6, 26, 52, 78, 104, 130},
    {6, 30, 56, 82, 108, 134}, {6, 34, 60, 86, 112, 138}, {6, 30, 58, 86, 114, 142}, {6, 34, 62, 90, 118, 146},
    {6, 30, 54, 78, 102, 126, 150}, {6, 24, 50, 76, 102, 128, 154}, {6, 28, 54, 80, 106, 132, 158},
    {6, 32, 58, 84, 110, 136, 162}, {6, 26, 54, 82, 110, 138, 166}, {6, 30, 58, 86, 114, 142, 170},
}

var rMqrCodeAlignments = map[int][]int{
    27: {}, 43: {21}, 59: {19, 39}, 77: {25, 51}, 99: {23, 49, 75}, 139: {27, 55, 83, 111},
}

func drawTiming(matrix *[][]uint8, kind QRCodeKind, ver QRCodeVersion) {
    switch kind {
    case MicroQRCode:
        for i := 0; i < len(*matrix); i++ {
            (*matrix)[0][i] = (uint8(i+1) & 1) | 4
            (*matrix)[i][0] = (uint8(i+1) & 1) | 4
        }
    case RMQRCode:
        width, height := qrCodeSize(kind, ver)
        for i := 0; i < width; i++ {
            (*matrix)[0][i] = (uint8(i+1) & 1) | 4
            (*matrix)[height-1][i] = (uint8(i+1) & 1) | 4
        }
        for i := 0; i < height; i++ {
            (*matrix)[i][0] = (uint8(i+1) & 1) | 4
            (*matrix)[i][width-1] = (uint8(i+1) & 1) | 4
            for _, j := range rMqrCodeAlignments[width] {
                (*matrix)[i][j] = (uint8(i+1) & 1) | 4
            }
        }
    case QRCode:
        for i := 0; i < len(*matrix); i++ {
            (*matrix)[6][i] = (uint8(i+1) & 1) | 4
            (*matrix)[i][6] = (uint8(i+1) & 1) | 4
        }
    }
}

func drawAlignment(matrix *[][]uint8, kind QRCodeKind, ver QRCodeVersion) {
    switch kind {
    case MicroQRCode:
        break
    case RMQRCode:
        width, height := qrCodeSize(kind, ver)
        for _, i := range []int{1, height - 2} {
            for _, j := range rMqrCodeAlignments[width] {
                for _, di := range []int{-1, 0, 1} {
                    for _, dj := range []int{-1, 0, 1} {
                        if di != 0 || dj != 0 {
                            (*matrix)[i+di][j+dj] = 9
                        } else {
                            (*matrix)[i+di][j+dj] = 8
                        }
                    }
                }
            }
        }
    case QRCode:
        ver--
        for _, i := range qrCodeAlignments[ver] {
            for _, j := range qrCodeAlignments[ver] {
                last := qrCodeAlignments[ver][len(qrCodeAlignments[ver])-1]
                if (i == 6 && j == 6) || (i == 6 && j == last) || (i == last && j == 6) {
                    continue
                }
                for _, di := range []int{-2, -1, 0, 1, 2} {
                    for _, dj := range []int{-2, -1, 0, 1, 2} {
                        if di == 2 || dj == 2 || di == -2 || dj == -2 || (di == 0 && dj == 0) {
                            (*matrix)[i+di][j+dj] = 9
                        } else {
                            (*matrix)[i+di][j+dj] = 8
                        }
                    }
                }
            }
        }
    }
}

func drawFinder(matrix *[][]uint8, x, y, width, height int) {
    for i := -1; i < 8; i++ {
        for j := -1; j < 8; j++ {
            r := y + i
            c := x + j
            if r >= 0 && r < height && c >= 0 && c < width {
                if (slices.Contains([]int{0, 6}, i) && j >= 0 && j <= 6) ||
                    (slices.Contains([]int{0, 6}, j) && i >= 0 && i <= 6) ||
                    (i > 1 && i < 5 && j > 1 && j < 5) {
                    (*matrix)[r][c] = 17
                } else {
                    (*matrix)[r][c] = 16
                }
            }
        }
    }
}

func drawFinders(matrix *[][]uint8, kind QRCodeKind, ver QRCodeVersion) {
    width, height := qrCodeSize(kind, ver)
    drawFinder(matrix, 0, 0, width, height)
    switch kind {
    case MicroQRCode:
    case RMQRCode:
        for i := 1; i <= 5; i++ {
            for j := 1; j <= 5; j++ {
                if i == 1 || i == 5 || j == 1 || j == 5 || (i == 3 && j == 3) {
                    (*matrix)[height-i][width-j] = 17
                } else {
                    (*matrix)[height-i][width-j] = 16
                }
            }
        }
        switch height {
        case 7:
            for i := 1; i <= 3; i++ {
                (*matrix)[0][width-i] = 17
            }
            (*matrix)[1][width-1] = 17
            (*matrix)[1][width-2] = 16
        case 9:
            for i := 1; i <= 3; i++ {
                (*matrix)[0][width-i] = 17
                (*matrix)[i-1][width-1] = 17
                (*matrix)[height-1][i-1] = 17
            }
            (*matrix)[1][width-2] = 16
        default:
            for i := 1; i <= 3; i++ {
                (*matrix)[0][width-i] = 17
                (*matrix)[i-1][width-1] = 17
                (*matrix)[height-1][i-1] = 17
                (*matrix)[height-i][0] = 17
            }
            (*matrix)[1][width-2] = 16
            (*matrix)[height-2][1] = 16
        }
    case QRCode:
        drawFinder(matrix, 0, height-7, width, height)
        drawFinder(matrix, width-7, 0, width, height)
    }
}

func drawFormat(matrix *[][]uint8, kind QRCodeKind, ver QRCodeVersion, ecc QRCodeEccLevel, mask int) {
    width, height := qrCodeSize(kind, ver)
    switch kind {
    case MicroQRCode:
        data := (map[QRCodeEccLevel][]int{QREccLowest: {0, 1, 3, 5}, QREccMedium: {0, 2, 4, 6},
            QREccQuality: {0, 0, 0, 7}}[ecc][ver-1] << 2) | mask&3
        data = ((data << 10) | bch(data, 10, 1335)) ^ 0x4445
        dots := []int{1, 2, 3, 4, 5, 6, 7, 8, 8, 8, 8, 8, 8, 8, 8}
        for i := 0; i < 15; i++ {
            (*matrix)[dots[i]][dots[14-i]] = uint8((data & 1) | 32)
            data >>= 1
        }
    case QRCode:
        data := (map[QRCodeEccLevel]int{QREccLowest: 1, QREccMedium: 0, QREccQuality: 3, QREccHighest: 2}[ecc] << 3) |
            (mask & 7)
        data = ((data << 10) | bch(data, 10, 1335)) ^ 0x5412
        rows := []int{0, 1, 2, 3, 4, 5, 7, 8, -7, -6, -5, -4, -3, -2, -1}
        cols := []int{-1, -2, -3, -4, -5, -6, -7, -8, 7, 5, 4, 3, 2, 1, 0}
        for i := 0; i < 15; i++ {
            row := rows[i]
            if row < 0 {
                row = height + row
            }
            (*matrix)[row][8] = uint8(data&1) | 32
            col := cols[i]
            if col < 0 {
                col = width + col
            }
            (*matrix)[8][col] = uint8(data&1) | 32
            data >>= 1
        }
        (*matrix)[height-8][8] = 33
    case RMQRCode:
        break
    }
}

func drawVersion(matrix *[][]uint8, kind QRCodeKind, ver QRCodeVersion, ecc QRCodeEccLevel) {
    width, height := qrCodeSize(kind, ver)
    switch kind {
    case QRCode:
        if ver >= 7 {
            data := (ver << 12) | bch(ver, 12, 7973)
            for i := 0; i < 18; i++ {
                x := width - 11 + (i % 3)
                y := i / 3
                (*matrix)[y][x] = uint8(data&1) | 64
                (*matrix)[x][y] = (*matrix)[y][x]
                data >>= 1
            }
        }
    case RMQRCode:
        data := ver - 1
        if ecc == QREccHighest {
            data |= 32
        }
        data = (data << 12) | bch(data, 12, 7973)
        ver1 := data ^ 0x1fab2
        for i := 0; i < 18; i++ {
            di := i % 5
            dj := i / 5
            (*matrix)[1+di][8+dj] = uint8(ver1&1) | 64
            ver1 >>= 1
        }
        ver2 := data ^ 0x20a7b
        for i := 0; i < 15; i++ {
            di := i % 5
            dj := i / 5
            (*matrix)[height-6+di][width-8+dj] = uint8(ver2&1) | 64
            ver2 >>= 1
        }
        for i := 0; i < 3; i++ {
            (*matrix)[height-6][width-5+i] = uint8(ver2&1) | 64
            ver2 >>= 1
        }
    case MicroQRCode:
        break
    }
}

func maskScore(matrix *[][]uint8, kind QRCodeKind, ver QRCodeVersion, mask int) (score int) {
    width, height := qrCodeSize(kind, ver)
    switch kind {
    case QRCode:
        applyDataMask(matrix, kind, ver, mask)
        for row := 0; row < height; row++ {
            var value uint8 = 3
            count := 0
            for col := 0; col <= width; col++ {
                if col == width {
                    if count > 5 {
                        score += count - 2
                    }
                } else if value != ((*matrix)[row][col] & 1) {
                    if count > 5 {
                        score += count - 2
                    }
                    count = 1
                    value = (*matrix)[row][col] & 1
                } else {
                    count++
                }
            }
        }
        for col := 0; col < width; col++ {
            var value uint8 = 3
            count := 0
            for row := 0; row <= height; row++ {
                if row == height {
                    if count > 5 {
                        score += count - 2
                    }
                } else if value != ((*matrix)[row][col] & 1) {
                    if count > 5 {
                        score += count - 2
                    }
                    count = 1
                    value = (*matrix)[row][col] & 1
                } else {
                    count++
                }
            }
        }
        for col := 0; col <= width-2; col++ {
            for row := 0; row <= height-2; row++ {
                count := 0
                count += int((*matrix)[row][col] & 1)
                count += int((*matrix)[row+1][col] & 1)
                count += int((*matrix)[row][col+1] & 1)
                count += int((*matrix)[row+1][col+1] & 1)
                if count <= 4 {
                    score += 3
                }
            }
        }
        for row := 0; row <= height-7; row++ {
            for col := 0; col < width; col++ {
                if ((*matrix)[row][col]&1 == 1) && ((*matrix)[row+1][col]&1 == 0) &&
                    ((*matrix)[row+2][col]&1 == 1) && ((*matrix)[row+3][col]&1 == 1) &&
                    ((*matrix)[row+4][col]&1 == 1) && ((*matrix)[row+5][col]&1 == 0) &&
                    ((*matrix)[row+6][col]&1 == 1) {
                    score += 40
                }
            }
        }
        for row := 0; row < height; row++ {
            for col := 0; col <= width-7; col++ {
                if ((*matrix)[row][col]&1 == 1) && ((*matrix)[row][col+1]&1 == 0) &&
                    ((*matrix)[row][col+2]&1 == 1) && ((*matrix)[row][col+3]&1 == 1) &&
                    ((*matrix)[row][col+4]&1 == 1) && ((*matrix)[row][col+5]&1 == 0) &&
                    ((*matrix)[row][col+6]&1 == 1) {
                    score += 40
                }
            }
        }
        count := 0
        for row := 0; row < height; row++ {
            for col := 0; col < width; col++ {
                count += int((*matrix)[row][col] & 1)
            }
        }
        score += int(math.Round(2 * math.Abs(float64((count*100)/(width*height)-50))))
        applyDataMask(matrix, kind, ver, mask)
    case MicroQRCode:
        applyDataMask(matrix, kind, ver, mask)
        sum1 := 0
        sum2 := 0
        for i := 0; i < height; i++ {
            sum1 += int(((*matrix)[height-1][i]) & 1)
            sum2 += int(((*matrix)[i][width-1]) & 1)
        }
        score = min(sum1, sum2)*16 + max(sum1, sum2)
        applyDataMask(matrix, kind, ver, mask)
    case RMQRCode:
        break
    }
    return
}

func getOptimalMask(matrix *[][]uint8, kind QRCodeKind, ver QRCodeVersion, ecc QRCodeEccLevel) (mask int) {
    switch kind {
    case QRCode:
        minScore := 0
        for i := 0; i < 8; i++ {
            drawFormat(matrix, kind, ver, ecc, i)
            score := maskScore(matrix, kind, ver, i)
            if (i == 0) || (score < minScore) {
                minScore = score
                mask = i
            }
        }
    case MicroQRCode:
        maxScore := 0
        for i := 0; i < 4; i++ {
            drawFormat(matrix, kind, ver, ecc, i)
            score := maskScore(matrix, kind, ver, i)
            if (i == 0) || (score > maxScore) {
                maxScore = score
                mask = i
            }
        }
    case RMQRCode:
        break
    }
    return
}

func maskValue(kind QRCodeKind, mask int, row, col int) uint8 {
    var r bool
    switch kind {
    case QRCode:
        switch mask {
        case 0:
            r = ((row + col) % 2) == 0
        case 1:
            r = (row % 2) == 0
        case 2:
            r = (col % 3) == 0
        case 3:
            r = ((row + col) % 3) == 0
        case 4:
            r = ((row/2 + col/3) % 2) == 0
        case 5:
            r = ((col*row)%2 + (col*row)%3) == 0
        case 6:
            r = ((col*row)%2+(col*row)%3)%2 == 0
        case 7:
            r = ((col+row)%2+(col*row)%3)%2 == 0
        }
    case RMQRCode:
        r = ((row/2 + col/3) % 2) == 0
    case MicroQRCode:
        switch mask {
        case 0:
            r = (row % 2) == 0
        case 1:
            r = ((row/2 + col/3) % 2) == 0
        case 2:
            r = ((col*row)%2+(col*row)%3)%2 == 0
        case 3:
            r = ((col+row)%2+(col*row)%3)%2 == 0
        }
    }
    if r {
        return 1
    } else {
        return 0
    }
}

func applyDataMask(matrix *[][]uint8, kind QRCodeKind, ver QRCodeVersion, mask int) {
    width, height := qrCodeSize(kind, ver)
    for i := 0; i < height; i++ {
        for j := 0; j < width; j++ {
            if ((*matrix)[i][j] & 0xfc) == 0 {
                (*matrix)[i][j] ^= maskValue(kind, mask, i, j)
            }
        }
    }
}

func placeMatrix(matrix [][]uint8, codes []byte, kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel) {
    index := 0
    bitId := 7
    dir := -1
    row := len(matrix) - 1
    col := len(matrix[0]) - 1
    for col >= 0 {
        switch kind {
        case QRCode:
            if col == 6 {
                col--
            }
        case MicroQRCode:
            if col == 0 {
                break
            }
        case RMQRCode:
            if col == 0 {
                break
            }
            if col == len(matrix[0])-1 {
                col--
            }
        }
        if col < 0 {
            break
        }
        for {
            for pair := 0; pair <= 1; pair++ {
                if (col-pair >= 0) && (matrix[row][col-pair]&0xfc == 0) {
                    if index >= totalWords(kind, version) {
                        matrix[row][col-pair] = 0x2
                    } else {
                        matrix[row][col-pair] = (codes[index] >> bitId) & 1
                    }
                    short := kind == MicroQRCode && ((version == 1 && index == 2) ||
                        (version == 3 && ((ecc == QREccLowest && index == 10) || (ecc != QREccLowest && index == 8))))
                    if bitId == 0 || (short && bitId <= 4) {
                        bitId = 7
                        index++
                    } else {
                        bitId--
                    }
                }
            }
            row += dir
            if row < 0 || row >= len(matrix) {
                row -= dir
                dir = -dir
                break
            }
        }
        col -= 2
    }
}

func buildMatrix(data []byte, kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel, escape bool,
    mirror bool) (*[][]byte, error, bool) {
    verValid := true
    eccValid := true
    up := true
    switch kind {
    case MicroQRCode:
        switch version {
        case 1:
            eccValid = ecc == QREccHighest
        case 2, 3:
            eccValid = ecc == QREccLowest || ecc == QREccMedium
        case 4:
            eccValid = ecc != QREccHighest
            up = false
        default:
            verValid = false
            up = false
        }
    case QRCode:
        verValid = version >= 1 && version <= 40
        up = false
    case RMQRCode:
        verValid = version >= 1 && version <= 32
        eccValid = ecc != QREccLowest && ecc != QREccQuality
        up = false
    default:
        return nil, fmt.Errorf("invalid QRCode kind: %s", kind), false
    }
    if !verValid {
        return nil, fmt.Errorf("invalid QRCode version: %d", version), up
    }
    if !eccValid {
        return nil, fmt.Errorf("invalid QRCode ECC level: %s", ecc), up
    }
    if codes, err, up, ec := encodeData(data, kind, version, ecc, escape); err != nil {
        return nil, err, up
    } else {
        matrix := qrCodeNewMatrix(kind, version)
        drawTiming(matrix, kind, version)
        drawAlignment(matrix, kind, version)
        drawFinders(matrix, kind, version)
        drawVersion(matrix, kind, version, ec)
        drawFormat(matrix, kind, version, ec, 0)
        placeMatrix(*matrix, codes, kind, version, ec)
        mask := getOptimalMask(matrix, kind, version, ec)
        drawFormat(matrix, kind, version, ec, mask)
        applyDataMask(matrix, kind, version, mask)
        if mirror {
            width := len((*matrix)[0])
            for i := 0; i < len(*matrix); i++ {
                line := make([]uint8, width)
                for j := 0; j < width; j++ {
                    line[width-j-1] = (*matrix)[i][j]
                }
                (*matrix)[i] = line
            }
        }
        return matrix, nil, false
    }
}

func drawQRDots(matrix *[][]byte, left, top, module, quietZone int, spcColor, dotColor string,
    drawDot func(left, top int, r, g, b, a uint8, background bool) error) error {
    height := len(*matrix)
    width := len((*matrix)[0])
    rgba := ColorHex2RGBA(spcColor, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
    for x := left; x < left+(width+quietZone*2)*module; x++ {
        for y := top; y < top+(height+quietZone*2)*module; y++ {
            if e := drawDot(x, y, rgba.R, rgba.G, rgba.B, rgba.A, true); e != nil {
                return e
            }
        }
    }
    rgba = ColorHex2RGBA(dotColor, color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff})
    for i := 0; i < height; i++ {
        for j := 0; j < width; j++ {
            if ((*matrix)[i][j] & 1) > 0 {
                for di := 0; di < module; di++ {
                    for dj := 0; dj < module; dj++ {
                        if e := drawDot(top+(quietZone+j)*module+dj, left+(quietZone+i)*module+di,
                            rgba.R, rgba.G, rgba.B, rgba.A, false); e != nil {
                            return e
                        }
                    }
                }
            }
        }
    }
    return nil
}

func colorPng(iPng *image.RGBA, matrix *[][]byte, module, quietZone int, invert bool) {
    rgbSpc := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
    rgbDot := color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
    colors := map[byte][]color.RGBA{
        0x00: {ColorHex2RGBA("#6b7280", rgbDot), ColorHex2RGBA("#f3f4f6", rgbSpc)}, // 数据位
        0x02: {ColorHex2RGBA("#b45309", rgbDot), ColorHex2RGBA("#fef3c7", rgbSpc)}, // 剩余位
        0x04: {ColorHex2RGBA("#b91c1c", rgbDot), ColorHex2RGBA("#fee2e2", rgbSpc)}, // 时序线
        0x08: {ColorHex2RGBA("#115e59", rgbDot), ColorHex2RGBA("#f0fdfa", rgbSpc)}, // 对齐框
        0x10: {ColorHex2RGBA("#3f6212", rgbDot), ColorHex2RGBA("#ecfccb", rgbSpc)}, // 发现框
        0x20: {ColorHex2RGBA("#7e22ce", rgbDot), ColorHex2RGBA("#f3e8ff", rgbSpc)}, // 格式位
        0x40: {ColorHex2RGBA("#1e40af", rgbDot), ColorHex2RGBA("#dbeafe", rgbSpc)}, // 版本位
    }
    height := len(*matrix)
    width := len((*matrix)[0])
    idx := 1
    if invert {
        idx = 0
    }
    for i := 0; i < height; i++ {
        for j := 0; j < width; j++ {
            if ((*matrix)[i][j] & 1) > 0 {
                color1 := colors[(*matrix)[i][j]&254][1-idx]
                for di := 0; di < module; di++ {
                    for dj := 0; dj < module; dj++ {
                        iPng.Set((quietZone+j)*module+dj, (quietZone+i)*module+di, color1)
                    }
                }
            } else {
                color0 := colors[(*matrix)[i][j]&254][idx]
                for di := 0; di < module; di++ {
                    for dj := 0; dj < module; dj++ {
                        iPng.Set((quietZone+j)*module+dj, (quietZone+i)*module+di, color0)
                    }
                }
            }
        }
    }
}

func totalBits(kind QRCodeKind, version QRCodeVersion) (result int) {
    width, height := qrCodeSize(kind, version)
    switch kind {
    case MicroQRCode:
        result = (width-1)*(height-1) - 64
    case QRCode:
        result = width*height - 64*3 - (width-16)*2 - 31
        if version >= 7 {
            result -= 36
        }
        rows := len(qrCodeAlignments[version-1])
        if rows >= 2 {
            result -= (rows*rows - 3) * 25
            result += (rows*2 - 4) * 5
        }
    case RMQRCode:
        algCols := len(rMqrCodeAlignments[width])
        if height == 7 {
            result = (width-19-algCols)*5 - 3
        } else if height == 9 {
            result = (width-9-algCols)*7 - 53
        } else {
            result = (width-2-algCols)*(height-2) - 103
        }
        result -= algCols * 8
    }
    return
}

func totalDataBits(kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel) (result int) {
    result = totalDataWords(kind, version, ecc) * 8
    if kind == MicroQRCode && (version == 1 || version == 3) {
        result -= 4
    }
    return
}

func totalWords(kind QRCodeKind, version QRCodeVersion) int {
    tolBits := totalBits(kind, version)
    if kind == MicroQRCode && (version == 1 || version == 3) {
        return (tolBits + 4) / 8
    }
    return tolBits / 8
}

func totalDataWords(kind QRCodeKind, version QRCodeVersion, ecc QRCodeEccLevel) int {
    block := eccDefines[kind][version-1][ecc]
    return block.count*block.data + block.next*(block.data+1)
}

func terminateBits(kind QRCodeKind, version QRCodeVersion) int {
    switch kind {
    case QRCode:
        return 4
    case RMQRCode:
        return 3
    case MicroQRCode:
        return version*2 + 1
    }
    return 0
}

type eccBlock struct {
    count int
    total int
    data  int
    next  int
}

var eccDefines = map[QRCodeKind][]map[QRCodeEccLevel]eccBlock{
    MicroQRCode: {
        {
            QREccLowest: eccBlock{count: 1, total: 5, data: 3, next: 0},
        }, // 1
        {
            QREccLowest: eccBlock{count: 1, total: 10, data: 5, next: 0},
            QREccMedium: eccBlock{count: 1, total: 10, data: 4, next: 0},
        }, // 2
        {
            QREccLowest: eccBlock{count: 1, total: 17, data: 11, next: 0},
            QREccMedium: eccBlock{count: 1, total: 17, data: 9, next: 0},
        }, // 3
        {
            QREccLowest:  eccBlock{count: 1, total: 24, data: 16, next: 0},
            QREccMedium:  eccBlock{count: 1, total: 24, data: 14, next: 0},
            QREccQuality: eccBlock{count: 1, total: 24, data: 10, next: 0},
        }, // 4
    },
    RMQRCode: {
        {
            QREccMedium:  eccBlock{count: 1, total: 13, data: 6, next: 0},
            QREccHighest: eccBlock{count: 1, total: 13, data: 3, next: 0},
        }, // 1
        {
            QREccMedium:  eccBlock{count: 1, total: 21, data: 12, next: 0},
            QREccHighest: eccBlock{count: 1, total: 21, data: 7, next: 0},
        }, // 2
        {
            QREccMedium:  eccBlock{count: 1, total: 32, data: 20, next: 0},
            QREccHighest: eccBlock{count: 1, total: 32, data: 10, next: 0},
        }, // 3
        {
            QREccMedium:  eccBlock{count: 1, total: 44, data: 28, next: 0},
            QREccHighest: eccBlock{count: 1, total: 44, data: 14, next: 0},
        }, // 4
        {
            QREccMedium:  eccBlock{count: 1, total: 68, data: 44, next: 0},
            QREccHighest: eccBlock{count: 2, total: 34, data: 12, next: 0},
        }, // 5
        {
            QREccMedium:  eccBlock{count: 1, total: 21, data: 12, next: 0},
            QREccHighest: eccBlock{count: 1, total: 21, data: 7, next: 0},
        }, // 6
        {
            QREccMedium:  eccBlock{count: 1, total: 33, data: 21, next: 0},
            QREccHighest: eccBlock{count: 1, total: 33, data: 11, next: 0},
        }, // 7
        {
            QREccMedium:  eccBlock{count: 1, total: 49, data: 31, next: 0},
            QREccHighest: eccBlock{count: 1, total: 24, data: 8, next: 1},
        }, // 8
        {
            QREccMedium:  eccBlock{count: 1, total: 66, data: 42, next: 0},
            QREccHighest: eccBlock{count: 2, total: 33, data: 11, next: 0},
        }, // 9
        {
            QREccMedium:  eccBlock{count: 1, total: 49, data: 31, next: 1},
            QREccHighest: eccBlock{count: 3, total: 33, data: 11, next: 0},
        }, // 10
        {
            QREccMedium:  eccBlock{count: 1, total: 15, data: 7, next: 0},
            QREccHighest: eccBlock{count: 1, total: 15, data: 5, next: 0},
        }, // 11
        {
            QREccMedium:  eccBlock{count: 1, total: 31, data: 19, next: 0},
            QREccHighest: eccBlock{count: 1, total: 31, data: 11, next: 0},
        }, // 12
        {
            QREccMedium:  eccBlock{count: 1, total: 47, data: 31, next: 0},
            QREccHighest: eccBlock{count: 1, total: 23, data: 7, next: 1},
        }, // 13
        {
            QREccMedium:  eccBlock{count: 1, total: 67, data: 43, next: 0},
            QREccHighest: eccBlock{count: 1, total: 33, data: 11, next: 1},
        }, // 14
        {
            QREccMedium:  eccBlock{count: 1, total: 44, data: 28, next: 1},
            QREccHighest: eccBlock{count: 1, total: 44, data: 14, next: 1},
        }, // 15
        {
            QREccMedium:  eccBlock{count: 2, total: 66, data: 42, next: 0},
            QREccHighest: eccBlock{count: 3, total: 44, data: 14, next: 0},
        }, // 16
        {
            QREccMedium:  eccBlock{count: 1, total: 21, data: 12, next: 0},
            QREccHighest: eccBlock{count: 1, total: 21, data: 7, next: 0},
        }, // 17
        {
            QREccMedium:  eccBlock{count: 1, total: 41, data: 27, next: 0},
            QREccHighest: eccBlock{count: 1, total: 41, data: 13, next: 0},
        }, // 18
        {
            QREccMedium:  eccBlock{count: 1, total: 60, data: 38, next: 0},
            QREccHighest: eccBlock{count: 2, total: 30, data: 10, next: 0},
        }, // 19
        {
            QREccMedium:  eccBlock{count: 1, total: 42, data: 26, next: 1},
            QREccHighest: eccBlock{count: 1, total: 42, data: 14, next: 1},
        }, // 20
        {
            QREccMedium:  eccBlock{count: 1, total: 56, data: 36, next: 1},
            QREccHighest: eccBlock{count: 1, total: 37, data: 11, next: 2},
        }, // 21
        {
            QREccMedium:  eccBlock{count: 2, total: 55, data: 35, next: 1},
            QREccHighest: eccBlock{count: 2, total: 41, data: 13, next: 2},
        }, // 22
        {
            QREccMedium:  eccBlock{count: 1, total: 51, data: 33, next: 0},
            QREccHighest: eccBlock{count: 1, total: 25, data: 7, next: 1},
        }, // 23
        {
            QREccMedium:  eccBlock{count: 1, total: 74, data: 48, next: 0},
            QREccHighest: eccBlock{count: 2, total: 37, data: 13, next: 0},
        }, // 24
        {
            QREccMedium:  eccBlock{count: 1, total: 51, data: 33, next: 1},
            QREccHighest: eccBlock{count: 2, total: 34, data: 10, next: 1},
        }, // 25
        {
            QREccMedium:  eccBlock{count: 2, total: 68, data: 44, next: 0},
            QREccHighest: eccBlock{count: 4, total: 34, data: 12, next: 0},
        }, // 26
        {
            QREccMedium:  eccBlock{count: 2, total: 66, data: 42, next: 1},
            QREccHighest: eccBlock{count: 1, total: 39, data: 13, next: 4},
        }, // 27
        {
            QREccMedium:  eccBlock{count: 1, total: 61, data: 39, next: 0},
            QREccHighest: eccBlock{count: 1, total: 30, data: 10, next: 1},
        }, // 28
        {
            QREccMedium:  eccBlock{count: 2, total: 44, data: 28, next: 0},
            QREccHighest: eccBlock{count: 2, total: 44, data: 14, next: 0},
        }, // 29
        {
            QREccMedium:  eccBlock{count: 2, total: 61, data: 39, next: 0},
            QREccHighest: eccBlock{count: 1, total: 40, data: 12, next: 2},
        }, // 30
        {
            QREccMedium:  eccBlock{count: 2, total: 53, data: 33, next: 1},
            QREccHighest: eccBlock{count: 4, total: 40, data: 14, next: 0},
        }, // 31
        {
            QREccMedium:  eccBlock{count: 4, total: 58, data: 38, next: 0},
            QREccHighest: eccBlock{count: 2, total: 38, data: 12, next: 4},
        }, // 32
    },
    QRCode: {
        {
            QREccLowest:  eccBlock{count: 1, total: 26, data: 19, next: 0},
            QREccMedium:  eccBlock{count: 1, total: 26, data: 16, next: 0},
            QREccQuality: eccBlock{count: 1, total: 26, data: 13, next: 0},
            QREccHighest: eccBlock{count: 1, total: 26, data: 9, next: 0},
        }, // 1
        {
            QREccLowest:  eccBlock{count: 1, total: 44, data: 34, next: 0},
            QREccMedium:  eccBlock{count: 1, total: 44, data: 28, next: 0},
            QREccQuality: eccBlock{count: 1, total: 44, data: 22, next: 0},
            QREccHighest: eccBlock{count: 1, total: 44, data: 16, next: 0},
        }, // 2
        {
            QREccLowest:  eccBlock{count: 1, total: 70, data: 55, next: 0},
            QREccMedium:  eccBlock{count: 1, total: 70, data: 44, next: 0},
            QREccQuality: eccBlock{count: 2, total: 35, data: 17, next: 0},
            QREccHighest: eccBlock{count: 2, total: 35, data: 13, next: 0},
        }, // 3
        {
            QREccLowest:  eccBlock{count: 1, total: 100, data: 80, next: 0},
            QREccMedium:  eccBlock{count: 2, total: 50, data: 32, next: 0},
            QREccQuality: eccBlock{count: 2, total: 50, data: 24, next: 0},
            QREccHighest: eccBlock{count: 4, total: 25, data: 9, next: 0},
        }, // 4
        {
            QREccLowest:  eccBlock{count: 1, total: 134, data: 108, next: 0},
            QREccMedium:  eccBlock{count: 2, total: 67, data: 43, next: 0},
            QREccQuality: eccBlock{count: 2, total: 33, data: 15, next: 2},
            QREccHighest: eccBlock{count: 2, total: 33, data: 11, next: 2},
        }, // 5
        {
            QREccLowest:  eccBlock{count: 2, total: 86, data: 68, next: 0},
            QREccMedium:  eccBlock{count: 4, total: 43, data: 27, next: 0},
            QREccQuality: eccBlock{count: 4, total: 43, data: 19, next: 0},
            QREccHighest: eccBlock{count: 4, total: 43, data: 15, next: 0},
        }, // 6
        {
            QREccLowest:  eccBlock{count: 2, total: 98, data: 78, next: 0},
            QREccMedium:  eccBlock{count: 4, total: 49, data: 31, next: 0},
            QREccQuality: eccBlock{count: 2, total: 32, data: 14, next: 4},
            QREccHighest: eccBlock{count: 4, total: 39, data: 13, next: 1},
        }, // 7
        {
            QREccLowest:  eccBlock{count: 2, total: 121, data: 97, next: 0},
            QREccMedium:  eccBlock{count: 2, total: 60, data: 38, next: 2},
            QREccQuality: eccBlock{count: 4, total: 40, data: 18, next: 2},
            QREccHighest: eccBlock{count: 4, total: 40, data: 14, next: 2},
        }, // 8
        {
            QREccLowest:  eccBlock{count: 2, total: 146, data: 116, next: 0},
            QREccMedium:  eccBlock{count: 3, total: 58, data: 36, next: 2},
            QREccQuality: eccBlock{count: 4, total: 36, data: 16, next: 4},
            QREccHighest: eccBlock{count: 4, total: 36, data: 12, next: 4},
        }, // 9
        {
            QREccLowest:  eccBlock{count: 2, total: 86, data: 68, next: 2},
            QREccMedium:  eccBlock{count: 4, total: 69, data: 43, next: 1},
            QREccQuality: eccBlock{count: 6, total: 43, data: 19, next: 2},
            QREccHighest: eccBlock{count: 6, total: 43, data: 15, next: 2},
        }, // 10
        {
            QREccLowest:  eccBlock{count: 4, total: 101, data: 81, next: 0},
            QREccMedium:  eccBlock{count: 1, total: 80, data: 50, next: 4},
            QREccQuality: eccBlock{count: 4, total: 50, data: 22, next: 4},
            QREccHighest: eccBlock{count: 3, total: 36, data: 12, next: 8},
        }, // 11
        {
            QREccLowest:  eccBlock{count: 2, total: 116, data: 92, next: 2},
            QREccMedium:  eccBlock{count: 6, total: 58, data: 36, next: 2},
            QREccQuality: eccBlock{count: 4, total: 46, data: 20, next: 6},
            QREccHighest: eccBlock{count: 7, total: 42, data: 14, next: 4},
        }, // 12
        {
            QREccLowest:  eccBlock{count: 4, total: 133, data: 107, next: 0},
            QREccMedium:  eccBlock{count: 8, total: 59, data: 37, next: 1},
            QREccQuality: eccBlock{count: 8, total: 44, data: 20, next: 4},
            QREccHighest: eccBlock{count: 12, total: 33, data: 11, next: 4},
        }, // 13
        {
            QREccLowest:  eccBlock{count: 3, total: 145, data: 115, next: 1},
            QREccMedium:  eccBlock{count: 4, total: 64, data: 40, next: 5},
            QREccQuality: eccBlock{count: 11, total: 36, data: 16, next: 5},
            QREccHighest: eccBlock{count: 11, total: 36, data: 12, next: 5},
        }, // 14
        {
            QREccLowest:  eccBlock{count: 5, total: 109, data: 87, next: 1},
            QREccMedium:  eccBlock{count: 5, total: 65, data: 41, next: 5},
            QREccQuality: eccBlock{count: 5, total: 54, data: 24, next: 7},
            QREccHighest: eccBlock{count: 11, total: 36, data: 12, next: 7},
        }, // 15
        {
            QREccLowest:  eccBlock{count: 5, total: 122, data: 98, next: 1},
            QREccMedium:  eccBlock{count: 7, total: 73, data: 45, next: 3},
            QREccQuality: eccBlock{count: 15, total: 43, data: 19, next: 2},
            QREccHighest: eccBlock{count: 3, total: 45, data: 15, next: 13},
        }, // 16
        {
            QREccLowest:  eccBlock{count: 1, total: 135, data: 107, next: 5},
            QREccMedium:  eccBlock{count: 10, total: 74, data: 46, next: 1},
            QREccQuality: eccBlock{count: 1, total: 50, data: 22, next: 15},
            QREccHighest: eccBlock{count: 2, total: 42, data: 14, next: 17},
        }, // 17
        {
            QREccLowest:  eccBlock{count: 5, total: 150, data: 120, next: 1},
            QREccMedium:  eccBlock{count: 9, total: 69, data: 43, next: 4},
            QREccQuality: eccBlock{count: 17, total: 50, data: 22, next: 1},
            QREccHighest: eccBlock{count: 2, total: 42, data: 14, next: 19},
        }, // 18
        {
            QREccLowest:  eccBlock{count: 3, total: 141, data: 113, next: 4},
            QREccMedium:  eccBlock{count: 3, total: 70, data: 44, next: 11},
            QREccQuality: eccBlock{count: 17, total: 47, data: 21, next: 4},
            QREccHighest: eccBlock{count: 9, total: 39, data: 13, next: 16},
        }, // 19
        {
            QREccLowest:  eccBlock{count: 3, total: 135, data: 107, next: 5},
            QREccMedium:  eccBlock{count: 3, total: 67, data: 41, next: 13},
            QREccQuality: eccBlock{count: 15, total: 54, data: 24, next: 5},
            QREccHighest: eccBlock{count: 15, total: 43, data: 15, next: 10},
        }, // 20
        {
            QREccLowest:  eccBlock{count: 4, total: 144, data: 116, next: 4},
            QREccMedium:  eccBlock{count: 17, total: 68, data: 42, next: 0},
            QREccQuality: eccBlock{count: 17, total: 50, data: 22, next: 6},
            QREccHighest: eccBlock{count: 19, total: 46, data: 16, next: 6},
        }, // 21
        {
            QREccLowest:  eccBlock{count: 2, total: 139, data: 111, next: 7},
            QREccMedium:  eccBlock{count: 17, total: 74, data: 46, next: 0},
            QREccQuality: eccBlock{count: 7, total: 54, data: 24, next: 16},
            QREccHighest: eccBlock{count: 34, total: 37, data: 12, next: 0},
        }, // 22
        {
            QREccLowest:  eccBlock{count: 4, total: 151, data: 121, next: 5},
            QREccMedium:  eccBlock{count: 4, total: 75, data: 47, next: 14},
            QREccQuality: eccBlock{count: 11, total: 54, data: 24, next: 14},
            QREccHighest: eccBlock{count: 16, total: 45, data: 15, next: 14},
        }, // 23
        {
            QREccLowest:  eccBlock{count: 6, total: 147, data: 117, next: 4},
            QREccMedium:  eccBlock{count: 6, total: 73, data: 45, next: 14},
            QREccQuality: eccBlock{count: 11, total: 54, data: 24, next: 16},
            QREccHighest: eccBlock{count: 30, total: 46, data: 16, next: 2},
        }, // 24
        {
            QREccLowest:  eccBlock{count: 8, total: 132, data: 106, next: 4},
            QREccMedium:  eccBlock{count: 8, total: 75, data: 47, next: 13},
            QREccQuality: eccBlock{count: 7, total: 54, data: 24, next: 22},
            QREccHighest: eccBlock{count: 22, total: 45, data: 15, next: 13},
        }, // 25
        {
            QREccLowest:  eccBlock{count: 10, total: 142, data: 114, next: 2},
            QREccMedium:  eccBlock{count: 19, total: 74, data: 46, next: 4},
            QREccQuality: eccBlock{count: 28, total: 50, data: 22, next: 6},
            QREccHighest: eccBlock{count: 33, total: 46, data: 16, next: 4},
        }, // 26
        {
            QREccLowest:  eccBlock{count: 8, total: 152, data: 122, next: 4},
            QREccMedium:  eccBlock{count: 22, total: 73, data: 45, next: 3},
            QREccQuality: eccBlock{count: 8, total: 53, data: 23, next: 26},
            QREccHighest: eccBlock{count: 12, total: 45, data: 15, next: 28},
        }, // 27
        {
            QREccLowest:  eccBlock{count: 3, total: 147, data: 117, next: 10},
            QREccMedium:  eccBlock{count: 3, total: 73, data: 45, next: 23},
            QREccQuality: eccBlock{count: 4, total: 54, data: 24, next: 31},
            QREccHighest: eccBlock{count: 11, total: 45, data: 15, next: 31},
        }, // 28
        {
            QREccLowest:  eccBlock{count: 7, total: 146, data: 116, next: 7},
            QREccMedium:  eccBlock{count: 21, total: 73, data: 45, next: 7},
            QREccQuality: eccBlock{count: 1, total: 53, data: 23, next: 37},
            QREccHighest: eccBlock{count: 19, total: 45, data: 15, next: 26},
        }, // 29
        {
            QREccLowest:  eccBlock{count: 5, total: 145, data: 115, next: 10},
            QREccMedium:  eccBlock{count: 19, total: 75, data: 47, next: 10},
            QREccQuality: eccBlock{count: 15, total: 54, data: 24, next: 25},
            QREccHighest: eccBlock{count: 23, total: 45, data: 15, next: 25},
        }, // 30
        {
            QREccLowest:  eccBlock{count: 13, total: 145, data: 115, next: 3},
            QREccMedium:  eccBlock{count: 2, total: 74, data: 46, next: 29},
            QREccQuality: eccBlock{count: 42, total: 54, data: 24, next: 1},
            QREccHighest: eccBlock{count: 23, total: 45, data: 15, next: 28},
        }, // 31
        {
            QREccLowest:  eccBlock{count: 17, total: 145, data: 115, next: 0},
            QREccMedium:  eccBlock{count: 10, total: 74, data: 46, next: 23},
            QREccQuality: eccBlock{count: 10, total: 54, data: 24, next: 35},
            QREccHighest: eccBlock{count: 19, total: 45, data: 15, next: 35},
        }, // 32
        {
            QREccLowest:  eccBlock{count: 17, total: 145, data: 115, next: 1},
            QREccMedium:  eccBlock{count: 14, total: 74, data: 46, next: 21},
            QREccQuality: eccBlock{count: 29, total: 54, data: 24, next: 19},
            QREccHighest: eccBlock{count: 11, total: 45, data: 15, next: 46},
        }, // 33
        {
            QREccLowest:  eccBlock{count: 13, total: 145, data: 115, next: 6},
            QREccMedium:  eccBlock{count: 14, total: 74, data: 46, next: 23},
            QREccQuality: eccBlock{count: 44, total: 54, data: 24, next: 7},
            QREccHighest: eccBlock{count: 59, total: 46, data: 16, next: 1},
        }, // 34
        {
            QREccLowest:  eccBlock{count: 12, total: 151, data: 121, next: 7},
            QREccMedium:  eccBlock{count: 12, total: 75, data: 47, next: 26},
            QREccQuality: eccBlock{count: 39, total: 54, data: 24, next: 14},
            QREccHighest: eccBlock{count: 22, total: 45, data: 15, next: 41},
        }, // 35
        {
            QREccLowest:  eccBlock{count: 6, total: 151, data: 121, next: 14},
            QREccMedium:  eccBlock{count: 6, total: 75, data: 47, next: 34},
            QREccQuality: eccBlock{count: 46, total: 54, data: 24, next: 10},
            QREccHighest: eccBlock{count: 2, total: 45, data: 15, next: 64},
        }, // 36
        {
            QREccLowest:  eccBlock{count: 17, total: 152, data: 122, next: 4},
            QREccMedium:  eccBlock{count: 29, total: 74, data: 46, next: 14},
            QREccQuality: eccBlock{count: 49, total: 54, data: 24, next: 10},
            QREccHighest: eccBlock{count: 24, total: 45, data: 15, next: 46},
        }, // 37
        {
            QREccLowest:  eccBlock{count: 4, total: 152, data: 122, next: 18},
            QREccMedium:  eccBlock{count: 13, total: 74, data: 46, next: 32},
            QREccQuality: eccBlock{count: 48, total: 54, data: 24, next: 14},
            QREccHighest: eccBlock{count: 42, total: 45, data: 15, next: 32},
        }, // 38
        {
            QREccLowest:  eccBlock{count: 20, total: 147, data: 117, next: 4},
            QREccMedium:  eccBlock{count: 40, total: 75, data: 47, next: 7},
            QREccQuality: eccBlock{count: 43, total: 54, data: 24, next: 22},
            QREccHighest: eccBlock{count: 10, total: 45, data: 15, next: 67},
        }, // 39
        {
            QREccLowest:  eccBlock{count: 19, total: 148, data: 118, next: 6},
            QREccMedium:  eccBlock{count: 18, total: 75, data: 47, next: 31},
            QREccQuality: eccBlock{count: 34, total: 54, data: 24, next: 34},
            QREccHighest: eccBlock{count: 20, total: 45, data: 15, next: 61},
        }, // 40
    },
}
