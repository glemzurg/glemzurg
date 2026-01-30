package parser_ai

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TreeSuite tests writing and reading a complete model tree to/from the filesystem.
type TreeSuite struct {
	suite.Suite
	tempDir string
}

func TestTreeSuite(t *testing.T) {
	suite.Run(t, new(TreeSuite))
}

func (suite *TreeSuite) SetupTest() {
	// Create a temporary directory for each test
	tempDir, err := os.MkdirTemp("", "parser_ai_tree_test_*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir
}

func (suite *TreeSuite) TearDownTest() {
	// Clean up temporary directory
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestWriteAndReadTree writes a complete populated model tree to the filesystem
// and reads it back, verifying the data roundtrips correctly.
func (suite *TreeSuite) TestWriteAndReadTree() {
	t := suite.T()

	// Build a complete model tree
	model := t_buildTestModelTree()

	// Write the tree to the filesystem
	modelDir := filepath.Join(suite.tempDir, "test_model")
	err := WriteModelTree(model, modelDir)
	require.NoError(t, err)

	// Read the tree back from the filesystem
	readModel, err := ReadModelTree(modelDir)
	require.NoError(t, err)

	// Verify the model
	assert.Equal(t, model.Name, readModel.Name)
	assert.Equal(t, model.Details, readModel.Details)

	// Verify actors
	require.Len(t, readModel.Actors, len(model.Actors))
	for key, actor := range model.Actors {
		readActor, ok := readModel.Actors[key]
		require.True(t, ok, "actor '%s' not found", key)
		assert.Equal(t, actor.Name, readActor.Name)
		assert.Equal(t, actor.Type, readActor.Type)
		assert.Equal(t, actor.Details, readActor.Details)
	}

	// Verify domains
	require.Len(t, readModel.Domains, len(model.Domains))
	for domainKey, domain := range model.Domains {
		readDomain, ok := readModel.Domains[domainKey]
		require.True(t, ok, "domain '%s' not found", domainKey)
		assert.Equal(t, domain.Name, readDomain.Name)
		assert.Equal(t, domain.Details, readDomain.Details)
		assert.Equal(t, domain.Realized, readDomain.Realized)

		// Verify subdomains
		require.Len(t, readDomain.Subdomains, len(domain.Subdomains))
		for subdomainKey, subdomain := range domain.Subdomains {
			readSubdomain, ok := readDomain.Subdomains[subdomainKey]
			require.True(t, ok, "subdomain '%s' not found", subdomainKey)
			assert.Equal(t, subdomain.Name, readSubdomain.Name)
			assert.Equal(t, subdomain.Details, readSubdomain.Details)

			// Verify classes
			require.Len(t, readSubdomain.Classes, len(subdomain.Classes))
			for classKey, class := range subdomain.Classes {
				readClass, ok := readSubdomain.Classes[classKey]
				require.True(t, ok, "class '%s' not found", classKey)
				assert.Equal(t, class.Name, readClass.Name)
				assert.Equal(t, class.Details, readClass.Details)
				assert.Equal(t, class.ActorKey, readClass.ActorKey)

				// Verify attributes
				require.Len(t, readClass.Attributes, len(class.Attributes))
				for attrKey, attr := range class.Attributes {
					readAttr, ok := readClass.Attributes[attrKey]
					require.True(t, ok, "attribute '%s' not found", attrKey)
					assert.Equal(t, attr.Name, readAttr.Name)
					assert.Equal(t, attr.DataTypeRules, readAttr.DataTypeRules)
				}

				// Verify state machine if present
				if class.StateMachine != nil {
					require.NotNil(t, readClass.StateMachine)
					assert.Len(t, readClass.StateMachine.States, len(class.StateMachine.States))
					assert.Len(t, readClass.StateMachine.Events, len(class.StateMachine.Events))
				}

				// Verify actions
				require.Len(t, readClass.Actions, len(class.Actions))
				for actionKey, action := range class.Actions {
					readAction, ok := readClass.Actions[actionKey]
					require.True(t, ok, "action '%s' not found", actionKey)
					assert.Equal(t, action.Name, readAction.Name)
				}

				// Verify queries
				require.Len(t, readClass.Queries, len(class.Queries))
				for queryKey, query := range class.Queries {
					readQuery, ok := readClass.Queries[queryKey]
					require.True(t, ok, "query '%s' not found", queryKey)
					assert.Equal(t, query.Name, readQuery.Name)
				}
			}

			// Verify generalizations
			require.Len(t, readSubdomain.Generalizations, len(subdomain.Generalizations))
			for genKey, gen := range subdomain.Generalizations {
				readGen, ok := readSubdomain.Generalizations[genKey]
				require.True(t, ok, "generalization '%s' not found", genKey)
				assert.Equal(t, gen.Name, readGen.Name)
				assert.Equal(t, gen.SuperclassKey, readGen.SuperclassKey)
				assert.Equal(t, gen.SubclassKeys, readGen.SubclassKeys)
			}

			// Verify subdomain associations
			require.Len(t, readSubdomain.Associations, len(subdomain.Associations))
		}

		// Verify domain associations
		require.Len(t, readDomain.Associations, len(domain.Associations))
	}

	// Verify model-level associations
	require.Len(t, readModel.Associations, len(model.Associations))
}

// t_buildTestModelTree creates a complete populated model tree for testing.
func t_buildTestModelTree() *inputModel {
	return &inputModel{
		Name:    "Web Books",
		Details: "An online bookstore application.",
		Actors: map[string]*inputActor{
			"customer": {
				Name:    "Customer",
				Type:    "person",
				Details: "A person who purchases books.",
			},
			"admin": {
				Name:    "Administrator",
				Type:    "person",
				Details: "System administrator.",
			},
		},
		Domains: map[string]*inputDomain{
			"order_fulfillment": {
				Name:     "Order Fulfillment",
				Details:  "Handles customer orders.",
				Realized: false,
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name:    "Default",
						Details: "Default subdomain for order fulfillment.",
						Classes: map[string]*inputClass{
							"book_order": {
								Name:     "Book Order",
								Details:  "Represents a customer order.",
								ActorKey: "customer",
								Attributes: map[string]*inputAttribute{
									"id": {
										Name:          "ID",
										DataTypeRules: "int",
									},
									"status": {
										Name:          "Status",
										DataTypeRules: "string",
									},
								},
								Indexes: [][]string{{"id"}, {"status"}},
								StateMachine: &inputStateMachine{
									States: map[string]*inputState{
										"pending": {
											Name:    "Pending",
											Details: "Order created but not confirmed.",
										},
										"confirmed": {
											Name:    "Confirmed",
											Details: "Order confirmed.",
										},
									},
									Events: map[string]*inputEvent{
										"confirm": {
											Name:    "confirm",
											Details: "Confirm the order.",
										},
									},
									Guards: map[string]*inputGuard{
										"has_items": {
											Name:    "hasItems",
											Details: "Order has at least one item.",
										},
									},
									Transitions: []inputTransition{
										{
											FromStateKey: strPtr("pending"),
											ToStateKey:   strPtr("confirmed"),
											EventKey:     "confirm",
											GuardKey:     strPtr("has_items"),
										},
									},
								},
								Actions: map[string]*inputAction{
									"calculate_total": {
										Name:    "Calculate Total",
										Details: "Sum up all line items.",
										Requires: []string{
											"Order has items",
										},
										Guarantees: []string{
											"Total is computed",
										},
									},
								},
								Queries: map[string]*inputQuery{
									"get_subtotal": {
										Name:    "Get Subtotal",
										Details: "Get order subtotal before tax.",
									},
								},
							},
							"book_order_line": {
								Name:    "Book Order Line",
								Details: "A line item in an order.",
								Attributes: map[string]*inputAttribute{
									"quantity": {
										Name:          "Quantity",
										DataTypeRules: "int",
									},
								},
								StateMachine: &inputStateMachine{
									States: map[string]*inputState{
										"active": {
											Name: "Active",
										},
									},
									Events: map[string]*inputEvent{
										"create": {
											Name: "create",
										},
									},
									Guards:      map[string]*inputGuard{},
									Transitions: []inputTransition{
										{
											ToStateKey: strPtr("active"),
											EventKey:   "create",
										},
									},
								},
								Actions: map[string]*inputAction{},
								Queries: map[string]*inputQuery{},
							},
							"product": {
								Name:    "Product",
								Details: "A product available for sale.",
								Attributes: map[string]*inputAttribute{
									"name": {
										Name:          "Name",
										DataTypeRules: "string",
									},
								},
								StateMachine: &inputStateMachine{
									States: map[string]*inputState{
										"active": {
											Name: "Active",
										},
									},
									Events: map[string]*inputEvent{
										"create": {
											Name: "create",
										},
									},
									Guards:      map[string]*inputGuard{},
									Transitions: []inputTransition{
										{
											ToStateKey: strPtr("active"),
											EventKey:   "create",
										},
									},
								},
								Actions: map[string]*inputAction{},
								Queries: map[string]*inputQuery{},
							},
							"book": {
								Name:    "Book",
								Details: "A physical book.",
								Attributes: map[string]*inputAttribute{
									"isbn": {
										Name:          "ISBN",
										DataTypeRules: "string",
									},
								},
								StateMachine: &inputStateMachine{
									States: map[string]*inputState{
										"active": {
											Name: "Active",
										},
									},
									Events: map[string]*inputEvent{
										"create": {
											Name: "create",
										},
									},
									Guards:      map[string]*inputGuard{},
									Transitions: []inputTransition{
										{
											ToStateKey: strPtr("active"),
											EventKey:   "create",
										},
									},
								},
								Actions: map[string]*inputAction{},
								Queries: map[string]*inputQuery{},
							},
							"ebook": {
								Name:    "EBook",
								Details: "An electronic book.",
								Attributes: map[string]*inputAttribute{
									"file_format": {
										Name:          "File Format",
										DataTypeRules: "string",
									},
								},
								StateMachine: &inputStateMachine{
									States: map[string]*inputState{
										"active": {
											Name: "Active",
										},
									},
									Events: map[string]*inputEvent{
										"create": {
											Name: "create",
										},
									},
									Guards:      map[string]*inputGuard{},
									Transitions: []inputTransition{
										{
											ToStateKey: strPtr("active"),
											EventKey:   "create",
										},
									},
								},
								Actions: map[string]*inputAction{},
								Queries: map[string]*inputQuery{},
							},
						},
						Generalizations: map[string]*inputGeneralization{
							"medium": {
								Name:          "Medium",
								Details:       "Different book formats.",
								SuperclassKey: "product",
								SubclassKeys:  []string{"book", "ebook"},
								IsComplete:    true,
							},
						},
						Associations: map[string]*inputAssociation{
							"order_has_lines": {
								Name:             "Order Has Lines",
								Details:          "An order contains line items.",
								FromClassKey:     "book_order",
								FromMultiplicity: "1",
								ToClassKey:       "book_order_line",
								ToMultiplicity:   "1..*",
							},
						},
					},
				},
				Associations: map[string]*inputAssociation{},
			},
		},
		Associations: map[string]*inputAssociation{},
	}
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}
