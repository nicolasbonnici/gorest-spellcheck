package spellcheck

import (
	"bufio"
	"os"
	"strings"
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

// NewSpellchecker creates a new spellchecker with the given configuration
func NewSpellchecker(config *Config) (*Spellchecker, error) {
	model := fuzzy.NewModel()

	model.SetDepth(4)
	model.SetThreshold(1)

	if err := loadEnglishDictionary(model); err != nil {
		return nil, err
	}

	if config.CustomDictionary != "" {
		if err := loadDictionaryFromFile(model, config.CustomDictionary); err != nil {
			return nil, err
		}
	}

	ignoredWords := make(map[string]bool)
	for _, word := range config.IgnoredWords {
		if config.CaseSensitive {
			ignoredWords[word] = true
		} else {
			ignoredWords[strings.ToLower(word)] = true
		}
	}

	return &Spellchecker{
		model:          model,
		config:         config,
		ignoredWords:   ignoredWords,
		caseSensitive:  config.CaseSensitive,
		minWordLength:  config.MinWordLength,
		maxSuggestions: config.MaxSuggestions,
	}, nil
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
	words := s.extractWords(text)

	for _, wordInfo := range words {
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

		if !s.isCorrect(word) {
			suggestions := s.getSuggestions(word)
			errors.Add(&SpellingError{
				Word:        word,
				Position:    position,
				Suggestions: suggestions,
			})
		}
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

// isCorrect checks if a word is spelled correctly
func (s *Spellchecker) isCorrect(word string) bool {
	checkWord := word
	if !s.caseSensitive {
		checkWord = strings.ToLower(word)
	}

	correction := s.model.SpellCheck(checkWord)
	correctionStr := string(correction)

	if correctionStr == "" || correctionStr == checkWord {
		return true
	}

	return false
}

// getSuggestions returns spelling suggestions for a misspelled word
func (s *Spellchecker) getSuggestions(word string) []string {
	checkWord := word
	if !s.caseSensitive {
		checkWord = strings.ToLower(word)
	}

	correction := s.model.SpellCheck(checkWord)
	correctionStr := string(correction)

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

	if len(filtered) > s.maxSuggestions {
		filtered = filtered[:s.maxSuggestions]
	}

	if len(word) > 0 && unicode.IsUpper(rune(word[0])) && !s.caseSensitive {
		for i, suggestion := range filtered {
			if len(suggestion) > 0 {
				filtered[i] = strings.ToUpper(suggestion[:1]) + suggestion[1:]
			}
		}
	}

	return filtered
}

// wordInfo holds information about a word's position in text
type wordInfo struct {
	word     string
	position int
}

// extractWords extracts words from text with their positions
func (s *Spellchecker) extractWords(text string) []wordInfo {
	var words []wordInfo
	var currentWord strings.Builder
	wordStart := -1

	for i, r := range text {
		if unicode.IsLetter(r) || r == '\'' || r == '-' {
			if currentWord.Len() == 0 {
				wordStart = i
			}
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				words = append(words, wordInfo{
					word:     currentWord.String(),
					position: wordStart,
				})
				currentWord.Reset()
				wordStart = -1
			}
		}
	}

	// Don't forget the last word if text doesn't end with punctuation
	if currentWord.Len() > 0 {
		words = append(words, wordInfo{
			word:     currentWord.String(),
			position: wordStart,
		})
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

	for _, word := range commonWords {
		if len(word) > 0 {
			capitalized := strings.ToUpper(word[:1]) + word[1:]
			model.Train([]string{capitalized})
		}
	}

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
