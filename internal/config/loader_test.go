package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
)

func TestLoadParams(t *testing.T) {
	// Create a temporary params file
	paramsJSON := []byte(`{
  "version": "1.0",
  "categories": [
    {
      "name": "Test",
      "params": [
        {"short_flag": "-m", "long_flag": "--model FNAME", "description_ru": "test"}
      ]
    }
  ],
  "total_params_count": 1,
  "source_docs": ["http://example.com"]
}`)

	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(paramsJSON)
	require.NoError(t, err)

	pf, err := LoadParams(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, pf)
	require.Len(t, pf.Categories, 1)
	require.Len(t, pf.Categories[0].Params, 1)
	assert.Equal(t, "-m", pf.Categories[0].Params[0].ShortFlag)
	assert.Equal(t, "--model FNAME", pf.Categories[0].Params[0].LongFlag)
	assert.Equal(t, "test", pf.Categories[0].Params[0].DescRU)
}

func TestLoadParams_InvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte("invalid json"))
	require.NoError(t, err)

	_, err = LoadParams(tmpFile.Name())
	assert.Error(t, err)
}

func TestLoadParams_NonexistentFile(t *testing.T) {
	_, err := LoadParams("/nonexistent/path")
	assert.Error(t, err)
}

func TestLoadParams_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(`{}`))
	require.NoError(t, err)

	pf, err := LoadParams(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, pf)
	assert.Empty(t, pf.Categories)
}

func TestLoadParams_MalformedJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write malformed JSON
	_, err = tmpFile.Write([]byte(`{ "categories": [ { "name": "Test" } } `))
	require.NoError(t, err)

	_, err = LoadParams(tmpFile.Name())
	assert.Error(t, err)
}

func TestLoadParams_EmptyCategories(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(`{
  "version": "1.0",
  "categories": []
}`))
	require.NoError(t, err)

	pf, err := LoadParams(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, pf)
	assert.Empty(t, pf.Categories)
}

func TestLoadParams_LargeCategories(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write params with 120 categories
	var b strings.Builder
	b.WriteString(`{
  "version": "1.0",
  "categories": [`)
	for i := 0; i < 120; i++ {
		name := fmt.Sprintf("Test%c", 'a'+i)
		desc := fmt.Sprintf("test%c", 'a'+i)
		fmt.Fprintf(&b, `{"name": "%s","params": [{"short_flag": "-m", "long_flag": "--model FNAME", "description_ru": "%s"}]}`, name, desc)
		if i < 119 {
			b.WriteByte(',')
		}
	}
	b.WriteString(`]}`)
	_, err = tmpFile.Write([]byte(b.String()))
	require.NoError(t, err)

	pf, err := LoadParams(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, pf)
	require.Len(t, pf.Categories, 120)
}
