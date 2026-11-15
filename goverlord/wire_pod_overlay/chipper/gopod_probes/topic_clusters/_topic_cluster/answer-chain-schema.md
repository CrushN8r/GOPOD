# Answer chain schema

- **about:** 2026-11-06 (operator: "Please do: Schemas for run cards and answer chains"): the master answer chain - every chain (`answer_chains`, 01_brain.json and the clusters) is `parts` in order, or `from` another chain. Each part is one of the known parts below (a new part is added here when the code learns it). The system check (`gopodSchemaCheck`, every restart) lists an unknown part or a missing `from` chain "to fix".

---

## Known parts
1. greeting
2. golden_line
3. start_snippet
4. docbs
5. persona_chain
6. colour:value_point
7. llm
8. closing_remark
9. point_bubbles
10. summary_bubbles
11. docbs_close
12. persona_chain_close
13. weather_kit
14. reminder
15. end_snippet
16. redirect
17. intro_wrap
18. intro_weather
19. rule
20. who_line
21. manager_take
