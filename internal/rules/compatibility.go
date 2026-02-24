package rules

import "github.com/ivanjtm/YunoChallenge/internal/model"

// RuleIndex provides O(1) lookup for compatibility rules by (method, country) key.
type RuleIndex struct {
	index map[string]model.CompatibilityRule
}

// NewRuleIndex builds an index from a slice of compatibility rules.
func NewRuleIndex(rules []model.CompatibilityRule) *RuleIndex {
	idx := &RuleIndex{
		index: make(map[string]model.CompatibilityRule, len(rules)),
	}
	for _, r := range rules {
		k := key(r.OriginalMethod, r.Country)
		idx.index[k] = r
	}
	return idx
}

func key(method model.PaymentMethod, country model.Country) string {
	return string(method) + ":" + string(country)
}

// Lookup returns the compatibility rule for the given method+country, or nil if none.
func (ri *RuleIndex) Lookup(method model.PaymentMethod, country model.Country) *model.CompatibilityRule {
	k := key(method, country)
	rule, ok := ri.index[k]
	if !ok {
		return nil
	}
	return &rule
}

// AllowedRefundMethods returns all allowed refund methods for a given original payment method and country.
func (ri *RuleIndex) AllowedRefundMethods(method model.PaymentMethod, country model.Country) []model.AllowedRefund {
	rule := ri.Lookup(method, country)
	if rule == nil {
		return nil
	}
	return rule.AllowedRefunds
}
