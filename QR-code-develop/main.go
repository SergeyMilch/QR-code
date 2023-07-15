package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/apex/gateway"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	qr "github.com/skip2/go-qrcode"
)

var answerCode = 200

type Encoder struct {
	AlphaThreshold int
	GreyThreshold  int
	QRLevel        qr.RecoveryLevel
}

// DefaultEncoder — кодировщик с настройками по умолчанию.
var DefaultEncoder = Encoder{
	AlphaThreshold: 2000,
	GreyThreshold:  30,
	QRLevel:        qr.Highest,
}

type AddLogo struct {
	AvFile image.Image
}

type Logo interface {
	UpLogoFile() image.Image
}

func setupRouter() *gin.Engine {

	router := gin.Default()

	router.POST("/url_logo", qrUrlLogo)
	router.POST("/text", qrGenText)
	router.POST("/url", qrURL)
	router.POST("/email", qrSendMail)
	router.POST("/phone", qrTel)
	router.POST("/sms", qrSms)
	router.POST("/wifi", qrWifi)
	router.POST("/maps", qrLocation)

	return router
}

// -----------------------------------------------QR для текста-------------------------------------------------//

func qrGenText(c *gin.Context) {

	defer returnError(c)

	//c.Writer.Header().Set("Content-type", "image/png")

	text := c.PostForm("text")

	if text == "" {
		panicWrapper("Text is empty", http.StatusBadRequest)
	}
	if len(text) > 4000 {
		panicWrapper("Text is too long", http.StatusBadRequest)
	}

	QRcode, err := qr.Encode(text, qr.Medium, 512)
	if err != nil {
		panicWrapper("404 Not Found", http.StatusNotFound)
	}

	Logo := AddLogo{}
	logoRes := Logo.UpLogoFile(c)

	if logoRes != nil {
		qr, err := Encode(text, logoRes, 512)
		if err != nil {
			panicWrapper("404 Not Found", http.StatusNotFound)
		}
		QRcode = qr.Bytes()
	}

	c.Data(http.StatusOK, "image/png", QRcode)
}

// -------------------------------------------------QR code для URL-----------------------------------------------//

func qrURL(c *gin.Context) {

	defer returnError(c)

	URL := c.PostForm("url")

	if !isValidUrl(URL) {
		panicWrapper("Wrong URL format", http.StatusBadRequest)
	}

	QRcode, err := qr.Encode(URL, qr.Medium, 512)
	if err != nil {
		panicWrapper("404 Not Found", http.StatusNotFound)
	}

	Logo := AddLogo{}
	logoRes := Logo.UpLogoFile(c)

	if logoRes != nil {
		qr, err := Encode(URL, logoRes, 512)
		if err != nil {
			panicWrapper("404 Not Found", http.StatusNotFound)
		}
		QRcode = qr.Bytes()
	}
	c.Data(http.StatusOK, "image/png", QRcode)
}

// ---------------------------------------------------QR для URL с логотипом ----------------------------------------//

func qrUrlLogo(c *gin.Context) {

	defer returnError(c)

	url := c.PostForm("url")

	if !isValidUrl(url) {
		panicWrapper("Wrong URL format", http.StatusBadRequest)
	}

	Logo := AddLogo{}
	logoRes := Logo.UpLogoFile(c)

	qr, err := Encode(url, logoRes, 512)
	if err != nil {
		panicWrapper("404 Not Found", http.StatusNotFound)
	}

	c.Data(http.StatusOK, "image/png", qr.Bytes())
}

// --------------------------------------------------QR код для E-mail----------------------------------------------//

func qrSendMail(c *gin.Context) {

	defer returnError(c)

	email := c.PostForm("email")

	if !validMail(email) {
		panicWrapper("E-mail is wrong", http.StatusBadRequest)
	}

	subject := c.PostForm("subject")

	// if len(subject) > 200 {
	// 	panicWrapper("Subject text is too long", http.StatusBadRequest)
	// }

	body := c.PostForm("body")

	// if len(subject) > 8000 {
	// 	panicWrapper("Body text is too long", http.StatusBadRequest)
	// }

	resultString := fmt.Sprintf("mailto:%s?subject=%s&body=%s", email, subject, body)

	QRcode, err := qr.Encode(resultString, qr.Medium, 512)
	if err != nil {
		panicWrapper("404 Not Found", http.StatusNotFound)
	}

	Logo := AddLogo{}
	logoRes := Logo.UpLogoFile(c)

	if logoRes != nil {
		qr, err := Encode(resultString, logoRes, 512)
		if err != nil {
			panicWrapper("404 Not Found", http.StatusNotFound)
		}
		QRcode = qr.Bytes()
	}

	c.Data(http.StatusOK, "image/png", QRcode)
}

// --------------------------------------------------QR код для номера телефона----------------------------------------------//

func qrTel(c *gin.Context) {

	defer returnError(c)

	tel := c.PostForm("tel")

	is_numeric := is_alphanum(tel)

	if is_numeric {
		resultString := fmt.Sprintf("tel:+%s", tel)

		QRcode, err := qr.Encode(resultString, qr.Medium, 512)
		if err != nil {
			panicWrapper("404 Not Found", http.StatusNotFound)
		}

		Logo := AddLogo{}
		logoRes := Logo.UpLogoFile(c)

		if logoRes != nil {
			qr, err := Encode(resultString, logoRes, 512)
			if err != nil {
				panicWrapper("404 Not Found", http.StatusNotFound)
			}
			QRcode = qr.Bytes()
		}

		c.Data(http.StatusOK, "image/png", QRcode)

	} else {
		panicWrapper("Phone number format is wrong", http.StatusBadRequest)
	}
}

// --------------------------------------------------QR код для SMS----------------------------------------------//

func qrSms(c *gin.Context) {

	defer returnError(c)

	tel := c.PostForm("phone")

	body := c.PostForm("body")

	if body == "" {
		panicWrapper("Text is empty", http.StatusBadRequest)
	}
	if len(body) > 160 {
		panicWrapper("Text is too long", http.StatusBadRequest)
	}

	is_numeric := is_alphanum(tel)

	if is_numeric {
		resultString := fmt.Sprintf("sms:+%v?&body=%v", tel, body)

		QRcode, err := qr.Encode(resultString, qr.Medium, 512)
		if err != nil {
			panicWrapper("404 Not Found", http.StatusNotFound)
		}

		Logo := AddLogo{}
		logoRes := Logo.UpLogoFile(c)

		if logoRes != nil {
			qr, err := Encode(resultString, logoRes, 512)
			if err != nil {
				panicWrapper("404 Not Found", http.StatusNotFound)
			}
			QRcode = qr.Bytes()
		}

		c.Data(http.StatusOK, "image/png", QRcode)
	} else {
		panicWrapper("Phone number format is wrong", http.StatusBadRequest)
	}
}

// --------------------------------------------------QR код для Wi-Fi----------------------------------------------//

func qrWifi(c *gin.Context) {

	defer returnError(c)

	const contentFmt = "WIFI:S:%s;T:%s;P:%s;;"

	ssid := c.PostForm("ssid")

	password := c.PostForm("password")

	encryptType := c.PostForm("encryptType")

	switch encryptType {
	case "WPA":
		encryptType = "WPA"
	case "WPA2":
		encryptType = "WPA2"
	case "WEP":
		encryptType = "WEP"
	case "wpa":
		encryptType = "WPA"
	case "wpa2":
		encryptType = "WPA2"
	case "wep":
		encryptType = "WEP"
	default:
		encryptType = "NONE"
	}

	resultString := fmt.Sprintf(contentFmt, ssid, encryptType, password)

	QRcode, err := qr.Encode(resultString, qr.Medium, 512)
	if err != nil {
		panicWrapper("404 Not Found", http.StatusNotFound)
	}

	Logo := AddLogo{}
	logoRes := Logo.UpLogoFile(c)

	if logoRes != nil {
		qr, err := Encode(resultString, logoRes, 512)
		if err != nil {
			panicWrapper("404 Not Found", http.StatusNotFound)
		}
		QRcode = qr.Bytes()
	}

	c.Data(http.StatusOK, "image/png", QRcode)
}

// --------------------------------------------------QR код для Location----------------------------------------------//

func qrLocation(c *gin.Context) {

	defer returnError(c)

	latitude := c.PostForm("latitude")

	if !isNumDot(latitude) {
		panicWrapper("Latitude format is wrong", http.StatusBadRequest)
	}

	longitude := c.PostForm("longitude")

	if !isNumDot(longitude) {
		panicWrapper("Longitude format is wrong", http.StatusBadRequest)
	}

	replaceCommaLat := strings.ReplaceAll(latitude, ",", ".")

	replaceCommaLon := strings.ReplaceAll(longitude, ",", ".")

	lat, _ := strconv.ParseFloat(replaceCommaLat, 64)

	lon, _ := strconv.ParseFloat(replaceCommaLon, 64)

	resultString := fmt.Sprintf("http://maps.google.com/maps?q=%g,%g", float64(lat), float64(lon))

	QRcode, err := qr.Encode(resultString, qr.Medium, 512)
	if err != nil {
		panicWrapper("404 Not Found", http.StatusNotFound)
	}

	Logo := AddLogo{}
	logoRes := Logo.UpLogoFile(c)

	if logoRes != nil {
		qr, err := Encode(resultString, logoRes, 512)
		if err != nil {
			panicWrapper("404 Not Found", http.StatusNotFound)
		}
		QRcode = qr.Bytes()
	}

	c.Data(http.StatusOK, "image/png", QRcode)
}

// Encode кодирует QR-изображение, добавляет наложение логотипа и отображает результат в формате PNG.
func Encode(str string, logo image.Image, size int) (*bytes.Buffer, error) {
	return DefaultEncoder.Encode(str, logo, size)
}

func (e Encoder) Encode(str string, logo image.Image, size int) (*bytes.Buffer, error) {

	var buf bytes.Buffer

	code, err := qr.New(str, e.QRLevel)
	if err != nil {
		return nil, err
	}

	img := code.Image(size)

	resultImg := image.NewRGBA(img.Bounds())
	e.overlayLogo(resultImg, img)
	e.overlayLogo(resultImg, logo)

	err = png.Encode(&buf, resultImg)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

// Размещение логотипа по-середине QR кода
func (e Encoder) overlayLogo(dst, src image.Image) {

	offset := dst.Bounds().Max.X/2 - src.Bounds().Max.X/2
	yOffset := dst.Bounds().Max.Y/2 - src.Bounds().Max.Y/2
	draw.Draw(dst.(draw.Image), dst.Bounds().Add(image.Pt(offset, yOffset)), src, image.Point{}, draw.Over)
}

// // Лого в центре, лого черно-белый пиксельный
// func (e Encoder) overlayLogo(dst, src image.Image) {
// 	grey := uint32(^uint16(0)) * uint32(e.GreyThreshold) / 100
// 	alphaOffset := uint32(e.AlphaThreshold)
// 	offset := dst.Bounds().Max.X/2 - src.Bounds().Max.X/2
// 	for x := 0; x < src.Bounds().Max.X; x++ {
// 		for y := 0; y < src.Bounds().Max.Y; y++ {

// 			if r, g, b, alpha := src.At(x, y).RGBA(); alpha > alphaOffset {
// 				col := color.Black
// 				if r > grey && g > grey && b > grey {
// 					col = color.White
// 				}
// 				dst.(*image.Paletted).Set(x+offset, y+offset, col)
// 			}
// 		}
// 	}
// }

// Загружает логотип PNG, JPG
func (a AddLogo) UpLogoFile(c *gin.Context) image.Image {

	AvFile, _ := c.FormFile("file")

	if AvFile == nil {
		log.Println("In QRcode won't be logo")
		return nil
	}

	log.Println(AvFile.Filename)

	file, err := AvFile.Open()
	if err != nil {
		panicWrapper("File format is wrong", http.StatusUnsupportedMediaType)
	}
	defer file.Close()

	logo, _, err := image.Decode(file)
	if err != nil {
		panicWrapper("404 Not Found", http.StatusNotFound)
	}

	// userValue / 5 = width
	// png 400*1000
	// 400/100 = 4
	// 1000/4 = 250
	// 100*250
	logoRes := resize.Resize(0, 100, logo, resize.Bilinear)

	defer returnError(c)

	return logoRes
}

// Проверка номера телефона
func is_alphanum(str string) bool {
	return regexp.MustCompile(`^[0-9+\(\)#\.\s\/ext-]+$`).MatchString(str)
}

// Проверка Email-адреса
func validMail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Проверка URL
func isValidUrl(str string) bool {
	_, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}
	u, err := url.Parse(str)
	if err != nil || u.Host == "" {
		return false
	}
	return true
}

// Проверка ввода широты и долготы
func isNumDot(s string) bool {
	dotFound := false
	for _, v := range s {
		if v == '.' || v == ',' {
			if dotFound {
				return false
			}
			dotFound = true
		} else if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

func main() {
	if inLambda() {
		fmt.Println("running aws lambda in aws")
		log.Fatal(gateway.ListenAndServe(":8080", setupRouter()))
	} else {
		fmt.Println("running aws lambda in local")
		log.Fatal(http.ListenAndServe(":8080", setupRouter()))
	}
}

func inLambda() bool {
	if lambdaTaskRoot := os.Getenv("LAMBDA_TASK_ROOT"); lambdaTaskRoot != "" {
		return true
	}
	return false
}

// Возврат кастомных ошибок
func returnError(c *gin.Context) {
	if err := recover(); err != nil {
		c.JSON(answerCode, gin.H{
			"status":  "error",
			"message": err,
		})
	}
}

func panicWrapper(err string, code int) {
	answerCode = code
	panic(err)
}
