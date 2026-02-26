

The general steps to follow are these...

-----------------------------

1. The Model (do first)

The heart of apps/requirements/req is the package apps/requirements/req/internal/req_model. Any code to be made that would be in that file tree should be done first. It is the source-of-truth for the data model in the system. Work confirmed with `go test ./internal/req_model/...`

A defined model is the full object tree that is passed around the system to be be used in different ways.

Then the test model needs to be updated to work. It is the input to the tests in other packages. Work confirmed with `go test ./internal/test_helper/...`

(All tests in this document run from `cd /workspaces/glemzurg/apps/requirements/req`)

-----------------------------

2. The Database (do second)

The database is the first grindy confirmation of the model. It forces the exactness of hte model to be pushed into a relational SQL shape completely. 

The apps/requirements/req/internal/database/sql/schema.sql is the schema, and any change to it should have comments that fit the pattern of the rest of the file. If there is an enum in the req_model, then it should be enumeration type in the schema, just like the rest of the model shows.

The database layer code should match the pattern of the rest of the data access calls and then there is a top-level round trip test to confirm the test model works. The database must be tested with a flag: `go test ./internal/database/... -dbtests`

There are some specific design choices in the database code:

- The INSERT in teh load never uses string placehodlers as parameter, only literally written strings.
- Each table has its own golang code file and unit test. The golang code files should not call other table's files. It will just be the top-level test and code that stitch values together from the dabase into objects.

After database work is done the documenation should be regenerated.

- Remove all the files in apps/requirements/req/docs/dbdoc
- Run apps/requirements/req/doc.sh

-----------------------------

3. The AI Parser (do third)

The AI parser (apps/requirements/req/internal/parser_ai) has more strict data requirements thatn teh rest of the system, so it uses the strict test model instead of the test mode. If the struct test model needs more objects defined to pass the ai parser round-trip then it should be updated to do so. Work confirmed with `go test ./internal/parser_ai/...`

The AI parser has a few design choices:

- All data from the req_model needs to be in the objects that are written to or read from json. This is enforced by the round trip test working.
- No error message from req_model should ever be eported by the parser_ai code. Instead parser_ai should find that error and report it itself, ensuring it has a distinct error code and a error md that instructs a calling ai how to correct the error.
- As much validation information should be put into the json schemas as possible, since it is a clean form of documentation for AI tooling.
- Every bit of description in the schemas should be helping instruct and teach and ai how to correctly fill out the data.

Before considering the parser_ai complete for a task, do an examination that every possible error that can be reported under a req_model class will be reported with an appropriate error code and md file from parser_ai.

-----------------------------

4. The human parser (do fourth)

The human parser (apps/requirements/req/internal/parser) is a custom markdown yaml data format. Ensure all the data that goes in and out is explored in the tests for a class. And then the round-trip test will confirm that everything works well together. It uses the test_helper.GetTestModel(), not the strict model, but that model should *not* be updated to fix a bug in the parser project. That model is meant to explore the constraints and the flexibility of the req_model tree. 

-----------------------------

5. The flattened requirements (do fifth)

The flattened requirements (apps/requirements/req/internal/req_flat) prepares lookups for the template generation. Add any new looksups needed. Confirm work with `go test ./internal/req_flat/...`

-----------------------------

6. The generation code (do sixth)

The generation code (apps/requirements/req/internal/generate) should be updated to include any new objects or data, fitting into the patterns that already exist. New lookups will likely need to be added to apps/requirements/req/internal/generate/template.go. The markdown can be generated for testing from apps/requirements/req/internal/generate/dump_test_model_test.go (if the Skip is temporarily disabled). Work here should be confirmed with `go test ./internal/generate/...`

-----------------------------

7. The simulator (do seventh)

The simulator (apps/requirements/req/internal/simulator) should be updated and confirmed with testing: `go test ./internal/simulator/...`


-----------------------------

Lastly, do a final check: `go test ./...`
