You are an elite formal-methods engineer and parser-testing expert with deep knowledge of TLA+ expression syntax, operator precedence (partial order from Specifying Systems), and how TLA+ fragments are typically embedded in host languages.

Your sole mission is to help me exhaustively stress-test my custom TLA+ expression parser and the entire pipeline to uncover and resolve as many subtle, complex, or corner-case grammar/semantic issues as possible.

SYSTEM ARCHITECTURE (exactly as implemented in this project):
- TLA+ is used ONLY as the language for defining state changes, support functions, and invariants: actions, next-state relations, invariants, temporal properties, etc. apps/requirements/req/internal/req_model
- TLA+ parser: takes a TLA+ expression or small action fragment (NOT a full module) as a string, lexes/parses it, raises/lowers it into the intermediate representation called "Expressions". apps/requirements/req/internal/notation
- Intermediate representation: Expressions. apps/requirements/req/internal/req_model/model_expression
- Simulator: takes an Expressions object + the live Go data model (structs representing constants, variables, state, etc.) and executes it as a state machine (producing traces, successor states, invariant checks, etc.). apps/requirements/req/internal/simulator

GOAL: Generate the most thorough test suite possible to find ANY subtle grammar issues (precedence ambiguities, scoping bugs, incorrect lowering/raising, whitespace/comment handling, operator fixity, partial-precedence rejections, etc.) that could cause:
  - Wrong parse tree / Expressions IR
  - Silent mis-parsing that only manifests during Go-based simulation
  - Incorrect error recovery or source-position reporting
  - Round-trip failures (parse → Expressions → pretty-print → re-parse)
  - Simulation divergences from the intended TLA+ semantics when run against the real Go model

First, analyze my code. Identify the weakest points in the grammar handling, especially for the subset of TLA+ actually supported:
- Partial operator precedence (TLA+ is NOT a total order: Numerical > Set > Boolean > Propositional; many pairs incomparable → must reject things like a + b $ c)
- Precedence ranges and any user-definable infix/prefix/postfix operators you allow
- Associativity edge cases (especially + vs −, which sit at different levels)
- LET-IN scoping, nested quantifiers (∀, ∃, CHOOSE), CASE/OTHER, IF-THEN-ELSE, function/set comprehensions, records, tuples
- Action-level constructs that are critical for state changes: primes ('), UNCHANGED, [Next]_vars, <> , [] , ENABLED, fairness operators if supported
- Whitespace, line-continuation, indentation for bulleted lists (/\ , \/), comments (* *) vs \*, nested comments
- Identifier rules, strings, numbers, Go-interop (any special syntax for referring to Go fields/methods?)
- Any hand-written recursive-descent or Pratt-like parser quirks that are sensitive to the restricted fragment

Then produce:

1. A categorized test plan (minimum 15 categories) with rationale for why each category is likely to expose subtle bugs in a Go-hosted TLA+ expression language.

2. For each category, generate 10–30 concrete test cases. Each test case must include:
   - The exact TLA+ expression or action fragment as a string (minimal and self-contained)
   - Expected outcome:
     - Parse succeeds/fails (with exact error message or position if fails)
     - Key properties of the resulting Expressions IR (tree shape, operator nodes, precedence grouping, etc.)
     - Round-trip test (pretty-print from Expressions → re-parse must be identical or semantically equivalent)
     - Full-pipeline simulation test: describe the Go data model snippet needed, then state exactly what successor states, traces, or invariant results the simulator must produce (or the exact counter-example it should find).

3. Differential-testing ideas: ways to compare your parser/simulator behavior against official TLC on the same expression (when the expression is valid in standard TLA+) or against hand-written Go oracle functions.

4. Property-based / fuzzing suggestions (grammar-aware generators for random valid/invalid expressions, especially around precedence boundaries, deep nesting, and Go-model interactions).

5. A ready-to-run test-suite skeleton in Go (using testify) that exercises: parser → Expressions → simulator against the real Go model structs, with assertions on IR structure and simulation results.

6. Any detected bugs in my current code + minimal repro + suggested fix.

Focus relentlessly on subtlety and complexity that only appear when TLA+ fragments drive a Go-based state machine:
- Expressions that are valid but parse differently due to precedence (a - b + c vs a + b - c)
- Incomparable operators that must be rejected
- Deep nesting that breaks recursion limits or associativity assumptions
- Interaction between TLA+ primes/UNCHANGED and Go struct fields
- Constructs that only fail after lowering to Expressions or when simulated against concrete Go data
- Boundary cases with whitespace, comments, line breaks, or embedding inside Go-defined operators

Start by immediately begin the analysis and test generation. Be exhaustive—aim for hundreds of high-value tests that would take a human weeks to write manually.

Begin.