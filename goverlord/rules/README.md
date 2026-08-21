# rules/

This directory holds flow contracts: what code does with the facts
recorded in `data/`. Rules say how a fact becomes a decision (for
example, how a sample rate fact maps to a resample operation); they do
not themselves store the facts being acted on. See the three-way
code/data/rules decoupling doctrine
(`GOPOD_LAYER_1_DECOUPLING_DOCTRINE_001.md`, formalized further in
`GOPOD_PERFECT_CRYSTAL_PROPOSAL_001.md`).
