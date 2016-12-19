package atom

import (
	"path/filepath"

	"github.com/qeedquan/go-media/sdl/sdlmixer"
)

type SFX struct {
	conf     *Config
	Title    *sdlmixer.Music
	End      *sdlmixer.Music
	Bzzz     *sdlmixer.Chunk
	Explode  *sdlmixer.Chunk
	Selected *sdlmixer.Chunk
}

func LoadSFX(conf *Config) *SFX {
	return &SFX{
		conf:     conf,
		Title:    loadMusic(conf, "title.ogg"),
		End:      loadMusic(conf, "end.ogg"),
		Bzzz:     loadSound(conf, "bzzz.wav"),
		Explode:  loadSound(conf, "explode.wav"),
		Selected: loadSound(conf, "selected.wav"),
	}
}

func loadMusic(conf *Config, name string) *sdlmixer.Music {
	name = filepath.Join(conf.Assets, "snd", name)
	mus, err := sdlmixer.LoadMUS(name)
	ek(err)
	return mus
}

func loadSound(conf *Config, name string) *sdlmixer.Chunk {
	name = filepath.Join(conf.Assets, "snd", name)
	snd, err := sdlmixer.LoadWAV(name)
	ek(err)
	return snd
}

func (sfx *SFX) PlayMusic(mus *sdlmixer.Music, fade int) {
	if sfx.conf.Sound && mus != nil {
		sdlmixer.FadeInMusic(mus, -1, fade)
	}
}

func (sfx *SFX) PlaySound(snd *sdlmixer.Chunk, loops int) {
	if sfx.conf.Sound && snd != nil {
		snd.PlayChannel(0, loops)
	}
}
