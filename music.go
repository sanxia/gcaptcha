package gcaptcha

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"sort"
)

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/sanxia/glib"
	"github.com/sanxia/gmusic"
)

/* ================================================================================
 * 五线谱图片
 * qq group: 582452342
 * email   : 2091938785@qq.com
 * author  : 美丽的地球啊 - mliu
 * ================================================================================ */
type (
	musicImage struct {
		title   string
		texts   []string //外部数据源
		head    string
		option  ImageOption
		itemMap map[int]string //数据映射
		cellMap map[int]string //文字映射
		colors  []*image.Uniform
		width   int
		height  int
		count   int
		music   *gmusic.Music
	}
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 初始化五线谱图
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func NewMusicImage(title string, texts []string, head string, count int) IImage {
	musicImage := &musicImage{
		option: ImageOption{
			FontSize: 12,
		},
	}

	musicImage.title = title
	musicImage.texts = texts
	musicImage.head = head
	musicImage.count = count

	//init
	musicImage.itemMap = make(map[int]string, 0)
	musicImage.cellMap = make(map[int]string, 0)

	musicImage.colors = make([]*image.Uniform, 0)
	musicImage.colors = append(musicImage.colors, &image.Uniform{color.RGBA{0x00, 0x64, 0x00, 0xff}})
	musicImage.colors = append(musicImage.colors, &image.Uniform{color.RGBA{0x00, 0x00, 0x8b, 0xff}})
	musicImage.colors = append(musicImage.colors, &image.Uniform{color.RGBA{0x55, 0x6b, 0x2f, 0xff}})
	musicImage.colors = append(musicImage.colors, &image.Uniform{color.RGBA{0x00, 0x64, 0x00, 0xff}})
	musicImage.colors = append(musicImage.colors, &image.Uniform{color.RGBA{0x00, 0x64, 0x00, 0xff}})

	musicImage.music = gmusic.NewMusic()

	return musicImage
}

func (s *musicImage) SetOption(option ImageOption) {
	s.option = option
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取图片数据
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *musicImage) GetImage() ([]byte, error) {
	var imageBuffer bytes.Buffer

	headerHeight := s.option.HeaderHeight
	width := s.option.CellWidth
	height := s.option.CellHeight
	texts := s.shuffle()

	s.width = s.count*(width+s.option.Gap) + s.option.Gap + ((s.count - 1) * s.option.Padding)
	s.height = 1*(height+s.option.Gap) + s.option.Gap + headerHeight + (2 * s.option.Padding)

	//偏移点
	offsetPoint := image.Point{s.option.Padding, s.option.Padding}

	//画布
	graphics := image.NewRGBA(image.Rect(0, 0, s.width, s.height))

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

	offsetPoint = image.Point{s.option.Padding, offsetPoint.Y}
	if len(s.title) > 0 {
		offsetPoint = image.Point{s.option.Padding, offsetPoint.Y + headerHeight}
	}

	//线条图
	musicImage, offsets, _ := s.getMusicLineImage()
	draw.Draw(graphics, musicImage.Bounds().Add(offsetPoint), musicImage, image.ZP, draw.Over)

	//音名图
	if circleImage, err := s.getMusicNameImage(texts, offsets); err == nil {
		offsetPoint.X = offsetPoint.X + glib.RandIntRange(30, 50)
		draw.Draw(graphics, circleImage.Bounds().Add(offsetPoint), circleImage, image.ZP, draw.Over)
	}

	//谱号图
	if len(s.head) > 0 {
		clefHightImage, _ := glib.GetImageFile(s.head)
		draw.Draw(graphics, clefHightImage.Bounds().Add(image.Point{10, 60}), clefHightImage, image.ZP, draw.Over)
	} else {
		if clefImage, err := s.getMusicClefImage(); err == nil {
			draw.Draw(graphics, clefImage.Bounds().Add(image.Point{10, 40}), clefImage, image.ZP, draw.Over)
		}
	}

	if err := png.Encode(&imageBuffer, graphics); err != nil {
		return nil, err
	}

	return imageBuffer.Bytes(), nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取标题图
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *musicImage) getTitleImage() (image.Image, error) {
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
	pt := freetype.Pt(0, 14)

	for _, s := range s.title {
		if _, err := ctx.DrawString(string(s), pt); err != nil {
			return nil, err
		}

		pt.X += ctx.PointToFixed(space)
	}

	return graphics, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取线图
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *musicImage) getMusicLineImage() (image.Image, []int, error) {
	dstImg := image.NewRGBA(image.Rect(0, 0, s.width, s.height))
	draw.Draw(dstImg, dstImg.Bounds(), image.Transparent, image.ZP, draw.Src)

	font, _ := s.getFont(s.option.FontPath)
	rowOffsets := make([]int, 0)

	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFontSize(6)
	ctx.SetFont(font)
	ctx.SetClip(dstImg.Bounds())
	ctx.SetDst(dstImg)
	ctx.SetSrc(image.Black)

	rowOffset := freetype.Pt(0, 0)
	for i := 0; i < 5; i++ {
		rowOffset.Y = ctx.PointToFixed(float64((i+1)*16) + float64(glib.RandIntRange(1, 2)))

		rowOffsets = append(rowOffsets, rowOffset.Y.Floor()-8)
		rowOffsets = append(rowOffsets, rowOffset.Y.Floor())

		if i == 4 {
			rowOffsets = append(rowOffsets, rowOffset.Y.Floor()+8)
		}

		lineOffset := freetype.Pt(0, 0)
		for j := 0; j < 94; j++ {
			if j > 0 {
				lineOffset.X += ctx.PointToFixed(float64(2.5))
			}
			lineOffset.Y = rowOffset.Y + ctx.PointToFixed(float64(glib.RandIntRange(0, 2)))
			if _, err := ctx.DrawString("-", lineOffset); err != nil {
				return nil, nil, err
			}
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(rowOffsets)))

	return dstImg, rowOffsets, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取音名图
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *musicImage) getMusicNameImage(texts []string, offsets []int) (image.Image, error) {
	dstImg := image.NewRGBA(image.Rect(0, 0, s.width, s.height))
	draw.Draw(dstImg, dstImg.Bounds(), image.Transparent, image.ZP, draw.Src)

	offsetPoint := image.Point{}

	log.Printf("texts: %#v", texts)

	for _, text := range texts {
		musicName := s.music.GetMusicNameByCode(text)

		musicLineIndexs := make([]int, 0)
		for _, musicLine := range s.music.GetMusicLinesByMusicName(musicName.Name) {
			musicLineIndexs = append(musicLineIndexs, musicLine.Index)
		}

		musicLineIndex := musicLineIndexs[0]
		if len(musicLineIndexs) > 1 {
			currentLocationIndex := glib.RandIntRange(0, len(musicLineIndexs))
			musicLineIndex = musicLineIndexs[currentLocationIndex]
		}

		offsetX := 15 + glib.RandIntRange(8, 15)
		offsetY := offsets[musicLineIndex] - 5

		if musicName.IsBlack() {
			offsetY = offsets[musicLineIndex] - 4
		}

		offsetPoint.X += offsetX
		offsetPoint.Y = offsetY

		log.Printf("musicName:%s, musicNameIndex:%d, musicLine index:%d, musicLine indexs:%#v", musicName.Name, musicName.Index, musicLineIndex, musicLineIndexs)

		if musicName.IsBlack() {
			//升降号
			font, _ := s.getFont(s.option.FontPath)
			ctx := freetype.NewContext()
			ctx.SetDPI(72)
			ctx.SetFontSize(10)
			ctx.SetFont(font)
			ctx.SetClip(dstImg.Bounds())
			ctx.SetDst(dstImg)
			ctx.SetSrc(image.Black)

			pt := freetype.Pt(offsetPoint.X-7, offsetPoint.Y+5)
			if _, err := ctx.DrawString(string(text[0]), pt); err != nil {
				return nil, err
			}
		}

		//音名
		/*
			srcImg := &image.Uniform{color.RGBA{120, 126, 60, 255}}
			if len(text) > 1 {
				srcImg = &image.Uniform{color.RGBA{200, 200, 200, 255}}
			}
		*/

		srcImg := s.colors[glib.RandIntRange(0, 5)]

		dstRect := image.Rect(0, 0, 5, 8).Add(offsetPoint)
		draw.Draw(dstImg, dstRect, srcImg, image.ZP, draw.Src)

	}

	return dstImg, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取谱号图
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *musicImage) getMusicClefImage() (image.Image, error) {
	dstImg := image.NewRGBA(image.Rect(0, 0, s.width, s.height))
	draw.Draw(dstImg, dstImg.Bounds(), image.Transparent, image.ZP, draw.Src)

	srcImg := &image.Uniform{color.RGBA{90, 50, 160, 255}}

	dstRect := image.Rect(0, 0, 30, 55)
	draw.Draw(dstImg, dstRect, srcImg, image.ZP, draw.Src)

	return dstImg, nil
}

func (s *musicImage) shuffle() []string {
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
func (s *musicImage) getTextWidth(fontSize int) int {
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
func (s *musicImage) getFont(fontPath string) (*truetype.Font, error) {
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
func (s *musicImage) GetText() []string {
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
