package main

import (
    _ "embed"
    "encoding/base64"
    "fmt"
    "github.com/signintech/gopdf"
    "image"
    "image/color"
    "image/draw"
    "image/png"
    "log"
    "macsvn.co/xqrcode"
    "net/http"
    "os"
    "strconv"
)

//go:embed index.html
var index []byte

func main() {
    // Generate SVG file
    _, _ = xqrcode.QRCodeToFile([]byte("abc1234567890"), xqrcode.QRCode, 15, xqrcode.QREccMedium, 3, 4,
        "#ffffff", "#000000", false, false, false, xqrcode.QRCodeFormatSvg, "./qr.svg")
    // Generate PNG file
    _, _ = xqrcode.QRCodeToFile([]byte("abc1234567890"), "rMQRCode", 15, "M", 3, 4, "#ffffff", "#000000", false, false,
        false, "png", "./qr.png")
    // Draw barcode in PDF
    pdf := gopdf.GoPdf{}
    pdf.Start(gopdf.Config{Unit: gopdf.UnitMM, PageSize: *gopdf.PageSizeA4})
    pdf.AddPage()
    _, _, _, _ = xqrcode.DrawQRCode1([]byte("abc1234567890"), xqrcode.QRCode, 15, xqrcode.QREccMedium, 4, false, false,
        false, "#ffffff", "#000000", 0.3, 80.0, 20.0, false, false,
        func(left, top, width, height float64, r, g, b, a uint8, background bool) error {
            pdf.SetFillColor(r, g, b)
            return pdf.Rectangle(left, top, left+width, top+height, "F", 0, 0)
        })
    _ = pdf.WritePdf("1.pdf")
    // Draw barcode in PNG
    iPng := image.NewRGBA(image.Rect(0, 0, 500, 500))
    _, _, _, _ = xqrcode.DrawQRCode1([]byte("abc1234567890"), xqrcode.QRCode, 7, xqrcode.QREccMedium, 2, false, false,
        false, "#ffffff", "#000000", 2, 50, 50, false, false,
        func(left, top, width, height float64, r, g, b, a uint8, background bool) error {
            dotColor := &image.Uniform{C: color.RGBA{R: r, G: g, B: b, A: a}}
            draw.Draw(iPng, image.Rect(int(left), int(top), int(left+width), int(top+height)), dotColor, image.Point{},
                draw.Over)
            return nil
        }) // method 1
    _, _, _, _ = xqrcode.DrawQRCode2([]byte("abc1234567890"), 250, 50, xqrcode.QRCode, 7, xqrcode.QREccMedium, 2, 4,
        "#ffffff", "#000000", false, false, false, false, false,
        func(left, top int, r, g, b, a uint8, background bool) error {
            iPng.Set(left, top, color.RGBA{R: r, G: g, B: b, A: a})
            return nil
        }) // method 2
    f, _ := os.Create("./3.png")
    defer func() { _ = f.Close() }()
    _ = png.Encode(f, iPng)
    // Display barcode in terminal:
    if text, err, _ := xqrcode.SprintQRCode([]byte("abc1234567890"), xqrcode.QRCode, 8, xqrcode.QREccMedium, 3,
        "\033[48;5;190m\033[38;5;19m", false, false); err == nil {
        fmt.Println(text)
    }
    // Generate GS1 barcode
    if data, err := xqrcode.EncodeGS1("(01)00625251888886(18)110201(10)456((D)CB(21)9876543210"); err == nil {
        _, _ = xqrcode.QRCodeToFile(data, "QRCode", 14, "M", 3, 4, "#ffffff", "#000000", true, false, false, "svg",
            "./gs1.svg")
    }
    // Listen HTTP request then generate and return a barcode
    mux := http.NewServeMux()
    mux.HandleFunc("/qrcode", func(w http.ResponseWriter, r *http.Request) {
        kind := xqrcode.QRCodeKind(r.PostFormValue("kind"))
        format := xqrcode.QRCodeFormat(r.PostForm.Get("format"))
        module, _ := strconv.Atoi(r.PostForm.Get("module"))
        escape := r.PostForm.Get("escape") == "escape"
        gs1 := r.PostForm.Get("gs1") == "gs1"
        structure := r.PostForm.Get("color") == "color"
        version, _ := strconv.Atoi(r.PostForm.Get("version"))
        ecc := xqrcode.QRCodeEccLevel(r.PostForm.Get("ecc"))
        quiet, _ := strconv.Atoi(r.PostForm.Get("quiet"))
        mirror := r.PostForm.Get("mirror") == "mirror"
        invert := r.PostForm.Get("invert") == "invert"
        bgColor := r.PostForm.Get("bgcolor")
        fgColor := r.PostForm.Get("fgcolor")
        barcode := r.PostForm.Get("barcode")
        data := []byte(barcode)
        if gs1 {
            if gs1Data, e := xqrcode.EncodeGS1(barcode); e != nil {
                w.Header().Set("QR-Error", base64.StdEncoding.EncodeToString([]byte(e.Error())))
                http.Error(w, e.Error(), http.StatusInternalServerError)
            } else {
                data = gs1Data
                escape = true
            }
        }
        if structure {
            fgColor += "#"
        }
        if _, e, _ := xqrcode.QRCodeToHTTP(data, kind, version, ecc, module, quiet, bgColor, fgColor, escape, mirror,
            invert, format, w); e != nil {
            w.Header().Set("QR-Error", base64.StdEncoding.EncodeToString([]byte(e.Error())))
            http.Error(w, e.Error(), http.StatusInternalServerError)
        }
    })
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        size := len(index)
        w.Header().Set("Content-Type", "text/html")
        w.Header().Set("Content-Length", strconv.Itoa(size))
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write(index)
    })
    fmt.Println("Listen on :9000")
    if err := http.ListenAndServe(":9000", mux); err != nil {
        log.Fatal(err)
    }
}
