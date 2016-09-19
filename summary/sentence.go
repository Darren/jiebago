package summary

import (
	"math"
)

//Sentence represents a sentence.
type Sentence struct {
	pos    int
	text   string
	words  set
	weight float64
}

// Text returns the setence's text.
func (s Sentence) Text() string {
	return s.text
}

// Weight returns the setence's weight.
func (s Sentence) Weight() float64 {
	return s.weight
}

func (s Sentence) Similarity(t Sentence) float64 {
	i := s.words.intersect(t.words)
	ww := float64(i.len())
	ss1 := math.Log(float64(s.words.len()))
	ss2 := math.Log(float64(t.words.len()))
	ss := ss1 + ss2

	if ss == 0 {
		return -1
	}

	return ww / (ss1 + ss2)
}

// Sentences represents a slice of Sentence.
type Sentences []Sentence

func (ss Sentences) Len() int {
	return len(ss)
}

func (ss Sentences) Less(i, j int) bool {
	if ss[i].weight == ss[j].weight {
		return ss[i].text < ss[j].text
	}

	return ss[i].weight < ss[j].weight
}

func (ss Sentences) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}
