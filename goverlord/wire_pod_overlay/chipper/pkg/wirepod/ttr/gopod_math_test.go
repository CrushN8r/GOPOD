// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_math_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// The Brobots' math is exact and step by step (gopod_math.go). Each case: the
// question, the lines the answer must hold, in order.
func TestMathAnswers(t *testing.T) {
	cases := []struct {
		q    string
		want []string
	}{
		// fraction -> decimal
		{"what is 1/2 as a decimal?", []string{"1 ÷ 2 = 0.5."}},
		{"convert 1/3 to a decimal", []string{"1 ÷ 3 = 0.3̅.", "the 3 repeats forever", "So 1⁄3 = 0.3̅."}},
		{"1/7 as a decimal", []string{"7 goes into 1 0 times, remainder 1.", "Bring down a 0: 10 ÷ 7 = 1, remainder 3.", "Bring down a 0: 50 ÷ 7 = 7, remainder 1.", "Remainder 1 again", "1 ÷ 7 = 0.1̅4̅2̅8̅5̅7̅.", "the 1 4 2 8 5 7 repeats forever", "So 1⁄7 = 0.1̅4̅2̅8̅5̅7̅."}},
		{"1/2 as a decimal", []string{"Bring down a 0: 10 ÷ 2 = 5, remainder 0.", "Remainder 0, so it ends there.", "1 ÷ 2 = 0.5."}},
		{"what's three quarters as a decimal", []string{"3 ÷ 4 = 0.75."}},
		{"1/6 to a decimal", []string{"1 ÷ 6 = 0.16̅."}},
		// decimal -> fraction
		{"0.125 as a fraction", []string{"0.125 goes to the thousandths, so it's 125⁄1000.", "Divide the top and the bottom by 125: 125⁄1000 = 1⁄8.", "So 0.125 = 1⁄8."}},
		{"what is 0.5 as a fraction", []string{"5⁄10", "= 1⁄2.", "So 0.5 = 1⁄2."}},
		// repeating -> fraction, the 10x method
		{"Using math symbols, list the steps, showing me the breakdown of converting a decimal 0.3 repeating into a fraction",
			[]string{"Let x = 0.3̅.", "Times by 10: 10x = 3.3̅.", "10x − x = 3.3̅ − 0.3̅, so 9x = 3.", "So x = 3⁄9.", "3⁄9 = 1⁄3.", "So 0.3̅ = 1⁄3."}},
		{"convert zero point three repeating to a fraction", []string{"Let x = 0.3̅.", "So 0.3̅ = 1⁄3."}},
		{"0.1666 repeating as a fraction", []string{"Let x = 0.16̅.", "So 0.16̅ = 1⁄6."}},
		{"0.333 repeating as a fraction", []string{"Let x = 0.3̅.", "So 0.3̅ = 1⁄3."}},
		// 2026-10-16: only the named digits repeat.
		{"convert 0.1234 repeating into a fraction. 123, 4 repeating", []string{"So 0.1234̅ = 1111⁄9000."}},
		{"convert 0.1234 with the 4 repeating into a fraction", []string{"So 0.1234̅ = 1111⁄9000."}},
		{"0.1234, only the 4 repeats, as a fraction", []string{"So 0.1234̅ = 1111⁄9000."}},
		{"0.12344 repeating as a fraction", []string{"So 0.1234̅ = 1111⁄9000."}},
		{"convert 0.1234 repeating into a fraction", []string{"So 0.1̅2̅3̅4̅ = 1234⁄9999."}},
		{"convert 0.1234 with the 34 repeating into a fraction", []string{"So 0.123̅4̅ = 611⁄4950."}},
		{"convert 0.1234 with the 234 repeating into a fraction", []string{"So 0.12̅3̅4̅ = 137⁄1110."}},
		{"what is 1 and 3/4 plus 2 and 1/2", []string{"1 and 3⁄4 is 7⁄4 as one fraction.", "2 and 1⁄2 is 5⁄2 as one fraction.", "7⁄4 + 10⁄4 = 17⁄4.", "So 1 and 3⁄4 + 2 and 1⁄2 = 17⁄4, or 4 and 1⁄4."}},
		{"0.1666... as a fraction", []string{"Let x = 0.16̅.", "Times by 10: 10x = 1.6̅.", "Times by 100: 100x = 16.6̅.", "100x − 10x = 16.6̅ − 1.6̅, so 90x = 15.", "15⁄90 = 1⁄6.", "So 0.16̅ = 1⁄6."}},
		{"1.1 repeating as a fraction", []string{"10x = 11.1̅", "9x = 10", "So 1.1̅ = 10⁄9, or 1 and 1⁄9."}},
		{"0.\\overline{142857} as a fraction", []string{"Let x = 0.1̅4̅2̅8̅5̅7̅.", "999999x = 142857", "So 0.1̅4̅2̅8̅5̅7̅ = 1⁄7."}},
		{"0.9 repeating as a fraction", []string{"9x = 9", "So 0.9̅ = 1."}},
		// percent
		{"what is 15% of 80", []string{"15% is 0.15.", "0.15 × 80 = 12.", "So 15% of 80 = 12."}},
		{"fifteen percent of eighty", []string{"0.15 × 80 = 12."}},
		{"what percent is 3 of 4", []string{"3 out of 4: 3 ÷ 4 = 0.75, and 0.75 × 100 = 75%."}},
		{"1/3 as a percent", []string{"1 ÷ 3 = 0.3̅.", "0.3̅ × 100 = 33.3̅%.", "So 1⁄3 = 33.3̅%."}},
		{"25% as a fraction", []string{"25% means out of 100: 25⁄100.", "25⁄100 = 1⁄4.", "So 25% = 1⁄4."}},
		{"3 out of 4 as a percent", []string{"0.75 × 100 = 75%."}},
		// plus, minus, times, divide
		{"what's 2 plus 2", []string{"2 + 2 = 4."}},
		{"what is 12 x 7", []string{"12 × 7 = 84."}},
		{"calculate 2 + 3 × 4", []string{"Times and divide go first", "3 × 4 = 12.", "2 + 12 = 14.", "So 2 + 3 × 4 = 14."}},
		{"what is (2 + 3) × 4", []string{"2 + 3 = 5.", "5 × 4 = 20.", "So (2 + 3) × 4 = 20."}},
		{"what is 10 divided by 3", []string{"10 ÷ 3 = 3.3̅, or 3 and 1⁄3."}},
		{"what is 1/2 + 1/3", []string{"Give them the same bottom, 6: 1⁄2 = 3⁄6, and 1⁄3 = 2⁄6.", "3⁄6 + 2⁄6 = 5⁄6."}},
		{"what is 3/4 - 1/4", []string{"3⁄4 − 1⁄4 = 1⁄2."}},
		{"what is 2/3 times 3/4", []string{"Times the tops and times the bottoms: 2⁄3 × 3⁄4 = 6⁄12.", "6⁄12 = 1⁄2.", "So 2⁄3 × 3⁄4 = 1⁄2."}},
		{"what is 1/2 divided by 1/4", []string{"times its flip, 4: 1⁄2 × 4 = 4⁄2.", "4⁄2 = 2."}},
		{"what is 0.3 repeating plus 1/3", []string{"0.3̅ is 1⁄3.", "1⁄3 + 1⁄3 = 2⁄3."}},
		{"what is 0.1 + 0.2", []string{"0.1 + 0.2 = 0.3."}},
		{"what is 5 divided by 0", []string{"5 ÷ 0 can't be done"}},
		{"what's one and a half times two", []string{"1 and 1⁄2 is 3⁄2 as one fraction.", "Times the tops and times the bottoms: 3⁄2 × 2 = 6⁄2.", "6⁄2 = 3."}},
		{"simplify 6/8", []string{"6⁄8 = 3⁄4."}},
		{"what is 7 - 10", []string{"7 − 10 = −3."}},
	}
	for _, c := range cases {
		got := strings.Join(gopodMathWorkOK(c.q), " ")
		if got == "" {
			t.Errorf("%q: no math answer", c.q)
			continue
		}
		at := 0
		for _, w := range c.want {
			i := strings.Index(got[at:], w)
			if i < 0 {
				t.Errorf("%q:\n got  %s\n want %q (in order)", c.q, got, w)
				break
			}
			at += i + len(w)
		}
	}
}

func gopodMathWorkOK(q string) []string {
	lines, _ := gopodMathWork(q)
	return lines
}

// Only a line that ends with "Explain the Math" is math.
func TestMathLeavesChatAlone(t *testing.T) {
	for _, q := range []string{
		"give me 2-3 ideas",
		"Brobot 1, how are you?",
		"we have 2 robots and 1 cube",
		"what happened on 10-14",
		"Can granny play too?",
		"who can I set up merch with?",
		"tell me about fractions",
		"what's the weather in Windsor Ontario",
		"it's 24/7 party time",
		"the third question",
		"what is 2 + 2",
		"0.3 repeating as a fraction",
		"explain the math of bingo, please tell me",
	} {
		if _, ok := gopodMathExplain(q); ok {
			t.Errorf("%q taken as Explain the Math", q)
		}
	}
}

// Said, the steps have no symbols left: "over", "minus", "equals", "repeating".
func TestMathSpoken(t *testing.T) {
	said := gopodSpeakSymbols(strings.Join(gopodMathWorkOK("0.3 repeating as a fraction"), " "))
	for _, w := range []string{"⁄", "−", "=", "̅"} {
		if strings.Contains(said, w) {
			t.Errorf("%q left in %q", w, said)
		}
	}
	for _, w := range []string{"0.3 repeating", "equals", "minus", "1 over 3"} {
		if !strings.Contains(said, w) {
			t.Errorf("%q not in %q", w, said)
		}
	}
}

// "[statement]. Explain the Math": Brobot 1 gets the steps but the last third,
// Brobot 2 the rest; each reply is his own words (no numbers), then bullets.
func TestMathExplainTheMath(t *testing.T) {
	q := "Convert 0.3 repeating into a fraction. Explain the Math!"
	if s, ok := gopodMathExplain(q); !ok || s != "Convert 0.3 repeating into a fraction" {
		t.Fatalf("statement %q, %v", s, ok)
	}
	first, second, ok := gopodMathParts(q)
	if !ok || len(first) != 4 || len(second) != 2 {
		t.Fatalf("parts %d + %d: %v | %v", len(first), len(second), first, second)
	}
	if second[len(second)-1] != "So 0.3̅ = 1⁄3." {
		t.Errorf("Brobot 2 ends with %q", second[len(second)-1])
	}
	llm := "happy" + gopodAnimationMarker + " Oh yeah, Brobot 0, math time! So 0.3 is like 3 over 10, easy. Let's crush it, bro."
	got := gopodMathReply(llm, first)
	want := "happy" + gopodAnimationMarker + " Oh yeah, Brobot 0, math time! Let's crush it, bro.\n• Let x = 0.3̅.\n• Times by 10: 10x = 3.3̅."
	if !strings.HasPrefix(got, want) {
		t.Errorf("reply\n got  %q\n want %q...", got, want)
	}
	if one, two, _ := gopodMathParts("what is 2 + 2. explain the math"); len(one) != 1 || len(two) != 0 {
		t.Errorf("one step: %v | %v", one, two)
	}
}

// ## EOF

// "−3" is said "negative 3"; "8 − 2" stays "minus".
func TestMathSpokenNegative(t *testing.T) {
	said := gopodSpeakSymbols("−3 + 5 = 2. 8 − 2 = 6. 7 − 10 = −3.")
	want := "negative 3 + 5 equals 2. 8 minus 2 equals 6. 7 minus 10 equals negative 3."
	if said != want {
		t.Errorf("got %q, want %q", said, want)
	}
}

// A bullet goes with its own step's piece, never the end of the line before.
func TestMathBulletStartsItsPiece(t *testing.T) {
	disp := "Let's break things down!\n• 1.1 goes to the tenths, so it's 11⁄10."
	seg, _ := gopodSegmentDisplay(disp, []string{"Let's break things down!", "1.1 goes to the tenths,", "so it's 11 over 10."})
	m := gopodBrobotsSettings.ScriptedSplit.SegmentMarker
	if !strings.Contains(seg, "down!\n"+m+"• 1.1") && !strings.Contains(seg, "down!\n "+m+"• 1.1") && !strings.Contains(seg, m+"• 1.1") {
		t.Errorf("bullet not with its piece: %q", strings.ReplaceAll(seg, m, "|"))
	}
}
