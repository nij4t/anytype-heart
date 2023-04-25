package rpcstore

import (
	"context"
	"fmt"
	"github.com/anytypeio/any-sync/accountservice/mock_accountservice"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/anytypeio/any-sync/commonfile/fileproto"
	"github.com/anytypeio/any-sync/commonfile/fileproto/fileprotoerr"
	"github.com/anytypeio/any-sync/commonspace/object/accountdata"
	"github.com/anytypeio/any-sync/net/rpc/rpctest"
	"github.com/anytypeio/any-sync/nodeconf"
	"github.com/golang/mock/gomock"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sort"
	"sync"
	"testing"
	"time"
)

var ctx = context.Background()

func TestStore_Put(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	bs := []blocks.Block{
		blocks.NewBlock([]byte{'1'}),
		blocks.NewBlock([]byte{'2'}),
		blocks.NewBlock([]byte{'3'}),
	}
	err := fx.Add(ctx, bs)
	assert.NoError(t, err)
	for _, b := range bs {
		assert.NotNil(t, fx.serv.data[string(b.Cid().Bytes())])
	}

}

func TestStore_Delete(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	bs := []blocks.Block{
		blocks.NewBlock([]byte{'1'}),
	}
	err := fx.Add(ctx, bs)
	require.NoError(t, err)
	assert.Len(t, fx.serv.data, 1)
	require.NoError(t, fx.Delete(ctx, bs[0].Cid()))
	assert.Len(t, fx.serv.data, 0)
}

func TestStore_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := []blocks.Block{
			blocks.NewBlock([]byte{'1'}),
		}
		err := fx.Add(ctx, bs)
		require.NoError(t, err)
		b, err := fx.Get(ctx, bs[0].Cid())
		require.NoError(t, err)
		assert.Equal(t, []byte{'1'}, b.RawData())
	})
	t.Run("not found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := []blocks.Block{
			blocks.NewBlock([]byte{'1'}),
		}
		b, err := fx.Get(ctx, bs[0].Cid())
		assert.Nil(t, b)
		assert.ErrorIs(t, err, fileblockstore.ErrCIDNotFound)
	})
}

func TestStore_GetMany(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	bs := []blocks.Block{
		blocks.NewBlock([]byte{'1'}),
		blocks.NewBlock([]byte{'2'}),
		blocks.NewBlock([]byte{'3'}),
	}
	err := fx.Add(ctx, bs)
	assert.NoError(t, err)

	res := fx.GetMany(ctx, []cid.Cid{
		bs[0].Cid(),
		bs[1].Cid(),
		bs[2].Cid(),
	})
	var resBlocks []blocks.Block
	for b := range res {
		resBlocks = append(resBlocks, b)
	}
	require.Len(t, resBlocks, 3)
	sort.Slice(resBlocks, func(i, j int) bool {
		return string(resBlocks[i].RawData()) < string(resBlocks[j].RawData())
	})
	assert.Equal(t, bs, resBlocks)
}

func TestStore_AddAsync(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	bs := []blocks.Block{
		blocks.NewBlock([]byte{'1'}),
		blocks.NewBlock([]byte{'2'}),
		blocks.NewBlock([]byte{'3'}),
	}
	err := fx.Add(ctx, bs[:1])
	assert.NoError(t, err)

	successCh := fx.AddAsync(ctx, bs)
	var successCids []cid.Cid
	for i := 0; i < len(bs); i++ {
		select {
		case <-time.After(time.Second):
			require.True(t, false, "timeout")
		case c := <-successCh:
			successCids = append(successCids, c)
		}
	}
	assert.Len(t, successCids, 3)
}

func newFixture(t *testing.T) *fixture {
	fx := &fixture{
		a: new(app.App),
		s: New().(*service),
		serv: &testServer{
			data: make(map[string][]byte),
		},
	}

	conf := &config{}

	for i := 0; i < 11; i++ {
		conf.Nodes = append(conf.Nodes, nodeconf.NodeConfig{
			PeerId: fmt.Sprint(i),
			Types:  []nodeconf.NodeType{nodeconf.NodeTypeFile},
		})
	}
	rserv := rpctest.NewTestServer()
	require.NoError(t, fileproto.DRPCRegisterFile(rserv.Mux, fx.serv))
	fx.ctrl = gomock.NewController(t)
	fx.a.Register(fx.s).
		Register(mock_accountservice.NewAccountServiceWithAccount(fx.ctrl, &accountdata.AccountData{})).
		Register(rpctest.NewTestPool().WithServer(rserv)).
		Register(nodeconf.New()).
		Register(conf)
	require.NoError(t, fx.a.Start(ctx))
	fx.store = fx.s.NewStore().(*store)
	return fx
}

type fixture struct {
	*store
	s    *service
	a    *app.App
	serv *testServer
	ctrl *gomock.Controller
}

func (fx *fixture) Finish(t *testing.T) {
	assert.NoError(t, fx.store.Close())
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}

type testServer struct {
	mu   sync.Mutex
	data map[string][]byte
}

func (t *testServer) BlockGet(ctx context.Context, req *fileproto.BlockGetRequest) (resp *fileproto.BlockGetResponse, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if data, ok := t.data[string(req.Cid)]; ok {
		return &fileproto.BlockGetResponse{
			Cid:  req.Cid,
			Data: data,
		}, nil
	} else {
		return nil, fileprotoerr.ErrCIDNotFound
	}
}

func (t *testServer) BlockPush(ctx context.Context, req *fileproto.BlockPushRequest) (*fileproto.BlockPushResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.data[string(req.Cid)] = req.Data
	return &fileproto.BlockPushResponse{}, nil
}

func (t *testServer) BlocksDelete(ctx context.Context, req *fileproto.BlocksDeleteRequest) (*fileproto.BlocksDeleteResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, c := range req.Cids {
		delete(t.data, string(c))
	}
	return &fileproto.BlocksDeleteResponse{}, nil
}

func (t *testServer) BlocksCheck(ctx context.Context, req *fileproto.BlocksCheckRequest) (resp *fileproto.BlocksCheckResponse, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	resp = &fileproto.BlocksCheckResponse{}
	for _, c := range req.Cids {
		status := fileproto.AvailabilityStatus_NotExists
		if _, ok := t.data[string(c)]; ok {
			status = fileproto.AvailabilityStatus_Exists
		}
		resp.BlocksAvailability = append(resp.BlocksAvailability, &fileproto.BlockAvailability{
			Cid:    c,
			Status: status,
		})
	}
	return
}

func (t *testServer) BlocksBind(ctx context.Context, request *fileproto.BlocksBindRequest) (*fileproto.BlocksBindResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (t *testServer) Check(ctx context.Context, req *fileproto.CheckRequest) (*fileproto.CheckResponse, error) {
	return &fileproto.CheckResponse{AllowWrite: true}, nil
}

type config struct {
	Nodes []nodeconf.NodeConfig
}

func (c config) Init(a *app.App) (err error) {
	return
}

func (c config) Name() string { return "config" }

func (c config) GetNodes() []nodeconf.NodeConfig {
	return c.Nodes
}