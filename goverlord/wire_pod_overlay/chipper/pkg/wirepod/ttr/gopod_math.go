// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_math.go
package wirepod_ttr

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The Brobots' math (operator 2026-10-14: "GOPOD is like as 'mathy' as math can
// get"; "gemma isn't for calculating"; "Build it all ... fractions, decimals,
// repeating decimals, same ground as fractions pack * Plus/minus/times/divide and
// percent * clean, correct & proper"). A math question in the live chat is
// worked out here, exactly (math/big, no rounding), and answered step by step,
// word for word - the LLM never calculates. What it answers:
//
//   - fraction -> decimal: long division; a remainder that comes back is a
//     repeating block, shown with a bar (1/3 = 0.3̅, 1/7 = 0.1̅4̅2̅8̅5̅7̅);
//   - decimal -> fraction: over 10, 100, 1000..., then simplified;
//   - repeating decimal -> fraction: the 10x method (x = 0.3̅, 10x = 3.3̅,
//     9x = 3, x = 1/3);
//   - percent: to and from fractions and decimals, "15% of 80", "what percent
//     is 3 of 4";
//   - plus, minus, times, divide - with fractions, decimals, repeating decimals
//     and percents, brackets and the order of operations, one line per step.
//
// It reads typed digits and symbols ("0.3̅", "\overline{3}", "0.333...", "3/4",
// "×", "÷") and the speech-to-text's words ("zero point three repeating", "three
// quarters", "fifteen percent of eighty"). Each step's wording is a line in
// math.lines in the settings file. Fractions show with the fraction slash (1⁄3)
// and minus as "−", so speech can say them ("over", "minus": spoken_symbols)
// without touching anything else the Brobots say.

// "Explain the Math" (operator 2026-10-14: "same as livestream contest format:
// \"[statement]. Explain the Math\""; "Brobot 1 to answer & Brobot 2 to follow
// up, Brobots shared-persona style"; "Wrap all that up as an \"Explain the
// Math\" pop-up"): a line that ends with math.trigger is a math challenge.
// Brobot 1 answers first, then Brobot 2 follows up (gopodMathFollowUp, kgsim.go):
// each says a line or two of his own (the LLM, any numbers in it dropped -
// gopodMathFlavour), then his share of the steps, one bullet each, word for word.

// gopodMathExplain: the statement before math.trigger, when the line ends with it.
func gopodMathExplain(text string) (string, bool) {
	m := gopodBrobotsSettings.Math
	trigger := strings.ToLower(strings.TrimSpace(m.Trigger))
	if !m.Enabled || trigger == "" {
		return "", false
	}
	line := strings.TrimRight(strings.TrimSpace(text), " .!?,")
	if !strings.HasSuffix(strings.ToLower(line), trigger) {
		return "", false
	}
	statement := strings.TrimRight(strings.TrimSpace(line[:len(line)-len(trigger)]), " .!?,:;-")
	if statement == "" {
		return "", false
	}
	return statement, true
}

// gopodMathParts: the steps for an "Explain the Math" line, split between the
// Brobots - Brobot 2 gets the last third (at least the last line), Brobot 1 the
// rest; one step is Brobot 1's alone.
func gopodMathParts(text string) (first, second []string, ok bool) {
	statement, ok := gopodMathExplain(text)
	// 2026-10-16 (operator, on "Explain the Math" as the math phrase: "Not
	// really..."; "Fraction ... covers either or both"): a line with one of
	// math.fraction_words is worked out too, whole.
	if !ok && gopodMathFractionLine(text) {
		statement, ok = text, true
	}
	if !ok {
		return nil, nil, false
	}
	lines, ok := gopodMathWork(statement)
	if !ok || len(lines) == 0 {
		return nil, nil, false
	}
	if len(lines) == 1 {
		return lines, []string{}, true
	}
	n := len(lines) / 3
	if n < 1 {
		n = 1
	}
	return lines[:len(lines)-n], lines[len(lines)-n:], true
}

// gopodMathFractionLine: the line has one of math.fraction_words.
func gopodMathFractionLine(text string) bool {
	lower := strings.ToLower(text)
	for _, w := range gopodBrobotsSettings.Math.FractionWords {
		if gopodHasWholeWords(lower, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

// gopodMathDecimalLine (operator 2026-10-16: "A decimal is a specialized kind
// of fraction ... This clarification is another good follow up for brobot 2"):
// Brobot 2's follow-up says math.decimal_follow_up_line when the line is about
// a decimal.
func gopodMathDecimalLine(text string) string {
	lower := strings.ToLower(text)
	if gopodHasWholeWords(lower, "decimal") || gopodHasWholeWords(lower, "decimals") {
		return strings.TrimSpace(gopodBrobotsSettings.Math.DecimalFollowUpLine)
	}
	return ""
}

// gopodMathReply: the Brobot's reply - his own line or two from the LLM (its
// animation cue kept, any sentence with a number dropped, at most
// math.flavour_sentences), then his steps as bullets.
func gopodMathReply(llm string, steps []string, extra ...string) string {
	m := gopodBrobotsSettings.Math
	cue := ""
	rest := strings.TrimSpace(llm)
	if loc := gopodAnimationMarkerPattern.FindStringIndex(rest); loc != nil && loc[0] == 0 {
		cue, rest = rest[:loc[1]], rest[loc[1]:]
	}
	flavour := gopodMathFlavour(rest, m.FlavourSentences)
	for _, e := range extra {
		if e = strings.TrimSpace(e); e != "" {
			flavour = strings.TrimSpace(flavour + " " + e)
		}
	}
	var b strings.Builder
	b.WriteString(cue)
	b.WriteString(flavour)
	for _, s := range steps {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(m.Bullet + s)
	}
	return strings.TrimSpace(b.String())
}

// gopodMathFlavour: up to max sentences of the LLM's words with no numbers in
// them - digits or number words - so only the code's steps carry the math.
func gopodMathFlavour(text string, max int) string {
	var keep []string
	for _, s := range gopodMathSentenceRe.FindAllString(strings.ReplaceAll(text, "\n", " "), -1) {
		s = strings.TrimSpace(s)
		if s == "" || strings.ContainsAny(gopodMathNormalize(s), "0123456789=×÷+−%") {
			continue
		}
		keep = append(keep, s)
		if len(keep) >= max {
			break
		}
	}
	return strings.Join(keep, " ")
}

var gopodMathSentenceRe = regexp.MustCompile(`[^.!?]+[.!?]*`)

// gopodMathFollowUp: the follow-up Brobot's turn after the lead's answer (kgsim.go).
var gopodMathFollowUp sync.Map

// gopodMathLeadPick (2026-10-24, operator: "'All' brobot 1 or 2 menu selections *
// keyboard & mic inputs, when 2 selected * brobot 2 then brobot 1 order, not being
// honoured"): a line sent to All with Brobot 2 first - he leads its math, Brobot 1
// follows up. Any other math line: Brobot 1 leads (2026-10-14).
var gopodMathLeadPick sync.Map

func gopodMathLead(text string) string {
	if v, ok := gopodMathLeadPick.Load(strings.TrimSpace(text)); ok {
		return v.(string)
	}
	return gopodBrobot1ESN
}

// gopodMathFollowUpStart: once the lead's answer is done, his brother follows up
// on the same line - his share of the steps, with the lead's answer in the
// shared chat.
func gopodMathFollowUpStart(text string) {
	follower := gopodBrotherOf(gopodMathLead(text))
	logger.Println("GOPOD_MATH " + gopodBrobotName(follower) + " follows up: " + text)
	gopodMathFollowUp.Store(follower, true)
	gopodAnswerNext.Store(follower, true)
	gopodFollowingUp.Store(follower, true)
	if _, err := StreamingKGSim(nil, follower, text, false); err != nil {
		logger.Println("LLM error (math follow-up): " + err.Error())
	}
	gopodMathLeadPick.Delete(strings.TrimSpace(text))
}

// gopodMathPopupMark: a pop-up item that is an "Explain the Math" challenge.
const gopodMathPopupMark = "math:"

// gopodMathPopup: the "Explain the Math" pop-up, live chat only (no draft, no
// run), up to topic_pictures.max_per_session times - Goverlord's
// congratulations over the challenge, Brobot 1's answer under it; Brobot 2's
// follow-up gets its own marquee bubble (the page's pop-up rule, 2026-10-15).
func gopodMathPopup(esn, text string) {
	tp := gopodBrobotsSettings.TopicPictures
	statement, ok := gopodMathExplain(text)
	if !ok || tp.Role == "" || gopodScriptPlaying.Load() || gopodRunActive() || !gopodPopupAllowed("math") {
		return
	}
	// 2026-10-16 (operator: '"AI Wordplay Explain the Math" = title of the
	// livestream show = easter egg'; "No Prize yet. Start with a 'Doc's Almost
	// Approved' Trophy"): math.egg_name, the trophy (popup_title) over the
	// submission.
	gopodShowEgg(gopodBrobotName(esn), gopodBrobotsSettings.Math.EggName, gopodMathPopupMark+strings.ReplaceAll(statement, "|", " "))
}

// gopodMathWork: the steps for a math question, one line each.
func gopodMathWork(text string) ([]string, bool) {
	norm := gopodMathNormalize(text)
	toks := gopodMathTokenize(norm)
	if lines, ok := gopodMathWhatPercent(norm); ok {
		return lines, true
	}
	if target, at := gopodMathTarget(toks); target != "" {
		if run := gopodMathRunBefore(toks, at); len(run) > 0 {
			if lines, ok := gopodMathConvert(run, target); ok {
				return lines, true
			}
		}
	}
	// "convert 0.3 repeating" / "simplify 6/8" with no target named.
	if gopodMathHasWord(toks, "convert") || gopodMathHasWord(toks, "simplify") {
		for _, run := range gopodMathRuns(toks) {
			if len(run) == 1 && run[0].kind == mathTokNum {
				target := map[string]string{"rep": "fraction", "dec": "fraction", "frac": "decimal", "pct": "fraction", "int": ""}[run[0].num.kind]
				if gopodMathHasWord(toks, "simplify") && run[0].num.kind == "frac" {
					target = "simplify"
				}
				if target != "" {
					if lines, ok := gopodMathConvert(run, target); ok {
						return lines, true
					}
				}
			}
		}
	}
	for _, run := range gopodMathRuns(toks) {
		if !gopodMathHasOp(run) {
			continue
		}
		if lines, ok := gopodMathArithmetic(run); ok {
			return lines, true
		}
	}
	return nil, false
}

// ---------------------------------------------------------------- numbers

// gopodMathNum: an exact number and how it was written.
type gopodMathNum struct {
	r    *big.Rat
	kind string // int, dec, frac, rep, pct
	show string // how a step shows it
}

const gopodMathBar = "\u0305"   // combining overline: the repeating digits
const gopodMathSlash = "\u2044" // fraction slash: 1⁄3
const gopodMathMinus = "\u2212" // minus sign

// gopodMathDecimalParts: r's decimal digits by long division - the whole part,
// the digits before the repeat, and the repeating block (none when it ends).
// done is false when no repeat or end shows within max digits.
func gopodMathDecimalParts(r *big.Rat, max int) (neg bool, whole, fixed, rep string, done bool) {
	num := new(big.Int).Abs(r.Num())
	den := new(big.Int).Set(r.Denom())
	neg = r.Sign() < 0
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	whole = q.String()
	seen := map[string]int{}
	var digits []byte
	ten := big.NewInt(10)
	for rem.Sign() != 0 && len(digits) < max {
		key := rem.String()
		if at, ok := seen[key]; ok {
			return neg, whole, string(digits[:at]), string(digits[at:]), true
		}
		seen[key] = len(digits)
		rem.Mul(rem, ten)
		d, r2 := new(big.Int).QuoRem(rem, den, new(big.Int))
		digits = append(digits, byte('0'+d.Int64()))
		rem = r2
	}
	if rem.Sign() == 0 {
		return neg, whole, string(digits), "", true
	}
	return neg, whole, string(digits), "", false
}

// gopodMathBarred: digits with a bar over each (3 -> 3̅).
func gopodMathBarred(digits string) string {
	var b strings.Builder
	for _, d := range digits {
		b.WriteRune(d)
		b.WriteString(gopodMathBar)
	}
	return b.String()
}

func gopodMathSign(neg bool) string {
	if neg {
		return gopodMathMinus
	}
	return ""
}

// gopodMathDecimal: r as a decimal, its repeating block barred. ok is false when
// the repeat is longer than math.max_repeat_digits (shown rounded instead).
func gopodMathDecimal(r *big.Rat) (string, bool) {
	maxRep := gopodBrobotsSettings.Math.MaxRepeatDigits
	if maxRep <= 0 {
		maxRep = 12
	}
	neg, whole, fixed, rep, done := gopodMathDecimalParts(r, 4*maxRep+20)
	if done && len(rep) <= maxRep {
		s := gopodMathSign(neg) + whole
		if fixed != "" || rep != "" {
			s += "." + fixed + gopodMathBarred(rep)
		}
		return s, true
	}
	f, _ := r.Float64()
	s := strconv.FormatFloat(f, 'f', 10, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return strings.Replace(s, "-", gopodMathMinus, 1) + "…", false
}

// gopodMathFraction: r as a fraction (3⁄4), a whole number as itself.
func gopodMathFraction(r *big.Rat) string {
	if r.IsInt() {
		return gopodMathInt(r.Num())
	}
	n := new(big.Int).Abs(r.Num())
	return gopodMathSign(r.Sign() < 0) + n.String() + gopodMathSlash + r.Denom().String()
}

func gopodMathInt(n *big.Int) string {
	if n.Sign() < 0 {
		return gopodMathMinus + new(big.Int).Abs(n).String()
	}
	return n.String()
}

// gopodMathMixed: an improper fraction as a mixed number ("1 and 1⁄9"), or "".
func gopodMathMixed(r *big.Rat) string {
	if r.IsInt() {
		return ""
	}
	n := new(big.Int).Abs(r.Num())
	q, rem := new(big.Int).QuoRem(n, r.Denom(), new(big.Int))
	if q.Sign() == 0 {
		return ""
	}
	return gopodMathSign(r.Sign() < 0) + q.String() + " and " + rem.String() + gopodMathSlash + r.Denom().String()
}

// gopodMathShow: a result shown the way its kind is written - a fraction for
// fractions and repeating decimals, a decimal otherwise.
func gopodMathShow(r *big.Rat, kind string) string {
	if r.IsInt() {
		return gopodMathInt(r.Num())
	}
	if kind == "frac" || kind == "rep" {
		return gopodMathFraction(r)
	}
	s, _ := gopodMathDecimal(r)
	return s
}

// gopodMathParseRepeat: whole.fixed + barred rep -> exact value.
func gopodMathParseRepeat(whole, fixed, rep string) *big.Rat {
	if rep == "" {
		r, _ := new(big.Rat).SetString(whole + "." + fixed + "0")
		return r
	}
	a, b := len(fixed), len(rep)
	all, _ := new(big.Int).SetString(whole+fixed+rep, 10)
	head, _ := new(big.Int).SetString(whole+fixed, 10)
	numer := new(big.Int).Sub(all, head)
	den := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(a+b)), nil), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(a)), nil))
	return new(big.Rat).SetFrac(numer, den)
}

// ---------------------------------------------------------------- reading

var (
	gopodMathLatexFracRe   = regexp.MustCompile(`\\?[dt]?frac\s*\{\s*(-?\d+)\s*\}\s*\{\s*(\d+)\s*\}`)
	gopodMathLatexBarRe    = regexp.MustCompile(`\\?(?:overline|bar|dot)\s*\{\s*(\d+)\s*\}`)
	gopodMathLatexCmdRe    = regexp.MustCompile(`\\(times|cdot|div|left|right|[()\[\]])`)
	gopodMathDotsRe        = regexp.MustCompile(`(\d+)\.(\d+)\s*(?:\.\.\.+|…)`)
	gopodMathRepeatWordRe  = regexp.MustCompile(`(\d+)\.(\d+)\s+(?:repeating|recurring|repeated|repeats)\b`)
	gopodMathPercentWordRe = regexp.MustCompile(`(\d)\s*(?:per\s*cent|percent)\b`)
	gopodMathOutOfRe       = regexp.MustCompile(`(\d)\s+out\s+of\s+(\d)`)
	gopodMathOverRe        = regexp.MustCompile(`(\d)\s+over\s+(\d)`)
	gopodMathTimesXRe      = regexp.MustCompile(`(\d)\s*[x*]\s*(\d|\()`)
	gopodMathMinusRe       = regexp.MustCompile(`(\d|\))\s+[-–]\s+(\d|\()`)
	gopodMathNameRe        = regexp.MustCompile(`\bbro\s*-?\s*bots?\s*[0-9]+\b|\bbro\s*-?\s*bots?\b`)
	gopodMathSpacesRe      = regexp.MustCompile(`\s+`)
)

// gopodMathNormalize: the line in one form - lower case, LaTeX and number words
// turned into digits and symbols, repeating decimals barred.
func gopodMathNormalize(text string) string {
	s := " " + strings.ToLower(text) + " "
	s = gopodMathNameRe.ReplaceAllString(s, " ")
	s = gopodMathLatexFracRe.ReplaceAllString(s, " $1/$2 ")
	s = gopodMathLatexBarRe.ReplaceAllStringFunc(s, func(m string) string {
		return gopodMathBarred(gopodMathLatexBarRe.FindStringSubmatch(m)[1])
	})
	s = gopodMathLatexCmdRe.ReplaceAllStringFunc(s, func(m string) string {
		switch m[1:] {
		case "times", "cdot":
			return " × "
		case "div":
			return " ÷ "
		}
		return " "
	})
	s = strings.NewReplacer("’", "'", "\u2212", " \u2212 ", "·", " × ", "\u0304", gopodMathBar).Replace(s)
	s = gopodMathWordsToNumbers(s)
	s = gopodMathOnlyRepeats(s)
	// "0.3 repeating" bars the 3; "0.1666 repeating" the 6, "0.333 repeating"
	// one 3 - the block that repeats at the end, else every digit.
	s = gopodMathRepeatWordRe.ReplaceAllStringFunc(s, func(m string) string {
		p := gopodMathRepeatWordRe.FindStringSubmatch(m)
		fixed, rep := gopodMathFindPeriod(p[2])
		return p[1] + "." + fixed + gopodMathBarred(rep)
	})
	s = gopodMathDotsRe.ReplaceAllStringFunc(s, func(m string) string {
		p := gopodMathDotsRe.FindStringSubmatch(m)
		fixed, rep := gopodMathFindPeriod(p[2])
		return p[1] + "." + fixed + gopodMathBarred(rep)
	})
	s = gopodMathPercentWordRe.ReplaceAllString(s, "$1%")
	s = gopodMathOutOfRe.ReplaceAllString(s, "$1/$2")
	s = gopodMathOverRe.ReplaceAllString(s, "$1/$2")
	for i := 0; i < 2; i++ { // twice: "2 x 3 x 4" shares its middle number
		s = gopodMathTimesXRe.ReplaceAllString(s, "$1 × $2")
		s = gopodMathMinusRe.ReplaceAllString(s, "$1 − $2")
	}
	for _, w := range gopodMathWordOps {
		s = w.re.ReplaceAllString(s, w.to)
	}
	s = strings.NewReplacer(
		" plus ", " + ", " minus ", " − ", " take away ", " − ",
		" multiplied by ", " × ", " times ", " × ",
		" divided by ", " ÷ ", " divide by ", " ÷ ",
	).Replace(s)
	return strings.TrimSpace(gopodMathSpacesRe.ReplaceAllString(s, " "))
}

// gopodMathRepeatSplitRe / gopodMathRepeatLaterRe: the digits that repeat,
// named after the number - "123, 4 repeating" (the digits that don't, then the
// ones that do); "the 4 repeating", "only the 4 repeats".
var gopodMathRepeatSplitRe = regexp.MustCompile(`\b(\d+),?\s+(\d+)\s+(?:repeating|recurring|repeated|repeats)\b`)
var gopodMathRepeatLaterRe = regexp.MustCompile(`\b(?:with\s+)?(?:only\s+)?(?:the\s+)?(\d+)\s+(?:repeating|recurring|repeated|repeats)\b`)

// gopodMathOnlyRepeats (2026-10-16, operator: "convert 0.1234 repeating into a
// fraction. 123, 4 repeating * only the 4 repeats"): a decimal whose repeating
// digits are named after it - the end of its digits - gets only those barred
// (0.1234 with "4 repeating" is 0.123 then 4s), and the naming words go.
func gopodMathOnlyRepeats(s string) string {
	for _, d := range gopodMathDecimalRe.FindAllStringSubmatchIndex(s, -1) {
		whole, digits := s[d[2]:d[3]], s[d[4]:d[5]]
		// "0.1234 repeating ... 4 repeating": a "repeating" right after the number goes.
		rest := s[d[1]:]
		after := strings.TrimLeft(rest, " ")
		skip := len(rest) - len(after)
		for _, w := range []string{"repeating", "recurring", "repeated", "repeats"} {
			if strings.HasPrefix(after, w) {
				skip += len(w)
				break
			}
		}
		later := s[d[1]+skip:]
		rep, from, to := "", -1, -1
		if m := gopodMathRepeatSplitRe.FindStringSubmatchIndex(later); m != nil && later[m[2]:m[3]]+later[m[4]:m[5]] == digits {
			rep, from, to = later[m[4]:m[5]], m[0], m[1]
		} else if m := gopodMathRepeatLaterRe.FindStringSubmatchIndex(later); m != nil {
			rep, from, to = later[m[2]:m[3]], m[0], m[1]
		}
		if rep == "" || len(rep) >= len(digits) || !strings.HasSuffix(digits, rep) {
			continue
		}
		barred := whole + "." + digits[:len(digits)-len(rep)] + gopodMathBarred(rep)
		return s[:d[0]] + barred + later[:from] + later[to:]
	}
	return s
}

// gopodMathDecimalRe: a decimal number, its whole part and its digits.
var gopodMathDecimalRe = regexp.MustCompile(`(\d+)\.(\d+)`)

// gopodMathFindPeriod: "0.333..." / "0.1666..." - the shortest block that
// repeats at least twice at the end; with none, every digit repeats.
func gopodMathFindPeriod(digits string) (fixed, rep string) {
	for p := 1; p <= len(digits)/2; p++ {
		block := digits[len(digits)-p:]
		start := len(digits) - p
		for start-p >= 0 && digits[start-p:start] == block {
			start -= p
		}
		if len(digits)-start >= 2*p {
			return digits[:start], block
		}
	}
	return "", digits
}

var gopodMathUnits = map[string]int{
	"zero": 0, "one": 1, "two": 2, "three": 3, "four": 4, "five": 5, "six": 6, "seven": 7, "eight": 8, "nine": 9,
	"ten": 10, "eleven": 11, "twelve": 12, "thirteen": 13, "fourteen": 14, "fifteen": 15, "sixteen": 16,
	"seventeen": 17, "eighteen": 18, "nineteen": 19,
}
var gopodMathTens = map[string]int{
	"twenty": 20, "thirty": 30, "forty": 40, "fourty": 40, "fifty": 50, "sixty": 60, "seventy": 70, "eighty": 80, "ninety": 90,
}
var gopodMathScales = map[string]int{"hundred": 100, "thousand": 1000, "million": 1000000}

// gopodMathDenominators: the fraction words ("three quarters", "a third").
var gopodMathDenominators = map[string]int{
	"half": 2, "halves": 2, "third": 3, "thirds": 3, "quarter": 4, "quarters": 4, "fourth": 4, "fourths": 4,
	"fifth": 5, "fifths": 5, "sixth": 6, "sixths": 6, "seventh": 7, "sevenths": 7, "eighth": 8, "eighths": 8,
	"ninth": 9, "ninths": 9, "tenth": 10, "tenths": 10, "eleventh": 11, "elevenths": 11, "twelfth": 12, "twelfths": 12,
	"twentieth": 20, "twentieths": 20, "hundredth": 100, "hundredths": 100, "thousandth": 1000, "thousandths": 1000,
}

// gopodMathWordsToNumbers: "zero point three" -> "0.3", "one hundred and five"
// -> "105", "three quarters" / "a third" -> "3/4" / "1/3", "one and a half" ->
// "1 1/2" (a mixed number). Digits already there are kept.
func gopodMathWordsToNumbers(s string) string {
	words := strings.Fields(strings.NewReplacer("-", " - ", ",", " , ", "?", " ? ", "!", " ! ").Replace(s))
	// Hyphenated tens ("twenty-five") come back together.
	var w []string
	for i := 0; i < len(words); i++ {
		if words[i] == "-" && i > 0 && i+1 < len(words) && gopodMathTens[words[i-1]] > 0 && gopodMathUnits[words[i+1]] > 0 && gopodMathUnits[words[i+1]] < 10 {
			continue
		}
		w = append(w, words[i])
	}
	isNumWord := func(x string) bool {
		_, u := gopodMathUnits[x]
		_, t := gopodMathTens[x]
		_, sc := gopodMathScales[x]
		return u || t || sc
	}
	var out []string
	for i := 0; i < len(w); i++ {
		// "a half", "a third", "an eighth"
		if (w[i] == "a" || w[i] == "an") && i+1 < len(w) && gopodMathDenominators[w[i+1]] > 0 && w[i+1] != "second" {
			out = append(out, "1/"+strconv.Itoa(gopodMathDenominators[w[i+1]]))
			i++
			continue
		}
		if !isNumWord(w[i]) || (gopodMathScales[w[i]] > 0 && (i == 0 || !isNumWord(w[i-1]))) {
			// a lone "hundred" is "100"; other words stay
			if gopodMathScales[w[i]] > 0 {
				out = append(out, strconv.Itoa(gopodMathScales[w[i]]))
				continue
			}
			out = append(out, w[i])
			continue
		}
		total, current := 0, 0
		j := i
		for ; j < len(w); j++ {
			x := w[j]
			if v, ok := gopodMathUnits[x]; ok {
				current += v
			} else if v, ok := gopodMathTens[x]; ok {
				current += v
			} else if v, ok := gopodMathScales[x]; ok {
				if current == 0 {
					current = 1
				}
				if v == 100 {
					current *= 100
				} else {
					total += current * v
					current = 0
				}
			} else if x == "and" && j+1 < len(w) && isNumWord(w[j+1]) && (current%100 == 0 && current > 0 || total > 0) && gopodMathDenominators[gopodMathAfterNumber(w, j+1)] == 0 {
				continue // "one hundred and five"
			} else {
				break
			}
		}
		num := strconv.Itoa(total + current)
		// "point three three"
		if j < len(w) && w[j] == "point" {
			k := j + 1
			var digits strings.Builder
			for ; k < len(w); k++ {
				if v, ok := gopodMathUnits[w[k]]; ok && v < 10 {
					digits.WriteString(strconv.Itoa(v))
				} else {
					break
				}
			}
			if digits.Len() > 0 {
				num += "." + digits.String()
				j = k
			}
		}
		// "three quarters"
		if j < len(w) && gopodMathDenominators[w[j]] > 0 && !strings.Contains(num, ".") {
			num += "/" + strconv.Itoa(gopodMathDenominators[w[j]])
			j++
		}
		out = append(out, num)
		i = j - 1
	}
	// "1 and 1/2" -> mixed "1 1/2"
	s = " " + strings.Join(out, " ") + " "
	return gopodMathMixedAndRe.ReplaceAllString(s, " $1 $2 ")
}

// gopodMathWordOps: "add 2 and 3", "subtract 2 from 5", "multiply 3 by 4",
// "divide 8 by 2", "the sum of ...", "the product of ...".
const gopodMathN = `([\d./%` + gopodMathBar + `]+)`

var gopodMathWordOps = []struct {
	re *regexp.Regexp
	to string
}{
	{regexp.MustCompile(`\badd ` + gopodMathN + ` (?:and|to|plus) ` + gopodMathN), "$1 + $2"},
	{regexp.MustCompile(`\bsubtract ` + gopodMathN + ` from ` + gopodMathN), "$2 − $1"},
	{regexp.MustCompile(`\bmultiply ` + gopodMathN + ` (?:by|and|times) ` + gopodMathN), "$1 × $2"},
	{regexp.MustCompile(`\bdivide ` + gopodMathN + ` (?:by|into) ` + gopodMathN), "$1 ÷ $2"},
	{regexp.MustCompile(`\bsum of ` + gopodMathN + ` and ` + gopodMathN), "$1 + $2"},
	{regexp.MustCompile(`\bproduct of ` + gopodMathN + ` and ` + gopodMathN), "$1 × $2"},
	{regexp.MustCompile(`\bdifference (?:of|between) ` + gopodMathN + ` and ` + gopodMathN), "$1 − $2"},
}

var gopodMathMixedAndRe = regexp.MustCompile(`\s(\d+)\s+and\s+(\d+/\d+)\s`)

// gopodMathAfterNumber: the first word after the number words from i.
func gopodMathAfterNumber(w []string, i int) string {
	for ; i < len(w); i++ {
		if _, ok := gopodMathUnits[w[i]]; ok {
			continue
		}
		if _, ok := gopodMathTens[w[i]]; ok {
			continue
		}
		return w[i]
	}
	return ""
}

// ---------------------------------------------------------------- tokens

const (
	mathTokNum = iota
	mathTokOp
	mathTokOpen
	mathTokClose
	mathTokPct
	mathTokOf
	mathTokWord
)

type gopodMathTok struct {
	kind int
	op   string // + − × ÷
	num  gopodMathNum
	word string
}

var gopodMathNumRe = regexp.MustCompile(`^(\d+)\s(\d+)/(\d+)|^(\d+)/(\d+)|^(\d+)\.((?:\d` + gopodMathBar + `?)+)|^\.((?:\d` + gopodMathBar + `?)+)|^(\d+)`)

// gopodMathTokenize: numbers, operators, brackets, "%", "of" and the other words.
func gopodMathTokenize(s string) []gopodMathTok {
	var toks []gopodMathTok
	for len(s) > 0 {
		r := []rune(s)[0]
		switch {
		case r == ' ':
			s = s[1:]
		case unicode.IsDigit(r) || (r == '.' && len(s) > 1 && s[1] >= '0' && s[1] <= '9'):
			m := gopodMathNumRe.FindStringSubmatch(s)
			if m == nil {
				s = s[len(string(r)):]
				continue
			}
			toks = append(toks, gopodMathTok{kind: mathTokNum, num: gopodMathNumFrom(m)})
			s = s[len(m[0]):]
			if strings.HasPrefix(s, "%") {
				t := &toks[len(toks)-1]
				t.num = gopodMathNum{r: new(big.Rat).Quo(t.num.r, big.NewRat(100, 1)), kind: "pct", show: t.num.show + "%"}
				s = s[1:]
			}
		case strings.ContainsRune("+−×÷", r):
			toks = append(toks, gopodMathTok{kind: mathTokOp, op: string(r)})
			s = s[len(string(r)):]
		case r == '-':
			toks = append(toks, gopodMathTok{kind: mathTokOp, op: gopodMathMinus})
			s = s[1:]
		case r == '/':
			toks = append(toks, gopodMathTok{kind: mathTokOp, op: "÷"})
			s = s[1:]
		case r == '(':
			toks = append(toks, gopodMathTok{kind: mathTokOpen})
			s = s[1:]
		case r == ')':
			toks = append(toks, gopodMathTok{kind: mathTokClose})
			s = s[1:]
		case r == '=':
			toks = append(toks, gopodMathTok{kind: mathTokWord, word: "="})
			s = s[1:]
		case unicode.IsLetter(r) || r == '\'':
			end := strings.IndexFunc(s, func(c rune) bool { return !(unicode.IsLetter(c) || c == '\'') })
			if end < 0 {
				end = len(s)
			}
			word := s[:end]
			if word == "of" {
				toks = append(toks, gopodMathTok{kind: mathTokOf})
			} else {
				toks = append(toks, gopodMathTok{kind: mathTokWord, word: word})
			}
			s = s[end:]
		default:
			toks = append(toks, gopodMathTok{kind: mathTokWord, word: string(r)})
			s = s[len(string(r)):]
		}
	}
	return toks
}

// gopodMathNumFrom: one number match -> its exact value and how it shows.
func gopodMathNumFrom(m []string) gopodMathNum {
	switch {
	case m[1] != "": // mixed 1 1/2
		w, _ := new(big.Int).SetString(m[1], 10)
		f := gopodMathFracRat(m[2], m[3])
		r := new(big.Rat).Add(new(big.Rat).SetInt(w), f)
		return gopodMathNum{r: r, kind: "frac", show: m[1] + " and " + m[2] + gopodMathSlash + m[3]}
	case m[4] != "":
		return gopodMathNum{r: gopodMathFracRat(m[4], m[5]), kind: "frac", show: m[4] + gopodMathSlash + m[5]}
	case m[6] != "" || m[8] != "":
		whole, digits := m[6], m[7]
		if m[8] != "" {
			whole, digits = "0", m[8]
		}
		fixed, rep := gopodMathSplitBars(digits)
		if rep != "" {
			return gopodMathNum{r: gopodMathParseRepeat(whole, fixed, rep), kind: "rep", show: whole + "." + fixed + gopodMathBarred(rep)}
		}
		r, _ := new(big.Rat).SetString(whole + "." + fixed)
		return gopodMathNum{r: r, kind: "dec", show: whole + "." + fixed}
	default:
		r, _ := new(big.Rat).SetString(m[9])
		return gopodMathNum{r: r, kind: "int", show: m[9]}
	}
}

func gopodMathFracRat(n, d string) *big.Rat {
	num, _ := new(big.Int).SetString(n, 10)
	den, _ := new(big.Int).SetString(d, 10)
	if den.Sign() == 0 {
		return nil
	}
	return new(big.Rat).SetFrac(num, den)
}

// gopodMathSplitBars: "16̅" -> fixed "1", rep "6". Barred digits must be the last
// ones (a bar in the middle is read as covering the digits from there on).
func gopodMathSplitBars(digits string) (fixed, rep string) {
	first := -1
	var plain strings.Builder
	n := 0
	runes := []rune(digits)
	for i := 0; i < len(runes); i++ {
		if runes[i] == []rune(gopodMathBar)[0] {
			continue
		}
		if i+1 < len(runes) && string(runes[i+1]) == gopodMathBar && first < 0 {
			first = n
		}
		plain.WriteRune(runes[i])
		n++
	}
	all := plain.String()
	if first < 0 {
		return all, ""
	}
	return all[:first], all[first:]
}

// gopodMathRuns: the stretches of numbers and operators in the line.
func gopodMathRuns(toks []gopodMathTok) [][]gopodMathTok {
	var runs [][]gopodMathTok
	var cur []gopodMathTok
	flush := func() {
		// no dangling operator or "of" at either end
		for len(cur) > 0 && (cur[0].kind == mathTokOp && cur[0].op != gopodMathMinus || cur[0].kind == mathTokOf || cur[0].kind == mathTokClose) {
			cur = cur[1:]
		}
		for len(cur) > 0 && (cur[len(cur)-1].kind == mathTokOp || cur[len(cur)-1].kind == mathTokOf || cur[len(cur)-1].kind == mathTokOpen) {
			cur = cur[:len(cur)-1]
		}
		if len(cur) > 0 {
			runs = append(runs, cur)
		}
		cur = nil
	}
	for _, t := range toks {
		if t.kind == mathTokWord {
			flush()
			continue
		}
		cur = append(cur, t)
	}
	flush()
	return runs
}

func gopodMathHasOp(run []gopodMathTok) bool {
	nums := 0
	ops := false
	for _, t := range run {
		if t.kind == mathTokNum {
			nums++
		}
		if t.kind == mathTokOp || t.kind == mathTokOf {
			ops = true
		}
	}
	return ops && nums >= 2
}

func gopodMathHasWord(toks []gopodMathTok, word string) bool {
	for _, t := range toks {
		if t.kind == mathTokWord && t.word == word {
			return true
		}
	}
	return false
}

// gopodMathTarget: "as a fraction" / "to a decimal" / "into a percent" /
// "in decimal form" - the target and where its phrase starts.
func gopodMathTarget(toks []gopodMathTok) (string, int) {
	targets := map[string]string{"fraction": "fraction", "fractions": "fraction", "decimal": "decimal", "decimals": "decimal", "percent": "percent", "percentage": "percent"}
	for i, t := range toks {
		if t.kind != mathTokWord {
			continue
		}
		target, ok := targets[t.word]
		if !ok {
			continue
		}
		// the words before it: as / to / into / in (/ a, an)
		j := i - 1
		for j >= 0 && toks[j].kind == mathTokWord && (toks[j].word == "a" || toks[j].word == "an" || toks[j].word == "its" || toks[j].word == "simplest") {
			j--
		}
		if j >= 0 && toks[j].kind == mathTokWord && (toks[j].word == "as" || toks[j].word == "to" || toks[j].word == "into" || toks[j].word == "in") {
			return target, j
		}
		// "what fraction is 0.25"
		if j >= 0 && toks[j].kind == mathTokWord && toks[j].word == "what" {
			return target, len(toks)
		}
	}
	return "", -1
}

// gopodMathRunBefore: the numbers right before the target phrase (or, for "what
// fraction is ...", the last stretch in the line).
func gopodMathRunBefore(toks []gopodMathTok, at int) []gopodMathTok {
	runs := gopodMathRuns(toks[:at])
	if len(runs) == 0 {
		return nil
	}
	return runs[len(runs)-1]
}

// ---------------------------------------------------------------- steps

// gopodMathLine: one step's line from math.lines, {key} filled in.
func gopodMathLine(name string, kv ...string) string {
	t := gopodBrobotsSettings.Math.Lines[name]
	if t == "" {
		return ""
	}
	return strings.NewReplacer(kv...).Replace(t)
}

type gopodMathSteps struct {
	lines   []string
	stopped bool // the sum can't be done (÷ 0); its line says why
}

func (s *gopodMathSteps) add(name string, kv ...string) {
	if l := gopodMathLine(name, kv...); l != "" {
		s.lines = append(s.lines, strings.ReplaceAll(l, "….", "…"))
	}
}

// gopodMathPlace: "tenths", "hundredths"... for n digits after the point.
func gopodMathPlace(n int) string {
	names := gopodBrobotsSettings.Math.PlaceNames
	if n >= 1 && n <= len(names) {
		return names[n-1]
	}
	return "1" + strings.Repeat("0", n) + "ths"
}

// gopodMathSimplify adds the simplify step for n/d (when there's one) and
// returns the simplest form.
func gopodMathSimplify(s *gopodMathSteps, n, d *big.Int) *big.Rat {
	r := new(big.Rat).SetFrac(n, d)
	g := new(big.Int).GCD(nil, nil, new(big.Int).Abs(n), new(big.Int).Abs(d))
	shown := gopodMathInt(n) + gopodMathSlash + d.String()
	if g.Cmp(big.NewInt(1)) > 0 {
		s.add("simplify", "{fraction}", shown, "{gcd}", g.String(), "{simple}", gopodMathFraction(r))
	}
	return r
}

// gopodMathToFraction: the steps from a number to its fraction.
func gopodMathToFraction(s *gopodMathSteps, n gopodMathNum) *big.Rat {
	switch n.kind {
	case "rep":
		// The digits as written (0.9̅ stays 0.9̅, though it is 1).
		parts := strings.SplitN(strings.TrimPrefix(n.show, gopodMathMinus), ".", 2)
		whole := parts[0]
		fixed, rep := gopodMathSplitBars(parts[1])
		a, b := len(fixed), len(rep)
		m1 := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(a)), nil)
		m2 := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(a+b)), nil)
		// x·10^a and x·10^(a+b), written from the digits so the bars stay put.
		v1 := strings.TrimLeft(whole+fixed, "0")
		if v1 == "" {
			v1 = "0"
		}
		v1 += "." + gopodMathBarred(rep)
		v2 := strings.TrimLeft(whole+fixed+rep, "0")
		if v2 == "" {
			v2 = "0"
		}
		v2 += "." + gopodMathBarred(rep)
		i1, _ := new(big.Int).SetString(whole+fixed, 10)
		i2, _ := new(big.Int).SetString(whole+fixed+rep, 10)
		if strings.HasPrefix(n.show, gopodMathMinus) {
			i1.Neg(i1)
			i2.Neg(i2)
			v1, v2 = gopodMathMinus+v1, gopodMathMinus+v2
		}
		k := new(big.Int).Sub(m2, m1)
		diff := new(big.Int).Sub(i2, i1)
		x := func(m *big.Int) string {
			if m.Cmp(big.NewInt(1)) == 0 {
				return "x"
			}
			return m.String() + "x"
		}
		s.add("repeat_let", "{decimal}", n.show)
		if a > 0 {
			s.add("repeat_multiply", "{by}", m1.String(), "{x}", x(m1), "{value}", v1)
		}
		s.add("repeat_multiply", "{by}", m2.String(), "{x}", x(m2), "{value}", v2)
		s.add("repeat_subtract", "{bigx}", x(m2), "{smallx}", x(m1), "{big}", v2, "{small}", v1, "{kx}", x(k), "{diff}", diff.String())
		s.add("repeat_divide", "{fraction}", diff.String()+gopodMathSlash+k.String())
		return gopodMathSimplify(s, diff, k)
	case "dec":
		parts := strings.SplitN(n.show, ".", 2)
		digits := parts[1]
		num, _ := new(big.Int).SetString(parts[0]+digits, 10)
		den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(digits))), nil)
		s.add("decimal_over", "{decimal}", n.show, "{place}", gopodMathPlace(len(digits)), "{fraction}", num.String()+gopodMathSlash+den.String())
		return gopodMathSimplify(s, num, den)
	case "pct":
		p := new(big.Rat).Mul(n.r, big.NewRat(100, 1))
		ps := strings.TrimSuffix(n.show, "%")
		if p.IsInt() {
			s.add("percent_over", "{percent}", n.show, "{fraction}", ps+gopodMathSlash+"100")
			return gopodMathSimplify(s, p.Num(), big.NewInt(100))
		}
		s.add("percent_divide", "{percent}", n.show, "{number}", ps, "{decimal}", gopodMathShow(n.r, "dec"))
		return gopodMathToFraction(s, gopodMathNum{r: n.r, kind: "dec", show: gopodMathShow(n.r, "dec")})
	case "frac":
		return gopodMathSimplify(s, n.r.Num(), n.r.Denom()) // n.r is already simplest
	}
	return n.r
}

// gopodMathToDecimal: the steps from a number to its decimal.
func gopodMathToDecimal(s *gopodMathSteps, n gopodMathNum) string {
	switch n.kind {
	case "frac":
		num, den := gopodMathFracParts(n)
		dec, exact := gopodMathDecimal(n.r)
		s.add("fraction_divide", "{fraction}", n.show, "{top}", num, "{bottom}", den)
		gopodMathLongDivision(s, n.r)
		_, _, _, rep, _ := gopodMathDecimalParts(n.r, 200)
		switch {
		case !exact:
			s.add("division_long", "{top}", num, "{bottom}", den, "{decimal}", dec)
		case rep != "":
			// 2026-10-15 (operator: "142857 should be said as numbers, not a
			// single number 142,857"): a longer block shows digit by digit.
			s.add("division_repeats", "{top}", num, "{bottom}", den, "{decimal}", dec, "{block}", strings.Join(strings.Split(rep, ""), " "))
		default:
			s.add("division_ends", "{top}", num, "{bottom}", den, "{decimal}", dec)
		}
		return dec
	case "pct":
		dec, _ := gopodMathDecimal(n.r)
		s.add("percent_divide", "{percent}", n.show, "{number}", strings.TrimSuffix(n.show, "%"), "{decimal}", dec)
		return dec
	}
	dec, _ := gopodMathDecimal(n.r)
	return dec
}

// gopodMathLongDivision (operator 2026-10-15: "tell me about yourself vs tell me
// all about yourself * this feels like it should fall into a long 'math'
// response"): "Explain the Math" shows the division itself, digit by digit -
// each bring-down, its digit and remainder - until it ends, a remainder comes
// back (the repeat), or math.long_division_max_steps.
func gopodMathLongDivision(s *gopodMathSteps, r *big.Rat) {
	max := gopodBrobotsSettings.Math.LongDivisionMaxSteps
	if max <= 0 {
		return
	}
	num := new(big.Int).Abs(r.Num())
	den := r.Denom()
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	s.add("long_first", "{top}", num.String(), "{bottom}", den.String(), "{times}", q.String(), "{rem}", rem.String())
	seen := map[string]bool{rem.String(): true}
	ten := big.NewInt(10)
	for step := 0; rem.Sign() != 0; step++ {
		if step >= max {
			s.add("long_more")
			return
		}
		down := new(big.Int).Mul(rem, ten)
		d, next := new(big.Int).QuoRem(down, den, new(big.Int))
		s.add("long_step", "{down}", down.String(), "{bottom}", den.String(), "{digit}", d.String(), "{rem}", next.String())
		rem = next
		if rem.Sign() == 0 {
			s.add("long_ends")
			return
		}
		if seen[rem.String()] {
			s.add("long_repeats", "{rem}", rem.String())
			return
		}
		seen[rem.String()] = true
	}
}

// gopodMathFracParts: the top and bottom as written (a mixed number as improper).
func gopodMathFracParts(n gopodMathNum) (string, string) {
	if p := strings.SplitN(n.show, gopodMathSlash, 2); len(p) == 2 && !strings.Contains(n.show, " and ") {
		return p[0], p[1]
	}
	return gopodMathInt(n.r.Num()), n.r.Denom().String()
}

// gopodMathConvert: one number (or a sum worked out first) to the target.
func gopodMathConvert(run []gopodMathTok, target string) ([]string, bool) {
	s := &gopodMathSteps{}
	var n gopodMathNum
	if len(run) == 1 && run[0].kind == mathTokNum {
		n = run[0].num
	} else {
		v, ok := gopodMathEvaluate(run, s)
		if !ok {
			return nil, false
		}
		n = v
	}
	if n.r == nil {
		return nil, false
	}
	switch target {
	case "fraction", "simplify":
		if n.kind == "int" {
			s.add("whole_fraction", "{number}", n.show)
			return s.lines, true
		}
		if n.kind == "frac" {
			if strings.Contains(n.show, " and ") {
				s.add("mixed_improper", "{mixed}", n.show, "{fraction}", gopodMathFraction(n.r))
				return s.lines, true
			}
			num, den := gopodMathFracParts(n)
			nb, _ := new(big.Int).SetString(strings.Replace(num, gopodMathMinus, "-", 1), 10)
			db, _ := new(big.Int).SetString(den, 10)
			r := gopodMathSimplify(s, nb, db)
			if len(s.lines) == 0 {
				s.add("already_simplest", "{fraction}", n.show)
				return s.lines, true
			}
			s.add("answer", "{from}", n.show, "{to}", gopodMathFraction(r))
			return s.lines, true
		}
		r := gopodMathToFraction(s, n)
		result := gopodMathFraction(r)
		if mixed := gopodMathMixed(r); mixed != "" {
			result += gopodMathLine("or_mixed", "{mixed}", mixed)
		}
		s.add("answer", "{from}", n.show, "{to}", result)
		return s.lines, true
	case "decimal":
		if n.kind == "int" || n.kind == "dec" || n.kind == "rep" {
			s.add("already_decimal", "{decimal}", n.show)
			return s.lines, true
		}
		dec := gopodMathToDecimal(s, n)
		s.add("answer", "{from}", n.show, "{to}", dec)
		return s.lines, true
	case "percent":
		if n.kind == "pct" {
			s.add("already_percent", "{percent}", n.show)
			return s.lines, true
		}
		dec := n.show
		if n.kind == "frac" {
			dec = gopodMathToDecimal(s, n)
		}
		p := new(big.Rat).Mul(n.r, big.NewRat(100, 1))
		pct := gopodMathShow(p, "dec") + "%"
		s.add("times_hundred", "{decimal}", dec, "{percent}", pct)
		s.add("answer", "{from}", n.show, "{to}", pct)
		return s.lines, true
	}
	return nil, false
}

var gopodMathWhatPercentRe = []*regexp.Regexp{
	regexp.MustCompile(`what (?:percent|percentage) is ([\d./]+) of ([\d./]+)`),
	regexp.MustCompile(`([\d./]+) is what (?:percent|percentage) of ([\d./]+)`),
}
var gopodMathWhatPercentOfRe = regexp.MustCompile(`what (?:percent|percentage) of ([\d./]+) is ([\d./]+)`)

// gopodMathWhatPercent: "what percent is 3 of 4" / "3 is what percent of 4" /
// "what percentage of 4 is 3".
func gopodMathWhatPercent(norm string) ([]string, bool) {
	var a, b string
	for _, re := range gopodMathWhatPercentRe {
		if m := re.FindStringSubmatch(norm); m != nil {
			a, b = m[1], m[2]
		}
	}
	if m := gopodMathWhatPercentOfRe.FindStringSubmatch(norm); m != nil {
		a, b = m[2], m[1]
	}
	if a == "" {
		return nil, false
	}
	ta, tb := gopodMathTokenize(a), gopodMathTokenize(b)
	if len(ta) != 1 || len(tb) != 1 || ta[0].kind != mathTokNum || tb[0].kind != mathTokNum || tb[0].num.r.Sign() == 0 {
		return nil, false
	}
	s := &gopodMathSteps{}
	q := new(big.Rat).Quo(ta[0].num.r, tb[0].num.r)
	dec, _ := gopodMathDecimal(q)
	pct := gopodMathShow(new(big.Rat).Mul(q, big.NewRat(100, 1)), "dec") + "%"
	s.add("what_percent", "{part}", ta[0].num.show, "{whole}", tb[0].num.show, "{decimal}", dec, "{percent}", pct)
	return s.lines, true
}

// ---------------------------------------------------------------- arithmetic

// gopodMathArithmetic: plus, minus, times, divide (and "of" after a percent or a
// fraction), in the order of operations, one line per step.
func gopodMathArithmetic(run []gopodMathTok) ([]string, bool) {
	s := &gopodMathSteps{}
	addSub, mulDiv, parens := false, false, false
	for _, t := range run {
		switch {
		case t.kind == mathTokOp && (t.op == "+" || t.op == gopodMathMinus):
			addSub = true
		case t.kind == mathTokOp || t.kind == mathTokOf:
			mulDiv = true
		case t.kind == mathTokOpen:
			parens = true
		}
	}
	if addSub && mulDiv && !parens {
		s.add("order_note")
	}
	v, ok := gopodMathEvaluate(run, s)
	if !ok {
		if s.stopped {
			return s.lines, true
		}
		return nil, false
	}
	// More than one step (or a top-heavy fraction): the whole sum and its
	// answer last.
	result := gopodMathShow(v.r, v.kind)
	mixed := ""
	if v.kind == "frac" {
		mixed = gopodMathMixed(v.r)
	}
	if len(s.lines) > 1 || mixed != "" {
		if mixed != "" {
			result += gopodMathLine("or_mixed", "{mixed}", mixed)
		}
		s.add("answer", "{from}", gopodMathShowRun(run), "{to}", result)
	}
	return s.lines, true
}

// gopodMathShowRun: the sum as it was asked ("15% of 80", "1⁄2 + 1⁄3").
func gopodMathShowRun(run []gopodMathTok) string {
	var parts []string
	for i, t := range run {
		switch t.kind {
		case mathTokNum:
			parts = append(parts, t.num.show)
		case mathTokOp:
			if t.op == gopodMathMinus && (i == 0 || run[i-1].kind == mathTokOp || run[i-1].kind == mathTokOpen) {
				parts = append(parts, gopodMathMinus)
				continue
			}
			parts = append(parts, t.op)
		case mathTokOf:
			parts = append(parts, "of")
		case mathTokOpen:
			parts = append(parts, "(")
		case mathTokClose:
			parts = append(parts, ")")
		}
	}
	s := strings.Join(parts, " ")
	return strings.NewReplacer("( ", "(", " )", ")", gopodMathMinus+" "+gopodMathMinus, gopodMathMinus+" "+gopodMathMinus).Replace(s)
}

// gopodMathEvaluate: a recursive-descent evaluator that writes each step.
func gopodMathEvaluate(run []gopodMathTok, s *gopodMathSteps) (gopodMathNum, bool) {
	p := &gopodMathParser{toks: run, s: s}
	v, ok := p.expr()
	if !ok || p.i != len(run) {
		return gopodMathNum{}, false
	}
	return v, true
}

type gopodMathParser struct {
	toks []gopodMathTok
	i    int
	s    *gopodMathSteps
	bad  bool
}

func (p *gopodMathParser) peek() *gopodMathTok {
	if p.i < len(p.toks) {
		return &p.toks[p.i]
	}
	return nil
}

func (p *gopodMathParser) expr() (gopodMathNum, bool) {
	a, ok := p.term()
	if !ok {
		return a, false
	}
	for t := p.peek(); t != nil && t.kind == mathTokOp && (t.op == "+" || t.op == gopodMathMinus); t = p.peek() {
		p.i++
		b, ok := p.term()
		if !ok {
			return b, false
		}
		if a, ok = gopodMathApply(p.s, a, t.op, b); !ok {
			return a, false
		}
	}
	return a, true
}

func (p *gopodMathParser) term() (gopodMathNum, bool) {
	a, ok := p.factor()
	if !ok {
		return a, false
	}
	for t := p.peek(); t != nil && (t.kind == mathTokOf || t.kind == mathTokOp && (t.op == "×" || t.op == "÷")); t = p.peek() {
		op := t.op
		if t.kind == mathTokOf {
			// "of" means times only after a percent or a fraction
			if a.kind != "pct" && a.kind != "frac" {
				return a, false
			}
			op = "of"
		}
		p.i++
		b, ok := p.factor()
		if !ok {
			return b, false
		}
		if a, ok = gopodMathApply(p.s, a, op, b); !ok {
			return a, false
		}
	}
	return a, true
}

func (p *gopodMathParser) factor() (gopodMathNum, bool) {
	t := p.peek()
	if t == nil {
		return gopodMathNum{}, false
	}
	switch {
	case t.kind == mathTokOp && t.op == gopodMathMinus:
		p.i++
		v, ok := p.factor()
		if !ok {
			return v, false
		}
		return gopodMathNum{r: new(big.Rat).Neg(v.r), kind: v.kind, show: gopodMathMinus + v.show}, true
	case t.kind == mathTokOpen:
		p.i++
		v, ok := p.expr()
		if !ok {
			return v, false
		}
		if c := p.peek(); c == nil || c.kind != mathTokClose {
			return v, false
		}
		p.i++
		return v, true
	case t.kind == mathTokNum:
		p.i++
		if t.num.r == nil {
			return t.num, false
		}
		return t.num, true
	}
	return gopodMathNum{}, false
}

// gopodMathApply: a op b, with its step line(s).
func gopodMathApply(s *gopodMathSteps, a gopodMathNum, op string, b gopodMathNum) (gopodMathNum, bool) {
	// A repeating decimal is worked as its fraction; a percent as its decimal.
	a = gopodMathOperand(s, a, op, b)
	b = gopodMathOperand(s, b, op, a)
	kind := "dec"
	if a.kind == "int" && b.kind == "int" {
		kind = "int"
	}
	if a.kind == "frac" || b.kind == "frac" {
		kind = "frac"
	}
	var r *big.Rat
	opShown := op
	switch op {
	case "+":
		r = new(big.Rat).Add(a.r, b.r)
	case gopodMathMinus:
		r = new(big.Rat).Sub(a.r, b.r)
	case "×", "of":
		r = new(big.Rat).Mul(a.r, b.r)
		opShown = "×"
	case "÷":
		if b.r.Sign() == 0 {
			s.add("divide_zero", "{number}", a.show)
			s.stopped = true
			return gopodMathNum{}, false
		}
		r = new(big.Rat).Quo(a.r, b.r)
	}
	// Fractions: the same bottom for plus and minus, top-by-top for times, the
	// flip for divide.
	if kind == "frac" {
		af, bf := gopodMathAsFraction(a), gopodMathAsFraction(b)
		switch op {
		case "+", gopodMathMinus:
			if !a.r.IsInt() && !b.r.IsInt() && a.r.Denom().Cmp(b.r.Denom()) != 0 {
				lcd := gopodMathLCM(a.r.Denom(), b.r.Denom())
				s.add("common_bottom", "{bottom}", lcd.String(), "{a}", af, "{a2}", gopodMathOverBottom(a.r, lcd), "{b}", bf, "{b2}", gopodMathOverBottom(b.r, lcd))
				s.add("op", "{a}", gopodMathOverBottom(a.r, lcd), "{op}", opShown, "{b}", gopodMathOverBottom(b.r, lcd), "{result}", gopodMathOverBottom(r, lcd))
				return gopodMathFinish(s, r, lcd, kind), true
			}
		case "×", "of":
			if !a.r.IsInt() || !b.r.IsInt() {
				tn := new(big.Int).Mul(a.r.Num(), b.r.Num())
				td := new(big.Int).Mul(a.r.Denom(), b.r.Denom())
				s.add("times_fractions", "{a}", af, "{b}", bf, "{result}", gopodMathInt(tn)+gopodMathSlash+td.String())
				return gopodMathFinishRaw(s, tn, td), true
			}
		case "÷":
			flip := new(big.Rat).Inv(b.r)
			tn := new(big.Int).Mul(a.r.Num(), flip.Num())
			td := new(big.Int).Mul(a.r.Denom(), flip.Denom())
			s.add("divide_fractions", "{a}", af, "{b}", bf, "{flip}", gopodMathFraction(flip), "{result}", gopodMathInt(tn)+gopodMathSlash+td.String())
			return gopodMathFinishRaw(s, tn, td), true
		}
	}
	result := gopodMathShow(r, kind)
	// Whole numbers that don't divide evenly: the decimal, and the fraction too.
	if op == "÷" && kind != "frac" && !r.IsInt() {
		if _, _, _, rep, _ := gopodMathDecimalParts(r, 200); rep != "" || kind == "int" {
			extra := gopodMathMixed(r)
			if extra == "" {
				extra = gopodMathFraction(r)
			}
			result += gopodMathLine("or_mixed", "{mixed}", extra)
			kind = "frac"
		}
	}
	s.add("op", "{a}", a.show, "{op}", opShown, "{b}", b.show, "{result}", result)
	return gopodMathNum{r: r, kind: kind, show: gopodMathShow(r, kind)}, true
}

// gopodMathOperand: a repeating decimal becomes its fraction, a percent its
// decimal (or its fraction, next to a fraction), each with its line.
func gopodMathOperand(s *gopodMathSteps, n gopodMathNum, op string, other gopodMathNum) gopodMathNum {
	if n.kind == "frac" && strings.Contains(n.show, " and ") {
		f := gopodMathFraction(n.r)
		s.add("mixed_improper", "{mixed}", n.show, "{fraction}", f)
		return gopodMathNum{r: n.r, kind: "frac", show: f}
	}
	switch n.kind {
	case "rep":
		f := gopodMathFraction(n.r)
		s.add("repeat_is", "{decimal}", n.show, "{fraction}", f)
		return gopodMathNum{r: n.r, kind: "frac", show: f}
	case "pct":
		if other.kind == "frac" || other.kind == "rep" {
			f := gopodMathFraction(n.r)
			s.add("percent_is", "{percent}", n.show, "{value}", f)
			return gopodMathNum{r: n.r, kind: "frac", show: f}
		}
		d, _ := gopodMathDecimal(n.r)
		s.add("percent_is", "{percent}", n.show, "{value}", d)
		return gopodMathNum{r: n.r, kind: "dec", show: d}
	}
	return n
}

func gopodMathAsFraction(n gopodMathNum) string {
	if n.r.IsInt() {
		return gopodMathInt(n.r.Num())
	}
	return gopodMathFraction(n.r)
}

func gopodMathLCM(a, b *big.Int) *big.Int {
	g := new(big.Int).GCD(nil, nil, a, b)
	return new(big.Int).Mul(new(big.Int).Quo(a, g), b)
}

// gopodMathOverBottom: r written over the bottom d (1⁄2 over 6 = 3⁄6).
func gopodMathOverBottom(r *big.Rat, d *big.Int) string {
	n := new(big.Int).Mul(r.Num(), new(big.Int).Quo(d, r.Denom()))
	return gopodMathInt(n) + gopodMathSlash + d.String()
}

// gopodMathFinish: r (written over d) simplified, as a fraction result.
func gopodMathFinish(s *gopodMathSteps, r *big.Rat, d *big.Int, kind string) gopodMathNum {
	n := new(big.Int).Mul(r.Num(), new(big.Int).Quo(d, r.Denom()))
	return gopodMathFinishRaw(s, n, d)
}

func gopodMathFinishRaw(s *gopodMathSteps, n, d *big.Int) gopodMathNum {
	r := gopodMathSimplify(s, n, d)
	return gopodMathNum{r: r, kind: "frac", show: gopodMathFraction(r)}
}

// ## EOF
