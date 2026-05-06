package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateParams(t *testing.T) {
	// Create a valid params file
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

	err = ValidateParams(pf)
	assert.NoError(t, err)
}

func TestValidateParams_NilParams(t *testing.T) {
	err := ValidateParams(nil)
	assert.Error(t, err)
}

func TestValidateParams_EmptyCategories(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write params with empty categories
	_, err = tmpFile.Write([]byte(`{"version": "1.0", "categories": []}`))
	require.NoError(t, err)

	_, err = LoadParams(tmpFile.Name())
	// LoadParams should succeed, then ValidateParams should fail
	assert.NoError(t, err)

	err = ValidateParams(nil)
	assert.Error(t, err)
}

func TestValidateParams_MissingDescription(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write params with empty description
	_, err = tmpFile.Write([]byte(`{
  "version": "1.0",
  "categories": [
    {
      "name": "Test",
      "params": [
        {"short_flag": "-m", "long_flag": "--model FNAME", "description_ru": ""}
      ]
    }
  ],
  "total_params_count": 1,
  "source_docs": ["http://example.com"]
}`))
	require.NoError(t, err)

	pf, err := LoadParams(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, pf)

	err = ValidateParams(pf)
	assert.Error(t, err)
}

func TestValidateParams_MultipleCategories(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write params with multiple categories
	_, err = tmpFile.Write([]byte(`{
  "version": "1.0",
  "categories": [
    {
      "name": "Test1",
      "params": [
        {"short_flag": "-m", "long_flag": "--model FNAME", "description_ru": "test1"}
      ]
    },
    {
      "name": "Test2",
      "params": [
        {"short_flag": "-t", "long_flag": "--threads N", "description_ru": "test2"}
      ]
    }
  ],
  "total_params_count": 2,
  "source_docs": ["http://example.com"]
}`))
	require.NoError(t, err)

	pf, err := LoadParams(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, pf)

	err = ValidateParams(pf)
	assert.NoError(t, err)
}

func TestValidateParams_EmptyDescriptionInMultipleCategories(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "params-test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write params with empty description in second category
	_, err = tmpFile.Write([]byte(`{
  "version": "1.0",
  "categories": [
    {
      "name": "Test1",
      "params": [
        {"short_flag": "-m", "long_flag": "--model FNAME", "description_ru": "test1"}
      ]
    },
    {
      "name": "Test2",
      "params": [
        {"short_flag": "-t", "long_flag": "--threads N", "description_ru": ""}
      ]
    }
  ],
  "total_params_count": 2,
  "source_docs": ["http://example.com"]
}`))
	require.NoError(t, err)

	pf, err := LoadParams(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, pf)

	err = ValidateParams(pf)
	assert.Error(t, err)
}
