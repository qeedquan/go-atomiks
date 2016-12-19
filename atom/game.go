package atom

import (
	"bufio"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"
)

const (
	LEVELS   = 30
	TILESIZE = 16
)

const (
	FREE  = 128
	ATOM  = 64
	WALL  = 192
	TYPE  = 192
	INDEX = 63
)

type Cursor struct {
	image.Point
	Mx, My int
	Sx, Sy int
	Ex, Ey int
	Type   int
	State  int
	Moving bool
}

type Grid struct {
	Squares [64][64]int
	Width   int
	Height  int
}

type Loosetile struct {
	image.Point
	Atom   int
	Mx, My int
	Ex, Ey int
	Dx, Dy int
}

func (g *Grid) Set(x, y, v int) {
	g.Squares[y][x] = v
}

func (g *Grid) At(x, y int) int {
	return g.Squares[y][x]
}

func (g *Grid) Type(x, y int) int {
	return g.Squares[y][x] & TYPE
}

func (g *Grid) Index(x, y int) int {
	return g.Squares[y][x] & INDEX
}

func (g *Grid) Tile(gfx *GFX, x, y int) *image.RGBA {
	var tile *image.RGBA
	index := g.At(x, y) & INDEX
	switch g.Type(x, y) {
	case ATOM:
		tile = gfx.Atom[index]
	case WALL:
		tile = gfx.Wall[index]
	case FREE:
		tile = gfx.Empty
	}
	return tile
}

type Game struct {
	conf        *Config
	screen      *Display
	gfx         *GFX
	Editor      bool
	Cursor      Cursor
	Field       Grid
	Solution    Grid
	BG          int
	Desc        [2][15]byte
	Offset      image.Point
	Level       int
	Score       int
	Hiscore     int
	TimeEnd     time.Time
	Duration    time.Duration
	PreviewTick time.Time
	Loose       Loosetile
	Loosing     bool
	Paused      bool
	PauseTime   time.Duration
}

func NewGame(conf *Config, screen *Display, gfx *GFX, editor bool) *Game {
	return &Game{
		conf:   conf,
		screen: screen,
		gfx:    gfx,
		Editor: editor,
	}
}

func (g *Game) MovedDistance(dir int) int {
	x := g.Cursor.X
	y := g.Cursor.Y
	if g.Field.Type(x, y) != ATOM {
		return 0
	}

	i := 0
	switch dir {
	case UP:
		y--
		for i = y; i >= 0; i-- {
			if g.Field.Type(x, i) != FREE {
				break
			}
		}
		return y - i
	case RIGHT:
		x++
		for i = x; i < 16; i++ {
			if g.Field.Type(i, y) != FREE {
				break
			}
		}
		return i - x
	case DOWN:
		y++
		for i = y; i < 16; i++ {
			if g.Field.Type(x, i) != FREE {
				break
			}
		}
		return i - y
	case LEFT:
		x--
		for i = x; i >= 0; i-- {
			if g.Field.Type(i, y) != FREE {
				break
			}
		}
		return x - i
	}

	return 0
}

func (g *Game) Load(level int) {
	defer func() {
		if g.Editor {
			g.Offset = image.Pt(32, 32)
			g.Cursor.Point = image.ZP
		}
	}()
	*g = Game{
		conf:    g.conf,
		screen:  g.screen,
		gfx:     g.gfx,
		Editor:  g.Editor,
		Offset:  image.Pt(80, 48),
		TimeEnd: time.Now().Add(60 * time.Second),
		Level:   level,
		Score:   500,
	}

	conf := g.conf
	if level < len(conf.Hiscores) {
		g.Hiscore = conf.Hiscores[level]
	}

	name := filepath.Join(conf.Assets, fmt.Sprintf("lev/lev%04d.dat", level))
	fd, err := os.Open(name)
	if err != nil {
		return
	}
	defer fd.Close()

	r := bufio.NewReader(fd)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			g.Field.Set(x, y, readByte(r))
			t := g.Field.Type(x, y)
			if (t == ATOM || t == FREE) && g.Cursor.Point == image.ZP {
				g.Cursor.Point = image.Pt(x, y)
			}
		}
	}

	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			g.Solution.Set(x, y, readByte(r))
		}
	}

	g.Duration = time.Duration(readShort(r))
	g.TimeEnd = time.Now().Add(g.Duration * time.Second)

	for i := range g.Desc {
		for x := range g.Desc[i] {
			g.Desc[i][x] = byte(readByte(r))
		}
	}

	g.Cursor.Type = readByte(r)
	g.BG = readByte(r)

	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if t := g.Field.Type(x, y); t == ATOM || t == WALL {
				if x+1 > g.Field.Width {
					g.Field.Width = x + 1
				}
				if y+1 > g.Field.Height {
					g.Field.Height = y + 1
				}
			}
		}
	}

	g.Offset.X += (15 - g.Field.Width) * 8
	g.Offset.Y = (15 - g.Field.Height) * 8

	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if t := g.Solution.Type(x, y); t == ATOM || t == WALL {
				if x+1 > g.Solution.Width {
					g.Solution.Width = x + 1
				}
				if y+1 > g.Solution.Height {
					g.Solution.Height = y + 1
				}
			}
		}
	}
}

func (g *Game) Save(level int) error {
	conf := g.conf
	name := filepath.Join(conf.Assets, fmt.Sprintf("lev%04d.dat", level))
	fd, err := os.Create(name)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(fd)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			w.WriteByte(byte(g.Field.At(x, y)))
		}
	}

	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			w.WriteByte(byte(g.Solution.At(x, y)))
		}
	}

	w.WriteByte(byte(g.Duration >> 8))
	w.WriteByte(byte(g.Duration))

	for i := range g.Desc {
		for x := 0; x < 15; x++ {
			w.WriteByte(g.Desc[i][x])
		}
	}

	w.WriteByte(byte(g.Cursor.Type))
	w.WriteByte(byte(g.BG))

	err = w.Flush()
	xerr := fd.Close()
	if err == nil {
		err = xerr
	}

	return err
}

func (g *Game) Won() bool {
	fx, fy := g.Field.Width, g.Field.Height
	sx, sy := g.Solution.Width, g.Solution.Height

	if fx == 0 || fy == 0 {
		return false
	}

	for y := 0; y <= fy-sy; y++ {
	check:
		for x := 0; x <= fx-sx; x++ {
			for yy := 0; yy < sy; yy++ {
				for xx := 0; xx < sx; xx++ {
					a := g.Solution.Type(xx, yy)
					b := g.Field.Type(x+xx, y+yy)
					if a == ATOM && a != b {
						continue check
					}
				}
			}
			return true
		}
	}
	return false
}

func (g *Game) DrawField() {
	g.drawGrid(&g.Field, 64, 64)
}

func (g *Game) DrawSolution() {
	g.drawGrid(&g.Solution, 32, 32)
}

func (g *Game) DrawTile(x, y int, tile *image.RGBA) {
	r := tile.Bounds()
	x = g.Offset.X + x*r.Dx()
	y = g.Offset.Y + y*r.Dy()
	DrawGFX(g.screen, tile, x, y)
}

func (g *Game) drawGrid(grid *Grid, width, height int) {
	screen := g.screen
	gfx := g.gfx

	DrawGFX(screen, gfx.BG[g.BG], 0, 0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			tile := grid.Tile(gfx, x, y)
			if tile == nil {
				continue
			}
			r := tile.Bounds()
			xx := g.Offset.X + x*r.Dx()
			yy := g.Offset.Y + y*r.Dy()
			DrawGFX(screen, gfx.Empty, xx, yy)
			DrawGFX(screen, tile, xx, yy)
		}
	}
}

func (g *Game) DrawPreview() {
	conf := g.conf
	screen := g.screen
	gfx := g.gfx

	s := &g.Solution
	DrawGFX(screen, gfx.Info, 0, 0)
	DrawGFX(screen, gfx.Levsel, 0, 0)
	if conf.MaxAuthLevel > 1 {
		DrawGFX(screen, gfx.Levsel2, 0, 0)
	}
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			if s.Type(x, y) != ATOM {
				continue
			}
			i := s.Index(x, y)
			t := gfx.Satom[i]
			oy := (7 - s.Height) * TILESIZE / 4
			px := x*TILESIZE/2 + WIDTH/4 - s.Width*TILESIZE/4
			py := 95 + oy + y*TILESIZE/2
			DrawGFX(screen, t, px, py)
		}
	}

	r := gfx.Font2[0].Bounds()
	x := WIDTH/4 - r.Dx()
	DrawGFX(screen, gfx.Font2[g.Level/10], x, 185)
	x += r.Dx()
	DrawGFX(screen, gfx.Font2[g.Level%10], x, 185)

	if g.Level < conf.MaxAuthLevel || conf.Unlocked {
		r := gfx.Completed.Bounds()
		DrawGFX(screen, gfx.Completed, 10+WIDTH/4-r.Dx()/2, 110)
	}
}

func (g *Game) DrawSmallPreview() {
	screen := g.screen
	gfx := g.gfx

	s := &g.Solution
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			if s.Type(x, y) != ATOM {
				continue
			}

			i := s.Index(x, y)
			t := gfx.Satom[i]
			px := ((8 - s.Width) * (TILESIZE / 2)) / 2
			py := ((7 - s.Height) * (TILESIZE / 2)) / 2
			if g.Desc[1][0] == 0 {
				py -= 8
			}
			rx := 4 + px + (x * TILESIZE / 2)
			ry := 16 + (240 - 71) + py + (y * TILESIZE / 2)
			DrawGFX(screen, t, rx, ry)
		}
	}
}
