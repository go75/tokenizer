package tokenizer

import (
	"fmt"
	"strings"
)

type BytePairEncoder struct {
	wsToken  string
	unkToken string
	// k: word, v: tokens
	wordToken map[string]*[]string
	// k: word, v: count
	wordCount map[string]int
	// k: token, v: count
	tokenCount map[string]int
	// k: id, v: token
	idToken map[int]string
	// k: token, v: id
	tokenId map[string]int
}

func DefaultBytePairEncoder() *BytePairEncoder {
	return NewBytePairEncoder("_", " ")
}

func NewBytePairEncoder(wsToken, unkToken string) *BytePairEncoder {
	return &BytePairEncoder{
		wsToken:    wsToken,
		unkToken:   unkToken,
		wordToken:  make(map[string]*[]string),
		wordCount:  make(map[string]int),
		tokenCount: make(map[string]int),
		idToken:   make(map[int]string),
		tokenId:   make(map[string]int),
	}
}

func (e *BytePairEncoder) wordToTokens(word string) *[]string {
	parts := []rune(word)
	n := len(parts)
	res := make([]string, n)
	for i := 0; i < n; i++ {
		token := string(parts[i])
		e.tokenCount[token]++
		res[i] = token
	}
	return &res
}

func (e *BytePairEncoder) processWord(word string) {
	e.wordToken[word] = e.wordToTokens(word)
	e.wordCount[word]++
}

func (e *BytePairEncoder) preprocess(text string) []string {
	text = strings.TrimSpace(text)
	return strings.Fields(text)
}

func (e *BytePairEncoder) initState(text string) {
	words := e.preprocess(text)
	for _, word := range words {
		e.processWord(e.wsToken + word)
	}
}

func (e *BytePairEncoder) merge_pair() {
	// k: token, v: count
	m := make(map[string]int)
	for word, tokens := range e.wordToken {
		n := len(*tokens) - 1
		for i := 0; i < n; i++ {
			m[(*tokens)[i]+(*tokens)[i+1]] += e.wordCount[word]
		}
	}

	maxToken := ""
	maxCount := 0
	for k, v := range m {
		if v > maxCount {
			maxToken = k
			maxCount = v
		}
	}

	if maxCount < 2 {
		return
	}

	e.tokenCount[maxToken] = maxCount

	for _, tokens := range e.wordToken {
		n := len(*tokens) - 1
		for i := 0; i < n; i++ {
			if (*tokens)[i]+(*tokens)[i+1] == maxToken {
				e.tokenCount[(*tokens)[i]]--
				e.tokenCount[(*tokens)[i+1]]--
				post := (*tokens)[i+1:]
				post[0] = maxToken
				*tokens = (*tokens)[:i]
				*tokens = append((*tokens), post...)
				*tokens = (*tokens)[:len(*tokens)]

				i--
				n -= 2
			}
		}
	}
}

func (e *BytePairEncoder) merge(steps int) {
	for i := 0; i < steps; i++ {
		e.merge_pair()
	}
}

func (e *BytePairEncoder) buildIndex() {
	e.tokenId[e.unkToken] = 0
	e.idToken[0] = e.unkToken
	i := 1
	for token := range e.tokenCount {
		e.tokenId[token] = i
		e.idToken[i] = token
		i++
	}
}

func (e *BytePairEncoder) Train(text string, steps int) {
	e.initState(text)
	e.merge(steps)
	e.buildIndex()
}

func (e *BytePairEncoder) segment(words []string) []int {
	res := make([]int, 0)
	for _, word := range words {
		parts := []rune(word)
	NEXT:
		for i := len(parts); i >= 1; i-- {
			if code, ok := e.tokenId[string(parts[:i])]; ok {
				parts = parts[i:]
				res = append(res, code)
				goto NEXT
			}
		}
		if len(parts) == 0 {
			continue
		}
		code := e.tokenId[string(parts[0])]
		res = append(res, code)
		parts = parts[1:]
		if len(parts) != 0 {
			goto NEXT
		}
	}

	return res
}

func (e *BytePairEncoder) Encode(text string) []int {
	words := e.preprocess(text)
	return e.segment(words)
}

func (e *BytePairEncoder) Decode(codes []int) []string {
	res := make([]string, 0)
	for _, code := range codes {
		res = append(res, e.idToken[code])
	}

	return res
}

func (e *BytePairEncoder) Dump() {
	fmt.Println("===== dump state ======")
	fmt.Println("===> dump wordToken <===")
	for word, tokens := range e.wordToken {
		fmt.Println(word, "=>", *tokens)
	}
	fmt.Println()
	fmt.Println("===> dump wordcnt <===")
	for word, count := range e.wordCount {
		fmt.Println(word, "=>", count)
	}
	fmt.Println()
	fmt.Println("===> dump tokenCount <===")
	for token, count := range e.tokenCount {
		fmt.Println(token, "=>", count)
	}
	fmt.Println()
	fmt.Println("===> dump idTokens <===")
	for code, token := range e.idToken {
		fmt.Println(code, "=>", token)
	}
	fmt.Println()
	fmt.Println("===> dump tokenIds <===")
	for token, code := range e.tokenId {
		fmt.Println(token, "=>", code)
	}
	fmt.Println()
}
