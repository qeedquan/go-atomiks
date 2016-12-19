package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlgfx"
	"github.com/qeedquan/go-media/sdl/sdlmixer"

	"github.com/qeedquan/go-atomiks/atom"
)

const (
	INTRO = iota + 1
	SELECT
	PLAY
	WON
	TIMEOUT
	CREDITS
	EXIT
)

var (
	conf   *atom.Config
	screen *atom.Display
	gfx    *atom.GFX
	sfx    *atom.SFX

	game    *atom.Game
	preview *atom.Game
	slider  atom.Slideshow
	won     struct {
		atoms []atom.Loosetile
		timer time.Time
		tick  time.Time
	}
	credits struct {
		y int
	}
	level    int
	state    int
	newstate int

	justStarted bool
	showCursor  bool

	fps sdlgfx.FPSManager
)

func main() {
	runtime.LockOSThread()
	rand.Seed(time.Now().UnixNano())
	conf = atom.NewConfig(false)
	screen = atom.NewDisplay(conf, "Atomiks", true)
	gfx = atom.LoadGFX(conf)
	sfx = atom.LoadSFX(conf)
	game = atom.NewGame(conf, screen, gfx, false)

	fps.Init()
	fps.SetRate(60)

	swtch(INTRO)
	for {
		if newstate != 0 {
			swtch(newstate)
			newstate = 0
		}
		event()
		update()
		blit()
		fps.Delay()
	}
}

func swtch(newstate int) {
	switch state = newstate; state {
	case INTRO:
		level = 1
		justStarted = true
		slider.Init(screen, []atom.Slide{
			{[]image.Image{gfx.Title}, 250 * 16 * time.Millisecond},
			{[]image.Image{gfx.Info, gfx.Intro[0]}, 0},
			{[]image.Image{gfx.Info, gfx.Intro[1]}, 0},
			{[]image.Image{gfx.Info, gfx.Intro[2]}, 0},
		}, true)
		sfx.PlayMusic(sfx.Title, 0)

	case SELECT:
		preview = atom.NewGame(conf, screen, gfx, true)
		preview.Load(level)

	case PLAY:
		sdlmixer.FadeOutMusic(2000)
		showCursor = true
		game.Load(level)

	case WON:
		if level == conf.MaxAuthLevel && conf.MaxAuthLevel < atom.LEVELS {
			conf.MaxAuthLevel++
		}

		won.atoms = won.atoms[:0]

		f := &game.Field
		for y := 0; y < f.Height; y++ {
			for x := 0; x < f.Width; x++ {
				if f.Type(x, y) == atom.ATOM {
					won.atoms = append(won.atoms, atom.Loosetile{
						Point: image.Pt(x, y),
					})
				}
			}
		}
		shuffle(won.atoms)
		won.tick = time.Now()
		won.timer = won.tick.Add(40 * time.Millisecond)
		showCursor = false

	case TIMEOUT:

	case CREDITS:
		credits.y = 0

	case EXIT:
		var slides []atom.Slide
		for i := 0; i < 16; i++ {
			alpha := 30 * i
			if alpha > 255 {
				alpha = 255
			}
			slides = append(slides, atom.Slide{
				[]image.Image{slider.Frame, image.NewUniform(color.RGBA{0, 0, 0, uint8(alpha)})},
				30 * time.Millisecond,
			})
		}
		slider.Init(screen, slides, false)
	}
}

func event() {
	for {
		ev := sdl.PollEvent()
		if ev == nil {
			break
		}
		switch ev := ev.(type) {
		case sdl.QuitEvent:
			os.Exit(0)

		case sdl.KeyDownEvent:
			switch key := atom.Key(ev.Sym); key {
			case atom.FULLSCREEN:
				if !conf.Fullscreen {
					screen.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP)
				} else {
					screen.SetFullscreen(0)
				}

			case atom.NONE:

			default:
				evState(key)
			}
		}
	}
}

func evState(key int) {
	switch state {
	case INTRO, EXIT:
		slider.Event(key)
	case SELECT:
		evSelect(key)
	case PLAY:
		evPlay(key)
	case WON:
		if key == atom.ESC {
			newstate = SELECT
		}
	case TIMEOUT:
		newstate = SELECT
	case CREDITS:
		if key == atom.ESC || key == atom.ENTER {
			newstate = SELECT
		}
	}
}

func evSelect(key int) {
	oldLevel := level
	switch key {
	case atom.ESC:
		atom.DrawGFX(slider.Frame, screen, 0, 0)
		newstate = EXIT
	case atom.LEFT:
		if level > 1 {
			level--
		}
	case atom.RIGHT:
		if level < atom.LEVELS && (level < conf.MaxAuthLevel || conf.Unlocked) {
			level++
		}
	case atom.HOME:
		level = 1
	case atom.END:
		level = conf.MaxAuthLevel
		if conf.Unlocked || level > atom.LEVELS {
			level = atom.LEVELS
		}
	case atom.ENTER:
		newstate = PLAY
	}

	if oldLevel != level {
		preview.Load(level)
	}
}

func evPlay(key int) {
	if justStarted {
		if key == atom.ESC {
			atom.DrawGFX(slider.Frame, screen, 0, 0)
			newstate = EXIT
		} else {
			justStarted = false
			game.TimeEnd = time.Now().Add(game.Duration * time.Second)
			game.PreviewTick = time.Now()
		}
		return
	}

	switch key {
	case atom.SPACE:
		game.Paused = !game.Paused
		if game.Paused {
			game.PauseTime = game.TimeEnd.Sub(time.Now())
		} else {
			game.TimeEnd = time.Now().Add(game.PauseTime)
		}
	}

	if game.Paused {
		return
	}

	switch key {
	case atom.ESC:
		newstate = SELECT
	case atom.LEFT:
		move(-1, 0, key)
	case atom.RIGHT:
		move(1, 0, key)
	case atom.UP:
		move(0, -1, key)
	case atom.DOWN:
		move(0, 1, key)
	case atom.ENTER:
		g := game
		if g.Field.Type(g.Cursor.X, g.Cursor.Y) == atom.ATOM {
			if g.Cursor.State == 0 {
				g.Cursor.State = g.Cursor.Type
				sfx.PlaySound(sfx.Selected, 0)
			} else {
				g.Cursor.State = 0
			}
		}
	}
}

func move(dx, dy, dir int) {
	if game.Cursor.Moving || game.Loosing {
		return
	}

	if game.Cursor.State == 0 {
		moveCursor(dx, dy)
	} else {
		moveAtom(dir)
	}
}

func moveAtom(dir int) {
	g := game
	d := g.MovedDistance(dir)
	if d == 0 {
		return
	}

	if !conf.NoLose {
		g.Score -= 5
		if g.Score < 0 {
			g.Score = 0
		}
	}

	c := &g.Cursor
	l := &g.Loose
	l.Atom = g.Field.Index(c.X, c.Y)
	l.X = g.Offset.X + c.X*atom.TILESIZE
	l.Y = g.Offset.Y + c.Y*atom.TILESIZE
	l.Ex, l.Ey = l.X, l.Y
	l.Mx, l.My = 0, 0
	l.Dx, l.Dy = c.X, c.Y
	g.Field.Set(c.X, c.Y, atom.FREE)
	g.Cursor.Sx = 0
	g.Cursor.Sy = 0

	dd := d
	d *= atom.TILESIZE
	switch dir {
	case atom.LEFT:
		l.Mx = -1
		l.Ex -= d
		l.Dx = c.X - dd
	case atom.RIGHT:
		l.Mx = 1
		l.Ex += d
		l.Dx = c.X + dd
	case atom.DOWN:
		l.My = 1
		l.Ey += d
		l.Dy = c.Y + dd
	case atom.UP:
		l.My = -1
		l.Ey -= d
		l.Dy = c.Y - dd
	}

	g.Loosing = true
	sfx.PlaySound(sfx.Bzzz, -1)
}

func moveCursor(mx, my int) {
	g := game
	f := &g.Field
	c := &g.Cursor

	x := c.X + mx
	y := c.Y + my
	if x < 0 || y < 0 || f.At(x, y) == 0 {
		return
	}

	c.Mx = mx
	c.My = my
	c.Sx = 0
	c.Sy = 0
	c.Ex = mx * atom.TILESIZE
	c.Ey = my * atom.TILESIZE

	c.Moving = true
}

func update() {
	switch state {
	case INTRO:
		if slider.Update() {
			if slider.Quit {
				newstate = EXIT
			} else {
				newstate = SELECT
			}
		}

	case PLAY:
		playUpdate()

	case WON:
		wonUpdate()

	case CREDITS:
		r := gfx.Credit.Bounds()
		if credits.y+181 < r.Dy() {
			credits.y++
		}

	case EXIT:
		if slider.Update() {
			os.Exit(0)
		}
	}
}

func playUpdate() {
	g := game
	c := &g.Cursor
	l := &g.Loose

	if g.Paused {
		return
	}

	if now := time.Now(); now.After(g.TimeEnd) && !conf.NoLose {
		newstate = TIMEOUT
	}

	switch {
	case c.Moving:
		c.Sx += c.Mx * 8
		c.Sy += c.My * 8
		if c.Sx == c.Ex && c.Sy == c.Ey {
			c.Moving = false
			c.X += c.Mx
			c.Y += c.My
			c.Sx = 0
			c.Sy = 0
		}

	case g.Loosing:
		mx := l.Mx * 8
		my := l.My * 8
		l.X += mx
		l.Y += my
		c.Sx += mx
		c.Sy += my
		if l.X == l.Ex && l.Y == l.Ey {
			x := l.Dx
			y := l.Dy
			c.Sx = 0
			c.Sy = 0
			g.Field.Set(x, y, l.Atom|atom.ATOM)
			l.Atom = 0
			c.X = x
			c.Y = y
			g.Loosing = false
			sdlmixer.HaltChannel(0)
		}

	case g.Won():
		newstate = WON
	}
}

func wonUpdate() {
	if len(won.atoms) == 0 {
		for i := 0; i < 3; i++ {
			game.Score += 10
			won.tick = won.tick.Add(1 * time.Second)
			if won.tick.After(game.TimeEnd) {
				if conf.MaxAuthLevel < atom.LEVELS {
					if game.Score >= game.Hiscore {
						game.Hiscore = game.Score
						if game.Hiscore >= conf.Hiscores[level-1] {
							conf.Hiscores[level-1] = game.Hiscore
						}
					}
					level++
					newstate = SELECT
				} else {
					newstate = CREDITS
				}
				err := conf.Save()
				if err == nil {
					sdl.Log("Saved config")
				} else {
					sdl.Log("%v", err)
				}
			}
		}
	} else {
		if now := time.Now(); now.After(won.timer) {
			won.timer = now.Add(40 * time.Millisecond)

			a := &won.atoms[0]
			if a.Atom++; a.Atom >= len(gfx.Explosion) {
				game.Field.Set(a.X, a.Y, atom.FREE)
				won.atoms = won.atoms[1:]
				if len(won.atoms) == 0 {
					won.timer = time.Now()
				}
			}
		}
	}
}

func blit() {
	screen.Clear()
	switch state {
	case INTRO, EXIT:
		slider.Draw()
	case SELECT:
		preview.DrawPreview()
	case PLAY:
		blitPlay(time.Now())
	case WON:
		blitPlay(won.tick)
		if len(won.atoms) > 0 {
			a := &won.atoms[0]
			game.DrawTile(a.X, a.Y, gfx.Explosion[a.Atom])
		}
	case TIMEOUT:
		atom.DrawGFX(screen, gfx.Timeout, 0, 0)
	case CREDITS:
		blitCredits()
	}
	screen.Flush()
}

func blitCredits() {
	atom.DrawGFX(screen, gfx.Info, 0, 0)
	r := gfx.Credit.Bounds()
	w := r.Dx()
	h := 181
	x := 160 - w/2
	y := 36
	atom.DrawGFXPartial(screen, gfx.Credit, x, credits.y, w, h, 0, y)
}

func blitPlay(now time.Time) {
	g := game

	if g.Paused {
		atom.DrawGFX(screen, gfx.Paused, 0, 0)
		return
	}

	var timeLeft time.Duration
	var previewTick int64

	if conf.NoLose {
		g.TimeEnd = time.Now().Add((g.Duration + 1) * time.Second)
	}

	if justStarted {
		timeLeft = g.Duration * time.Second
	} else {
		if now.Before(g.TimeEnd) {
			timeLeft = g.TimeEnd.Sub(now)
		}
		previewTick = ((now.Sub(g.PreviewTick).Nanoseconds() / 1e6) % 1600) / 800
	}

	g.DrawField()
	if g.Loose.Atom != 0 {
		atom.DrawGFX(screen, gfx.Atom[g.Loose.Atom], g.Loose.X, g.Loose.Y)
	}

	if showCursor {
		r := gfx.Cursor[0].Bounds()
		x := g.Offset.X + g.Cursor.X*r.Dx() + g.Cursor.Sx
		y := g.Offset.Y + g.Cursor.Y*r.Dy() + g.Cursor.Sy
		atom.DrawGFX(screen, gfx.Cursor[g.Cursor.State], x, y)
	}

	r3 := gfx.Font3[0].Bounds()
	x := atom.TILESIZE / 2
	y := atom.TILESIZE / 2
	blitString("HISCORE", x, y)

	y += int(float64(r3.Dy()) * 1.4)
	blitNumber(g.Hiscore, x, y)

	y += atom.TILESIZE * 2
	blitString("SCORE", x, y)

	y += int(float64(r3.Dy()) * 1.4)
	blitNumber(g.Score, x, y)

	y += atom.TILESIZE * 2
	blitString("LEVEL", x, y)

	y += int(float64(r3.Dy()) * 1.4)
	atom.DrawGFX(screen, gfx.Font2[g.Level/10], x, y)
	r2 := gfx.Font2[0].Bounds()
	x += r2.Dx()
	atom.DrawGFX(screen, gfx.Font2[g.Level%10], x, y)

	x = atom.TILESIZE / 2
	y += atom.TILESIZE * 2
	blitString("TIME", x, y)

	min := int(timeLeft.Minutes())
	sec := int(timeLeft.Seconds())
	y += int(float64(r3.Dy()) * 1.4)
	atom.DrawGFX(screen, gfx.Font2[min], x, y)
	x += r2.Dx()
	atom.DrawGFX(screen, gfx.Font2[10], x, y)
	x += r2.Dx()
	atom.DrawGFX(screen, gfx.Font2[(sec%60)/10], x, y)
	x += r2.Dx()
	atom.DrawGFX(screen, gfx.Font2[sec%10], x, y)

	x = 0
	y = 240 - 71
	atom.DrawGFX(screen, gfx.Preview[previewTick], x, y)

	r1 := gfx.Font1[0].Bounds()
	y = 181
	blitDesc(g.Desc[0][:], y)
	y += r1.Dy() + 2
	blitDesc(g.Desc[1][:], y)

	g.DrawSmallPreview()

	if justStarted && g.Level == 1 && conf.MaxAuthLevel == 1 {
		atom.DrawGFX(screen, gfx.Instructions, 0, 0)
	}
}

func blitString(text string, x, y int) {
	for _, ch := range text {
		ch -= 'A'
		if !(0 <= ch && ch < rune(len(gfx.Font3))) {
			ch = 27
		}
		atom.DrawGFX(screen, gfx.Font3[ch], x, y)

		r := gfx.Font3[0].Bounds()
		x += r.Dx()
	}
}

func blitNumber(n, x, y int) {
	str := fmt.Sprint(n)
	for _, ch := range str {
		atom.DrawGFX(screen, gfx.Font2[ch-'0'], x, y)
		r := gfx.Font2[0].Bounds()
		x += r.Dx()
	}
}

func blitDesc(desc []byte, y int) {
	x := 36 - font1Size(desc)/2
	for _, ch := range desc {
		if ch == 0 {
			break
		}

		n := ascii2font1(ch)
		atom.DrawGFX(screen, gfx.Font1[n], x, y)
		x += font1Width[n]
	}
}

var font1Width = []int{
	5, 5, 4, 5, 4, 4, 5, 5, 2, 4, 4, 4, 6, 5, 5, 5, 5, 5,
	5, 4, 5, 4, 6, 4, 5, 4, 5, 4, 4, 4, 4, 4, 5, 4, 5, 5,
}

func font1Size(buf []byte) int {
	w := 0
	for _, ch := range buf {
		if ch == 0 {
			break
		}
		w += font1Width[ascii2font1(ch)]
	}
	return w
}

func ascii2font1(ch byte) int {
	switch c := int(ch); {
	case 'A' <= c && c <= 'Z':
		return c - 'A'
	case 'a' <= c && c <= 'z':
		return c - 'a'
	case '0' <= c && c <= '9':
		return 26 + (c - '0')
	}
	return 35
}

func shuffle(l []atom.Loosetile) {
	for i := len(l) - 1; i >= 1; i-- {
		j := rand.Intn(i + 1)
		l[i], l[j] = l[j], l[i]
	}
}
