package arangolite

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransactionRun runs tests on the Transaction Run method.
func TestTransactionRun(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	db := New(false)
	db.Connect("http://arangodb:8000", "dbName", "foo", "bar")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://arangodb:8000/_db/dbName/_api/transaction",
		func(r *http.Request) (*http.Response, error) {
			buffer, _ := ioutil.ReadAll(r.Body)
			str := strings.Replace(string(buffer), `"`, `'`, -1)
			return httpmock.NewStringResponse(200, `{"error": false, "errorMessage": "", "result": "`+str+`"}`), nil
		})

	result, err := NewTransaction(nil, nil).Run(nil)
	r.Error(err)
	a.Nil(result)

	result, err = NewTransaction([]string{"foo"}, []string{"bar"}).
		AddQuery("var1", "FOR c IN customer RETURN c").
		AddQuery("var2", "FOR c IN {{.var1}} RETURN c").
		Return("var1").Run(db)
	r.NoError(err)
	a.Equal(`"{'collections':{'read':['foo'],'write':['bar']},'action':'function () {var db = require('internal').db; var var1 = db._query('FOR c IN customer RETURN c'); var var2 = db._query('FOR c IN ' + JSON.stringify(var1._documents) + ' RETURN c'); return var1;'}"`, string(result))
}