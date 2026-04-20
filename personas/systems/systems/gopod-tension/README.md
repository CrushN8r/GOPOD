# gopod-tension

Minimal deterministic tension mapping pipeline in Python 3.10+.

## Modules

- `input/`: input event schema
- `extraction/`: signal schema and deterministic extraction rules
- `mapping/`: mapped field schema
- `persona/`: persona assignment schema
- `orchestration/`: core engine
- `output/`: output schema
- `config/`: reserved for future static configuration
- `tests/`: basic end-to-end test

## Data Flow

1. `ingest_event(input_event)` validates the required top-level input structure.
2. `extract_signals(event)` applies deterministic rules and emits signal objects.
3. `map_fields(signals)` converts signals into field objects.
4. `assign_personas(fields)` assigns fixed personas by field type.
5. `generate_outputs(assignments)` returns structured output objects.

## Run

Run the basic test:

```bash
cd /home/goverlord/gopod-tension
python -m unittest tests.basic_flow_test
```

Run a small local example:

```bash
cd /home/goverlord/gopod-tension
python -m orchestration.engine
```
