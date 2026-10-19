// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_spoken_symbols.go
package wirepod_ttr

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// How a symbol is said (spoken_symbols in gopod_brobots_settings.json; operator
// 2026-10-08: "add the missing symbols, the old 'times' and 'divided by', and the
// LaTeX cleanup to the speech cleanup", "$" = "dollars"). From the older
// special_characters_vars.go, with LaTeX math cleaned first so a model's
// "\frac{1}{2}" or "0.\overline{3}" is said, not spelled. Speech only.
var (
	gopodLatexBracketsRe  = regexp.MustCompile(`\\[()\[\]]`)
	gopodLatexFracRe      = regexp.MustCompile(`\\?[dt]?frac\s*\{([^{}]+)\}\s*\{([^{}]+)\}`)
	gopodLatexOverlineRe  = regexp.MustCompile(`\\?(?:overline|bar)\s*\{(\d+)\}`)
	gopodCombiningBarRe   = regexp.MustCompile(`((?:\d[\x{0305}\x{0304}])+)`)
	gopodLatexCommandRe   = regexp.MustCompile(`\\(times|cdot|div|pm|ldots|dots|left|right)\b`)
	gopodDollarAmountRe   = regexp.MustCompile(`\$\s?(\d+(?:,\d{3})*(?:\.\d+)?)`)
	gopodSpokenSymbolsGap = regexp.MustCompile(`[ \t]{2,}`)
	gopodHmmRe            = regexp.MustCompile(`(?i)\bh+m+\b`)
	gopodMarkdownLinkRe   = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
	gopodMilitaryTimeRe   = regexp.MustCompile(`\b([01]?\d|2[0-3]):?([0-5]\d)\b`)
)

// gopodSpeakMilitaryTime (2026-10-14, operator: "18:14 (Military Time) ...
// generate a 'Military Time' statement for robot safe say"; then "\"(Military
// Time)\" (hours) = \"Hours Military Time\" (hours military time)"): "1814 Hours
// Military Time" is said "eighteen fourteen hours military time", "0800 Hours
// Military Time" "zero eight hundred hours military time"
// (live_weather.military_label). Speech only - the bubble keeps the digits.
func gopodSpeakMilitaryTime(text string) string {
	label := strings.TrimSpace(gopodBrobotsSettings.LiveWeather.MilitaryLabel)
	if label == "" {
		return text
	}
	re := regexp.MustCompile(gopodMilitaryTimeRe.String() + `\s+(?i:` + regexp.QuoteMeta(label) + `)`)
	return re.ReplaceAllStringFunc(text, func(m string) string {
		p := re.FindStringSubmatch(m)
		h, _ := strconv.Atoi(p[1])
		min, _ := strconv.Atoi(p[2])
		return gopodMilitaryTimeSpoken(h, min) + " " + strings.ToLower(label)
	})
}

// gopodSpeakYears (2026-10-17, operator: "say year similar as 20 26"): a year
// after its date ("October 17, 2026") is said in two halves, "20 26". Speech
// only; 2000-2009 stay as they are.
var gopodDateYearRe = regexp.MustCompile(`\b((?:January|February|March|April|May|June|July|August|September|October|November|December) \d{1,2}, )(\d{2})([1-9]\d)\b`)

func gopodSpeakYears(text string) string {
	return gopodDateYearRe.ReplaceAllString(text, "$1$2 $3")
}

// gopodSpeakFlags (2026-10-17, operator: "Windsor / Ontario / Canada [flag]"):
// a country flag is said as its country's name (its longest name in
// clean_politics.countries), unless that country was just named before it.
// gopodCountryName: a flag code's country, its longest name, in capitals
// ("CA" = "Canada", "US" = "United States"); "" when not listed.
func gopodCountryName(code string) string {
	long := ""
	for name, c := range gopodBrobotsSettings.CleanPolitics.Countries {
		if strings.EqualFold(c, code) && (len(name) > len(long) || len(name) == len(long) && name < long) {
			long = name
		}
	}
	words := strings.Fields(long)
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

func gopodSpeakFlags(text string) string {
	return gopodFlagRE.ReplaceAllStringFunc(text, func(flag string) string {
		code := gopodFlagCode(flag)
		name := gopodCountryName(code)
		if name == "" {
			return flag
		}
		at := strings.Index(text, flag)
		before := strings.ToLower(text[:at])
		if len(before) > 40 {
			before = before[len(before)-40:]
		}
		for n, c := range gopodBrobotsSettings.CleanPolitics.Countries {
			if c == code && strings.Contains(before, n) {
				return ""
			}
		}
		return name
	})
}

// gopodSayNegatives: "−" right before a digit, with no number or ")" before it
// (spaces aside), is said word ("negative"); between two numbers it stays minus.
func gopodSayNegatives(text, word string) string {
	r := []rune(text)
	var b strings.Builder
	for i, c := range r {
		if c == '\u2212' && i+1 < len(r) && r[i+1] >= '0' && r[i+1] <= '9' {
			j := i - 1
			for j >= 0 && r[j] == ' ' {
				j--
			}
			if j < 0 || !(r[j] >= '0' && r[j] <= '9' || r[j] == ')' || r[j] == '\u0305') {
				b.WriteString(word + " ")
				continue
			}
		}
		b.WriteRune(c)
	}
	return b.String()
}

var gopodLatexCommandSymbol = map[string]string{
	"times": "×", "cdot": "×", "div": "÷", "pm": "±", "ldots": "...", "dots": "...", "left": "", "right": "",
}

func gopodSpeakSymbols(text string) string {
	ss := gopodBrobotsSettings.SpokenSymbols
	// Every "hmm" (Hm, Hmm, Hmmmm...) is said hmm_said (operator 2026-10-08).
	if ss.HmmSaid != "" {
		text = gopodHmmRe.ReplaceAllString(text, ss.HmmSaid)
	}
	text = gopodSpeakMilitaryTime(text)
	text = gopodSpeakYears(text)
	text = gopodSpeakFlags(text)
	// 2026-10-15: a link ("[GOPOD Weather Kit](/gopod_weather_kit.html)") is
	// said by its words; the bubble shows it as a link.
	text = gopodMarkdownLinkRe.ReplaceAllString(text, "$1")
	text = gopodLatexBracketsRe.ReplaceAllString(text, "")
	text = gopodLatexFracRe.ReplaceAllString(text, "$1/$2")
	repeating := " " + ss.RepeatingWord
	text = gopodLatexOverlineRe.ReplaceAllString(text, "$1"+repeating)
	text = gopodCombiningBarRe.ReplaceAllStringFunc(text, func(m string) string {
		return strings.NewReplacer("̅", "", "̄", "").Replace(m) + repeating
	})
	// 2026-10-14: the math's minus sign (gopod_math.go) in front of a number,
	// not between two, is said "negative" (−3 + 5).
	if word := ss.NegativeWord; word != "" {
		text = gopodSayNegatives(text, word)
	}
	text = gopodLatexCommandRe.ReplaceAllStringFunc(text, func(m string) string {
		return gopodLatexCommandSymbol[strings.TrimPrefix(m, `\`)]
	})
	if word := ss.Words["$"]; word != "" {
		text = gopodDollarAmountRe.ReplaceAllString(text, "$1 "+word)
	}
	keys := make([]string, 0, len(ss.Words))
	for k := range ss.Words {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, k := range keys {
		if strings.Contains(text, k) {
			text = strings.ReplaceAll(text, k, " "+ss.Words[k]+" ")
		}
	}
	text = gopodSpokenSymbolsGap.ReplaceAllString(text, " ")
	return strings.TrimSpace(strings.NewReplacer(" ,", ",", " .", ".", " !", "!", " ?", "?").Replace(text))
}
