package main

import "fmt"

// TrainingItem is a single item of training data
type TrainingItem struct {
	Props map[string]interface{}
	Label string
}

func uniqueProps(data []TrainingItem) []string {
	var out []string
	props := map[string]int{}
	for _, item := range data {
		for prop := range item.Props {
			_, ok := props[prop]
			if !ok {
				out = append(out, prop)
			}
			props[prop] = 1
		}
	}

	return out
}

func uniqueValuesForProp(data []TrainingItem, prop string) []interface{} {
	var out []interface{}
	values := map[interface{}]int{}
	for _, item := range data {
		value, ok := item.Props[prop]
		if !ok {
			continue
		}

		_, ok = values[value]
		if !ok {
			values[value] = 1
			out = append(out, value)
		}
	}

	return out
}

func countLabels(data []TrainingItem) map[string]int {
	labels := map[string]int{}
	for _, item := range data {
		count, ok := labels[item.Label]
		if !ok {
			labels[item.Label] = 1
		} else {
			labels[item.Label] = count + 1
		}
	}

	return labels
}

type question struct {
	prop  string
	value interface{}
}

func newQuestion(prop string, value interface{}) *question {
	return &question{prop, value}
}

func (q *question) match(props map[string]interface{}) (bool, error) {
	target, ok := props[q.prop]
	if !ok {
		// target doesn't have this prop set at all, let's consider this a
		// non-match. Alternatively you could also error here if you want to
		// enforce that all props must always be set
		return false, nil
	}

	switch q.value.(type) {
	case string:
		return q.matchString(target)
	case int:
		return q.matchInt(target)
	case float64:
		return q.matchFloat64(target)
	default:
		return false, fmt.Errorf("unsupported type %T", q.value)
	}
}

func (q *question) matchString(target interface{}) (bool, error) {
	targetString, ok := target.(string)
	if !ok {
		return false, fmt.Errorf(
			"question expected prop '%s' to be string, but got: %T",
			q.prop, target,
		)
	}

	return q.value.(string) == targetString, nil
}

func (q *question) matchInt(target interface{}) (bool, error) {
	targetInt, ok := target.(int)
	if !ok {
		return false, fmt.Errorf(
			"question expected prop '%s' to be int, but got: %T",
			q.prop, target,
		)
	}

	return targetInt >= q.value.(int), nil
}

func (q *question) matchFloat64(target interface{}) (bool, error) {
	targetFloat, ok := target.(float64)
	if !ok {
		return false, fmt.Errorf(
			"question expected prop '%s' to be int, but got: %T",
			q.prop, target,
		)
	}

	return targetFloat >= q.value.(float64), nil
}

type partitionResult struct {
	True  []TrainingItem
	False []TrainingItem
}

func partition(data []TrainingItem, q *question) (partitionResult, error) {
	res := partitionResult{}
	for i, item := range data {
		match, err := q.match(item.Props)
		if err != nil {
			return res, fmt.Errorf("element %d: %v", i, err)
		}

		if match {
			res.True = append(res.True, item)
		} else {
			res.False = append(res.False, item)
		}
	}

	return res, nil
}

func BuildTree(data []TrainingItem) {

}
