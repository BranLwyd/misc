package main

import "os"
import "io"

type Rot13Reader struct {
	source io.Reader
}

func (r Rot13Reader) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		switch {
		case 'a' <= p[i] && p[i] <= 'z':
			p[i] = (p[i]-'a'+13)%26 + 'a'
		case 'A' <= p[i] && p[i] <= 'Z':
			p[i] = (p[i]-'A'+13)%26 + 'A'
		}
	}

	return n, err
}

func main() {
	reader := Rot13Reader{os.Stdin}

	io.Copy(os.Stdout, reader)
}
