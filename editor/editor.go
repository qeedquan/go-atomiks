package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlgfx"

	"github.com/qeedquan/go-atomiks/atom"
)

var (
	conf   *atom.Config
	screen *atom.Display
	gfx    *atom.GFX
	game   *atom.Game
	view   int
	level  int
	char   int
	line   int
	item   int
	fps    sdlgfx.FPSManager
)

func main() {
	runtime.LockOSThread()
	flag.Usage = usage
	conf = atom.NewConfig(true)
	if flag.NArg() < 1 {
		usage()
	}
	level, _ = strconv.Atoi(flag.Arg(0))
	screen = atom.NewDisplay(conf, "Editor", false)
	gfx = atom.LoadGFX(conf)
	game = atom.NewGame(conf, screen, gfx, true)
	game.Load(level)
	line = 1

	fps.Init()
	fps.SetRate(60)
	for {
		event()
		blit()
		fps.Delay()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: editor [options] level")
	flag.PrintDefaults()
	os.Exit(2)
}

func event() {
	g := game
	for {
		ev := sdl.PollEvent()
		if ev == nil {
			break
		}
		switch ev := ev.(type) {
		case sdl.QuitEvent:
			os.Exit(0)
		case sdl.KeyDownEvent:
			switch ev.Sym {
			case sdl.K_ESCAPE:
				os.Exit(0)
			case sdl.K_LEFT:
				if g.Cursor.X > 0 {
					g.Cursor.X--
				}
			case sdl.K_RIGHT:
				if g.Cursor.X < 63 {
					g.Cursor.X++
				}
			case sdl.K_UP:
				if g.Cursor.Y > 0 {
					g.Cursor.Y--
				}
			case sdl.K_DOWN:
				if g.Cursor.Y < 63 {
					g.Cursor.Y++
				}
			case sdl.K_SPACE:
				setItem()
			case sdl.K_INSERT:
				if view == 0 {
					g.Field.Set(g.Cursor.X, g.Cursor.Y, item)
				} else {
					g.Solution.Set(g.Cursor.X, g.Cursor.Y, item)
				}
			case sdl.K_RETURN:
				if view == 0 {
					x := g.Cursor.X
					y := g.Cursor.Y
					f := &g.Field
					switch f.Type(x, y) {
					case atom.FREE:
						f.Set(x, y, atom.WALL)
					case atom.WALL:
						f.Set(x, y, atom.ATOM)
					case atom.ATOM:
						f.Set(x, y, 0)
					default:
						f.Set(x, y, atom.FREE)
					}
					item = f.At(x, y)
				}
			case sdl.K_DELETE:
				if view == 0 {
					g.Field.Set(g.Cursor.X, g.Cursor.Y, 0)
				} else {
					g.Solution.Set(g.Cursor.X, g.Cursor.Y, 0)
				}
			case sdl.K_TAB:
				view = 1 - view
			case sdl.K_KP_MINUS:
				if g.Duration > 0 {
					g.Duration--
				}
			case sdl.K_KP_PLUS:
				if g.Duration < 3600 {
					g.Duration++
				}
			case sdl.K_F1:
				if char++; char >= 15 {
					line = 3 - line
					char = 0
				}
			case sdl.K_F2:
				if g.Cursor.Type++; g.Cursor.Type > 2 {
					g.Cursor.Type = 1
				}
			case sdl.K_F3:
				g.BG = (g.BG + 1) % 3
			case sdl.K_F5:
				err := g.Save(level)
				if err != nil {
					sdl.Log("Failed to save to file: %v", err)
				} else {
					sdl.Log("Saved")
				}
			default:
				key := sdlk2char(ev.Sym)
				switch {
				case key == '.':
					key = 0
					fallthrough
				case 'A' <= key && key <= 'Z':
					g.Desc[line-1][char] = byte(key)
				}
			}
		}
	}
}

func setItem() {
	g := game
	f := &g.Field
	if view != 0 {
		f = &g.Solution
	}
	x := g.Cursor.X
	y := g.Cursor.Y
	t := f.At(x, y)
	switch typ := f.Type(x, y); {
	case typ == atom.ATOM && view == 0:
		t = (((t & atom.INDEX) + 1) % 48) | atom.ATOM
	case typ == atom.WALL && view == 0:
		t = (((t & atom.INDEX) + 1) % 18) | atom.WALL
	case typ == atom.ATOM && view == 1:
		t &= atom.INDEX
		if t++; t > 48 {
			t = 0
		} else {
			t |= atom.ATOM
		}
	case typ != atom.ATOM && view == 1:
		t = atom.ATOM
	}
	f.Set(x, y, t)
	item = t
}

func blit() {
	screen.Clear()
	switch view {
	case 0:
		game.DrawField()
	case 1:
		game.DrawSolution()
	}
	blitTimer()
	blitDesc()
	blitCursor()
	screen.Flush()
}

func blitTimer() {
	t := game.Duration
	x := 260
	y := 20
	atom.DrawGFX(screen, gfx.Font2[t/60], x, y)

	r := gfx.Font2[0].Bounds()
	x += r.Dx()
	atom.DrawGFX(screen, gfx.Font2[10], x, y)

	x += r.Dx()
	atom.DrawGFX(screen, gfx.Font2[(t%60)/10], x, y)

	x += r.Dx()
	atom.DrawGFX(screen, gfx.Font2[t%10], x, y)
}

func blitDesc() {
	g := game

	x := 210
	y := 220
	r := gfx.Font3[0].Bounds()
	w := r.Dx() * 2
	h := r.Dy() * 2

	for i := range g.Desc {
		for j, c := range g.Desc[i] {
			if line == i+1 && char == j {
				atom.DrawRect(screen, x*2, y*2, w, h, 255, 0, 0, 255)
			} else {
				atom.DrawRect(screen, x*2, y*2, w, h, 0x30, 0x30, 0x30, 255)
			}
			if 'A' <= c && c <= 'Z' {
				atom.DrawGFX(screen, gfx.Font3[c-'A'], x, y)
			}
			x += r.Dx()
		}
		x = 210
		y += r.Dy() + 2
	}
}

func blitCursor() {
	g := game

	x := 300
	y := 130
	atom.DrawGFX(screen, gfx.Cursor[g.Cursor.Type], x, y)

	r := gfx.Cursor[0].Bounds()
	x = 32 + g.Cursor.X*r.Dx()
	y = 32 + g.Cursor.Y*r.Dy()
	atom.DrawGFX(screen, gfx.Cursor[0], x, y)
}

func sdlk2char(key sdl.Keycode) int {
	switch key {
	case sdl.K_a:
		return 'A'
	case sdl.K_b:
		return 'B'
	case sdl.K_c:
		return 'C'
	case sdl.K_d:
		return 'D'
	case sdl.K_e:
		return 'E'
	case sdl.K_f:
		return 'F'
	case sdl.K_g:
		return 'G'
	case sdl.K_h:
		return 'H'
	case sdl.K_i:
		return 'I'
	case sdl.K_j:
		return 'J'
	case sdl.K_k:
		return 'K'
	case sdl.K_l:
		return 'L'
	case sdl.K_m:
		return 'M'
	case sdl.K_n:
		return 'N'
	case sdl.K_o:
		return 'O'
	case sdl.K_p:
		return 'P'
	case sdl.K_q:
		return 'Q'
	case sdl.K_r:
		return 'R'
	case sdl.K_s:
		return 'S'
	case sdl.K_t:
		return 'T'
	case sdl.K_u:
		return 'U'
	case sdl.K_v:
		return 'V'
	case sdl.K_w:
		return 'W'
	case sdl.K_x:
		return 'X'
	case sdl.K_y:
		return 'Y'
	case sdl.K_z:
		return 'Z'
	case sdl.K_PERIOD:
		return '.'
	default:
		return ' '
	}
}
