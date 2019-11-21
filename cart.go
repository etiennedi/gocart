package main

import (
	"fmt"
	"math"
)

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

func (q *question) String() string {
	operator := ">="
	if _, ok := q.value.(string); ok {
		operator = "=="
	}
	return fmt.Sprintf("Is %s %s %v?", q.prop, operator, q.value)

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

func giniImpurity(set []TrainingItem) float64 {
	counts := countLabels(set)
	imp := 1.0
	for _, count := range counts {
		probability := float64(count) / float64(len(set))
		imp -= float64(math.Pow(probability, 2))
	}

	return imp
}

func informationGain(parts partitionResult,
	previousUncertainty float64) float64 {
	weightTrue := float64(len(parts.True)) / float64((len(parts.True) + len(parts.False)))

	return previousUncertainty - weightTrue*giniImpurity(parts.True) -
		(1-weightTrue)*giniImpurity(parts.False)
}

func findBestQuestion(set []TrainingItem) (*question, float64, error) {
	var bestGain float64
	var bestQuestion *question
	previousImpurity := giniImpurity(set)

	for _, prop := range uniqueProps(set) {
		values := uniqueValuesForProp(set, prop)

		for _, value := range values {
			q := newQuestion(prop, value)
			parts, err := partition(set, q)
			if err != nil {
				return nil, 0, fmt.Errorf("prop '%v', value '%v': partition: %v",
					prop, value, err)
			}

			if len(parts.True) == 0 || len(parts.False) == 0 {
				// skip if no split occurred
				continue
			}

			gain := informationGain(parts, previousImpurity)
			if gain >= bestGain {
				bestQuestion, bestGain = q, gain
			}
		}
	}

	return bestQuestion, bestGain, nil
}

// A Node can either be a Leaf or a Branch
type Node interface {
	IsLeaf() bool
	String(indent string) string
}

// BuildTree recursively builds a decision tree, starting with the question
// with the highest possible information gain at the root
func BuildTree(data []TrainingItem) (Node, error) {
	question, gain, err := findBestQuestion(data)
	if err != nil {
		return nil, fmt.Errorf("find best question: %v", err)
	}

	if gain == 0 {
		// exit condition for recursion, there is no more information to be gained,
		// return a leaf
		return newLeaf(data), nil
	}

	parts, err := partition(data, question)
	if err != nil {
		return nil, fmt.Errorf("partition on question '%s': %v", question, err)
	}

	trueBranch, err := BuildTree(parts.True)
	if err != nil {
		return nil, fmt.Errorf("question '%s' true branch: %v", question, err)
	}

	falseBranch, err := BuildTree(parts.False)
	if err != nil {
		return nil, fmt.Errorf("question '%s' true branch: %v", question, err)
	}

	return newDecisionNode(question, trueBranch, falseBranch), nil
}

// Leaf is an endpoint in the tree. Opposite of Branch.
type Leaf struct {
	data []TrainingItem
}

// IsLeaf is always true for a leave
func (l *Leaf) IsLeaf() bool {
	return true
}

func (l *Leaf) String(indent string) string {
	return fmt.Sprintf("%sPrediction: %v", indent, countLabels(l.data))
}

func newLeaf(data []TrainingItem) *Leaf {
	return &Leaf{data}
}

// DecisionNode splits up into a true and false branch based on the attached
// quetion
type DecisionNode struct {
	TrueBranch  Node
	FalseBranch Node
	Question    *question
}

// IsLeaf is never true for a decision node
func (d *DecisionNode) IsLeaf() bool {
	return false
}

func (d *DecisionNode) String(indent string) string {
	return fmt.Sprintf("%s%s\n%s-->True: \n%s\n%s-->False: \n%s",
		indent, d.Question, indent, d.TrueBranch.String(indent+"  "),
		indent, d.FalseBranch.String(indent+"  "))

}

func newDecisionNode(question *question, trueBranch,
	falseBranch Node) *DecisionNode {
	return &DecisionNode{
		TrueBranch:  trueBranch,
		FalseBranch: falseBranch,
		Question:    question,
	}
}
