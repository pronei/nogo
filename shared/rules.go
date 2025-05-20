package structs

type LimitRequest struct {
	//entity name -> attribute name -> attribute value
	Parameters map[string]EntityParameters `json:"parameters"`
	RequestId  string                      `json:"requestId,omitempty"`
}

type EntityParameters struct {
	EntityType    string            `json:"entityType"`
	AttributesMap map[string]string `json:"attributesMap"`
}

type RuleImport struct {
	// key for this map should be the type of entity
	EntityRuleMap map[string]EntityRules `json:"ruleMap"`
}

type EntityRules struct {
	EntityName       string          `json:"name" bson:",omitempty"`
	EntityType       string          `json:"type"`
	EntityAttributes []AttributeRule `json:"attributes"`

	// TODO: implement as ALL attribute rule
	EntityLimit int `json:"limit"`
}

type AttributeRule struct {
	Description    string `json:"description,omitempty"`
	AttributeType  string `json:"type"`
	AttributeValue string `json:"value"`

	Rates  []Rate `json:"rates,omitempty"`
	Bucket Bucket `json:"bucket"`
}

type Rate struct {
	// Duration is as specified by the configuration during initialization
	Duration int64 `json:"duration"`
	Limit    int   `json:"limit"`
}

type Bucket struct {
	Duration int64 `json:"duration"`
	Refill   int64 `json:"refill"`
	Cost     int64 `json:"cost"`
	Maximum  int64 `json:"maximum"`
}
