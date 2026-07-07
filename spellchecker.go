package spellcheck

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/sajari/fuzzy"
)

// Spellchecker provides spell checking functionality using fuzzy matching
type Spellchecker struct {
	model          *fuzzy.Model
	config         *Config
	ignoredWords   map[string]bool
	caseSensitive  bool
	minWordLength  int
	maxSuggestions int
}

// modelCache memoizes trained fuzzy models keyed by custom-dictionary path.
// Training the dictionary is by far the most expensive part of constructing a
// Spellchecker, and a trained model is read-only afterwards (fuzzy guards
// lookups with an RWMutex), so a single model can be shared safely across
// every checker — including the per-request checkers derived in the handler.
var (
	modelCacheMu sync.Mutex
	modelCache   = map[string]*fuzzy.Model{}
)

func loadModel(customDictionary string) (*fuzzy.Model, error) {
	modelCacheMu.Lock()
	defer modelCacheMu.Unlock()

	if model, ok := modelCache[customDictionary]; ok {
		return model, nil
	}

	model := fuzzy.NewModel()
	model.SetDepth(4)
	model.SetThreshold(1)

	if err := loadEnglishDictionary(model); err != nil {
		return nil, err
	}

	if customDictionary != "" {
		if err := loadDictionaryFromFile(model, customDictionary); err != nil {
			return nil, err
		}
	}

	modelCache[customDictionary] = model
	return model, nil
}

// NewSpellchecker creates a new spellchecker with the given configuration
func NewSpellchecker(config *Config) (*Spellchecker, error) {
	model, err := loadModel(config.CustomDictionary)
	if err != nil {
		return nil, err
	}

	return &Spellchecker{
		model:          model,
		config:         config,
		ignoredWords:   buildIgnoredWords(config.IgnoredWords, config.CaseSensitive),
		caseSensitive:  config.CaseSensitive,
		minWordLength:  config.MinWordLength,
		maxSuggestions: config.MaxSuggestions,
	}, nil
}

func buildIgnoredWords(words []string, caseSensitive bool) map[string]bool {
	ignored := make(map[string]bool, len(words))
	for _, word := range words {
		if caseSensitive {
			ignored[word] = true
		} else {
			ignored[strings.ToLower(word)] = true
		}
	}
	return ignored
}

// derive returns a lightweight checker sharing this checker's trained model but
// with per-request overrides applied. It exists so on-demand requests carrying
// context words or options don't pay to retrain the dictionary; the model is
// identical, only the ignore set and scalar options differ.
func (s *Spellchecker) derive(contextWords []string, caseSensitive *bool, maxSuggestions *int) *Spellchecker {
	cs := s.caseSensitive
	if caseSensitive != nil {
		cs = *caseSensitive
	}

	ms := s.maxSuggestions
	if maxSuggestions != nil {
		ms = *maxSuggestions
	}

	// Rekey the ignore set to the effective case sensitivity so results match
	// what NewSpellchecker would have produced for the merged config.
	merged := make([]string, 0, len(s.config.IgnoredWords)+len(contextWords))
	merged = append(merged, s.config.IgnoredWords...)
	merged = append(merged, contextWords...)

	return &Spellchecker{
		model:          s.model,
		config:         s.config,
		ignoredWords:   buildIgnoredWords(merged, cs),
		caseSensitive:  cs,
		minWordLength:  s.minWordLength,
		maxSuggestions: ms,
	}
}

// Check checks the spelling of the given text and returns any errors found
func (s *Spellchecker) Check(text string) (*SpellingErrors, error) {
	if len(text) > s.config.MaxTextLength {
		return nil, &TextTooLongError{
			Length:    len(text),
			MaxLength: s.config.MaxTextLength,
		}
	}

	errors := &SpellingErrors{}

	st := checkStatePool.Get().(*checkState)
	defer func() {
		st.words = st.words[:0]
		checkStatePool.Put(st)
	}()

	st.words = s.extractWordsInto(text, st.words)

	for _, wordInfo := range st.words {
		word := wordInfo.word
		position := wordInfo.position

		if len(word) < s.minWordLength {
			continue
		}

		checkWord := word
		if !s.caseSensitive {
			checkWord = strings.ToLower(word)
		}
		if s.ignoredWords[checkWord] {
			continue
		}

		if containsDigit(word) {
			continue
		}

		// One SpellCheck round-trip decides correctness and seeds the
		// suggestion list, halving model lookups for misspelled words.
		correction := string(s.model.SpellCheck(checkWord))
		if correction == "" || correction == checkWord {
			continue
		}

		errors.Add(&SpellingError{
			Word:        word,
			Position:    position,
			Suggestions: s.buildSuggestions(word, checkWord, correction),
		})
	}

	return errors, nil
}

// CheckField checks the spelling of text in a specific field (for middleware)
func (s *Spellchecker) CheckField(fieldName, text string) (*SpellingErrors, error) {
	errors, err := s.Check(text)
	if err != nil {
		return nil, err
	}

	for _, e := range errors.Errors {
		e.Field = fieldName
	}

	return errors, nil
}

// getSuggestions returns spelling suggestions for a misspelled word
func (s *Spellchecker) getSuggestions(word string) []string {
	checkWord := word
	if !s.caseSensitive {
		checkWord = strings.ToLower(word)
	}

	return s.buildSuggestions(word, checkWord, string(s.model.SpellCheck(checkWord)))
}

// buildSuggestions assembles the ranked suggestion list from an already-computed
// correction, so callers that have run SpellCheck don't repeat it.
func (s *Spellchecker) buildSuggestions(word, checkWord, correctionStr string) []string {
	var filtered []string
	if correctionStr != "" && correctionStr != checkWord {
		filtered = append(filtered, correctionStr)
	}

	suggestions := s.model.Suggestions(checkWord, false)
	for _, suggestion := range suggestions {
		if suggestion != checkWord && suggestion != correctionStr {
			filtered = append(filtered, suggestion)
			if len(filtered) >= s.maxSuggestions {
				break
			}
		}
	}

	// Preserve capitalization if original word was capitalized
	shouldCapitalize := !s.caseSensitive && len(word) > 0 && unicode.IsUpper(rune(word[0]))

	if shouldCapitalize {
		for i, suggestion := range filtered {
			if len(suggestion) > 0 && len(filtered) <= s.maxSuggestions {
				filtered[i] = strings.ToUpper(suggestion[:1]) + suggestion[1:]
			}
		}
	}

	if len(filtered) > s.maxSuggestions {
		filtered = filtered[:s.maxSuggestions]
	}

	return filtered
}

// wordInfo holds information about a word's position in text
type wordInfo struct {
	word     string
	position int
}

// checkState carries the reusable per-Check scratch buffer. Pooling it keeps the
// word slice off the hot-path allocator; each Check owns its instance for the
// duration of the call, so reuse stays race-free.
type checkState struct {
	words []wordInfo
}

var checkStatePool = sync.Pool{
	New: func() any {
		return &checkState{words: make([]wordInfo, 0, 64)}
	},
}

// extractWords extracts words from text with their positions
func (s *Spellchecker) extractWords(text string) []wordInfo {
	return s.extractWordsInto(text, nil)
}

// extractWordsInto appends extracted words into buf, reusing its capacity. Word
// strings are sub-slices of text rather than freshly built, avoiding a copy per
// word; the shared backing array is immutable so this is safe to retain.
func (s *Spellchecker) extractWordsInto(text string, buf []wordInfo) []wordInfo {
	words := buf[:0]
	wordStart := -1

	for i, r := range text {
		if unicode.IsLetter(r) || r == '\'' || r == '-' {
			if wordStart < 0 {
				wordStart = i
			}
		} else if wordStart >= 0 {
			words = append(words, wordInfo{word: text[wordStart:i], position: wordStart})
			wordStart = -1
		}
	}

	if wordStart >= 0 {
		words = append(words, wordInfo{word: text[wordStart:], position: wordStart})
	}

	return words
}

// containsDigit checks if a string contains any digit
func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// loadEnglishDictionary loads a basic English dictionary into the model
func loadEnglishDictionary(model *fuzzy.Model) error {
	commonWords := getCommonEnglishWords()

	model.Train(commonWords)

	capitalized := make([]string, 0, len(commonWords))
	for _, word := range commonWords {
		if len(word) > 0 {
			capitalized = append(capitalized, strings.ToUpper(word[:1])+word[1:])
		}
	}
	model.Train(capitalized)

	return nil
}

// loadDictionaryFromFile loads words from a custom dictionary file
func loadDictionaryFromFile(model *fuzzy.Model, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "#") {
			model.Train([]string{strings.ToLower(word)})
		}
	}

	return scanner.Err()
}

// getCommonEnglishWords returns a list of common English words for the dictionary
// This is an expanded set for better spell checking. In production, you'd load from a comprehensive dictionary file.
func getCommonEnglishWords() []string {
	return []string{
		// Common articles, prepositions, conjunctions
		"the", "a", "an", "and", "or", "but", "if", "then", "else", "when", "at", "by", "for",
		"with", "about", "as", "into", "through", "after", "over", "between", "out", "against",
		"during", "without", "before", "under", "around", "among", "of", "to", "from", "in",
		"on", "off", "upon", "within", "near", "beyond", "across", "behind", "beside", "below",
		"above", "inside", "outside", "toward", "towards", "unto", "via",

		// Common pronouns
		"i", "you", "he", "she", "it", "we", "they", "them", "their", "this", "that", "these",
		"those", "my", "your", "his", "her", "its", "our", "me", "him", "us", "who", "whom",
		"whose", "which", "what", "myself", "yourself", "himself", "herself", "itself", "ourselves",
		"themselves", "anyone", "someone", "everyone", "nobody", "somebody", "anybody", "everybody",

		// Common verbs
		"is", "are", "was", "were", "be", "been", "being", "have", "has", "had", "do", "does",
		"did", "will", "would", "could", "should", "may", "might", "must", "can", "get", "got",
		"make", "made", "go", "went", "gone", "take", "took", "taken", "come", "came", "see",
		"saw", "seen", "know", "knew", "known", "think", "thought", "look", "looked", "want",
		"wanted", "give", "gave", "given", "use", "used", "find", "found", "tell", "told",
		"ask", "asked", "work", "worked", "seem", "seemed", "feel", "felt", "try", "tried",
		"leave", "left", "call", "called", "become", "became", "put", "set", "keep", "kept",
		"begin", "began", "show", "showed", "shown", "hear", "heard", "play", "played", "run",
		"ran", "move", "moved", "live", "lived", "believe", "believed", "bring", "brought",
		"happen", "happened", "write", "wrote", "written", "sit", "sat", "stand", "stood",
		"lose", "lost", "pay", "paid", "meet", "met", "include", "included", "continue",
		"continued", "learn", "learned", "learnt", "change", "changed", "lead", "led",
		"understand", "understood", "watch", "watched", "follow", "followed", "stop", "stopped",
		"create", "created", "speak", "spoke", "spoken", "read", "reading", "allow", "allowed",
		"add", "added", "spend", "spent", "grow", "grew", "grown", "open", "opened", "walk",
		"walked", "win", "won", "offer", "offered", "remember", "remembered", "love", "loved",
		"consider", "considered", "appear", "appeared", "buy", "bought", "wait", "waited",
		"serve", "served", "die", "died", "send", "sent", "expect", "expected", "build",
		"built", "stay", "stayed", "fall", "fell", "fallen", "cut", "cutting", "reach",
		"reached", "kill", "killed", "remain", "remained", "suggest", "suggested", "raise",
		"raised", "pass", "passed", "sell", "sold", "require", "required", "report", "reported",
		"decide", "decided", "pull", "pulled",

		// Common adjectives
		"good", "new", "first", "last", "long", "great", "little", "own", "other", "old",
		"right", "big", "high", "different", "small", "large", "next", "early", "young",
		"important", "few", "public", "bad", "same", "able", "best", "better", "sure", "free",
		"real", "true", "full", "special", "certain", "clear", "whole", "white", "legal",
		"short", "major", "nice", "serious", "human", "local", "recent", "available", "likely",
		"late", "hard", "low", "easy", "strong", "possible", "national", "international",
		"common", "particular", "social", "political", "economic", "financial", "current",
		"specific", "general", "simple", "modern", "medical", "commercial", "traditional",
		"military", "natural", "happy", "sad", "beautiful", "wonderful", "amazing", "quick",
		"slow", "fast", "hot", "cold", "warm", "cool", "ready", "correct", "incorrect", "wrong",
		"black", "blue", "red",
		"green", "yellow", "orange", "purple", "brown", "gray", "grey", "dark", "light",
		"bright", "heavy", "fine", "dead", "close", "open", "wide", "narrow", "deep",
		"high", "low", "rich", "poor", "clean", "dirty", "quiet", "loud", "soft", "hard",

		// Common nouns
		"time", "person", "year", "way", "day", "thing", "man", "world", "life", "hand",
		"part", "child", "eye", "woman", "place", "work", "week", "case", "point", "government",
		"company", "number", "group", "problem", "fact", "people", "water", "room", "money",
		"story", "home", "night", "area", "book", "word", "business", "issue", "side", "kind",
		"head", "house", "service", "friend", "father", "power", "hour", "game", "line",
		"end", "member", "law", "car", "city", "community", "name", "president", "team",
		"minute", "idea", "kid", "body", "information", "back", "parent", "face", "others",
		"level", "office", "door", "health", "art", "war", "history", "party", "result",
		"change", "morning", "reason", "research", "girl", "guy", "moment", "air", "teacher",
		"force", "education", "state", "country", "nation", "family", "student", "system",
		"program", "question", "school", "law", "policy", "market", "rate", "music", "lot",
		"right", "society", "death", "mother", "stuff", "course", "job", "form", "event",
		"street", "report", "decision", "process", "month", "century", "husband", "wife",
		"son", "daughter", "brother", "sister", "foot", "heart", "example", "interest",
		"food", "land", "window", "phone", "period", "activity", "institution", "dog",
		"picture", "video", "film", "experience", "ground", "town", "center", "tree",
		"voice", "light", "color", "cost", "price", "value", "image", "god", "action",

		// Common adverbs
		"not", "also", "very", "so", "just", "how", "now", "where", "here", "there", "then",
		"up", "down", "only", "too", "well", "back", "even", "still", "never", "really",
		"most", "much", "always", "often", "sometimes", "today", "however", "away", "ago",
		"maybe", "perhaps", "rather", "quite", "almost", "already", "yet", "later", "soon",
		"enough", "far", "long", "more", "less", "once", "ever", "again", "instead",
		"together", "probably", "actually", "especially", "exactly", "certainly", "clearly",
		"completely", "nearly", "almost", "usually", "particularly", "quickly", "slowly",
		"recently", "finally", "suddenly", "currently", "directly", "immediately", "simply",

		// Numbers and quantifiers
		"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten",
		"first", "second", "third", "fourth", "fifth", "sixth", "seventh", "eighth", "ninth",
		"tenth", "hundred", "thousand", "million", "billion", "many", "several", "all",
		"both", "each", "every", "another", "some", "any", "more", "most", "few", "less",
		"least", "much", "enough", "half", "quarter", "dozen",

		// Common contractions base words
		"it's", "isn't", "aren't", "wasn't", "weren't", "haven't", "hasn't", "hadn't",
		"don't", "doesn't", "didn't", "won't", "wouldn't", "shouldn't", "couldn't",
		"can't", "mustn't", "i'm", "you're", "he's", "she's", "we're", "they're",
		"i've", "you've", "we've", "they've", "i'll", "you'll", "he'll", "she'll",
		"we'll", "they'll", "i'd", "you'd", "he'd", "she'd", "we'd", "they'd",
		"that's", "there's", "here's", "who's", "what's", "where's", "when's",
		"why's", "how's", "let's",

		// Tech/API/Web common words
		"user", "id", "name", "email", "password", "data", "api", "http", "https", "url",
		"json", "xml", "html", "css", "javascript", "code", "function", "method", "class",
		"object", "array", "string", "number", "boolean", "null", "undefined", "true", "false",
		"return", "value", "key", "field", "property", "attribute", "element", "node",
		"request", "response", "server", "client", "database", "query", "table", "column",
		"row", "record", "index", "search", "filter", "sort", "page", "limit", "offset",
		"create", "read", "update", "delete", "insert", "select", "where", "from", "join",
		"post", "get", "put", "patch", "status", "error", "message", "success", "fail",
		"failed", "valid", "invalid", "token", "session", "cookie", "cache", "file",
		"upload", "download", "version", "release", "latest", "timestamp", "date", "time",
		"format", "type", "size", "length", "count", "total", "list", "items", "results",
		"content", "text", "title", "description", "body", "header", "footer", "link",
		"image", "video", "audio", "media", "document", "pdf", "csv", "txt", "markdown",
		"config", "configuration", "settings", "options", "parameters", "arguments", "input",
		"output", "source", "destination", "target", "origin", "path", "directory", "folder",
		"endpoint", "route", "handler", "controller", "model", "view", "template", "component",
		"module", "package", "library", "framework", "plugin", "extension", "interface",
		"implementation", "abstract", "concrete", "public", "private", "protected", "static",
		"const", "var", "let", "async", "await", "promise", "callback", "event", "listener",
		"trigger", "dispatch", "emit", "subscribe", "publish", "observer", "pattern",
		"singleton", "factory", "builder", "adapter", "decorator", "proxy", "facade",
		"strategy", "command", "iterator", "composite", "visitor", "memento", "state",
		"hello", "world", "test", "example", "sample", "demo", "tutorial", "guide",
		"documentation", "reference", "manual", "help", "support", "about", "contact",
		"welcome", "homepage", "website", "application", "app", "software", "program",
		"tool", "utility", "service", "platform", "system", "solution", "product",
		"feature", "functionality", "capability", "requirement", "specification",

		// Common words for animals, nature
		"dog", "cat", "fox", "bird", "fish", "animal", "pet", "wild", "nature", "tree",
		"flower", "plant", "grass", "leaf", "forest", "mountain", "river", "lake", "ocean",
		"sea", "sky", "sun", "moon", "star", "cloud", "rain", "snow", "wind", "weather",
		"season", "spring", "summer", "autumn", "fall", "winter", "brown", "lazy", "quick",
		"jumps", "lazy", "over",
	}
}
