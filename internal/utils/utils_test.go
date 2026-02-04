package utils

import (
	"testing"
)

func BenchmarkGenerateShortID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateShortID()
		if err != nil {
			b.Fatal(err)
		}
	}
}
