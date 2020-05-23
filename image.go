package gcaptcha

/* ================================================================================
 * 五线谱图片
 * qq group: 582452342
 * email   : 2091938785@qq.com
 * author  : 美丽的地球啊 - mliu
 * ================================================================================ */
type (
	IImage interface {
		GetText() []string
		GetImage() ([]byte, error)
		SetOption(ImageOption)
	}

	ImageOption struct {
		HeaderHeight int
		CellWidth    int
		CellHeight   int
		Gap          int
		Padding      int
		Backgroud    string
		FontPath     string
		FontSize     float64
	}
)
