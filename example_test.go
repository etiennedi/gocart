package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Gocart(t *testing.T) {
	trainingData := []TrainingItem{
		TrainingItem{
			Props: map[string]interface{}{
				"color":    "green",
				"diameter": 3,
			},
			Label: "Apple",
		},
		TrainingItem{
			Props: map[string]interface{}{
				"color":    "yellow",
				"diameter": 3,
			},
			Label: "Apple",
		},
		TrainingItem{
			Props: map[string]interface{}{
				"color":    "red",
				"diameter": 1,
			},
			Label: "Grape",
		},
		TrainingItem{
			Props: map[string]interface{}{
				"color":    "red",
				"diameter": 1,
			},
			Label: "Grape",
		},
		TrainingItem{
			Props: map[string]interface{}{
				"color":    "yellow",
				"diameter": 3,
			},
			Label: "Lemon",
		},
	}

	t.Run("helpers", func(t *testing.T) {
		// these helper methods are internals and I usually wouldn't recommend
		// testing them on their own. In fact, they're not even exported. However,
		// since the overall usecase is rather complex those tests help me assure
		// that these indiviual components work and that there is no bug hiding
		// among them which would make the overall use case harder to debug.
		//
		// One could argue that this sepearate value and testability is reason
		// enough for them to become their own use case, however, we're not
		// expecting any complex project structure, so let's keep it simple and
		// have them in here
		t.Run("uniqueProps", func(t *testing.T) {
			assert.ElementsMatch(t, []string{"color", "diameter"},
				uniqueProps(trainingData))
		})

		t.Run("uniqueValuesForProp", func(t *testing.T) {
			assert.ElementsMatch(t, []interface{}{"red", "green", "yellow"},
				uniqueValuesForProp(trainingData, "color"))
		})

		t.Run("countLabels", func(t *testing.T) {
			assert.Equal(t, map[string]int{
				"Apple": 2, "Grape": 2, "Lemon": 1,
			}, countLabels(trainingData))
		})

		t.Run("question match on string prop", func(t *testing.T) {
			q := newQuestion("color", "red")

			assert.Equal(t, false, questionMustMatch(q, trainingData[0].Props))
			assert.Equal(t, false, questionMustMatch(q, trainingData[1].Props))
			assert.Equal(t, true, questionMustMatch(q, trainingData[2].Props))
			assert.Equal(t, true, questionMustMatch(q, trainingData[3].Props))
			assert.Equal(t, false, questionMustMatch(q, trainingData[4].Props))
		})

		t.Run("question match on int prop", func(t *testing.T) {
			q := newQuestion("diameter", 3)

			assert.Equal(t, true, questionMustMatch(q, trainingData[0].Props))
			assert.Equal(t, true, questionMustMatch(q, trainingData[1].Props))
			assert.Equal(t, false, questionMustMatch(q, trainingData[2].Props))
			assert.Equal(t, false, questionMustMatch(q, trainingData[3].Props))
			assert.Equal(t, true, questionMustMatch(q, trainingData[4].Props))
		})

		t.Run("partition on question", func(t *testing.T) {
			q := newQuestion("diameter", 3)
			res, err := partition(trainingData, q)
			require.Nil(t, err)

			expectedTrue := []TrainingItem{
				trainingData[0], trainingData[1], trainingData[4],
			}
			assert.ElementsMatch(t, res.True, expectedTrue)

			expectedFalse := []TrainingItem{
				trainingData[2], trainingData[3],
			}
			assert.ElementsMatch(t, res.False, expectedFalse)

		})
	})

	BuildTree(trainingData)

}

func questionMustMatch(q *question, props map[string]interface{}) bool {
	match, err := q.match(props)
	if err != nil {
		panic(err)
	}

	return match
}
