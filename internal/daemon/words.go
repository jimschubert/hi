package daemon

import (
	"math/rand/v2"
	"strings"
)

// words is a small corpus used to generate random-ish responses.
// these words are pulled from a paragraph of lorem ipsum, lowercased, deduped, and sorted (with some random commas remaining)
var words = []string{
	"ad",
	"adipiscing",
	"aliqua",
	"aliquip",
	"amet,",
	"anim",
	"aute",
	"cillum",
	"commodo",
	"consectetur",
	"consequat",
	"culpa",
	"cupidatat",
	"deserunt",
	"do",
	"dolor",
	"dolore",
	"duis",
	"ea",
	"eiusmod",
	"elit,",
	"enim",
	"esse",
	"est",
	"et",
	"eu",
	"ex",
	"excepteur",
	"exercitation",
	"fugiat",
	"id",
	"in",
	"incididunt",
	"ipsum",
	"irure",
	"knostrud",
	"labore",
	"laboris",
	"laborum",
	"lorem",
	"magna",
	"minim",
	"mollit",
	"nisi",
	"non",
	"nulla",
	"occaecat",
	"officia",
	"pariatur",
	"proident",
	"qui",
	"quis",
	"reprehenderit",
	"sed",
	"sint",
	"sit",
	"sunt",
	"tempor",
	"ullamco",
	"ut",
	"velit",
	"veniam",
	"voluptate",
}

func randomWord() string {
	return words[rand.IntN(len(words))]
}

func randomSentence(wordCount int) string {
	ws := make([]string, wordCount)
	for i := range wordCount {
		word := randomWord()
		if i == 0 {
			word = strings.ToTitle(word[:1]) + word[1:]
		}
		ws[i] = word
	}
	return strings.Join(ws, " ") + "."
}
