package engine

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestFilterExtractor(t *testing.T) {

	u := &url.URL{
		RawPath:  `https://127.0.0.1/api/v1/areas?filter={"ids":[1494749745,1494749740,1494749741],"level":"WARN","q":"test_search_string"}&range=[0,9]&sort=["id","ASC"]`,
		RawQuery: `filter={"ids":[1494749745,1494749740,1494749741],"level":"WARN","q":"test_search_string"}&range=[0,9]&sort=["id","ASC"]`,
	}

	f, err := FilterFromUrlExtractor(u)

	assert.NoError(t, err)
	assert.Len(t, f.Sort, 2)
	assert.Equal(t, f.Range[1], int64(10)) // max range value +1, because last index exclude from fetched data set
	assert.Equal(t, f.Filters["level"], "WARN")
	assert.Equal(t, f.Filters["q"], "test_search_string")
	require.Len(t, f.IDs, 3)

	// test with error
	u = &url.URL{
		RawPath:  `https://127.0.0.1/api/v1/areas?filter={"ids":["1494749745"],"level":"WARN","q":"test_search_string"}&range=[0,9]&sort=["id","ASC"]`,
		RawQuery: `filter={"ids":["1494749745"],"level":"WARN","q":"test_search_string"}&range=[0,9]&sort=["id","ASC"]`,
	}

	f, err = FilterFromUrlExtractor(u)
	assert.Error(t, err)

}
