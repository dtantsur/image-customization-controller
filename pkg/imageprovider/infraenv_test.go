package imageprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonEqual(t *testing.T) {
	first := `
{
    "version": "3.2.0",
    "nested": [
	{
	    "field1": "value1",
	    "field2": "value2"
	},
	{
	    "field3": null
	}
    ]
}
`
	second := `
{
    "nested": [
	{
	    "field2": "value2",
	    "field1": "value1"
	},
	{
	    "field3": null
	}
    ],
    "version": "3.2.0"
}
`

	equal, err := jsonEqual(first, second)
	assert.NoError(t, err)
	assert.True(t, equal)
}

func TestJsonNotEqual(t *testing.T) {
	first := `
{
    "version": "3.2.0",
    "nested": [
	{
	    "field1": "value1",
	    "field2": "value2"
	},
	{
	    "field3": null
	}
    ]
}
`
	second := `
{
    "nested": [
	{
	    "field2": "value2"
	},
	{
	    "field3": null
	}
    ],
    "version": "3.2.0"
}
`

	equal, err := jsonEqual(first, second)
	assert.NoError(t, err)
	assert.False(t, equal)
}
