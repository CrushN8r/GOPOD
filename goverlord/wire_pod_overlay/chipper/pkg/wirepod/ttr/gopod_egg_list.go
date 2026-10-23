// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_egg_list.go
package wirepod_ttr

import "strings"

// The Brobots' egg list (2026-10-17, operator: "'Hey that reminds me' thought
// bubbles need to be updated * updated to center around a list of easter eggs
// (chapter/topics/value-points) * during live chat, whatever easter eggs not
// discovered, are left on the list to mention * so brobots in their responses
// also finish with helping to keep the conversation flowing * like a fallback
// list of sellable features the brobots have, as GOPOD reps"). Every easter egg,
// in the chapters' order: each topic with a prize (its pictures or card,
// topic_pictures; named by egg_names, else interview_chapters.topic_labels),
// then value_points.reminder_eggs (the weather, Explain the Math, Get Photo). An
// egg found since this Wire-Pod start (its pop-up shown, gopodPopupAllowed) is
// off the list. "Hey, that reminds me..." names the eggs still on it - each in
// turn; once all have been named, round again.

type gopodListEgg struct {
	Key, Label string
}

// gopodEggList: every egg, found or not. 2026-10-17 (operator: "easter eggs
// are live run chapters"): each chapter (by its title), then each topic in no
// chapter with a prize (the roast battle), then value_points.reminder_eggs.
func gopodEggList() []gopodListEgg {
	tp, ic := gopodBrobotsSettings.TopicPictures, gopodBrobotsSettings.InterviewChapters
	var out []gopodListEgg
	seen := map[string]bool{}
	for _, c := range ic.Chapters {
		if title := strings.TrimSpace(c.Title); title != "" && !seen["chapter:"+title] && !gopodNotEgg("chapter:"+title) {
			seen["chapter:"+title] = true
			out = append(out, gopodListEgg{"chapter:" + title, title})
		}
	}
	// 2026-10-18 (operator: "leave halloween easter egg, out of interview lists"):
	// topic_pictures.not_listed pop up when found, but aren't on the list.
	for _, n := range tp.NotListed {
		seen[n] = true
	}

	for _, p := range gopodTopicRegistry() {
		if seen[p.Name] || gopodNotEgg(p.Name) || (len(tp.Pictures[p.Name]) == 0 && strings.TrimSpace(tp.Cards[p.Name]) == "") {
			continue
		}
		if key, _ := gopodTopicChapter(p.Name); key != "" {
			continue
		}
		label := strings.TrimSpace(tp.EggNames[p.Name])
		if label == "" {
			label = ic.TopicLabels[p.Name]
		}
		if label == "" {
			continue
		}
		seen[p.Name] = true
		out = append(out, gopodListEgg{p.Name, label})
	}
	for _, e := range gopodBrobotsSettings.ValuePoints.ReminderEggs {
		if e.Key != "" && e.Label != "" && !seen[e.Key] && !gopodNotEgg(e.Key) {
			seen[e.Key] = true
			out = append(out, gopodListEgg{e.Key, e.Label})
		}
	}
	return out
}

// gopodNotEgg (2026-10-18, operator: "confusing overlap between system managed &
// easter egg * remove easter egg: 'Websites, themes and the channel'"): the egg
// is taken out (topic_pictures.not_eggs).
// 2026-10-22 (operator: "fewer easter eggs would be nice * Do 1 and 2 together";
// CORE_TOPICS_002.md): with topic_pictures.eggs_keep set, only those are eggs.
func gopodNotEgg(key string) bool {
	if keep := gopodBrobotsSettings.TopicPictures.EggsKeep; len(keep) > 0 {
		in := false
		for _, k := range keep {
			if k == key {
				in = true
			}
		}
		if !in {
			return true
		}
	}
	for _, n := range gopodBrobotsSettings.TopicPictures.NotEggs {
		if n == key {
			return true
		}
	}
	return false
}

// gopodEggFound: its pop-up has shown since this Wire-Pod start.
func gopodEggFound(key string) bool {
	gopodInterviewShown.Lock()
	defer gopodInterviewShown.Unlock()
	return gopodInterviewShown.pictures[key] > 0
}

// gopodEggsToMention: up to want eggs not found, not named yet this round (a
// new round once all of them have been named), skipping skip's keys.
func gopodEggsToMention(want int, skip map[string]bool) []gopodListEgg {
	var open []gopodListEgg
	for _, e := range gopodEggList() {
		if !gopodEggFound(e.Key) && !skip[e.Key] {
			open = append(open, e)
		}
	}
	pick := func() []gopodListEgg {
		var out []gopodListEgg
		for _, e := range open {
			if len(out) < want && !gopodReminded("egg:"+e.Key) {
				out = append(out, e)
			}
		}
		return out
	}
	out := pick()
	if len(out) == 0 && len(open) > 0 {
		gopodCoverage.Lock()
		for k := range gopodCoverage.reminded {
			if strings.HasPrefix(k, "egg:") {
				delete(gopodCoverage.reminded, k)
			}
		}
		gopodCoverage.Unlock()
		out = pick()
	}
	for _, e := range out {
		gopodRemind("egg:" + e.Key)
	}
	return out
}

// gopodNameList: "a", "a and b", "a, b and c".
func gopodNameList(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// ## EOF
