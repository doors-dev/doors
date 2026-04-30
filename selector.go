package doors

import "github.com/doors-dev/doors/internal/front"

// Selector chooses which DOM nodes an [Indicators] should target.
type Selector = front.Selector

// SelectorTarget selects the element that triggered the event.
func SelectorTarget() Selector {
	return front.SelectTarget()
}

// SelectorQuery selects the first element matching query.
func SelectorQuery(query string) Selector {
	return front.SelectQuery(query)
}

// SelectorQueryAll selects every element matching query.
func SelectorQueryAll(query string) Selector {
	return front.SelectQueryAll(query)
}

// SelectorQueryParent selects the closest ancestor matching query.
func SelectorQueryParent(query string) Selector {
	return front.SelectQueryParent(query)
}
