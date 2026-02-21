
---------------------------

todo:
  - examine test coverage of database package
    - specifically foreing key tests
  - use ai to check comemnts and forieng keys in schema
    - coverage of tests
  - update the yaml parsing
    - update hte parser ai parsing
      - add the extra structures to parser ai
    - update the md file output
      - include tla+
    - update the simulator
      - finish the todo processing
    - join simulator into req with parameters
  - replace all fmt.Errorf() with errors.Errorf()
  
feb (in parallel)
  - create complete model
    - add TLA+ parsing to create complete model
    - move parameters to actions and parsed into database
    - remvoe the json markup in req_model tree, and any json handling code
      - move to the parser_ai package
  - prepare the md output to be complete
    - update the scenarios to use the logic
    - update the use cases if needed
  - enter the evenplay model

march (art workshops)
  - steven tockey model freivew
  - generate data models compilers
  - generate protocols compilers

when:
  - generate ui designs compilers
  - generate to the level of architecture diagrams
    - model an AST with an Adjacency Listf and CTE query



april (in parallel)
  - generate model compilers
  - manage design models
  - update simulators to run models together


---------------------------

simulator todo:

- LET in TLA (it's is started by AI)
- Bags in TLA
- Simulator code
- data type in simulators
- unit test all the tla+ calls
- remove the monkey stubs in ast/old_stubs.go
- fix object float constructor to use the strings avlues from tla
- migrate into main repo
  - extend into database and input/outputs
- implement stubbed simulator logic

---------------------------

  - make generalizations have subdomain parents: apps/requirements/req/internal/database/generalization.go

- restrict names so that tla+ do not have conflics (like _Stack:Pop, no state attribute)

  - make object members private
- cleanup regex must compile code
- import examples models from steve's examples
- fix the nested sequence diagram display issue (move to d2)
- examine d2 diagramming:
  - https://github.com/terrastruct/d2?tab=readme-ov-file#d2-as-a-library
  - examine license of other libraries in use

- make generalization a class object, and make a use case one too

- schema has commented out columsn that represent undone features

- last half of year: complete all the functionality from text book
  - and create working example for documentation

- tla plus peg parser
- design the simuilator - chained with derived simulators
  - inspect the existing simulator and create a model for it
  - study library and grammar https://pkg.go.dev/github.com/mna/pigeon

update to https://github.com/go-playground/validator

- examine the recursive postgres capabilities
  - store full simualator as recursive rows without json blobs 

- work out the source of section and destination of section
- work out the event parameter source field
- duplicate the text book diagram
- update all the other class diagrams to be a reduced form of it
- create a solution for ordering of events, states, actions, etc. for reability.
- same work for state machines
- update parsing of actions
- updating parsing of data types
- add data dictionary of the textbook format
  - work out how to name and how to add details
- add domains
- make state class members innate and not specified

- use godoc to review exported methods

- revisit irrationals and how they are handled in a model:
  - https://grok.com/share/c2hhcmQtMw_524fb597-1e4a-4906-b1ce-37de14fe80af
  - https://grok.com/share/c2hhcmQtMw_e2b88d89-32b7-4b71-bd3a-a37d5251e03f

- add tla invariants

- postgres more features:
  - constraints
  - work through study
  - posgres domain and composites and ranges best practices

- Work through how to handle actions with tla:
  - data types for action parameters would be in use, right now unused
  - remvoe event parameters
  - the tla itself would deifne the parameters for events, actions, etc. 
    - all data flow 

- make a tla prover?
  - https://proofs.tlapl.us/doc/web/content/Home.html
  - isolate which tla grammar is really for the prover:
    - https://lamport.azurewebsites.net/tla/tla2.html

- reinvistion the data flow
  - correct the issue with outward flows being based on data types

- restore the foreign key from object references to classes
  - -- ALTER TABLE data_type_atomic2 ADD CONSTRAINT fk_atomic_class FOREIGN KEY (model_key, object_class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE;

- in database:
  - add state generalizations
  - add invariants to class (on state right now)
  - add invariants to state
  - add invariants to actions (requires/guanratees)
  - add history transitions

- for md/yaml generation create a custom yaml exporter (instead of just using string construction)

- setup best practices github repo for modeler

- go test updates
  - update tests to be a table format with names

- move to devcontainers
  - move scripts to make files
  - remove linux users and local postgres
  - remove postgres?
  - https://www.cyberciti.biz/faq/linux-list-users-command/
  - remove mesasge queues
  - tilt and cattle prod https://github.com/tilt-dev/ctlptl

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

parser:

- https://github.com/mna/pigeon

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

- optimize for speed of modeling
  - work with experienced ui designer
- design data models explicitly for digestion by AI
  - give the models self referential hints for "thrashing" through code options
  - small data tranformations that contain meaningful code patterns that can be applied iteratively
  - json schemas that dictate available choices that can be iterated on
- start gitbooks documentation
  - use grammar level checking (https://x.com/i/grok?conversation=1998900138715263329)
- interaction diagrams:
  - update node creation and deletion style
- attribute interaction lines are:
  - multiple attributes
  - conditional
  - multiple, nested multiples (data flow)
- attribute interaction lines are:
  - records of attributes with multiples
  - conditiononals
  - indication of whether it is many responses
  - same with events, showing the parameters
  - (perhaps this should be resolved with data flow diagrams)
- formal verification: https://martin.kleppmann.com/2025/12/08/ai-formal-verification.html
- clean up looping transitions to be inside the body as: event [guard] / action
- post gres schemas, domians, compositions, records, etc
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
- build a custom svg graphing library

- process tracking:
  - Benchmark of files changed equals a bug
- come up with solution for performance-optimized code generation
  - different inputs

consider gremlins testing:
- https://gremlins.dev/latest/

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