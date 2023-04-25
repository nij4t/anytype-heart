package syncstatus

import (
	"context"
	"sync"
	"time"

	"github.com/anytypeio/any-sync/app"
	"go.uber.org/zap"

	"github.com/anytypeio/go-anytype-middleware/core/anytype/config"
	"github.com/anytypeio/go-anytype-middleware/core/block/getblock"
	"github.com/anytypeio/go-anytype-middleware/core/filestorage/filesync"
	"github.com/anytypeio/go-anytype-middleware/pb"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/core"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/core/smartblock"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/localstore/filestore"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/logging"
	"github.com/anytypeio/go-anytype-middleware/space"
	"github.com/anytypeio/go-anytype-middleware/space/typeprovider"
)

var log = logging.Logger("anytype-mw-status")

const CName = "status"

type Service interface {
	Watch(id string, fileFunc func() []string) (new bool, err error)
	Unwatch(id string)
	app.ComponentRunnable
}

var _ Service = (*service)(nil)

type service struct {
	typeProvider typeprovider.SmartBlockTypeProvider
	spaceService space.Service

	coreService core.Service

	fileWatcher        *fileWatcher
	objectWatcher      *objectWatcher
	subObjectsWatcher  *subObjectsWatcher
	linkedFilesWatcher *linkedFilesWatcher

	isRunning bool

	sync.Mutex
}

func New(
	typeProvider typeprovider.SmartBlockTypeProvider,
	spaceService space.Service,
	coreService core.Service,
	fileSyncService filesync.FileSync,
	fileStore filestore.FileStore,
	picker getblock.Picker,
	cfg *config.Config,
	sendEvent func(event *pb.Event),
	fileWatcherUpdateInterval time.Duration,
) Service {
	fileStatusRegistry := newFileStatusRegistry(fileSyncService, fileStore, picker, fileWatcherUpdateInterval)
	linkedFilesWatcher := newLinkedFilesWatcher(spaceService, fileStatusRegistry)
	subObjectsWatcher := newSubObjectsWatcher()
	updateReceiver := newUpdateReceiver(coreService, linkedFilesWatcher, subObjectsWatcher, cfg, sendEvent)
	fileWatcher := newFileWatcher(fileStatusRegistry, updateReceiver, fileWatcherUpdateInterval)
	objectWatcher := newObjectWatcher(spaceService, updateReceiver)
	return &service{
		spaceService:       spaceService,
		typeProvider:       typeProvider,
		coreService:        coreService,
		fileWatcher:        fileWatcher,
		objectWatcher:      objectWatcher,
		subObjectsWatcher:  subObjectsWatcher,
		linkedFilesWatcher: linkedFilesWatcher,
	}
}

func (s *service) Init(a *app.App) (err error) {
	return
}

func (s *service) Run(ctx context.Context) (err error) {
	s.Lock()
	defer s.Unlock()
	s.isRunning = true

	go s.fileWatcher.run()

	if err = s.objectWatcher.run(ctx); err != nil {
		return err
	}

	_, err = s.watch(s.coreService.PredefinedBlocks().Account, nil)
	return
}

func (s *service) Name() string {
	return CName
}

func (s *service) Watch(id string, filesGetter func() []string) (new bool, err error) {
	s.Lock()
	defer s.Unlock()
	return s.watch(id, filesGetter)
}

func (s *service) Unwatch(id string) {
	s.Lock()
	defer s.Unlock()
	s.unwatch(id)
}

func (s *service) watch(id string, filesGetter func() []string) (new bool, err error) {
	if !s.isRunning {
		return false, nil
	}
	sbt, err := s.typeProvider.Type(id)
	if err != nil {
		log.Debug("failed to get type of", zap.String("objectID", id))
	}
	switch sbt {
	case smartblock.SmartBlockTypeFile:
		s.fileWatcher.Watch(s.spaceService.AccountId(), id)
		return false, nil
	case smartblock.SmartBlockTypeSubObject:
		s.subObjectsWatcher.Watch(id)
		return true, nil
	default:
		s.linkedFilesWatcher.WatchLinkedFiles(id, filesGetter)
		if err = s.objectWatcher.Watch(id); err != nil {
			return false, err
		}
		return true, nil
	}
}

func (s *service) unwatch(id string) {
	if !s.isRunning {
		return
	}
	sbt, err := s.typeProvider.Type(id)
	if err != nil {
		log.Debug("failed to get type of", zap.String("objectID", id))
	}
	switch sbt {
	case smartblock.SmartBlockTypeFile:
		s.fileWatcher.Unwatch(s.spaceService.AccountId(), id)
	case smartblock.SmartBlockTypeSubObject:
		s.subObjectsWatcher.Unwatch(id)
	default:
		s.linkedFilesWatcher.UnwatchLinkedFiles(id)
		s.objectWatcher.Unwatch(id)
	}
}

func (s *service) Close(ctx context.Context) (err error) {
	s.Lock()
	defer s.Unlock()
	s.isRunning = false
	s.unwatch(s.coreService.PredefinedBlocks().Account)
	s.fileWatcher.close()
	s.linkedFilesWatcher.close()

	return nil
}