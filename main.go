package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"

	"github.com/golang/freetype/truetype"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func main() {
	http.HandleFunc("/view", view)
	http.HandleFunc("/download", download)

	http.ListenAndServe(":"+*flag.String("port", "8000", "Port server HTTP"), nil)
	flag.Parse()
}

func view(w http.ResponseWriter, r *http.Request) {
	fmt.Println("anda menakses view")

	text := r.FormValue("text")
	label := r.FormValue("label")
	kode := r.FormValue("kode")

	qr, err := generateQR(text, label, kode)
	if err != nil {
		http.Error(w, "Error Generate QR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "image/jpeg")
	jpeg.Encode(w, qr, &jpeg.Options{Quality: 100})
}

func download(w http.ResponseWriter, r *http.Request) {
	fmt.Println("anda mengakses download")

	text := r.FormValue("text")
	label := r.FormValue("label")
	kode := r.FormValue("kode")

	fileName := label + ".jpeg"

	qr, _ := generateQR(text, label, kode)

	w.Header().Set("Content-Disposition", "attachment; filename=smartlink.jpeg")
	w.Header().Set("Content-Type", "Image/image")

	png.Encode(w, qr)
	http.ServeFile(w, r, fileName)
}

func generateQR(text string, label string, kode string) (*image.RGBA, error) {

	qrCode, err := qrcode.New(text, qrcode.Highest)
	if err != nil {
		return nil, err
	}

	qrCode.DisableBorder = true
	labelWidth := len(label) * 7
	labelImg1, _, err_labelImg1 := makeLabelImg(label, labelWidth, 37, "Helvetica-Font/Helvetica-Bold.ttf", 500, "label1")
	if err_labelImg1 != nil {
		return nil, err_labelImg1
	}

	labelImg2, widhtLabel, err_labelImg2 := makeLabelImg("QRIS by NOBU", labelWidth, 37, "Helvetica-Font/Helvetica-Bold.ttf", 300, "label2")
	if err_labelImg2 != nil {
		return nil, err_labelImg2
	}

	labelImg3, _, err_labelImg3 := makeLabelImg("AIOT Laundry System", labelWidth, 40, "Helvetica-Font/Helvetica.ttf", 500, "label3")
	if err_labelImg3 != nil {
		return nil, err_labelImg3
	}

	labelImg4, _, err_labelImg4 := makeLabelImg("by Smartlink.id", labelWidth, 40, "Helvetica-Font/Helvetica-Bold.ttf", 500, "label4")
	if err_labelImg4 != nil {
		return nil, err_labelImg4
	}

	labelImg5, widhtLabel5, err_labelImg5 := makeLabelImg(kode, labelWidth, 100, "Helvetica-Font/Helvetica-Bold.ttf", 500, "label5")
	if err_labelImg5 != nil {
		return nil, err_labelImg5
	}

	importImg, err_importImg := os.Open("washer.png")
	if err_importImg != nil {
		return nil, err
	}

	baseBackground, _, _ := image.Decode(importImg)
	// fmt.Println(baseBackground.Bounds().Dx()) => 998
	// fmt.Println(baseBackground.Bounds().Dy()) => 1228
	qrImage := qrCode.Image(baseBackground.Bounds().Dx() - 200)
	qrImageDraw := image.NewRGBA(baseBackground.Bounds())

	x := (baseBackground.Bounds().Dx() - (baseBackground.Bounds().Dx() - 200)) / 2
	y := (baseBackground.Bounds().Dx() - (baseBackground.Bounds().Dx() - 150)) / 2
	draw.Draw(qrImageDraw, qrImageDraw.Bounds(), baseBackground, image.Point{}, draw.Over)
	draw.Draw(qrImageDraw, qrImage.Bounds().Add(image.Pt(x, y)), qrImage, image.Point{}, draw.Over)
	draw.Draw(qrImageDraw, qrImageDraw.Bounds().Add(image.Pt(97, qrImageDraw.Bounds().Dy()-315)), labelImg1, image.Point{}, draw.Over)
	draw.Draw(qrImageDraw, qrImageDraw.Bounds().Add(image.Pt(qrImageDraw.Bounds().Dx()-(widhtLabel+97), qrImageDraw.Bounds().Dy()-315)), labelImg2, image.Point{}, draw.Over)
	draw.Draw(qrImageDraw, qrImageDraw.Bounds().Add(image.Pt(50, qrImageDraw.Bounds().Dy()-170)), labelImg3, image.Point{}, draw.Over)
	draw.Draw(qrImageDraw, qrImageDraw.Bounds().Add(image.Pt(50, qrImageDraw.Bounds().Dy()-120)), labelImg4, image.Point{}, draw.Over)
	draw.Draw(qrImageDraw, qrImageDraw.Bounds().Add(image.Pt(qrImageDraw.Bounds().Dx()-(widhtLabel5+78), 1045)), labelImg5, image.Point{}, draw.Over)
	return qrImageDraw, nil
}

func makeLabelImg(label string, width int, fontSize int, fontFile string, maxWidth int, toLabel string) (*image.RGBA, int, error) {
	fontFamily, err := callFont(fontFile, float64(fontSize))
	if err != nil {
		return nil, 0, err
	}

	widthLabelSettings := settingWidthLabel(label, fontFamily)
	fmt.Println("ini widht label setting", widthLabelSettings)

	labelImg := image.NewRGBA(image.Rect(0, 0, maxWidth, fontSize*2))
	draw.Draw(labelImg, labelImg.Bounds().Add(image.Pt(0, 0)), image.Transparent, image.Point{}, draw.Over)

	if toLabel == "label1" || toLabel == "label2" {
		d := &font.Drawer{
			Dst:  labelImg,
			Src:  image.Black,
			Face: fontFamily,
			Dot:  fixed.P(0, fontSize+5),
		}
		d.DrawString(label)
	}

	if toLabel == "label3" || toLabel == "label4" || toLabel == "label5" {
		d := &font.Drawer{
			Dst:  labelImg,
			Src:  image.White,
			Face: fontFamily,
			Dot:  fixed.P(0, fontSize+5),
		}
		d.DrawString(label)
	}
	return labelImg, widthLabelSettings, nil
}

func callFont(fontFile string, fontSize float64) (font.Face, error) {
	dataFont, err := os.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}

	fontFormat, err := truetype.Parse(dataFont)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(fontFormat, &truetype.Options{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	}), nil

}

func settingWidthLabel(label string, face font.Face) int {
	width := 5

	for _, r := range label {
		_, advance, ok := face.GlyphBounds(r)
		if !ok {
			advance, _ = face.GlyphAdvance(r)
		}
		width += int(advance.Ceil())
	}

	return width
}
