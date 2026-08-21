owui-env() {
  # BACKUP COPY NOTE: the real path here is scrubbed to a placeholder for this
  # public mirror - the live ~/.gopod_alias_lib/suits.sh keeps the real vault
  # path. Adapt this path before running from this copy.
  secure_env="$HOME/path/to/your/secure/vault/openwebui.local.env"
  [ -f "$secure_env" ] && source "$secure_env"
  [ -n "${OWUI_API_KEY:-}" ] && echo "OWUI_API_KEY_PRESENT yes" || { echo "OWUI_API_KEY_MISSING"; return 1; }
}

owui-key-save() {
  # BACKUP COPY NOTE: same scrub as owui-env() above.
  mkdir -p "$HOME/path/to/your/secure/vault"
  secure_env="$HOME/path/to/your/secure/vault/openwebui.local.env"
  [ -n "${OWUI_API_KEY:-}" ] || { echo "OWUI_API_KEY_MISSING"; return 1; }
  umask 077
  printf "export OWUI_API_KEY='%s'\n" "$OWUI_API_KEY" > "$secure_env"
  chmod 600 "$secure_env"
  echo "OWUI_KEY_SAVED yes"
}

owui-chat-bridge-json() {
  gorepo
  mkdir -p goverlord/runtime/suits/openwebui_import_001

  cat > goverlord/runtime/suits/openwebui_import_001/chat_bridge_skill.json <<'JSON'
{
  "id": "chat-bridge",
  "name": "GOPOD Chat Bridge",
  "description": "GOPOD multi-chat participant bridge",
  "meta": {"tags": ["GOPOD", "chat-bridge", "multi-chat"]},
  "is_active": true,
  "access_grants": [],
  "content": "GOPOD Chat Bridge Skill\n\nYou are participating in a GOPOD multi-chat SESSION.\n\nGoverlord conducts the session.\nYou are a participant body.\n\nUse the active suit.\nKeep responses packet-sized.\nRespect current role, destination, and filter.\nDo not mutate source meanings.\nDo not turn waste into knowledge.\nReturn JSON when asked.\nExpose PASS/BLOCKED when relevant.\n\nDefault packet fields:\nparticipant\nrole\nsession_read\nresponse_packet\nblocked\nblock_reason\nnext_action"
}
JSON

  python3 -m json.tool goverlord/runtime/suits/openwebui_import_001/chat_bridge_skill.json >/dev/null
  echo "CHAT_BRIDGE_SKILL_JSON_CREATED yes"
}

owui-suit-auto() {
  gorepo
  owui-env || return 1
  mkdir -p goverlord/runtime/suits/openwebui_import_001

  body="goverlord/runtime/suits/openwebui_import_001/suit_changer_update.json"
  response="goverlord/runtime/suits/openwebui_import_001/suit_changer_update_response.json"

  cat > "$body" <<'JSON'
{
  "id": "suit-changer",
  "base_model_id": "llama3.2:3b",
  "name": "Suit Changer",
  "params": {
    "system": "You are wearing the GOPOD Goverlord Flow Guide JSON Suit. raw llm response = live LLM responses later recorded/archived in sessions. filter = live scope for Goverlord flow-guide output. Use current repo truth first. Return JSON with PASS/BLOCKED and next action. Do not mutate source meanings.",
    "temperature": 0.35,
    "top_p": 0.9,
    "repeat_penalty": 1.08,
    "max_tokens": 384,
    "num_ctx": 4096,
    "keep_alive": "15m",
    "format": "json",
    "think": false
  },
  "meta": {
    "profile_image_url": "/static/favicon.png",
    "description": "GOPOD suit changer / Goverlord flow-guide JSON suit",
    "toolIds": ["goverlord_tool", "robot_status", "robot_say"],
    "skillIds": ["gopod-skill", "chat-bridge"],
    "tags": [{"name":"GOPOD"},{"name":"Suit Changer"},{"name":"Goverlord"},{"name":"Chat Bridge"}]
  },
  "access_grants": [],
  "is_active": true
}
JSON

  curl -fsS -X POST \
    -H "Authorization: Bearer $OWUI_API_KEY" \
    -H "Content-Type: application/json" \
    "http://127.0.0.1:3000/api/v1/models/model/update?id=suit-changer" \
    --data-binary @"$body" > "$response" || { echo "OWUI_SUIT_UPDATE_BLOCKED"; return 1; }

  grep -q '"id":"suit-changer"' "$response" \
    && grep -q '"chat-bridge"' "$response" \
    && echo "OWUI_SUIT_CHANGER_CHAT_BRIDGE_ATTACH_PASS yes" \
    || { echo "OWUI_SUIT_VERIFY_BLOCKED"; return 1; }
}

owui-chat-bridge-close() {
  owui-key-save
  owui-chat-bridge-json
  owui-suit-auto
}

# NOT pha0b (with a zero, ~/.gopod_alias_lib/brobots.sh) - that is a
# different, unrelated alias (the studio-wide Playhead Point A/0/B song
# cockpit, picks a song and plays a step_id slice through a real robot
# runner). One character apart, two different tools - if you meant to
# rehearse a song slice, you want pha0b, not this.
phaob() {
  gorepo
  cat <<'TXT'
PLAYHEAD WHERE ARE WE?

Frame 0:
gopod_alias tool fitting

Point A:
PASS — gopod_alias import file was cut and imported into Open WebUI.

Point 0:
PASS — CRUSHN8R attached gopod_alias to Suit Changer and saved.

Point B:
PASS — saved Suit Changer export shows meta.toolIds includes gopod_alias.

Frame 1:
Run Suit Changer tool call:
gopod_alias
name = suit_check

Expected:
SUIT_ALIAS_READY yes
TXT
}

