// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_egg_list.go
package wirepod_ttr

import "strings"

type gopodListEgg struct {
	Key, Label string
}

// gopodEggList: every egg, found or not.
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

// gopodNotEgg: the egg is taken out (topic_pictures.not_eggs).
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

// gopodEggsToMention: up to want eggs not found, not named yet this round (a new round once all of
// them have been named), skipping skip's keys.
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
