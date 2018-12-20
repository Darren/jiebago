package summary

import (
	"math"
	"sort"
	"strings"

	"unicode/utf8"

	"github.com/darren/jiebago/posseg"
)

const dampingFactor = 0.85

var (
	defaultAllowPOS = []string{"ns", "n", "vn", "v"}
)

var posFilter set

func isFiltered(text string) bool {
	return posFilter.has(text)
}

type edge struct {
	start, end string
	weight     float64
}

type edges []edge

func (es edges) Len() int           { return len(es) }
func (es edges) Less(i, j int) bool { return es[i].weight < es[j].weight }
func (es edges) Swap(i, j int)      { es[i], es[j] = es[j], es[i] }

type undirectWeightedGraph struct {
	graph map[string]edges
	keys  sort.StringSlice
}

func newUndirectWeightedGraph() *undirectWeightedGraph {
	u := new(undirectWeightedGraph)
	u.graph = make(map[string]edges)
	u.keys = make(sort.StringSlice, 0)
	return u
}

func (u *undirectWeightedGraph) addEdge(start, end string, weight float64) {
	e := edge{start, end, weight}
	r := edge{end, start, weight}

	if _, ok := u.graph[start]; !ok {
		u.keys = append(u.keys, start)
		u.graph[start] = edges{e}
	} else {
		u.graph[start] = append(u.graph[start], e)
	}

	if _, ok := u.graph[end]; !ok {
		u.keys = append(u.keys, end)
		u.graph[end] = edges{r}
	} else {
		u.graph[end] = append(u.graph[end], r)
	}
}

func (u *undirectWeightedGraph) rank() Sentences {
	if !sort.IsSorted(u.keys) {
		sort.Sort(u.keys)
	}

	ws := make(map[string]float64)
	outSum := make(map[string]float64)

	wsdef := 1.0
	if len(u.graph) > 0 {
		wsdef /= float64(len(u.graph))
	}
	for n, out := range u.graph {
		ws[n] = wsdef
		sum := 0.0
		for _, e := range out {
			sum += e.weight
		}
		outSum[n] = sum
	}

	for x := 0; x < 10; x++ {
		for _, n := range u.keys {
			s := 0.0
			inedges := u.graph[n]
			for _, e := range inedges {
				s += e.weight / outSum[e.end] * ws[e.end]
			}
			ws[n] = (1 - dampingFactor) + dampingFactor*s
		}
	}
	minRank := math.MaxFloat64
	maxRank := math.SmallestNonzeroFloat64
	for _, w := range ws {
		if w < minRank {
			minRank = w
		} else if w > maxRank {
			maxRank = w
		}
	}
	result := make(Sentences, 0)
	for n, w := range ws {
		result = append(result, Sentence{text: n, weight: (w - minRank/10.0) / (maxRank - minRank/10.0)})
	}
	sort.Sort(sort.Reverse(result))
	return result
}

func isSetenctStop(r rune) bool {
	return r == '。' || r == '！' || r == '？' || r == '\n' || r == '\r'
}

func (t *TextRanker) Summary(text string, limit int) string {
	ss := t.TextRank(text, limit)
	ts := make([]string, len(ss))
	for i, s := range ss {
		ts[i] = s.text + "。"
	}
	return strings.Join(ts, "")
}

func (t *TextRanker) TextRank(text string, limit int) Sentences {
	var sentences Sentences

	var sm = make(map[string]int, 0)

	i := 0
	for _, s := range strings.FieldsFunc(text, isSetenctStop) {
		if _, ok := sm[s]; !ok {
			var sentence Sentence
			var words = newSet()
			for seg := range t.seg.Cut(s, true) {
				word := seg.Text()
				if !isFiltered(seg.Pos()) {
					words.add(word)
				}
			}
			sentence = Sentence{
				pos:   i,
				text:  s,
				words: words,
			}

			sentences = append(sentences, sentence)
			sm[s] = i
			i++
		}
	}

	g := newUndirectWeightedGraph()
	cm := make(map[[2]string]float64)

	for i, s1 := range sentences {
		if i < len(sentences)+1 {
			for _, s2 := range sentences {
				var pair = [2]string{s1.Text(), s2.Text()}

				if sim := s1.Similarity(s2); sim > 0 {
					cm[pair] = sim
				}
			}
		}
	}

	for startEnd, weight := range cm {
		g.addEdge(startEnd[0], startEnd[1], weight)
	}

	tags := g.rank()

	var poses = make([]int, 0, len(tags))

	totalLength := 0
	for _, t := range tags {
		c := utf8.RuneCountInString(t.Text())
		if c > limit {
			continue
		}
		pos := sm[t.Text()]
		totalLength += c + 1
		if totalLength > limit {
			break
		}

		poses = append(poses, pos)
	}

	sort.Sort(sort.IntSlice(poses))

	var res Sentences

	for _, v := range poses {
		s := sentences[v]
		res = append(res, s)
	}

	return res
}

type TextRanker struct {
	seg *posseg.Segmenter
}

func (t *TextRanker) SetSegmenter(seg *posseg.Segmenter) {
	t.seg = seg
}

func (t *TextRanker) LoadDictionary(fileName string) error {
	t.seg = new(posseg.Segmenter)
	return t.seg.LoadDictionary(fileName)
}

func init() {
	posFilter = newSet()
	for _, p := range defaultAllowPOS {
		posFilter.add(p)
	}
}
