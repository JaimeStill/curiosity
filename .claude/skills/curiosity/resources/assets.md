# Asset Resources

Curated resources for acquiring external assets. Organized by category. Updated as the landscape shifts. Not subject to the documentation-decay discipline that governs `design/`.

## Notes Before Using

**Voxel-specific priorities.** Block textures (small, tileable, palette-consistent), particle and effect sprites, UI elements, fonts, sound effects, ambience, and music are the dominant needs. 3D models are lower priority - when needed, voxel volumes are typically authored in MagicaVoxel (free, native `.vox` format) rather than acquired as meshes.

**Placeholder-first posture.** Early engine work needs assets that exist and are usable, not assets that are aesthetically coherent. Solid colors per block ID, procedurally generated noise textures, and CC0 placeholder sets are all sufficient for validating engine foundations. Custom asset authoring becomes a worthwhile investment once the engine is mature enough that the work has somewhere stable to land. Until then, prefer the lowest-friction acquisition path.

**Licensing matters.** Every asset acquired should have its license recorded with the asset itself (or in a project-level attribution file when CC-BY or similar requires it). License posture matrix:

- **CC0 / Public Domain** - no strings, preferred default.
- **CC-BY** - attribution required; maintain an attribution file.
- **CC-BY-SA** - share-alike propagates; avoid for assets that would couple with engine code.
- **GPL** - share-alike with code implications; usable for assets but requires care.
- **Custom / proprietary** - read the license; commercial use is not always granted.

**AI-generated content.** Commercial use policies vary by tool and shift frequently. Confirm current terms at acquisition time, not at first use of the tool. Some tools assert no rights to outputs; others reserve rights or require paid tiers for commercial use. Self-hosted open-weight models (e.g. Stable Diffusion via local tooling) avoid platform-side licensing concerns but still inherit any model-card restrictions.

---

## Free and Open Source

### Multi-category hubs

- **Kenney.nl** - CC0, ~40,000 assets across 2D, 3D, voxel, UI, audio, fonts. The single highest-value source for indie projects. Voxel asset packs included.
- **OpenGameArt.org** - mixed licenses (CC0, CC-BY, CC-BY-SA, GPL), large and varied. Filter by license. Quality varies but depth is unmatched.
- **itch.io free assets** (`itch.io/game-assets/free`) - mixed licenses, often indie creators. Read each pack's license carefully.

### Textures

- **ambientCG** - CC0 PBR textures up to 8K+. Strong on natural surfaces, materials.
- **Poly Haven** - CC0 textures, HDRIs, and 3D models. High quality, no signup.
- **CGBookcase** - CC0 PBR textures. Smaller library than ambientCG, comparable quality.

### Audio - sound effects

- **Freesound.org** - CC-licensed sounds (mixed licenses, filter by CC0). Massive library, search is functional. Quality varies by upload.
- **Sonniss GDC bundle** - released annually around GDC, professional game audio, royalty-free for commercial use. Past bundles archived.
- **GameSounds.xyz** - curated free sound effects oriented at games.
- **Mixkit** - free sound effects with their own license terms (commercial use generally allowed; verify).

### Audio - music

- **Free Music Archive** - CC-licensed music, mixed licenses.
- **Incompetech (Kevin MacLeod)** - CC-BY music, very large catalog, attribution required.
- **FreePD** - public domain music.

### Fonts and UI iconography

- **Nerd Fonts** (`nerdfonts.com`) - **default choice for engine UI.** Patches popular open-source monospace fonts (JetBrains Mono, FiraCode, Hack, Iosevka, and others) with 10,000+ glyphs polyfilled into the Unicode private-use area. Glyphs include Font Awesome, Devicons, Octicons, Material Design, Powerline, weather, and more. Patched fonts inherit their original licenses (typically OFL) and remain free for commercial use.

  The strategic value for early engine work: UI iconography is delivered through the text rendering pipeline, so status indicators, navigation affordances, and action icons do not require a separate icon asset system. The UI subsystem can render glyphs as part of normal text. An icon system can be deferred until the engine is mature enough to need one - if ever.

- **Google Fonts** - open-licensed (typically OFL or Apache), broad selection. Useful for non-monospace needs (display text, headings).
- **Font Squirrel** - filtered for commercial-use-allowed fonts.

### Voxel models

- **MagicaVoxel community / Sketchfab** - many `.vox` and voxel-derived models under CC0 or CC-BY. Sketchfab requires filtering by "downloadable" and per-asset license check.

### Procedural sound generators

- **Bfxr / sfxr / jsfxr** - retro-style sound effect generators, browser or desktop. Output is yours to use.
- **ChipTone** - similar, browser-based, friendly UI.

---

## AI Generated

### Local / open-weight (recommended for control and licensing clarity)

- **Stable Diffusion** (via Automatic1111, ComfyUI, or similar) - local, open weights. Full control over outputs and no platform-side commercial restrictions beyond the model license. Strong for textures, particle sprites, concept reference.

### Hosted - images and textures

- **Scenario.gg** - game-art focused, sprites, textures, concept art. Commercial tiers available.
- **Midjourney** - high-quality outputs, commercial use generally permitted on paid tiers. Less prompt control than open-weight tools.

### Hosted - audio

- **ElevenLabs** - sound effects and voice. Production-quality outputs. Commercial tiers required for commercial use.
- **Stable Audio** - music and SFX. Open and hosted variants exist; check current terms.
- **Suno / Udio** - music generation. Commercial use policies vary by tier and have shifted recently; verify before committing.
- **AIVA** - cinematic music with mood and genre control.

### Hosted - 3D (low priority for voxel)

- **Meshy** - text/image to 3D, includes auto-rigging. Outputs typically need cleanup before production use. Less relevant for voxel pipeline unless using as reference for voxelization.
- **Tripo / Luma / Rodin** - alternatives in the same space.

### Considerations

- **Output ownership and commercial use** vary by tool and tier. Confirm at acquisition.
- **Style consistency** is the practical weak point - generating a coherent set of assets that feel like one game requires either careful prompt discipline, fine-tuned models, or post-processing.
- **Quality is uneven**. AI-generated assets often need cleanup, retouching, or selective regeneration.

---

## Paid

### Marketplaces

- **Unity Asset Store** - large variety. Assets are usable outside Unity for most license categories; check per-asset terms. Quality varies.
- **Fab** (Epic's unified marketplace, replaces Unreal Marketplace and Quixel) - large catalog, includes former Quixel Megascans content.
- **itch.io paid assets** - indie creators, often excellent value, varied licenses.
- **GameDev Market** - game-focused asset marketplace.
- **CraftPix** - game-focused, sprites and tilesets, premium and free tiers.

### Audio - subscription libraries

- **Epidemic Sound** - music and SFX subscription, commercial use included.
- **Artlist** - music and SFX subscription.
- **Soundsnap** - sound effects subscription.
- **Sonniss** - professional sound libraries, perpetual licenses.

### Bundles

- **Humble Bundle** - periodic game asset bundles, often strong value when active.

### Texture libraries

- **Quixel Megascans** (now part of Fab) - high-fidelity scanned materials. Largely free for Unreal use; check terms for other engines.
- **Textures.com** - mixed free and paid PBR textures.

---

## Selection Heuristics

When choosing among options for a given asset need:

1. **Start with CC0 sources** (Kenney, ambientCG, Poly Haven, Sonniss bundles). Lowest friction, no attribution overhead.
2. **Check Freesound, OpenGameArt, itch.io free** for specific gaps. Filter by license.
3. **Reach for AI generation** when a specific asset doesn't exist in free libraries and acquisition friction or cost is the primary obstacle. Prefer local/open-weight tools when output ownership matters.
4. **Pay** when a curated, coherent, production-quality asset pack saves enough time to justify the cost, or when subscription libraries cover ongoing needs cheaper than per-asset licensing.

For each acquired asset, record source, license, and any attribution requirement at acquisition time. This is significantly cheaper than reconstructing it later.
