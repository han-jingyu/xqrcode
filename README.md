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
// Generate SVG file:
_, _ = xqrcode.QRCodeToFile(
    []byte("abc1234567890"), // barcode data
    xqrcode.QRCode,          // barcode type
    15,                      // version
    xqrcode.QREccMedium,     // error correction level
    3,                       // module
    4,                       // quiet zone
    "#ffffff",               // background color
    "#000000",               // foreground color
    false,                   // escape
    false,                   // mirror
    false,                   // invert
    xqrcode.QRCodeFormatSvg, // file format
    "./qr.svg")              // file path and name

// Generate PNG file:
_, _ = xqrcode.QRCodeToFile(
    []byte("abc1234567890"), // barcode data
    "QRCode",                // barcode type
    15,                      // version
    "M",                     // error correction level
    3,                       // module
    4,                       // quiet zone
    "#ffffff",               // background color
    "#000000",               // foreground color
    false,                   // escape
    false,                   // mirror
    false,                   // invert
    "png",                   // file format
    "./qr.png")              // file path and name
```

### Draw barcode in PDF:

```go
pdf := gopdf.GoPdf{}
pdf.Start(gopdf.Config{Unit: gopdf.UnitMM, PageSize: *gopdf.PageSizeA4})
pdf.AddPage()
_, _, _, _ = xqrcode.DrawQRCode1(
    []byte("abc1234567890"),  // barcode data
    xqrcode.QRCode,           // barcode type
    15,                       // version
    xqrcode.QREccMedium,      // error correction level
    4,                        // quiet zone
    false,                    // escape
    false,                    // mirror
    false,                    // invert
    "#ffffff",                // background color
    "#000000",                // foreground color
    0.3,                      // module
    80.0,                     // left
    20.0,                     // top
    false,                    // right alignment
    false,                    // bottom alignment
    func(left, top, width, height float64, r, g, b, a uint8, background bool) error {
        pdf.SetFillColor(r, g, b)
        return pdf.Rectangle(left, top, left+width, top+height, "F", 0, 0)
    })                        // function for drawing a rectangle
_ = pdf.WritePdf("1.pdf")
```

### Draw barcode in PNG:

#### Method 1

```go
iPng := image.NewRGBA(image.Rect(0, 0, 500, 500))
_, _, _, _ = xqrcode.DrawQRCode1(
    []byte("abc1234567890"),  // barcode data
    xqrcode.QRCode,           // barcode type
    7,                        // version
    xqrcode.QREccMedium,      // error correction level
    2,                        // quiet zone
    false,                    // escape
    false,                    // mirror
    false,                    // invert
    "#ffffff",                // background color
    "#000000",                // foreground color
    2,                        // module
    50,                       // left
    50,                       // top
    false,                    // right alignment
    false,                    // bottom alignment
    func(left, top, width, height float64, r, g, b, a uint8, background bool) error {
        dotColor := &image.Uniform{C: color.RGBA{R: r, G: g, B: b, A: a}}
        draw.Draw(iPng, image.Rect(int(left), int(top), int(left+width), int(top+height)),
            dotColor, image.Point{}, draw.Over) // function for drawing a rectangle
        return nil
    })
f, _ := os.Create("./1.png")
defer func() { _ = f.Close() }()
_ = png.Encode(f, iPng)
```

#### Method 2

```go
iPng := image.NewRGBA(image.Rect(0, 0, 500, 500))
_, _, _, _ = xqrcode.DrawQRCode2(
    []byte("abc1234567890"), // barcode data
    250,                     // left
    50,                      // top
    xqrcode.QRCode,          // barcode type
    7,                       // version
    xqrcode.QREccMedium,     // error correction level
    2,                       // module
    4,                       // quiet zone
    "#ffffff",               // background
    "#000000",               // foreground
    false,                   // escape
    false,                   // mirror
    false,                   // invert
    false,                   // right alignment
    false,                   // bottom alignment
    func(left, top int, r, g, b, a uint8, background bool) error {
        iPng.Set(left, top, color.RGBA{R: r, G: g, B: b, A: a})
        return nil
    })                       // function for setting a pixel
f, _ := os.Create("./2.png")
defer func() { _ = f.Close() }()
_ = png.Encode(f, iPng)
```

### Display barcode in terminal:

```go
if text, err, _ := xqrcode.SprintQRCode(
    []byte("abc1234567890"), // barcode data
    xqrcode.QRCode,          // barcode type
    8,                       // version
    xqrcode.QREccMedium,     // error correction level
    3,                       // quiet zone
    "",                      // foreground and(or) background color
    false,                   // escape
    false); err == nil {     // mirror
    fmt.Println(text)
}

// Customize foreground and(or) background color
if text, err, _ := xqrcode.SprintQRCode(
    []byte("abc1234567890"),       // barcode data
    xqrcode.QRCode,                // barcode type
    8,                             // version
    xqrcode.QREccMedium,           // error correction level
    3,                             // quiet zone
    "\033[48;5;190m\033[38;5;19m", // background and(or) foreground color
    false,                         // escape
    false); err == nil {           // mirror
    fmt.Println(text)
}
```

### Listen HTTP request then generate and return a barcode:

```go
mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
    kind := query.Get("kind")
    barcode := query.Get("barcode")
    _, _, _ = xqrcode.QRCodeToHTTP(
        []byte(barcode),            // barcode data
        xqrcode.QRCodeKind(kind),   // barcode type
        15,                         // version
        "M",                        // error correction level
        3,                          // module
        4,                          // quiet zone
        "#ffffff",                  // background color
        "#000000",                  // foreground color
        false,                      // escape
        false,                      // mirror
        false,                      // invert
        "svg",                      // output format
        w)                          // HTTP response
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
if data, err := xqrcode.EncodeGS1(
    "(01)00625251888886(18)110201(10)456((D)CB(21)9876543210"); // GS1 barcode
    err == nil {
    _, _ = xqrcode.QRCodeToFile(
        data,               // barcode data
        "rMQRCode",         // barcode type
        14,                 // version
        "M",                // error correction level
        3,                  // module
        4,                  // quiet zone
        "#ffffff",          // background color
        "#000000",          // foreground color
        true,               // escape (must be set to true)
        false,              // mirror
        false,              // invert
        "svg",              // output format
        "./gs1.svg")        // file to save
}
```

Note:

- The `escape` parameter must be set to `true`.
- When inserting `"("` into a GS1 AI element value, please use `"(("`.
