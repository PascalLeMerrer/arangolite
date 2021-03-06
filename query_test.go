package arangolite

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const shortQuery = `
    FOR d
    IN documents
    RETURN {
        document: d,
        metaData: (
            FOR m
            IN metaData
            FILTER m.documentId == d._id
            RETURN m
        )
    }
`

// TestQueryRun runs tests on the Query Run method.
func TestQueryRun(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	db := New().LoggerOptions(false, false, false)
	db.Connect("http://arangodb:8000", "dbName", "foo", "bar")

	result, err := db.Run(NewQuery(shortQuery))
	r.Error(err)
	a.Nil(result)

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/_api/cursor",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": []}`))

	result, err = db.Run(NewQuery(""))
	r.NoError(err)
	a.Equal("[]", string(result))

	result, err = db.Run(NewQuery(shortQuery).Cache(true).BatchSize(500))
	r.NoError(err)
	a.Equal("[]", string(result))

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/_api/cursor",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": [{}], "hasMore":true, "id":"1000"}`))

	httpmock.RegisterResponder("PUT", "http://arangodb:8000/_db/dbName/_api/cursor/1000",
		httpmock.NewStringResponder(200, `{"error": false, "errorMessage": "", "result": [{}], "hasMore":false}`))

	result, err = db.Run(NewQuery(""))
	r.NoError(err)
	a.Equal("[{},{}]", string(result))

	q := NewQuery(`
	    FOR d
	    IN documents
	    FILTER d._key == @key
	    RETURN d
	    `)
	q.Bind("key", 1000)
	result, err = db.Run(q)

	r.NoError(err)
	a.Equal("[{}]", string(result))

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/_api/cursor",
		httpmock.NewStringResponder(500, `{"error": true, "errorMessage": "error !"}`))

	result, err = db.Run(NewQuery(shortQuery))
	r.Error(err)
	a.Nil(result)
}
