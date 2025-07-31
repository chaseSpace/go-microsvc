package uhttp

import (
	"os"
	"testing"
)

func TestDownFile(t *testing.T) {
	buf, err := DownFile("https://scontent-sin6-4.cdninstagram.com/v/t51.29350-15/458388290_1213479493018392_2599660234419147879_n.jpg?stp=dst-jpg_e35&efg=eyJ2ZW5jb2RlX3RhZyI6ImltYWdlX3VybGdlbi45NTV4OTU1LnNkci5mMjkzNTAuZGVmYXVsdF9pbWFnZSJ9&_nc_ht=scontent-sin6-4.cdninstagram.com&_nc_cat=100&_nc_ohc=zH4hvk2TTpUQ7kNvgEPoodB&_nc_gid=d7fc4ce7aba04d9dbb86591905655f9a&edm=AA24wl0BAAAA&ccb=7-5&ig_cache_key=MzQ0OTE2NzIwNzQ1NTEwNDIwNA%3D%3D.3-ccb7-5&oh=00_AYDJbPEjmDt4exSqNsCqbzlODljbicuSeiLM99BYs0rWHw&oe=66E25933&_nc_sid=c3221d")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile("img.png", buf, 0666)
}
