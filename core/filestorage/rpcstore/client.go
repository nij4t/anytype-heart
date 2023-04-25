package rpcstore

import (
	"context"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/anytypeio/any-sync/commonfile/fileproto"
	"github.com/anytypeio/any-sync/net/rpc/rpcerr"
	"github.com/cheggaaa/mb/v3"
	"github.com/ipfs/go-cid"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"sync"
	"time"
)

func newClient(ctx context.Context, s *service, peerId string, tq *mb.MB[*task]) (*client, error) {
	c := &client{
		peerId:     peerId,
		taskQueue:  tq,
		opLoopDone: make(chan struct{}),
		stat:       newStat(),
		s:          s,
	}
	if err := c.checkConnectivity(ctx); err != nil {
		return nil, err
	}
	log.Debug("starting client for peer", zap.String("peer", peerId), zap.Strings("spaces", c.spaceIds))
	var runCtx context.Context
	runCtx, c.opLoopCtxCancel = context.WithCancel(context.Background())
	go c.opLoop(runCtx)
	return c, nil
}

// client gets and executes tasks from taskQueue
// it has an internal queue for a waiting CIDs
type client struct {
	peerId          string
	spaceIds        []string
	allowWrite      bool
	taskQueue       *mb.MB[*task]
	opLoopDone      chan struct{}
	opLoopCtxCancel context.CancelFunc
	stat            *stat
	s               *service
	mu              sync.Mutex
}

// opLoop gets tasks from taskQueue
func (c *client) opLoop(ctx context.Context) {
	defer close(c.opLoopDone)
	c.mu.Lock()
	spaceIds := c.spaceIds
	allowWrite := c.allowWrite
	c.mu.Unlock()
	cond := c.taskQueue.NewCond().WithFilter(func(t *task) bool {
		if t.write && !allowWrite {
			return false
		}
		if slices.Index(t.denyPeerIds, c.peerId) != -1 {
			return false
		}
		if len(spaceIds) > 0 && slices.Index(spaceIds, t.spaceId) == -1 {
			return false
		}
		return true
	})
	for {
		t, err := cond.WithPriority(c.stat.Score()).WaitOne(ctx)
		if err != nil {
			return
		}
		t.execWithClient(c)
	}
}

func (c *client) delete(ctx context.Context, cids ...cid.Cid) (err error) {
	p, err := c.s.pool.Get(ctx, c.peerId)
	if err != nil {
		return
	}
	cidsB := make([][]byte, 0, len(cids))
	for _, c := range cids {
		cidsB = append(cidsB, c.Bytes())
	}
	if _, err = fileproto.NewDRPCFileClient(p).BlocksDelete(ctx, &fileproto.BlocksDeleteRequest{
		SpaceId: fileblockstore.CtxGetSpaceId(ctx),
		Cids:    cidsB,
	}); err != nil {
		return rpcerr.Unwrap(err)
	}
	c.stat.UpdateLastUsage()
	return
}

func (c *client) put(ctx context.Context, cd cid.Cid, data []byte) (err error) {
	p, err := c.s.pool.Get(ctx, c.peerId)
	if err != nil {
		return
	}
	st := time.Now()
	if _, err = fileproto.NewDRPCFileClient(p).BlockPush(ctx, &fileproto.BlockPushRequest{
		SpaceId: fileblockstore.CtxGetSpaceId(ctx),
		Cid:     cd.Bytes(),
		Data:    data,
	}); err != nil {
		return rpcerr.Unwrap(err)
	}
	log.Debug("put cid", zap.String("cid", cd.String()))
	c.stat.Add(st, len(data))
	return
}

// get sends the get request to the stream and adds task to waiting list
func (c *client) get(ctx context.Context, cd cid.Cid) (data []byte, err error) {
	p, err := c.s.pool.Get(ctx, c.peerId)
	if err != nil {
		return
	}
	st := time.Now()
	resp, err := fileproto.NewDRPCFileClient(p).BlockGet(ctx, &fileproto.BlockGetRequest{
		SpaceId: fileblockstore.CtxGetSpaceId(ctx),
		Cid:     cd.Bytes(),
	})
	if err != nil {
		return nil, rpcerr.Unwrap(err)
	}
	log.Debug("get cid", zap.String("cid", cd.String()))
	c.stat.Add(st, len(resp.Data))
	return resp.Data, nil
}

func (c *client) checkBlocksAvailability(ctx context.Context, cids ...cid.Cid) ([]*fileproto.BlockAvailability, error) {
	p, err := c.s.pool.Get(ctx, c.peerId)
	if err != nil {
		return nil, err
	}
	var cidsB = make([][]byte, len(cids))
	for i, c := range cids {
		cidsB[i] = c.Bytes()
	}
	resp, err := fileproto.NewDRPCFileClient(p).BlocksCheck(ctx, &fileproto.BlocksCheckRequest{
		SpaceId: fileblockstore.CtxGetSpaceId(ctx),
		Cids:    cidsB,
	})
	if err != nil {
		return nil, err
	}
	return resp.BlocksAvailability, nil
}

func (c *client) checkConnectivity(ctx context.Context) (err error) {
	p, err := c.s.pool.Get(ctx, c.peerId)
	if err != nil {
		return
	}
	resp, err := fileproto.NewDRPCFileClient(p).Check(ctx, &fileproto.CheckRequest{})
	if err != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.spaceIds = resp.SpaceIds
	c.allowWrite = resp.AllowWrite
	return
}

func (c *client) LastUsage() time.Time {
	return c.stat.LastUsage()
}

func (c *client) Close() error {
	c.opLoopCtxCancel()
	<-c.opLoopDone
	return nil
}