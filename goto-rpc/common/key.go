package common

var keyChar = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func genKey(n int) string {
	if n == 0 {
		return string(keyChar[0])
	}

	b := make([]byte, 20)
	i := len(b)
	l := len(keyChar)
	for i >= 0 && n > 0 {
		i--
		j := n % l
		n = (n - j) / l
		b[i] = keyChar[j]
	}
	return string(b[i:])
}
