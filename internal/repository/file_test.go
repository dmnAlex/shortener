package repository

import (
	"fmt"
	"testing"
)

func BenchmarkFileRepo_Save(b *testing.B) {
	r, err := NewFileRepo("")
	if err != nil {
		b.Fatal(err)
	}
	defer r.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.Save("user1", "https://example.com/"+fmt.Sprint(i))
		if err != nil {
			b.Fatal(err)
		}
	}
}
