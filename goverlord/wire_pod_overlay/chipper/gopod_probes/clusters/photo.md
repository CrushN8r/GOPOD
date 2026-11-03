# The photo
- **words:** photo, picture, what do you see, camera
- **ask:** What do the Brobots say about a photo?
- **chain:** Other - the photo egg (what a Brobot sees)
- **see also:** easter_eggs, topic-pictures
- **about:** 2026-11-03 (step 6 of GOPOD_CLUSTERS_PLAN_002.md): from 03_chat_page.json photo. A value in backticks is that text exactly ("\n" a break). The numbers, switches, roles and nested groups stay in the settings for the format step.
---

## Photo (photo)
1. prompt: Describe what you see in one short, simple sentence, naming a few main things.
2. answer_text: I see {saw}
3. recall_text: I saw {saw}
4. nothing_text: I haven't taken a photo yet.
5. fail_text: I couldn't take a photo.
6. unclear_text: I took a photo, but I couldn't make out what's in it.
7. quick_text: I see {objects}.
8. quick_none_text: Hmm, I need a closer look.
9. gathering_text: Gathering details. Please wait!
10. detail_text: I see {saw}
11. detail_unclear_text: That's all I can make out.
12. detail_prompt: ``
13. stall_lines: My eyes are basically potato cams, bro. | Hold on, my lens is buffering. | Zoom and enhance... like on TV.
14. detail_strip_prefixes: The image shows | The image features | The image depicts | The image captures | This image shows | The photo shows | The picture shows | In the image, | In this image, | I see
