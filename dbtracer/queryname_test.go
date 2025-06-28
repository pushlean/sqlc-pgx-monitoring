package dbtracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryNameExtractionEdgeCases(t *testing.T) {
	testCases := []struct {
		name         string
		sql          string
		expectedName string
		expectedType string
	}{
		{
			name:         "query without name comment",
			sql:          "SELECT * FROM users",
			expectedName: "unknown",
			expectedType: "unknown",
		},
		{
			name:         "query with malformed comment",
			sql:          "-- name: malformed\nSELECT * FROM users",
			expectedName: "unknown",
			expectedType: "unknown",
		},
		{
			name:         "empty sql",
			sql:          "",
			expectedName: "unknown",
			expectedType: "unknown",
		},
		{
			name:         "block comment style",
			sql:          "/* name: block_comment_query :one */\nSELECT * FROM users",
			expectedName: "block_comment_query",
			expectedType: "one",
		},
	}

	for _, tc := range testCases {
		name, queryType := queryNameFromSQL(tc.sql)
		assert.Equal(t, tc.expectedName, name)
		assert.Equal(t, tc.expectedType, queryType)
	}
}
