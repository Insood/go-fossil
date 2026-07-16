# TODO

Low-priority follow-up items from the current review:

- Cache the chunk slice and its sort order in `ChunkManager.Chunks()` instead of allocating and sorting a new slice every frame.
- Avoid rebinding the same shadow depth texture onto every receiver material on every frame; cache the binding or only refresh it when the framebuffer changes.
