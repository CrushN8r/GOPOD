# Answer chain schema

- **about:** 2026-11-06 (operator: "Please do: Schemas for run cards and answer chains"): the master answer chain - every chain (`answer_chains`, 01_brain.json and the clusters) is `parts` in order, or `from` another chain. Each part is one of the known parts below (a new part is added here when the code learns it). The system check (`gopodSchemaCheck`, every restart) lists an unknown part or a missing `from` chain "to fix".

---

## Known parts
1. golden_line
2. start_snippet
3. docbs
4. persona_chain
5. colour:value_point
6. llm
7. closing_remark
8. point_bubbles
9. summary_bubbles
10. docbs_close
11. persona_chain_close
12. weather_kit
13. reminder
14. end_snippet
15. redirect
16. intro_wrap
17. intro_weather
18. rule
19. who_line
20. manager_take
