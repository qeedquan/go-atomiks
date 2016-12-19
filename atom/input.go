package atom

import "github.com/qeedquan/go-media/sdl"

const (
	UP = iota + 1
	RIGHT
	DOWN
	LEFT
	FULLSCREEN
	HOME
	END
	ESC
	SPACE
	ENTER
	NONE
	UNKNOWN
)

func Key(key sdl.Keycode) int {
	var mod int
	switch key {
	case sdl.K_LEFT, sdl.K_KP_4:
		mod = LEFT
	case sdl.K_RIGHT, sdl.K_KP_6:
		mod = RIGHT
	case sdl.K_UP, sdl.K_KP_8:
		mod = UP
	case sdl.K_DOWN, sdl.K_KP_2:
		mod = DOWN
	case sdl.K_RETURN, sdl.K_KP_5, sdl.K_KP_ENTER:
		if sdl.GetModState()&sdl.KMOD_ALT != 0 {
			mod = FULLSCREEN
		}
		mod = ENTER
	case sdl.K_HOME, sdl.K_KP_7:
		mod = HOME
	case sdl.K_END, sdl.K_KP_1:
		mod = END
	case sdl.K_ESCAPE:
		mod = ESC
	case sdl.K_SPACE:
		mod = SPACE
	case sdl.K_LALT, sdl.K_RALT:
		mod = NONE
	default:
		mod = UNKNOWN
	}

	if mod != FULLSCREEN && sdl.GetModState()&sdl.KMOD_ALT != 0 {
		mod = NONE
	}

	return mod
}
