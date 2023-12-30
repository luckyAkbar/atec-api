package model

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"math"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// maxImageWidth are made to limit the maximum image output width
const maxImageWitdh = 1080

// optimumTextLength was calculated by counting how much chars can be written on maxImageWidth
// without overflowing. The original is 70, but made to 65 to give the room between text and image
// border
const optimumTextLength = 65

// ImageResult will be the result of image generation.
// Needed to be able to return the content-type if more than 1 image type
// generation is supported.
type ImageResult struct {
	ContentType string
	Buffer      bytes.Buffer
}

// SDResultImageGenerator interface
type SDResultImageGenerator interface {
	GenerateJPEG() *ImageResult
}

// SDResultImageGenerationOpts options to generate image for sd test result
type SDResultImageGenerationOpts struct {
	Title          string
	Result         SDTestResult
	TestID         uuid.UUID
	IndicationText string

	rgba         *image.RGBA
	ttp          []string
	width        int
	height       int
	titleDrawer  *font.Drawer
	textDrawer   *font.Drawer
	sampleDrawer *font.Drawer
	font         *truetype.Font
	dpi          float64
	textSize     float64
	titleSize    float64
	spacing      float64
}

// NewResultGenerator factory to make image generator
func NewResultGenerator(f *truetype.Font, opts *SDResultImageGenerationOpts) SDResultImageGenerator {
	dpi := float64(208)
	size := float64(12)
	spacing := float64(1.5)
	titleSize := float64(18)

	initialTitleDrawer := &font.Drawer{
		Face: truetype.NewFace(f, &truetype.Options{
			Size:    titleSize,
			DPI:     dpi,
			Hinting: font.HintingFull,
		}),
	}

	initialTextDrawer := &font.Drawer{
		Face: truetype.NewFace(f, &truetype.Options{
			Size:    size,
			DPI:     dpi,
			Hinting: font.HintingNone,
		}),
	}

	genOpts := &SDResultImageGenerationOpts{
		Title:          opts.Title,
		Result:         opts.Result,
		TestID:         opts.TestID,
		IndicationText: opts.IndicationText,
		sampleDrawer:   initialTextDrawer,
		spacing:        spacing,
		font:           f,
		textSize:       size,
		titleSize:      titleSize,
		dpi:            dpi,
	}

	genOpts.generateTTP()
	genOpts.countOptimumImageWidth(initialTextDrawer, initialTitleDrawer)
	genOpts.countOptimumImageHeight()

	rgba := image.NewRGBA(image.Rect(0, 0, genOpts.width, genOpts.height))
	draw.Draw(rgba, rgba.Bounds(), image.White, image.Point{}, draw.Src)
	genOpts.rgba = rgba
	genOpts.generateTextDrawer()
	genOpts.generateTitleDrawer()

	return genOpts
}

// GenerateJPEG will generate jpeg image for the test result
func (o *SDResultImageGenerationOpts) GenerateJPEG() *ImageResult {
	y := 10 + int(math.Ceil(o.textSize*o.dpi/72))
	dy := int(math.Ceil(o.textSize * o.spacing * o.dpi / 72))
	o.textDrawer.Dot = fixed.Point26_6{
		X: (fixed.I(o.width) - o.textDrawer.MeasureString(o.Title)) / 2,
		Y: fixed.I(y),
	}

	ty := 10 + int(math.Ceil(o.titleSize*o.dpi/72))
	tdy := int(math.Ceil(o.titleSize * o.spacing * o.dpi / 72))

	tx := (fixed.I(o.width) - o.titleDrawer.MeasureString(o.Title)) / 2

	o.titleDrawer.Dot = fixed.Point26_6{
		X: tx,
		Y: fixed.I(ty),
	}

	o.titleDrawer.DrawString(o.Title)
	y += tdy
	for _, s := range o.ttp {
		center := (fixed.I(o.width) - o.textDrawer.MeasureString(s)) / 2
		o.textDrawer.Dot = fixed.P(center.Ceil(), y)
		o.textDrawer.DrawString(s)
		y += dy
	}

	var imgBuf bytes.Buffer
	if err := jpeg.Encode(&imgBuf, o.rgba, nil); err != nil {
		logrus.WithError(err).Error("failed to encode image")
	}

	return &ImageResult{
		ContentType: "image/jpeg",
		Buffer:      imgBuf,
	}
}

func (o *SDResultImageGenerationOpts) generateTTP() {
	for _, r := range o.Result.Result {
		o.appendTTP(fmt.Sprintf("%s: %d", r.GroupName, r.Result))
	}
	o.appendTTP(fmt.Sprintf("Total: %d", o.Result.Total))
	o.appendTTP(fmt.Sprintf("Indikasi: %s", o.IndicationText))
	o.appendTTP(fmt.Sprintf("Test ID: %s", o.TestID))
}

func (o *SDResultImageGenerationOpts) appendTTP(s string) {
	arrStr := o.ensureSafeLongText(s, o.sampleDrawer)
	o.ttp = append(o.ttp, arrStr...)
}

func (o *SDResultImageGenerationOpts) countOptimumImageWidth(initialTextDrawer, initialTitleDrawer *font.Drawer) {
	maxWidth := initialTitleDrawer.MeasureString(o.Title)
	for _, t := range o.ttp {
		ms := initialTextDrawer.MeasureString(t)
		if ms > maxWidth {
			maxWidth = ms
		}
	}

	if maxWidth >= maxImageWitdh {
		o.width = maxImageWitdh
	} else {
		o.width = maxWidth.Ceil() + 5*maxWidth.Ceil()/100
	}
}

// ensureSafeLongText will try to check if writing s will cause text overflow
// or if the text length more than optimumTextLength.
// If text deemed to long, will call wordWrapper to wrap the text and prevent overflow
func (o *SDResultImageGenerationOpts) ensureSafeLongText(s string, drawer *font.Drawer) []string {
	width := drawer.MeasureString(s)
	if width.Ceil() >= maxImageWitdh || len(s) >= optimumTextLength {
		return wordWrapper(s)
	}

	return []string{s}
}

func wordWrapper(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{s}
	}

	result := []string{}

	// convert string to slice
	ss := strings.Fields(s)
	lastIdx := 0
	var temp string
	appended := false
	for {
		temp += ss[lastIdx] + " "
		if len(temp) >= optimumTextLength {
			result = append(result, temp)
			temp = ""
			appended = true
		}

		if lastIdx == len(ss)-1 && !appended {
			result = append(result, temp)
			break
		}

		lastIdx++
		appended = false
	}
	return result

}

func (o *SDResultImageGenerationOpts) countOptimumImageHeight() {
	y := 10 + int(math.Ceil(o.textSize*o.dpi/72))
	tdy := int(math.Ceil(o.titleSize * o.spacing * o.dpi / 72))
	y += tdy

	incrementor := int(math.Ceil(o.textSize * o.spacing * o.dpi / 72))
	y += incrementor * len(o.ttp)

	o.height = y
}

func (o *SDResultImageGenerationOpts) generateTextDrawer() {
	o.textDrawer = &font.Drawer{
		Dst: o.rgba,
		Src: image.Black,
		Face: truetype.NewFace(o.font, &truetype.Options{
			Size:    o.textSize,
			DPI:     o.dpi,
			Hinting: font.HintingNone,
		}),
	}
}

func (o *SDResultImageGenerationOpts) generateTitleDrawer() {
	o.titleDrawer = &font.Drawer{
		Dst: o.rgba,
		Src: image.Black,
		Face: truetype.NewFace(o.font, &truetype.Options{
			Size:    o.titleSize,
			DPI:     o.dpi,
			Hinting: font.HintingFull,
		}),
	}
}
