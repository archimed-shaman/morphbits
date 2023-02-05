package keyboard

type Layout []string

func QWERTY() Layout {
	return []string{
		"1234567890-=",
		"qwertyuiop",
		"asdfghjkl",
		"zxcvbnm",
	}
}
