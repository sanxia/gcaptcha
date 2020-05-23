package gcaptcha

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"sort"
	"strings"
)

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/sanxia/glib"
)

/* ================================================================================
 * 文字图片
 * qq group: 582452342
 * email   : 2091938785@qq.com
 * author  : 美丽的地球啊 - mliu
 * ================================================================================ */
type (
	textImage struct {
		title   string
		texts   []string //外部数据源
		option  ImageOption
		itemMap map[int]string //数据映射
		cellMap map[int]string //文字映射
		colors  []*image.Uniform
		width   int
		height  int
		count   int
	}
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 初始化文字图
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func NewTextImage(title string, texts []string, count int) IImage {
	textImage := &textImage{
		option: ImageOption{
			FontSize: 12,
		},
	}

	textImage.title = title
	textImage.texts = texts
	textImage.count = count

	//init
	textImage.itemMap = make(map[int]string, 0)
	textImage.cellMap = make(map[int]string, 0)

	textImage.colors = make([]*image.Uniform, 0)
	textImage.colors = append(textImage.colors, &image.Uniform{color.RGBA{120, 120, 50, 255}})
	textImage.colors = append(textImage.colors, &image.Uniform{color.RGBA{120, 126, 60, 255}})
	textImage.colors = append(textImage.colors, &image.Uniform{color.RGBA{120, 132, 40, 255}})
	textImage.colors = append(textImage.colors, &image.Uniform{color.RGBA{120, 120, 50, 255}})
	textImage.colors = append(textImage.colors, &image.Uniform{color.RGBA{120, 127, 14, 255}})

	return textImage
}

func (s *textImage) SetOption(option ImageOption) {
	s.option = option
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取图片数据
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *textImage) GetImage() ([]byte, error) {
	var imageBuffer bytes.Buffer

	headerHeight := s.option.HeaderHeight
	width := s.option.CellWidth
	height := s.option.CellHeight
	texts := s.shuffle()

	s.width = s.count*(width+s.option.Gap) + s.option.Gap + ((s.count - 1) * s.option.Padding)
	s.height = 1*(height+s.option.Gap) + s.option.Gap + headerHeight + (2 * s.option.Padding)

	//偏移点
	graphics := image.NewRGBA(image.Rect(0, 0, s.width, s.height))
	offsetPoint := image.Point{s.option.Padding, s.option.Padding}

	//背景图
	if s.option.Backgroud != "" {
		backgroundImage, _ := glib.GetImageFile(s.option.Backgroud)
		draw.Draw(graphics, graphics.Bounds(), backgroundImage, image.ZP, draw.Over)
	} else {
		white := color.RGBA{255, 255, 255, 255}
		draw.Draw(graphics, graphics.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)
	}

	//标题图
	if len(s.title) > 0 {
		if titleImage, err := s.getTitleImage(); err == nil {
			draw.Draw(graphics, titleImage.Bounds().Add(offsetPoint), titleImage, image.ZP, draw.Over)
		}
	}

	//文字图
	offsetPoint = image.Point{s.option.Padding, offsetPoint.Y}
	if len(s.title) > 0 {
		offsetPoint = image.Point{s.option.Padding, offsetPoint.Y + headerHeight}
	}

	textImage, _ := s.getTextImage(texts)
	draw.Draw(graphics, textImage.Bounds().Add(offsetPoint), textImage, image.ZP, draw.Over)

	if err := png.Encode(&imageBuffer, graphics); err != nil {
		return nil, err
	}

	return imageBuffer.Bytes(), nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取标题图
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *textImage) getTitleImage() (image.Image, error) {
	graphics := image.NewRGBA(image.Rect(0, 0, s.width, s.height))
	draw.Draw(graphics, graphics.Bounds(), image.Transparent, image.ZP, draw.Src)

	font, _ := s.getFont(s.option.FontPath)

	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFontSize(s.option.FontSize)
	ctx.SetFont(font)
	ctx.SetClip(graphics.Bounds())
	ctx.SetDst(graphics)
	ctx.SetSrc(image.Black)

	space := float64(12)
	pt := freetype.Pt(2, 14)

	for _, s := range s.title {
		if _, err := ctx.DrawString(string(s), pt); err != nil {
			return nil, err
		}

		pt.X += ctx.PointToFixed(space)
	}

	return graphics, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取文字图
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *textImage) getTextImage(texts []string) (image.Image, error) {
	graphics := image.NewRGBA(image.Rect(0, 0, s.width, s.height))
	draw.Draw(graphics, graphics.Bounds(), image.Transparent, image.ZP, draw.Src)

	font, _ := s.getFont(s.option.FontPath)

	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(font)
	ctx.SetClip(graphics.Bounds())
	ctx.SetDst(graphics)

	textPoint := freetype.Pt(2, 5)
	flags := make(map[int]bool, 0)
	var nextIndex int

	newTexts := strings.Join(texts, "")
	for _, text := range newTexts {
		colorIndex := glib.RandInt(len(s.colors))
		ctx.SetSrc(s.colors[colorIndex])

		fontSize := glib.RandIntRange(int(s.option.FontSize), int(s.option.FontSize)+2)
		offsetX := 14 + glib.RandIntRange(-2, 2)
		offsetY := glib.RandIntRange(12, 18)

		if string(text) == "#" || string(text) == "b" {
			fontSize = glib.RandIntRange(int(s.option.FontSize)-6, int(s.option.FontSize)-2)
			flags[nextIndex] = true
		}

		if flags[nextIndex] {
			offsetY = glib.RandIntRange(8, 14)
		}

		if nextIndex > 0 && flags[nextIndex-1] {
			offsetX = 8 + glib.RandIntRange(-5, 0)
		}

		ctx.SetFontSize(float64(fontSize))

		textPoint.X += ctx.PointToFixed(float64(offsetX))
		textPoint.Y = ctx.PointToFixed(float64(offsetY))

		if _, err := ctx.DrawString(string(text), textPoint); err != nil {
			return nil, err
		}

		nextIndex++
	}

	return graphics, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 随机文字
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *textImage) shuffle() []string {
	//随机打散texts到cellMap
	for index, text := range s.texts {
		s.itemMap[index] = text
	}

	index := glib.RandIntRange(0, s.count)
	s.cellMap[index] = s.itemMap[index]

	count := s.count

	//第一个已提前写入，所以是大于1
	for count > 1 {
		index = glib.RandIntRange(0, len(s.itemMap))
		for {
			if _, ok := s.cellMap[index]; !ok {
				break
			} else {
				index = glib.RandIntRange(0, len(s.itemMap))
			}
		}

		s.cellMap[index] = s.itemMap[index]

		count--
	}

	return s.GetText()
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取文字宽度
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *textImage) getTextWidth(fontSize int) int {
	font, _ := s.getFont(s.option.FontPath)
	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFontSize(float64(fontSize))
	ctx.SetFont(font)
	space := float64(fontSize)

	return int(ctx.PointToFixed(space))
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取字体
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *textImage) getFont(fontPath string) (*truetype.Font, error) {
	absolutePath := glib.GetAbsolutePath(fontPath)

	fontBytes, err := ioutil.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}

	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}

	return font, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取文字
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *textImage) GetText() []string {
	//维持字典索引有序
	keys := make([]int, 0)
	for keyIndex := range s.cellMap {
		keys = append(keys, keyIndex)
	}
	sort.Ints(keys)

	texts := make([]string, 0)
	for _, key := range keys {
		texts = append(texts, string(s.cellMap[key]))
	}

	return texts
}
