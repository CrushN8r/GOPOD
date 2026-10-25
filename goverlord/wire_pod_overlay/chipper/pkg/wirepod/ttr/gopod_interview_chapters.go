// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_interview_chapters.go
package wirepod_ttr

import (
	"sort"
	"strings"
	"sync"
	"unicode"
)

// The chat as a long interview video (2026-10-06, operator: "think of it like a
// long interview video on youtube, segmented with chapters ... or a video
// playlist of topics, then within each segment, ordered value points are
// revealed"). The topics are grouped into chapters (interview_chapters in the
// settings file). When a question's topic is in a different chapter than the
// last one shown, the chapter's title goes to the chat page as its own line
// (interview_chapters.role), before the reply. Within a topic, the value points
// come out in order (each pack's "reveal these, in order"); one point leads to
// the next topic ("Hey, that reminds me..." in scripted_snippets).
//
// Topic pictures (2026-10-06, operator: "Pictures in the chat"): a topic with
// pictures (topic_pictures) shows them once per Wire-Pod start, after the reply.

var gopodInterviewShown = struct {
	sync.Mutex
	chapter  string
	pictures map[string]int
	revealed []string // 2026-10-25: each egg's title, in the order found (gopodEggsRevealed)
}{pictures: map[string]int{}}

// gopodChapterOf returns the title of the chapter a topic is in, or "".
func gopodChapterOf(topic string) string {
	for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
		for _, t := range c.Topics {
			if t == topic {
				return c.Title
			}
		}
	}
	return ""
}

// gopodShowChapter puts the chapter title of the question's first topic in the
// chat, when it's a new chapter, while a draft plays. Both Brobots answering one question show it once.
func gopodShowChapter(userMessage string) {
	role := gopodBrobotsSettings.InterviewChapters.Role
	// 2026-10-09 (operator: "disabled until 'start draft' * disabled again upon
	// draft run complete"): chapter titles show only while a draft plays.
	if role == "" || !gopodScriptPlaying.Load() {
		return
	}
	gopodScript.Lock()
	own, ownTitle := gopodScript.own, gopodScript.chapter
	gopodScript.Unlock()
	packs := []loadedScenarioPacket{{}}
	if !own {
		packs = loadScenarioPacketsForMessage(userMessage)
	}
	for _, p := range packs {
		title := gopodChapterOf(p.Name)
		if own {
			// 2026-10-12: a song that names its chapters shows its own.
			title = ownTitle
		}
		if title == "" {
			continue
		}
		gopodInterviewShown.Lock()
		changed := title != gopodInterviewShown.chapter
		gopodInterviewShown.chapter = title
		gopodInterviewShown.Unlock()
		if changed {
			// 2026-10-23 (operator: "Chapter: What GOPOD is - no generating +
			// hourglass * ANY generated voice file goes through the same template"):
			// the chapter is Goverlord's voiced line, shown and said the one way -
			// Generating and the hourglass while its voice is made, then its words
			// lit as it's said (gopodSayShown).
			if sr := gopodBrobotsSettings.SignOff.SystemRole; sr != "" {
				gopodReporterSay(sr, gopodChapterSaid(title), true, false)
			} else {
				GopodChatFeedAdd(role, title)
			}
		}
		return
	}
}

// gopodShowTopicPictures puts the pictures of the question's topics in the chat,
// each topic's once per Wire-Pod start. The page gets the files' paths, one
// line, separated by "|".
//
// 2026-10-14 (operator: "can you combine a brobot's response bubble with a
// pop-up wrapper? ... like more of a surprise prize ... make the next brobot
// response bubble with the pop-up surprise announcement, then thinking please
// wait, then the llm [response]"): a pop-up names the Brobot who answers
// ("popup:Brobot 2::<pictures>"), so the page wraps his reply bubble with it,
// and Goverlord's congratulations is heard before the reply goes on.
func gopodShowTopicPictures(esn, userMessage string) {
	who := gopodBrobotName(esn)
	// 2026-10-16: an "Explain the Math" line is answered by Brobot 1 first, so
	// its eggs are his (it used to skip every other egg - the operator's "is it
	// granny weather for bingo? Explain the Math" showed none).
	_, mathLine := gopodMathExplain(userMessage)
	if _, _, ok := gopodMathParts(userMessage); ok || mathLine {
		who = "Brobot 1"
	}
	tp := gopodBrobotsSettings.TopicPictures
	if tp.Role == "" {
		return
	}
	// 2026-10-12: a song that names its chapters shows only its own pictures,
	// each time its question is asked.
	gopodScript.Lock()
	own, ownPictures := gopodScript.own && gopodScriptPlaying.Load(), gopodScript.pictures
	gopodScript.pictures = nil
	gopodScript.Unlock()
	// 2026-10-14 (operator: "treat these as found easter eggs throughout a live
	// chat * ensure live pop-ups are not triggered by a song run"): a pop-up
	// only in the live chat - no draft playing and no run going (a song, the
	// interview, a replay). Locked by TestPopupsOnlyLive.
	popup := !gopodScriptPlaying.Load() && !gopodRunActive()
	if own {
		var paths []string
		for _, f := range ownPictures {
			paths = append(paths, "/"+tp.Folder+"/"+f)
		}
		if len(paths) > 0 {
			GopodChatFeedAdd(tp.Role, strings.Join(paths, "|"))
		}
		return
	}
	// GOLDEN, the pinnacle egg process (operator 2026-10-16: "Mentioned easter
	// eggs * display in the order mentioned * in their newest bubbles * then
	// brobot response * this is the pinnacle process * \"Congratulations, Master
	// CrushN8r! You found the easter egg, [easter egg name/trigger]\""): every egg
	// the line finds - a topic's pictures, the weather, Explain the Math - shows in
	// the order its word comes in the line, each its own egg bubble, each with
	// Goverlord's topic_pictures.egg_text for it ({name}: the word that found it),
	// heard before the next; the reply comes after them all.
	lower := strings.ToLower(userMessage)
	type gopodEgg struct {
		at    int
		show  func()
		files []string
	}
	var eggs []gopodEgg
	// 2026-10-17 (operator: "easter eggs are live run chapters * each easter egg
	// should have a set title (chapter title) = [easter egg title] found * topics
	// in chapters are the easter egg triggers * then the value points the
	// brobots speak"; "Rule: If no image for easter egg, skip image & mouse over
	// message"): the topics the line found are grouped by their chapter - one
	// egg per chapter, titled by it, with those topics' pictures (none: no
	// pictures, no zoom line). A topic in no chapter (_always: the roast battle)
	// is its own egg when it has a prize.
	type chapterEgg struct {
		key, title string
		at         int
		paths      []string
		files      []string
		card       string
	}
	var found []*chapterEgg
	byKey := map[string]*chapterEgg{}
	for _, p := range loadScenarioPacketsForMessage(userMessage) {
		files := tp.Pictures[p.Name]
		card := strings.TrimSpace(tp.Cards[p.Name])
		var paths []string
		for _, f := range files {
			// 2026-10-20 (operator: "cam/camera feed(s) = easter egg"): a path
			// starting with "/" (a live camera stream) is used as it is.
			if strings.HasPrefix(f, "/") {
				paths = append(paths, f)
				continue
			}
			paths = append(paths, "/"+tp.Folder+"/"+f)
		}
		if !popup {
			if len(paths) > 0 {
				GopodChatFeedAdd(tp.Role, strings.Join(paths, "|"))
			}
			continue
		}
		at, word := gopodEggWord(lower, gopodTopicTriggers(p.Name))
		key, title := gopodTopicChapter(p.Name)
		if key == "" {
			if len(files) == 0 && card == "" {
				continue
			}
			key, title = p.Name, word
			if n := strings.TrimSpace(tp.EggNames[p.Name]); n != "" {
				title = n
			}
		}
		if gopodNotEgg(key) {
			continue
		}
		e := byKey[key]
		if e == nil {
			e = &chapterEgg{key: key, title: title, at: at}
			byKey[key] = e
			found = append(found, e)
		}
		if at >= 0 && (e.at < 0 || at < e.at) {
			e.at = at
		}
		e.paths = append(e.paths, paths...)
		for _, f := range files {
			if !strings.HasPrefix(f, "/") { // a live stream isn't looked at
				e.files = append(e.files, f)
			}
		}
		if card != "" {
			e.card = card
		}
	}
	for _, e := range found {
		e := e
		line := strings.Join(e.paths, "|")
		// 2026-10-16: a card prize (the roast battle) - its words in the egg.
		if e.card != "" {
			line = strings.Trim(line+"|"+gopodEggCardMark+strings.ReplaceAll(e.card, "|", " "), "|")
		}
		eggs = append(eggs, gopodEgg{at: e.at, files: e.files, show: func() {
			// 2026-10-21 (operator: "upon 1st reveal, easter egg bubble outside
			// brobots bubble + summary inside brobots window * upon more same egg
			// reveals, summary in brobots bubble only"): the egg's summary goes in
			// the answering Brobot's bubble every time it's found.
			defer gopodEggSummaryPost(who, e.key, e.title, e.card)
			// 2026-10-13: up to topic_pictures.max_per_session times per Wire-Pod start.
			if !gopodPopupAllowed(e.key) {
				return
			}
			gopodPictureDescribeStart(e.files)
			// 2026-10-18 (operator: "CHALK (display) now is to take ownership of
			// ... hidden/surprise egg cards"): a surprise egg on CHALK's board.
			eggWho := who
			for _, n := range tp.NotListed {
				if n == e.key && gopodBrobotsSettings.Managers.DisplayHost != "" {
					eggWho = gopodBrobotsSettings.Managers.DisplayHost
				}
			}
			gopodShowEgg(eggWho, e.title, line)
		}})
	}
	// 2026-10-17 (operator: "no longer easter eggs: Weather, Get Photo"): the
	// weather isn't the Brobots' egg; 2026-10-18 CHALK takes it (gopodChalkWeather):
	// "I'll help you with that", the weather egg on a chalkboard, the weather line.
	// (gopod_managers.go: PLAYHEAD takes the weather and the time and place.)
	// 2026-10-22 (operator's Brobot Bubble template, notice line 3): a manager
	// takes his line from the Brobot's notice instead, right after his "Got it!"
	// and value point line (gopodManagerStepIn) - not here, before or after him.
	// 2026-10-14: the Explain the Math egg (gopodMathPopup), with the others.
	if popup && mathLine {
		at, _ := gopodEggWord(lower, []string{strings.ToLower(gopodBrobotsSettings.Math.Trigger)})
		eggs = append(eggs, gopodEgg{at: at, show: func() { gopodMathPopup(gopodBrobot1ESN, userMessage) }})
	}
	sort.SliceStable(eggs, func(i, j int) bool { return eggs[i].at < eggs[j].at })
	for _, e := range eggs {
		e.show()
	}
}

// gopodBubbleNoticeMark: the bubble notice template (2026-10-21, operator:
// "Same as message inside a brobot chat bubble. Template it") - a small titled
// notice at the top of a Brobot's bubble ("notice:<title>|<line>"): an egg's
// summary, a native Wire-Pod intent.
const gopodBubbleNoticeMark = "notice:"

// gopodBubbleNotice posts a notice for brobot's bubble.
func gopodBubbleNotice(brobot, title, line string) {
	GopodChatFeedAddTo(gopodBrobotsSettings.TopicPictures.Role, brobot,
		gopodBubbleNoticeMark+strings.ReplaceAll(strings.TrimSpace(title), "|", " ")+"|"+strings.ReplaceAll(strings.TrimSpace(line), "|", " "))
}

// gopodEggSummaryPost: the egg's summary to the Brobot answering - a chapter's
// summary (interview_chapters), a card egg's card, else its topic's first say line.
func gopodEggSummaryPost(who, key, title, card string) {
	if who != "Brobot 1" && who != "Brobot 2" {
		return
	}
	summary := ""
	if strings.HasPrefix(key, "chapter:") {
		for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
			if "chapter:"+c.Title == key {
				summary = strings.TrimSpace(c.Summary)
			}
		}
	} else if strings.TrimSpace(card) != "" {
		summary = strings.TrimSpace(card)
	} else {
		_, always, _ := gopodReadChaptersFolder()
		for _, t := range always {
			if t.Name == key {
				summary = strings.TrimSpace(t.SayLines[1])
			}
		}
	}
	if summary == "" {
		return
	}
	// 2026-10-22: a line of his one Value Points notice (gopodNoticeEgg).
	if gopodNoticeEgg(who, title, summary) {
		return
	}
	gopodNoticeShow(who, gopodBrobotsSettings.BubbleNotices.Egg, map[string]string{"title": title, "summary": summary})
}

// gopodShowEgg: one egg bubble ("popup:Brobot 1::msg=<congrats>::<prize>") and
// Goverlord's congratulations for it, heard before the next.
func gopodShowEgg(who, word, prize string) {
	tp := gopodBrobotsSettings.TopicPictures
	gopodEggRevealed(gopodEggName(word))
	text := gopodEggText(word)
	// 2026-10-16 (operator: "Congratulations! Easter Egg [Title] found. /
	// MouseOver to Zoom View"): a prize that zooms (a picture, a photo, a weather
	// face) gets topic_pictures.egg_zoom_text on its own line.
	zoomSaid := ""
	if zoom := strings.TrimSpace(tp.EggZoomText); zoom != "" && gopodEggZooms(prize) {
		text += "\n" + zoom
		zoomSaid = zoom
	}
	// 2026-10-16 (operator: "Congratulations! Easter Egg 'Roast Battle' found: /
	// 🔥 Roast Battle 🔥 / ..."): before a card, the line ends with a colon.
	if strings.Contains(prize, gopodEggCardMark) && strings.HasSuffix(text, ".") {
		text = strings.TrimSuffix(text, ".") + ":"
	}
	GopodChatFeedAdd(tp.Role, gopodPopupMark+who+gopodPopupWhoMark+gopodEggMsgMark+text+gopodPopupWhoMark+prize)
	// 2026-10-24: said from its kept pieces, only the title made live (gopod_egg_voice.go).
	if !gopodEggSay(text, gopodEggsRevealed(", ", false), zoomSaid) {
		gopodSystemSay(text, true)
	}
}

// gopodEggName: the title as egg_text shows it (its first letter capital).
func gopodEggName(word string) string {
	name := strings.TrimSpace(word)
	if name != "" {
		r := []rune(name)
		name = strings.ToUpper(string(r[0])) + string(r[1:])
	}
	return name
}

// gopodEggZooms: the prize has something to zoom - a picture or photo path, or a
// weather face (not a card or the math trophy alone).
func gopodEggZooms(prize string) bool {
	for _, item := range strings.Split(prize, "|") {
		item = strings.TrimSpace(item)
		if strings.HasPrefix(item, "/") || strings.HasPrefix(item, gopodWeatherPopupMark) {
			return true
		}
	}
	return false
}

// gopodEggText: topic_pictures.egg_text for the word that found the egg ("Granny").
func gopodEggText(word string) string {
	tp := gopodBrobotsSettings.TopicPictures
	if strings.TrimSpace(tp.EggText) == "" {
		return tp.PopupText
	}
	return strings.NewReplacer("{name}", gopodEggName(word), "{eggs}", gopodEggsRevealed(", ", true)).Replace(tp.EggText)
}

// gopodEggRevealed: an egg's title, added to the ones revealed this session.
func gopodEggRevealed(title string) {
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}
	gopodInterviewShown.Lock()
	defer gopodInterviewShown.Unlock()
	for _, t := range gopodInterviewShown.revealed {
		if t == title {
			return
		}
	}
	gopodInterviewShown.revealed = append(gopodInterviewShown.revealed, title)
}

// gopodEggsRevealed: the eggs revealed this session, in the order found, joined
// by sep - each in quotes when quoted (shown: 'Bingo and foodmath', 'Roast
// Battle'; said: Bingo and foodmath, Roast Battle).
func gopodEggsRevealed(sep string, quoted bool) string {
	gopodInterviewShown.Lock()
	defer gopodInterviewShown.Unlock()
	out := make([]string, 0, len(gopodInterviewShown.revealed))
	for _, t := range gopodInterviewShown.revealed {
		if quoted {
			t = "'" + t + "'"
		}
		out = append(out, t)
	}
	return strings.Join(out, sep)
}

// gopodEggCardMark: a card prize in a pop-up line ("card:<words>").
const gopodEggCardMark = "card:"

// gopodEggMsgMark: before the egg's own congratulations in a pop-up line.
const gopodEggMsgMark = "msg="

// gopodEggWord: where the first of words comes in the line (whole words), and
// that word as written in the line; -1 when none.
func gopodEggWord(lower string, words []string) (int, string) {
	best, found := -1, ""
	for _, w := range words {
		w = strings.ToLower(strings.TrimSpace(w))
		if w == "" {
			continue
		}
		for start := 0; start < len(lower); {
			i := strings.Index(lower[start:], w)
			if i < 0 {
				break
			}
			i += start
			before := i == 0 || !gopodWordRune(lower[:i], true)
			after := i+len(w) >= len(lower) || !gopodWordRune(lower[i+len(w):], false)
			if before && after {
				if best < 0 || i < best {
					best, found = i, w
				}
				break
			}
			start = i + 1
		}
	}
	return best, found
}

func gopodWordRune(s string, last bool) bool {
	r := []rune(s)
	if len(r) == 0 {
		return false
	}
	c := r[0]
	if last {
		c = r[len(r)-1]
	}
	return unicode.IsLetter(c) || unicode.IsDigit(c)
}

// gopodTopicChapter: the egg key ("chapter:<title>") and title of the chapter
// a topic is in; "" when it's in none.
func gopodTopicChapter(topic string) (string, string) {
	for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
		for _, t := range c.Topics {
			if t == topic && strings.TrimSpace(c.Title) != "" {
				return "chapter:" + c.Title, c.Title
			}
		}
	}
	return "", ""
}

// gopodTopicTriggers: a topic's words (its pack's "words:" line).
func gopodTopicTriggers(name string) []string {
	for _, packet := range gopodTopicRegistry() {
		if packet.Name == name {
			return packet.Triggers
		}
	}
	return nil
}

// gopodPopupAllowed counts one more showing of the named pop-up and says
// whether it may show: up to topic_pictures.max_per_session per Wire-Pod start.
func gopodPopupAllowed(name string) bool {
	max := gopodBrobotsSettings.TopicPictures.MaxPerSession
	if max <= 0 {
		max = 1
	}
	gopodInterviewShown.Lock()
	defer gopodInterviewShown.Unlock()
	shown := gopodInterviewShown.pictures[name]
	if shown >= max {
		return false
	}
	gopodInterviewShown.pictures[name] = shown + 1
	return true
}

// gopodPopupMark: a picture line the page shows as a live pop-up.
const gopodPopupMark = "popup:"

// gopodPopupWhoMark: after the Brobot's name in a pop-up line.
const gopodPopupWhoMark = "::"

// gopodWeatherPopupMark: a pop-up item that is the weather ("weather:<description>@<hour>").
const gopodWeatherPopupMark = "weather:"

// gopodChapterReset: a new draft run shows its first chapter's title again.
func gopodChapterReset() {
	gopodInterviewShown.Lock()
	gopodInterviewShown.chapter = ""
	gopodInterviewShown.Unlock()
}

// ## EOF
