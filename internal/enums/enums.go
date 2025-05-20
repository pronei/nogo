package enums

type RuleAction uint8

const (
	RuleAdd RuleAction = iota
	RuleUpdate
	RuleDelete
)

type Strategy uint8

const (
	StrategyUnknown Strategy = iota
	StrategyStatic
	StrategyRolling
	StrategyFixedBucket
)

func GetStrategy(s string) Strategy {
	switch s {
	case "static_window":
		return StrategyStatic
	case "rolling_window":
		return StrategyRolling
	case "fixed_bucket":
		return StrategyFixedBucket
	default:
		return StrategyUnknown
	}
}
