// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_alias.go
package wirepod_ttr

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// gopodAliasCard: one alias studio card - its title, its commandments (alias -> what it does) and
// the aliases it marks for the chat.
type gopodAliasCard struct {
	Title    string
	Aliases  [][2]string
	FromChat map[string]bool
}

var gopodAliasArg = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// gopodAliasCards reads every alias card in alias_play_studio/*/README.md, live - a new card or
// alias needs no code.
func gopodAliasCards() []gopodAliasCard {
	files, _ := filepath.Glob(filepath.Join(gopodRepoRoot(), "alias_play_studio", "*", "README.md"))
	sort.Strings(files)
	var cards []gopodAliasCard
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		c := gopodAliasCard{FromChat: map[string]bool{}}
		section := ""
		sc := bufio.NewScanner(fh)
		for sc.Scan() {
			t := strings.TrimSpace(sc.Text())
			switch {
			case strings.HasPrefix(t, "# ") && c.Title == "":
				c.Title = strings.TrimPrefix(t, "# ")
			case strings.HasPrefix(t, "## "):
				section = strings.ToLower(strings.TrimPrefix(t, "## "))
			case len(t) > 2 && t[0] >= '0' && t[0] <= '9' && strings.Contains(t, ". "):
				item := strings.TrimSpace(t[strings.Index(t, ". ")+2:])
				name, about, _ := strings.Cut(item, ", ")
				name = strings.TrimSpace(name)
				if section == "commandments" {
					c.Aliases = append(c.Aliases, [2]string{name, strings.TrimSpace(about)})
				} else if section == "from the chat" {
					c.FromChat[name] = true
				}
			}
		}
		fh.Close()
		if len(c.Aliases) > 0 {
			cards = append(cards, c)
		}
	}
	return cards
}

// gopodSystemAlias: "system-alias" lists the aliases from their cards; "system-alias <name>
// [args]" runs one its card marks for the chat - the terminal's own function, one code path.
func gopodSystemAlias(text string) string {
	words := strings.Fields(strings.TrimSpace(text))
	cards := gopodAliasCards()
	if len(words) < 2 {
		out := []string{gopodSystemText("system_alias.list_title")}
		for _, c := range cards {
			out = append(out, c.Title+":")
			for _, a := range c.Aliases {
				mark := ""
				if c.FromChat[a[0]] {
					mark = " " + gopodSystemText("system_alias.chat_mark")
				}
				out = append(out, "  "+a[0]+"  "+a[1]+mark)
			}
		}
		return gopodJoinLines(out)
	}
	name, args := words[1], words[2:]
	var card *gopodAliasCard
	for i := range cards {
		for _, a := range cards[i].Aliases {
			if a[0] == name {
				card = &cards[i]
			}
		}
	}
	if card == nil {
		return gopodSystemText("system_alias.unknown", "name", name)
	}
	if !card.FromChat[name] {
		return gopodSystemText("system_alias.terminal_only", "name", name, "card", card.Title)
	}
	if gopodRunActive() {
		return gopodSystemText("system_alias.run_going", "name", name)
	}
	for _, a := range args {
		if !gopodAliasArg.MatchString(a) {
			return gopodSystemText("system_alias.bad_arg", "arg", a)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	lib := gopodHomePath("~/.gopod_alias_lib/brobots.sh")
	cmd := exec.CommandContext(ctx, "bash", append([]string{"-c", `source "$0" >/dev/null 2>&1; "$@"`, lib, name}, args...)...)
	out, err := gopodAsOperator(cmd).CombinedOutput()
	answer := strings.TrimSpace(string(out))
	if err != nil && answer == "" {
		answer = err.Error()
	}
	return gopodSystemText("system_alias.ran", "name", strings.TrimSpace(name+" "+strings.Join(args, " "))) + "\n" + answer
}

// ## EOF
