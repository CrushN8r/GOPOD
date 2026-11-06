# Topic schema
- **about:** 2026-11-06 (operator: "build schemas for the other places in GOPOD where a schema would earn its keep, most useful first"): the master game card - every topic (a "Game card:" in a root doc, `topic_clusters.root_docs`) fills these lines; README.md here tells what each one is. The system check (`gopodSchemaCheck`, every restart) lists a card missing a required line "to fix". Nothing reads this file as a topic.

---

## Required
1. label: `# <label>` - how the Brobots name the topic
2. chapter: the chapter it belongs to - a chapter's `# <title>` (chapter-schema.md)
3. words: the words in a question that load the topic
4. summary: what a summary ask shows, pieces split by |
5. source: the repo file that tells the topic in full

## Optional
1. value points - `say N`, `notes N`, or numbered points after `---`; a card with none is an explainer, nothing to win
2. ask
3. pictures
4. reminder
5. pre
6. post
7. closing
8. sentences
9. start line
10. end line
