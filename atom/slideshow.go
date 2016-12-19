package atom

import (
	"image"
	"image/color"
	"time"
)

type Slide struct {
	Images   []image.Image
	Duration time.Duration
}

type Slideshow struct {
	screen  *Display
	advance bool
	events  bool
	Quit    bool
	Slides  []Slide
	Index   int
	Start   time.Time
	Frame   *image.RGBA
}

func (s *Slideshow) Init(screen *Display, slides []Slide, events bool) {
	*s = Slideshow{
		screen: screen,
		events: events,
		Slides: slides,
		Start:  time.Now(),
		Frame:  image.NewRGBA(image.Rect(0, 0, WIDTH, HEIGHT)),
	}
}

func (s *Slideshow) Event(key int) {
	if !s.events {
		return
	}

	if key == ESC {
		s.Index = len(s.Slides)
		s.Quit = true
	} else {
		s.advance = true
	}
}

func (s *Slideshow) Update() bool {
	if s.Index >= len(s.Slides) {
		return true
	}

	d := s.Slides[s.Index].Duration
	if d > 0 {
		t := s.Start.Add(d)
		n := time.Now()
		if n.After(t) {
			s.advance = true
		}
	}

	if s.advance {
		s.advance = false
		if s.Index < len(s.Slides) {
			s.Index++
		}
		s.Start = time.Now()
	}

	return false
}

func (s *Slideshow) Draw() {
	if s.Index >= len(s.Slides) {
		DrawGFX(s.screen, s.Frame, 0, 0)
		return
	}

	DrawGFX(s.Frame, image.NewUniform(color.Black), 0, 0)
	for _, m := range s.Slides[s.Index].Images {
		DrawGFX(s.screen, m, 0, 0)
		DrawGFX(s.Frame, m, 0, 0)
	}
}
