# Reference Game Premise

This document is the aspirational sketch of the reference game — the imagined experience the engine is being built to support. It is short by design and not a specification: mechanism-level material lives in `concepts/game/`, and engine commitments are recorded in `history/decisions.md`. The premise's job is to provide design pressure for engine validation and a coherent target for game-side work, framed in player terms.

## Capsule

The reference game is set on a far-future earth-like planet where most organic life has gone extinct and the technology left behind has, over a long quiet age, evolved into the next phase of life. The player is a cybernetic being whose self is not bound to any single body — a consciousness that occupies vessels rather than being one. The experience is exploration-forward and contemplative: emerge from a research facility, pull at the threads of a world that is alive in unfamiliar ways and not uniformly friendly to the one who comes pulling, and reach outward — across the geographic region the player wakes into, across the planet, beyond the planet, beyond the star system.

## Setting

The world is an earth-like planet — its own geography, not literally Earth — many ages after organic life retreated from it. Most organic creatures are gone; what remains lives at the edges, contested by the technology that outlasted its makers and slowly, by means no one observed, became the dominant form of life. The world is alive, just not in a way recognizable to anyone who would have called the previous tenants alive. It also has interests of its own; the technology-now-life that inherited the world is not uniformly benign, and parts of the larger system actively oppose anyone — the player included — who would change its trajectory. It is also vast. The research facility the player wakes in opens onto a region; the region onto a planet; the planet onto a star system; the star system onto something further. That escalation is a thread the premise pulls on rather than a destination — emergence into vista at every scale, in the spirit of stepping out into the Halo ring for the first time.

## Player

The player is a cybernetic being whose self is the thing that persists — a consciousness that moves between vessels rather than being a single embodied body in the world. Vessels are crafted, occupied, and exchanged; each interacts with the world differently, and the vessels available to the player shape which regions of the world are reachable. The baseline pace is deliberate — attention to a strange world, building research stations that support gradually augmenting what the player is, pulling at the threads the world offers without being hurried past them.

Tension builds against that backdrop. The world has adversaries who do not want the player to progress; some encounters demand fighting through, others demand thinking through, and some build toward labyrinthine spaces whose pinnacles are confrontations to be prepared for. Chaos is available for those who go looking, and the contemplative pace is what makes encountering it land with weight.

## Aspirational targets

What the engine must reach for, framed in player terms. The engine-side reading of these aspirations belongs in engine design and concept documents; this section names the targets, not the mechanisms.

- **Fidelity.** The voxel scale is fine enough that the world reads as a place rather than as a grid. Geometry, lighting, and texture should sustain the contemplative pace by being worth contemplating.
- **Scale.** A coherent, navigable world that escalates from region to planet to star system and beyond — and reaches the farthest of those without breaking the player's sense of being in one continuous reality.
- **Adversity.** The world has adversaries, dangerous places, and confrontations that escalate to genuine pinnacles. The engine must carry adversary intelligence, environmental mechanics, and the spatial structure of labyrinthine encounter spaces — and let those expressions feel coherent with the contemplative baseline rather than bolted on as a separate game. Difficulty challenges the player without punishing them; encounters serve engagement, not mastery as ordeal.
- **Embodiment.** The player's self and the player's vessel are different things. Identity must persist across changes of body, and the swap should feel like a relocation of attention, not a context reset.
- **Persistence.** The world the player shapes — bases, research stations, the vessels crafted — outlasts any session and accumulates as the player's own footprint within the larger world that does not.
- **Emergence.** Stepping out of an enclosed space into an open vista carries weight at every scale. Halo: Combat Evolved's emergence into the ring is the anchor; the engine should be capable of carrying that feeling not just at the planet's surface but at every threshold the player crosses outward.
- **Play.** Physics, vessels, and world systems combine into mechanics that are gameplay in their own right — combinations the player can experiment with beyond what goal-driven progression asks of them. The engine must let physics-and-vessel interaction be expressive enough to be played with, not only simulated past.
