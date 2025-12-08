
next:

- create the initial svgs
- add the svg links to the use case md
- work on the diagram generation

---------------------------


- work out the source of section and destination of section
- work out the event parameter source field
- use case diagram (graphviz)
  - Work through testing images with graphviz by hand.
  - all three appendices
- sequence diagrams: https://github.com/aorith/svg-sequence
- duplicate the text book diagram
- update all the other class diagrams to be a reduced form of it
- create a solution for ordering of events, states, actions, etc. for reability.


- an event that doesn't change state should be in the state as: EventB / ActionY (Just put the event name, the ‘/’, and the action name in the ‘event action’ zone. Note that the eventname/ action notation is somewhat more limited than explicit self-transition notation. Specifically, eventname/ action notation does not allow a guard, and only one action can be specified.)

- same work for state machines
- update parsing of actions
- updating parsing of data types
- add data dictionary of the textbook format
- add use cases
- add domains
- add interaction diagrams
- make state class members innate and not specified

- add use case level svg images
  - work with an artist to make mud, sea, sky (and mabye fish and kite)
  - create the computer svg and user svg as well
  - no white padding around the images
- add data flow diagram of all classes in model, every attribute

validation work:

- update to https://github.com/go-playground/validator
- update all fmt. to structured logs log/slog

database work:

- update to https://github.com/golang-migrate/migrate
- update to sqlc
- update to https://github.com/go-playground/validator
- any godog testing worth adding?

graphviz migration:

- classes:
  - add a model uml diagram using graphviz
  - tune labeldistance (head tail labels)
  - use invisible subgraph clusters to group the class inheritance
  - use invisible subgraph clusters to group association classes

- work out in graphviz:
  - how to draw a svg inside a node, of a class: lines and text
    - wait until the wasm is more mature (note from 10/2025)
  - how to ensure all classes are large enough for text
  - examine other aspects of the graphviz dot language

- TLA+ inside tool:
  - https://github.com/tlaplus-community/tree-sitter-tlaplus


Long time targets:

- post gres schemas, domians, compositions, records, etc
- move to devcontainers
- full TLA+ as initial pim overlay
  - Consider e ch node represents either a leaf (code line with FK) or a composite structure (sequence, branch, loop). It leverages PostgreSQL's recursive CTEs for traversal and supports modern indexing.
- full support of entire text book models
- cleaned up diagraming, allow person to configure
- generate minimal tool so that someone can get started easily
  - hire a talented designer for the tool
- generate two divergently different implementations of the tool
- generate communication diagrams
- feedback and hints on cleaner requirements writing
  - like amgiguity language
- support all known requirement methodologies
- build out interaction diagrams from existing data and just have users constrain them

================================

what is the best practices for an open source github repo

================================

Abiguity Review of requirements. Richard Bender

================================

https://dreampuf.github.io/GraphvizOnline/

================================

Some loose ideas:

peg parser generator:

https://github.com/mna/pigeon

render with mermaid.js

docker version of:
    https://github.com/mermaid-js/mermaid-cli