# xqrcode
QRCode, rMQRCode, and MicroQRCode library.

## Features

- Support QRCode, MicroQRCode, and rMQRCode.
- Support full features, including FNC1 character, Extended Channel Interpretation (ECI),
  Structured Append.
- Draw barcode in PDF, PNG, etc.
- Generate SVG, PNG file.
- Response HTTP request.
- Support GS1 barcode.

## Installatiom

```
go get -u macsvn.co/xqrcode
```

## Usage

### Generate SVG or PNG file:

```go
_, _ = xqrcode.QRCodeToFile([]byte("abc1234567890"), xqrcode.QRCode, 15, xqrcode.QREccMedium, 3, 4,
    "#ffffff", "#000000", false, false, false, xqrcode.QRCodeFormatSvg, "./qr.svg")

_, _ = xqrcode.QRCodeToFile([]byte("abc1234567890"), "QRCode", 15, "M", 3, 4,
    "#ffffff", "#000000", false, false, false, "png", "./qr.png")
```

### Draw barcode in PDF:

```go
pdf := gopdf.GoPdf{}
pdf.Start(gopdf.Config{Unit: gopdf.UnitMM, PageSize: *gopdf.PageSizeA4})
pdf.AddPage()
_, _, _, _ = xqrcode.DrawQRCode1([]byte("abc1234567890"), xqrcode.QRCode, 15, xqrcode.QREccMedium,
    4, false, false, false, "#ffffff", "#000000", 0.3, 80.0, 20.0, false, false,
    func(left, top, width, height float64, r, g, b, a uint8, background bool) error {
        pdf.SetFillColor(r, g, b)
        return pdf.Rectangle(left, top, left+width, top+height, "F", 0, 0)
    })
_ = pdf.WritePdf("1.pdf")
```

### Draw barcode in PNG:

#### Method 1

```go
iPng := image.NewRGBA(image.Rect(0, 0, 500, 500))
_, _, _, _ = xqrcode.DrawQRCode1([]byte("abc1234567890"), xqrcode.QRCode, 7, xqrcode.QREccMedium, 2,
    false, false, false, "#ffffff", "#000000", 2, 50, 50, false, false,
    func(left, top, width, height float64, r, g, b, a uint8, background bool) error {
        dotColor := &image.Uniform{C: color.RGBA{R: r, G: g, B: b, A: a}}
        draw.Draw(iPng, image.Rect(int(left), int(top), int(left+width), int(top+height)), dotColor,
            image.Point{}, draw.Over)
        return nil
    })
f, _ := os.Create("./1.png")
defer func() { _ = f.Close() }()
_ = png.Encode(f, iPng)
```

#### Method 2

```go
iPng := image.NewRGBA(image.Rect(0, 0, 500, 500))
_, _, _, _ = xqrcode.DrawQRCode2([]byte("abc1234567890"), 250, 50, xqrcode.QRCode, 7,
    xqrcode.QREccMedium, 2, 4, "#ffffff", "#000000", false, false, false, false, false,
    func(left, top int, r, g, b, a uint8, background bool) error {
        iPng.Set(left, top, color.RGBA{R: r, G: g, B: b, A: a})
        return nil
    })
f, _ := os.Create("./2.png")
defer func() { _ = f.Close() }()
_ = png.Encode(f, iPng)
```

### Display barcode in terminal:

```go
if text, err, _ := xqrcode.SprintQRCode([]byte("abc1234567890"), xqrcode.QRCode, 8,
    xqrcode.QREccMedium, 3, "", false, false); err == nil {
    fmt.Println(text)
}

// Customize foreground and(or) background color
if text, err, _ := xqrcode.SprintQRCode([]byte("abc1234567890"), xqrcode.QRCode, 8,
    xqrcode.QREccMedium, 3, "\033[48;5;190m\033[38;5;19m", false, false); err == nil {
    fmt.Println(text)
}
```

### Listen HTTP request then grnerate and reture a barcode:

```go
mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
    kind := query.Get("kind")
    barcode := query.Get("barcode")
    _, _, _ = xqrcode.QRCodeToHTTP([]byte(barcode), xqrcode.QRCodeKind(kind), 15, "M", 3, 4,
        "#ffffff", "#000000", false, false, false, "svg", w)
})
fmt.Println("Listen on :9000")
if err := http.ListenAndServe(":9000", mux); err != nil {
    log.Fatal(err)
}
```

### Generate GS1 barcode

Use the "`EncodeGS1`" function to convert GS1 barcode text to `[]byte` data, then generate the
barcode symbol as before. For example:

```go
if data, err := xqrcode.EncodeGS1("(01)00625251888886(18)110201(10)456((D)CB(21)9876543210"); err == nil {
    _, _ = xqrcode.QRCodeToFile(data, "rMQRCode", 14, "M", 3, 4, "#ffffff", "#000000", true, false,
        false, "svg", "./gs1.svg")
}
```

Note: When inserting `"("` into a GS1 AI element value, please use `"(("`.
