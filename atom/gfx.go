package atom

import (
	"image"
	"image/color"
	"path/filepath"

	"github.com/qeedquan/go-media/image/imageutil"
	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlimage"
	"github.com/qeedquan/go-media/sdl/sdlimage/sdlcolor"
	"github.com/qeedquan/go-media/sdl/sdlmixer"
	"golang.org/x/image/draw"
)

const (
	WIDTH  = 640
	HEIGHT = 480
)

type Display struct {
	*sdl.Window
	*sdl.Renderer
	*sdl.Texture
	*image.RGBA
	conf *Config
}

type GFX struct {
	Title        *image.RGBA
	Credit       *image.RGBA
	Timeout      *image.RGBA
	Info         *image.RGBA
	Paused       *image.RGBA
	Instructions *image.RGBA
	Intro        [3]*image.RGBA
	Levsel       *image.RGBA
	Levsel2      *image.RGBA
	Atom         [49]*image.RGBA
	Satom        [49]*image.RGBA
	Wall         [19]*image.RGBA
	Explosion    [8]*image.RGBA
	Empty        *image.RGBA
	Preview      [2]*image.RGBA
	BG           [3]*image.RGBA
	Completed    *image.RGBA
	Black        *image.RGBA
	Cursor       [3]*image.RGBA
	Font1        [37]*image.RGBA
	Font2        [11]*image.RGBA
	Font3        [26]*image.RGBA
}

func NewDisplay(conf *Config, title string, icon bool) *Display {
	err := sdl.Init(sdl.INIT_EVERYTHING &^ sdl.INIT_AUDIO)
	ck(err)

	err = sdl.InitSubSystem(sdl.INIT_AUDIO)
	ek(err)

	err = sdlmixer.OpenAudio(44100, sdl.AUDIO_S16, 2, 8192)
	ek(err)

	_, err = sdlmixer.Init(sdlmixer.INIT_OGG)
	ek(err)

	sdlmixer.AllocateChannels(128)

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "best")

	width, height := WIDTH, HEIGHT
	wflag := sdl.WINDOW_RESIZABLE
	if conf.Fullscreen {
		wflag |= sdl.WINDOW_FULLSCREEN_DESKTOP
	}
	window, renderer, err := sdl.CreateWindowAndRenderer(width, height, wflag)
	ck(err)

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, width, height)
	ck(err)

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))

	window.SetTitle(title)
	if icon {
		surface := loadSurface(conf, "atomiks.png")
		if surface != nil {
			window.SetIcon(surface)
			surface.Free()
		}
	}

	renderer.SetLogicalSize(width, height)
	sdl.ShowCursor(sdl.DISABLE)

	return &Display{window, renderer, texture, canvas, conf}
}

func LoadGFX(conf *Config) *GFX {
	g := &GFX{}
	g.Title = loadImage(conf, "title.png")
	g.Credit = loadImage(conf, "credits.png")
	g.Timeout = loadImage(conf, "timeout.png")
	g.Info = loadImage(conf, "infoscreen.png")
	g.Paused = loadImage(conf, "pausedscreen.png")
	g.Instructions = loadImage(conf, "instructs.png")
	g.Intro[0] = loadImage(conf, "intro1.png")
	g.Intro[1] = loadImage(conf, "intro2.png")
	g.Intro[2] = loadImage(conf, "intro3.png")
	g.Levsel = loadImage(conf, "levsel.png")
	g.Levsel2 = loadImage(conf, "levsel2.png")
	g.Completed = loadImage(conf, "completed.png")

	loadSheet(conf, "bg.png", 320, 240, g.BG[:])
	g.Black = loadImage(conf, "black.png")
	g.Preview[0] = loadImage(conf, "preview.png")
	g.Preview[1] = loadImage(conf, "preview2.png")
	g.Empty = loadImage(conf, "empty.png")
	loadSheet(conf, "atoms.png", 16, 16, g.Atom[:])
	loadSheet(conf, "satoms.png", 8, 8, g.Satom[:])
	loadSheet(conf, "explosion.png", 16, 16, g.Explosion[:])
	loadSheet(conf, "walls.png", 16, 16, g.Wall[:])
	loadSheet(conf, "cursors.png", 16, 16, g.Cursor[:])
	loadSheet(conf, "font1.png", 5, 5, g.Font1[:])
	loadSheet(conf, "font2.png", 14, 16, g.Font2[:])
	loadSheet(conf, "font3.png", 7, 8, g.Font3[:])

	return g
}

func loadSheet(conf *Config, name string, width, height int, sheet []*image.RGBA) *image.RGBA {
	rgba := loadImage(conf, name)
	for i := range sheet {
		sheet[i] = rgba.SubImage(image.Rect(i*width, 0, (i+1)*width, height)).(*image.RGBA)
	}
	return rgba
}

func loadImage(conf *Config, name string) *image.RGBA {
	name = filepath.Join(conf.Assets, "img", name)
	rgba, err := imageutil.LoadFile(name)
	ck(err)
	return rgba
}

func loadSurface(conf *Config, name string) *sdl.Surface {
	name = filepath.Join(conf.Assets, "img", name)
	surface, err := sdlimage.LoadSurfaceFile(name)
	ek(err)
	return surface
}

func (d *Display) Clear() {
	draw.Draw(d, d.Bounds(), image.Black, image.ZP, draw.Src)
}

func (d *Display) Flush() {
	d.SetDrawColor(sdlcolor.Black)
	d.Renderer.Clear()
	d.Update(nil, d.Pix, d.Stride)
	d.Copy(d.Texture, nil, nil)
	d.Present()
}

func DrawGFXPartial(dst draw.Image, src image.Image, x, y, w, h, xx, yy int) {
	s := 2
	xx, yy = xx*s, yy*s
	sr := image.Rect(x, y, x+w, y+h)
	dr := image.Rect(xx, yy, xx+w*s, yy+h*s)
	draw.NearestNeighbor.Scale(dst, dr, src, sr, draw.Over, nil)
}

func DrawGFX(dst draw.Image, src image.Image, x, y int) {
	s := 2
	sr := src.Bounds()
	if sr.Dx() == WIDTH && sr.Dy() == HEIGHT {
		s = 1
	}
	dr := image.Rect(x*s, y*s, x*s+sr.Dx()*s, y*s+sr.Dy()*s)
	draw.NearestNeighbor.Scale(dst, dr, src, sr, draw.Over, nil)
}

func DrawRect(dst draw.Image, x, y, w, h int, r, g, b, a uint8) {
	c := image.NewUniform(color.RGBA{r, g, b, a})
	dr := image.Rect(x, y, x+w, y+h)
	draw.Draw(dst, dr, c, image.ZP, draw.Over)
}
