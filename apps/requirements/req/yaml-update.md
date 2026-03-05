You are an expert senior Go engineer specializing in large-scale parser/generator packages, clean multi-package architecture, and maintainable file structures in Go.

Project context (do NOT change these facts):

- Package: `apps/requirements/req/internal/parser_human/`  ← THIS is the package you will update
  - Handles parsing and generation of a human-friendly hybrid markup + YAML DSL.
  - TLA+ parsing now goes into an Intermediate Representation (IR) in `apps/requirements/req/internal/req_model/model_expression/` and on generation it comes back out. The input TLA+ could come out as a different TLA+ so tests should be updated for a TLA+ that round-trips fine.
  - The model now has new logics that have to be explored in the tests. And data types can now be linked to expression types.
 

- Package: `apps/requirements/req/internal/req_model` ← do NOT modify
  - Contains the canonical, fully detailed Go structs.
  - Recently received several new object types (e.g. new top-level or nested structs for additional domain concepts). You can assume the new types are already exported and documented in datamodel.
  - If a change seems to need to be made her, stop and prompt me.

Task:

We need to update the apps/requirements/req/internal/parser_human to handle every aspect of the model in apps/requirements/req/internal/req_model (keeping in mind that the IR is really TLA+ in the parser_human)

Specific requirements for the update:

1. Add support for the new datamodel objects by:
   - Implementing parser functions that turn the human markup+YAML syntax for these new objects into the existing IR via TLA+.
   - Implementing generator functions that turn IR back into clean human-readable markup+YAML+TLA for the new objects.
   - Extending the central entry points (or the internal dispatcher/registry) so the new objects are automatically routed to their handlers.

2. apps/requirements/req/internal/parser_human/top_level_round_trip_test.go works.


Best-practice constraints:
- Use the same style, naming conventions, and error-handling patterns already present in the package.
- Prefer explicit registration (e.g. a `objectHandlers` map or interface-based dispatcher) over giant switch statements for future scalability.
- All new code must have full godoc.
- No changes to TLA+ parsing, IR structs, or the distill logic for existing objects.
