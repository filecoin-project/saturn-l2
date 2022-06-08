package carstore

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/filecoin-project/saturn-l2/testutils"

	cid "github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

var root = "QmfMYyn8LUWEfRXfijKFjBAshSsPVRUgwLZzsD7kcTtX1A"

func TestGatewayAPI(t *testing.T) {
	ctx := context.Background()

	bz := []byte("hello")
	svc := testutils.GetTestServer(t, root, bz)
	defer svc.Close()

	gw := &gatewayAPI{
		baseURL: svc.URL,
	}

	c, err := cid.Decode(root)
	require.NoError(t, err)

	rd, err := gw.Fetch(ctx, c)
	require.NoError(t, err)
	require.NotEmpty(t, rd)

	out, err := ioutil.ReadAll(rd)
	require.NoError(t, err)
	require.EqualValues(t, bz, out)
}

func TestGatewayAPIFailure(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svc.Close()

	ctx := context.Background()
	gw := &gatewayAPI{
		baseURL: svc.URL,
	}

	c, err := cid.Decode(root)
	require.NoError(t, err)

	rd, err := gw.Fetch(ctx, c)
	require.Error(t, err)
	require.Empty(t, rd)
}
