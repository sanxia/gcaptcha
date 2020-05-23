package gcaptcha

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"sort"
)

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/sanxia/glib"
)

/* ================================================================================
 * 网格图片
 * qq group: 582452342
 * email   : 2091938785@qq.com
 * author  : 美丽的地球啊 - mliu
 * ================================================================================ */
type (
	gridImage struct {
		Title         string
		HeaderHeight  int
		CellWidth     int
		CellHeight    int
		Gap           int
		PaddingWidth  int
		PaddingHeight int
		Backgroud     string
		FontPath      string
		ImagePath     string
		datas         []*GridItem       //外部数据源
		itemMap       map[int]*GridItem //数据映射
		cellMap       map[int]string    //格子图片文件名映射
		width         int
		height        int
		targetIndex   int //当前目标项目索引
		count         int
	}

	GridItem struct {
		Title          string
		Path           string
		Words          []string
		Filenames      []int
		SelectedIndexs []int
	}
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取网格图实例
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func NewGridImage(count int, datas []*GridItem) *gridImage {
	gridImage := new(gridImage)
	gridImage.count = count
	gridImage.datas = datas
	//gridImage.FontPath = "assets/font/华文仿宋.ttf"
	//gridImage.ImagePath = "assets/img/verify"

	//init
	for _, item := range gridImage.datas {
		item.SelectedIndexs = make([]int, 0)
	}

	gridImage.targetIndex = glib.RandIntRange(0, len(gridImage.datas))
	gridImage.itemMap = map[int]*GridItem{
		gridImage.targetIndex: gridImage.datas[gridImage.targetIndex],
	}
	gridImage.cellMap = make(map[int]string, 0)

	gridImage.generateItems(3)
	gridImage.generateSelectedIndexs()
	gridImage.generateCellIndexs()

	return gridImage
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取数据索引
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *gridImage) GetData() []int {
	cellIndexs := make([]int, 0)
	for k, v := range s.itemMap {
		if k == s.targetIndex {
			for _, selectedIndex := range v.SelectedIndexs {
				filename := fmt.Sprintf("%s/%d", v.Path, v.Filenames[selectedIndex])
				for cellIndex, cellName := range s.cellMap {
					if cellName == filename {
						cellIndexs = append(cellIndexs, cellIndex)
					}
				}
			}
		}
	}

	return cellIndexs
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取图片数据
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *gridImage) GetImage() ([]byte, error) {
	var buf bytes.Buffer
	if s.Title == "" {
		s.Title = "找出所有的："
	}
	rows, columns := 0, 0
	headerHeight := s.HeaderHeight
	width := s.CellWidth
	height := s.CellHeight
	gap := s.Gap

	s.width = 3*(width+gap) + gap + (2 * s.PaddingWidth)
	s.height = 3*(width+gap) + gap + headerHeight + (2 * s.PaddingHeight)

	//画布偏移点
	offsetPoint := image.Point{s.PaddingWidth, s.PaddingHeight}

	graphics := image.NewRGBA(image.Rect(0, 0, s.width, s.height))

	//维持字典索引有序
	keys := make([]int, 0)
	for k := range s.cellMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//背景图
	if s.Backgroud != "" {
		backgroundImage, _ := glib.GetImageFile(glib.GetAbsolutePath(s.Backgroud))
		draw.Draw(graphics, graphics.Bounds(), backgroundImage, image.ZP, draw.Over)
	} else {
		white := color.RGBA{255, 255, 255, 255}
		draw.Draw(graphics, graphics.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)
	}

	//标题
	titleImage, _ := s.getTitleImage()
	draw.Draw(graphics, titleImage.Bounds().Add(offsetPoint), titleImage, image.ZP, draw.Over)

	//图片单元格
	for _, cellIndex := range keys {
		fontPath := fmt.Sprintf("%s%s%s.png", s.ImagePath, string(os.PathSeparator), s.cellMap[cellIndex])
		img, err := glib.GetImageFile(glib.GetAbsolutePath(fontPath))
		if err != nil {
			log.Printf("GetImageFile err: %v", err)
			return nil, err
		}

		if cellIndex == 3 || cellIndex == 6 {
			rows++
			columns = 0
		}

		x := columns*(width+gap) + gap
		y := rows*(height+gap) + gap + headerHeight
		r := image.Rect(x, y, s.width, s.height).Add(offsetPoint)

		draw.Draw(graphics, r, img, img.Bounds().Min, draw.Over)

		columns++
	}

	if err := png.Encode(&buf, graphics); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取字体
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *gridImage) getFont(fontPath string) (*truetype.Font, error) {
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
 * 获取标题图
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *gridImage) getTitleImage() (image.Image, error) {
	graphics := image.NewRGBA(image.Rect(0, 0, s.width, s.height))

	//white := color.RGBA{255, 255, 255, 255}
	//draw.Draw(graphics, graphics.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)
	draw.Draw(graphics, graphics.Bounds(), image.Transparent, image.ZP, draw.Src)

	font, _ := s.getFont(s.FontPath)
	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFontSize(12)
	ctx.SetFont(font)
	ctx.SetClip(graphics.Bounds())
	ctx.SetDst(graphics)

	ctx.SetSrc(image.White)

	space := float64(12)

	offsetY := s.HeaderHeight - 5
	if offsetY < 0 {
		offsetY = 0
	}

	pt := freetype.Pt(2, offsetY)
	for _, s := range s.Title {
		_, err := ctx.DrawString(string(s), pt)
		if err != nil {
			return nil, err
		}

		pt.X += ctx.PointToFixed(space)
	}

	targetTitle := s.itemMap[s.targetIndex].Title
	ctx.SetFontSize(16)
	for _, s := range targetTitle {
		_, err := ctx.DrawString(string(s), pt)
		if err != nil {
			return nil, err
		}
		pt.X += ctx.PointToFixed(space)
	}

	return graphics, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 生成项目映射
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *gridImage) generateItems(count int) {
	for count > 0 {
		index := glib.RandIntRange(0, len(s.datas))
		for {
			if _, ok := s.itemMap[index]; !ok {
				break
			} else {
				index = glib.RandIntRange(0, len(s.datas))
			}
		}

		s.itemMap[index] = s.datas[index]
		log.Printf("generateItems index: %d", index)
		count--
	}
	log.Printf("generateItems s.itemMap: %v", s.itemMap)

}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 生成每个项目的选中索引集合
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *gridImage) generateSelectedIndexs() {
	for k, v := range s.itemMap {
		filenameCount := 2
		if k == s.targetIndex {
			filenameCount = 3
		}

		maps := make(map[int]int, 0)

		for filenameCount > 0 {
			index := glib.RandIntRange(0, len(v.Filenames))

			for {
				if _, ok := maps[index]; !ok {
					break
				} else {
					index = glib.RandIntRange(0, len(v.Filenames))
				}
			}

			maps[index] = index

			//每个选中项的选中索引集合
			v.SelectedIndexs = append(v.SelectedIndexs, index)

			log.Printf("generateSelectedIndexs index: %d", index)
			log.Printf("generateSelectedIndexs k:%d, v: %v", k, v)

			filenameCount--
		}

		log.Printf("generateSelectedIndexs v.SelectedIndexs: %v", v.SelectedIndexs)
	}
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 生成单元格集合
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *gridImage) generateCellIndexs() {
	for _, v := range s.itemMap {

		//项目选中集合里的每个文件名, SelectedIndex对应着文件名映射
		for _, selectedIndex := range v.SelectedIndexs {
			index := glib.RandIntRange(0, s.count)
			for {
				if _, ok := s.cellMap[index]; !ok {
					break
				} else {
					index = glib.RandIntRange(0, s.count)
				}
			}
			filename := fmt.Sprintf("%s/%d", v.Path, v.Filenames[selectedIndex])
			s.cellMap[index] = filename

			log.Printf("generateCellIndexs index: %d", index)
		}

		log.Printf("generateCellIndexs s.cellMap: %v", s.cellMap)
	}
}
