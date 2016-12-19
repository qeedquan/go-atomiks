package atom

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"

	"github.com/qeedquan/go-media/sdl"
)

type Config struct {
	Assets       string
	Pref         string
	Fullscreen   bool
	Sound        bool
	NoLose       bool
	Unlocked     bool
	MaxAuthLevel int
	Hiscores     [LEVELS]int
}

func NewConfig(editor bool) *Config {
	c := &Config{
		Assets: filepath.Join(sdl.GetBasePath(), "assets"),
		Pref:   sdl.GetPrefPath("", "atomiks"),
	}
	flag.StringVar(&c.Assets, "assets", c.Assets, "assets directory")
	flag.StringVar(&c.Pref, "pref", c.Pref, "preference directory")
	flag.BoolVar(&c.Fullscreen, "fullscreen", false, "fullscreen mode")
	if !editor {
		flag.BoolVar(&c.Sound, "sound", true, "enable sound")
		flag.BoolVar(&c.NoLose, "no-lose", false, "can't lose")
		flag.BoolVar(&c.Unlocked, "unlocked", false, "unlock all levels")
	}
	flag.Parse()
	c.Load()
	return c
}

func (c *Config) Save() error {
	name := filepath.Join(c.Pref, "Atomiks")
	fd, err := os.Create(name)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(fd)
	w.WriteByte(byte(c.MaxAuthLevel))
	for _, s := range c.Hiscores {
		w.WriteByte(byte(s >> 8))
		w.WriteByte(byte(s))
	}

	err = w.Flush()
	xerr := fd.Close()
	if err == nil {
		err = xerr
	}

	return err
}

func (c *Config) Load() {
	c.MaxAuthLevel = 1
	for i := range c.Hiscores {
		c.Hiscores[i] = 0
	}

	name := filepath.Join(c.Pref, "Atomiks")
	fd, err := os.Open(name)
	if err != nil {
		return
	}

	r := bufio.NewReader(fd)
	c.MaxAuthLevel = readByte(r)
	for i := range c.Hiscores {
		c.Hiscores[i] = readShort(r)
	}
	if c.MaxAuthLevel < 1 {
		c.MaxAuthLevel = 1
	}
}
