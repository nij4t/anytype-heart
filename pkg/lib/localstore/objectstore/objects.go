package objectstore

import (
	"encoding/binary"
	"fmt"
	"github.com/anytypeio/go-anytype-middleware/app"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/datastore"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/localstore"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/localstore/addr"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/logging"
	"github.com/anytypeio/go-anytype-middleware/util/slice"
	"strings"
	"sync"
	"time"

	"github.com/anytypeio/go-anytype-middleware/pkg/lib/bundle"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/core/smartblock"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/database"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/localstore/ftsearch"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/pb"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/pb/model"
	pbrelation "github.com/anytypeio/go-anytype-middleware/pkg/lib/pb/relation"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/schema"
	"github.com/anytypeio/go-anytype-middleware/util/pbtypes"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

var log = logging.Logger("anytype-localstore")

const CName = "objectstore"

var (
	// ObjectInfo is stored in db key pattern:
	pagesPrefix        = "pages"
	pagesDetailsBase   = ds.NewKey("/" + pagesPrefix + "/details")
	pagesRelationsBase = ds.NewKey("/" + pagesPrefix + "/relations")

	pagesSnippetBase       = ds.NewKey("/" + pagesPrefix + "/snippet")
	pagesInboundLinksBase  = ds.NewKey("/" + pagesPrefix + "/inbound")
	pagesOutboundLinksBase = ds.NewKey("/" + pagesPrefix + "/outbound")
	indexQueueBase         = ds.NewKey("/" + pagesPrefix + "/index")

	relationsPrefix = "relations"
	// /relations/options/<relOptionId>: option model
	relationsOptionsBase = ds.NewKey("/" + relationsPrefix + "/options")
	// /relations/relations/<relKey>: relation model
	relationsBase = ds.NewKey("/" + relationsPrefix + "/relations")

	indexObjectTypeRelationObjectId = localstore.Index{
		Prefix: relationsPrefix,
		Name:   "objtype_relkey_objid",
		Keys: func(val interface{}) []localstore.IndexKeyParts {
			if v, ok := val.(*relationObjectType); ok {
				var indexes []localstore.IndexKeyParts
				for _, rk := range v.relationKeys {
					for _, ot := range v.objectTypes {
						otCompact, err := objTypeCompactEncode(ot)
						if err != nil {
							log.Errorf("objtype_relkey_objid index construction error(ot '%s'): %s", ot, err.Error())
							continue
						}

						indexes = append(indexes, localstore.IndexKeyParts([]string{otCompact, rk}))
					}
				}
				return indexes
			}
			return nil
		},
		Unique:             false,
		SplitIndexKeyParts: true,
	}

	indexObjectTypeRelationSetId = localstore.Index{
		Prefix: relationsPrefix,
		Name:   "objtype_relkey_setid",
		Keys: func(val interface{}) []localstore.IndexKeyParts {
			if v, ok := val.(*relationObjectType); ok {
				var indexes []localstore.IndexKeyParts
				for _, rk := range v.relationKeys {
					for _, ot := range v.objectTypes {
						otCompact, err := objTypeCompactEncode(ot)
						if err != nil {
							log.Errorf("objtype_relkey_setid index construction error('%s'): %s", ot, err.Error())
							continue
						}

						indexes = append(indexes, localstore.IndexKeyParts([]string{otCompact, rk}))
					}
				}
				return indexes
			}
			return nil
		},
		Unique:             false,
		SplitIndexKeyParts: true,
	}

	indexRelationOptionObject = localstore.Index{
		Prefix: pagesPrefix,
		Name:   "relkey_optid",
		Keys: func(val interface{}) []localstore.IndexKeyParts {
			if v, ok := val.(*pbrelation.Relation); ok {
				var indexes []localstore.IndexKeyParts
				if v.Format != pbrelation.RelationFormat_tag && v.Format != pbrelation.RelationFormat_status {
					return nil
				}
				if len(v.SelectDict) == 0 {
					return nil
				}

				for _, opt := range v.SelectDict {
					indexes = append(indexes, localstore.IndexKeyParts([]string{v.Key, opt.Id}))
				}
				return indexes
			}
			return nil
		},
		Unique:             false,
		SplitIndexKeyParts: true,
	}

	indexRelationObject = localstore.Index{
		Prefix: pagesPrefix,
		Name:   "relkey",
		Keys: func(val interface{}) []localstore.IndexKeyParts {
			if v, ok := val.(*pbrelation.Relation); ok {
				return []localstore.IndexKeyParts{[]string{v.Key}}
			}
			return nil
		},
		Unique: false,
	}

	indexFormatOptionObject = localstore.Index{
		Prefix: pagesPrefix,
		Name:   "format_relkey_optid",
		Keys: func(val interface{}) []localstore.IndexKeyParts {
			if v, ok := val.(*pbrelation.Relation); ok {
				var indexes []localstore.IndexKeyParts
				if v.Format != pbrelation.RelationFormat_tag && v.Format != pbrelation.RelationFormat_status {
					return nil
				}
				if len(v.SelectDict) == 0 {
					return nil
				}

				for _, opt := range v.SelectDict {
					indexes = append(indexes, localstore.IndexKeyParts([]string{v.Format.String(), v.Key, opt.Id}))
				}
				return indexes
			}
			return nil
		},
		Unique:             false,
		SplitIndexKeyParts: true,
	}

	indexObjectTypeObject = localstore.Index{
		Prefix: pagesPrefix,
		Name:   "type",
		Keys: func(val interface{}) []localstore.IndexKeyParts {
			if v, ok := val.(*model.ObjectDetails); ok {
				var indexes []localstore.IndexKeyParts
				types := pbtypes.GetStringList(v.Details, bundle.RelationKeyType.String())

				for _, ot := range types {
					otCompact, err := objTypeCompactEncode(ot)
					if err != nil {
						log.Errorf("type index construction error('%s'): %s", ot, err.Error())
						continue
					}
					indexes = append(indexes, localstore.IndexKeyParts([]string{otCompact}))
				}
				return indexes
			}
			return nil
		},
		Unique: false,
		Hash:   false,
	}

	_ ObjectStore = (*dsObjectStore)(nil)
)

func New() ObjectStore {
	return &dsObjectStore{}
}

func (ls *dsObjectStore) Init(a *app.App) (err error) {
	ls.dsIface = a.MustComponent(datastore.CName).(datastore.Datastore)

	fts := a.Component(ftsearch.CName)
	if fts == nil {
		log.Warnf("init objectstore without fulltext")
	} else {
		ls.fts = fts.(ftsearch.FTSearch)
	}
	return nil
}

func (ls *dsObjectStore) Name() (name string) {
	return CName
}

type ObjectStore interface {
	app.ComponentRunnable
	localstore.Indexable
	database.Reader

	CreateObject(id string, details *types.Struct, relations *pbrelation.Relations, links []string, snippet string) error
	UpdateObjectDetails(id string, details *types.Struct, relations *pbrelation.Relations) error
	UpdateObjectLinksAndSnippet(id string, links []string, snippet string) error

	StoreRelations(relations []*pbrelation.Relation) error

	DeleteObject(id string) error
	RemoveRelationFromCache(key string) error

	UpdateRelationsInSet(setId, objTypeBefore, objTypeAfter string, relationsBefore, relationsAfter *pbrelation.Relations) error

	GetWithLinksInfoByID(id string) (*model.ObjectInfoWithLinks, error)
	GetWithOutboundLinksInfoById(id string) (*model.ObjectInfoWithOutboundLinks, error)
	GetDetails(id string) (*model.ObjectDetails, error)
	GetAggregatedOptions(relationKey string, relationFormat pbrelation.RelationFormat, objectType string) (options []*pbrelation.RelationOption, err error)

	GetByIDs(ids ...string) ([]*model.ObjectInfo, error)
	List() ([]*model.ObjectInfo, error)
	ListIds() ([]string, error)

	QueryObjectInfo(q database.Query, objectTypes []smartblock.SmartBlockType) (results []*model.ObjectInfo, total int, err error)
	AddToIndexQueue(id string) error
	IndexForEach(f func(id string, tm time.Time) error) error
	FTSearch() ftsearch.FTSearch
}

type relationOption struct {
	relationKey string
	optionId    string
}

type relationObjectType struct {
	relationKeys []string
	objectTypes  []string
}

var ErrNotAnObject = fmt.Errorf("not an object")

var filterNotSystemObjects = &filterObjectTypes{
	objectTypes: []smartblock.SmartBlockType{
		smartblock.SmartBlockTypeArchive,
		smartblock.SmartBlockTypeHome,
	},
	not: true,
}

type filterObjectTypes struct {
	objectTypes []smartblock.SmartBlockType
	not         bool
}

type RelationWithObjectType struct {
	objectType string
	relation   *pbrelation.Relation
}

func (m *filterObjectTypes) Filter(e query.Entry) bool {
	keyParts := strings.Split(e.Key, "/")
	id := keyParts[len(keyParts)-1]

	t, err := smartblock.SmartBlockTypeFromID(id)
	if err != nil {
		log.Errorf("failed to detect smartblock type for %s: %s", id, err.Error())
		return false
	}

	for _, ot := range m.objectTypes {
		if t == ot {
			return !m.not
		}
	}
	return m.not
}

type dsObjectStore struct {
	// underlying storage
	ds      ds.TxnDatastore
	dsIface datastore.Datastore

	fts ftsearch.FTSearch

	// serializing page updates
	l sync.Mutex

	subscriptions    []database.Subscription
	depSubscriptions []database.Subscription
}

func (m *dsObjectStore) Run() (err error) {
	m.ds, err = m.dsIface.LocalstoreDS()
	return
}

func (m *dsObjectStore) Close() (err error) {
	return nil
}

func (m *dsObjectStore) AggregateObjectIdsByOptionForRelation(relationKey string) (objectsByOptionId map[string][]string, err error) {
	txn, err := m.ds.NewTransaction(true)
	defer txn.Discard()

	res, err := localstore.GetKeysByIndexParts(txn, pagesPrefix, indexRelationOptionObject.Name, []string{relationKey}, "/", false, 100)
	if err != nil {
		return nil, err
	}

	keys, err := localstore.ExtractKeysFromResults(res)
	if err != nil {
		return nil, err
	}

	objectsByOptionId = make(map[string][]string)

	for _, key := range keys {
		optionId, err := localstore.CarveKeyParts(key, -2, -1)
		if err != nil {
			return nil, err
		}
		objId, err := localstore.CarveKeyParts(key, -1, 0)
		if err != nil {
			return nil, err
		}

		if _, exists := objectsByOptionId[optionId]; !exists {
			objectsByOptionId[optionId] = []string{}
		}

		objectsByOptionId[optionId] = append(objectsByOptionId[optionId], objId)
	}
	return
}

func (m *dsObjectStore) getAggregatedOptionsForFormat(format pbrelation.RelationFormat) (options []relationOption, err error) {
	txn, err := m.ds.NewTransaction(true)
	defer txn.Discard()

	res, err := localstore.GetKeysByIndexParts(txn, pagesPrefix, indexFormatOptionObject.Name, []string{format.String()}, "/", false, 100)
	if err != nil {
		return nil, err
	}

	keys, err := localstore.ExtractKeysFromResults(res)
	if err != nil {
		return nil, err
	}

	var ex = make(map[string]struct{})
	for _, key := range keys {
		optionId, err := localstore.CarveKeyParts(key, -2, -1)
		if err != nil {
			return nil, err
		}
		relKey, err := localstore.CarveKeyParts(key, -3, -2)
		if err != nil {
			return nil, err
		}

		if _, exists := ex[optionId]; exists {
			continue
		}
		ex[optionId] = struct{}{}
		options = append(options, relationOption{
			relationKey: relKey,
			optionId:    optionId,
		})
	}
	return
}

// GetAggregatedOptions returns aggregated options for specific relation and format. Options have a specific scope
func (m *dsObjectStore) GetAggregatedOptions(relationKey string, relationFormat pbrelation.RelationFormat, objectType string) (options []*pbrelation.RelationOption, err error) {
	objectsByOptionId, err := m.AggregateObjectIdsByOptionForRelation(relationKey)
	if err != nil {
		return nil, err
	}

	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, err
	}

	for optId, objIds := range objectsByOptionId {
		var scope = pbrelation.RelationOption_relation
		for _, objId := range objIds {
			exists, err := isObjectBelongToType(txn, objId, objectType)
			if err != nil {
				return nil, err
			}

			if exists {
				scope = pbrelation.RelationOption_local
				break
			}
		}
		opt, err := getOption(txn, optId)
		if err != nil {
			return nil, err
		}
		opt.Scope = scope
		options = append(options, opt)
	}

	relationOption, err := m.getAggregatedOptionsForFormat(relationFormat)
	for _, opt := range relationOption {
		if opt.relationKey == relationKey {
			// skip options for the same relation key because we already have them in the local/relation scope
			continue
		}

		opt2, err := getOption(txn, opt.optionId)
		if err != nil {
			return nil, err
		}
		opt2.Scope = pbrelation.RelationOption_format
		options = append(options, opt2)
	}

	return
}

func (m *dsObjectStore) GetAggregatedOptionsForFormat(format pbrelation.RelationFormat) ([]*pbrelation.RelationOption, error) {
	ros, err := m.getAggregatedOptionsForFormat(format)
	if err != nil {
		return nil, err
	}

	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, err
	}

	var options []*pbrelation.RelationOption
	for _, ro := range ros {
		opt, err := getOption(txn, ro.optionId)
		if err != nil {
			return nil, err
		}

		options = append(options, opt)
	}

	return options, nil
}

func (m *dsObjectStore) QueryAndSubscribeForChanges(schema *schema.Schema, q database.Query, sub database.Subscription) (records []database.Record, close func(), total int, err error) {
	m.l.Lock()
	defer m.l.Unlock()

	records, total, err = m.Query(schema, q)

	var ids []string
	for _, record := range records {
		ids = append(ids, pbtypes.GetString(record.Details, bundle.RelationKeyId.String()))
	}

	sub.Subscribe(ids)
	m.addSubscriptionIfNotExists(sub)
	close = func() {
		m.closeAndRemoveSubscription(sub)
	}

	return
}

// unsafe, use under mutex
func (m *dsObjectStore) addSubscriptionIfNotExists(sub database.Subscription) (existed bool) {
	for _, s := range m.subscriptions {
		if s == sub {
			return true
		}
	}
	log.Debugf("objStore subscription add %p", sub)
	m.subscriptions = append(m.subscriptions, sub)
	return false
}

func (m *dsObjectStore) closeAndRemoveSubscription(sub database.Subscription) {
	m.l.Lock()
	defer m.l.Unlock()
	sub.Close()

	for i, s := range m.subscriptions {
		if s == sub {
			log.Debugf("objStore subscription remove %p", s)
			m.subscriptions = append(m.subscriptions[:i], m.subscriptions[i+1:]...)
			break
		}
	}
}

func (m *dsObjectStore) QueryByIdAndSubscribeForChanges(ids []string, sub database.Subscription) (records []database.Record, close func(), err error) {
	m.l.Lock()
	defer m.l.Unlock()

	sub.Subscribe(ids)
	records, err = m.QueryById(ids)

	close = func() {
		m.closeAndRemoveSubscription(sub)
	}

	m.addSubscriptionIfNotExists(sub)

	return
}

func (m *dsObjectStore) Query(sch *schema.Schema, q database.Query) (records []database.Record, total int, err error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	dsq, err := q.DSQuery(sch)
	if err != nil {
		return
	}
	dsq.Offset = 0
	dsq.Limit = 0
	dsq.Prefix = pagesDetailsBase.String() + "/"
	if !q.WithSystemObjects {
		dsq.Filters = append([]query.Filter{filterNotSystemObjects}, dsq.Filters...)
	}
	if q.FullText != "" {
		if dsq, err = m.makeFTSQuery(q.FullText, dsq); err != nil {
			return
		}
	}
	for _, f := range dsq.Filters {
		log.Warnf("query filter: %+v", f)
	}

	res, err := txn.Query(dsq)
	if err != nil {
		return nil, 0, fmt.Errorf("error when querying ds: %w", err)
	}

	var (
		results []database.Record
		offset  = q.Offset
	)

	// We use own limit/offset implementation in order to find out
	// total number of records matching specified filters. Query
	// returns this number for handy pagination on clients.
	for rec := range res.Next() {
		total++

		if offset > 0 {
			offset--
			continue
		}

		if q.Limit > 0 && len(results) >= q.Limit {
			continue
		}

		var details model.ObjectDetails
		if err = proto.Unmarshal(rec.Value, &details); err != nil {
			log.Errorf("failed to unmarshal: %s", err.Error())
			total--
			continue
		}

		key := ds.NewKey(rec.Key)
		keyList := key.List()
		id := keyList[len(keyList)-1]

		if details.Details == nil || details.Details.Fields == nil {
			details.Details = &types.Struct{Fields: map[string]*types.Value{}}
		} else {
			pb.StructDeleteEmptyFields(details.Details)
		}

		details.Details.Fields[database.RecordIDField] = pb.ToValue(id)
		results = append(results, database.Record{Details: details.Details})
	}

	return results, total, nil
}

func (m *dsObjectStore) QueryObjectInfo(q database.Query, objectTypes []smartblock.SmartBlockType) (results []*model.ObjectInfo, total int, err error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	dsq, err := q.DSQuery(nil)
	if err != nil {
		return
	}
	dsq.Offset = 0
	dsq.Limit = 0
	dsq.Prefix = pagesDetailsBase.String() + "/"
	if len(objectTypes) > 0 {
		dsq.Filters = append([]query.Filter{&filterObjectTypes{objectTypes: objectTypes}}, dsq.Filters...)
	}
	if q.FullText != "" {
		if dsq, err = m.makeFTSQuery(q.FullText, dsq); err != nil {
			return
		}
	}
	res, err := txn.Query(dsq)
	if err != nil {
		return nil, 0, fmt.Errorf("error when querying ds: %w", err)
	}

	var (
		offset = q.Offset
	)

	// We use own limit/offset implementation in order to find out
	// total number of records matching specified filters. Query
	// returns this number for handy pagination on clients.
	for rec := range res.Next() {
		total++

		if offset > 0 {
			offset--
			continue
		}

		if q.Limit > 0 && len(results) >= q.Limit {
			continue
		}

		key := ds.NewKey(rec.Key)
		keyList := key.List()
		id := keyList[len(keyList)-1]
		oi, err := getObjectInfo(txn, id)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, oi)
	}
	return results, total, nil
}

func (m *dsObjectStore) QueryById(ids []string) (records []database.Record, err error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	for _, id := range ids {
		v, err := txn.Get(pagesDetailsBase.ChildString(id))
		if err != nil {
			log.Errorf("QueryByIds failed to find id: %s", id)
			continue
		}

		var details model.ObjectDetails
		if err = proto.Unmarshal(v, &details); err != nil {
			log.Errorf("QueryByIds failed to unmarshal id: %s", id)
			continue
		}

		if details.Details == nil || details.Details.Fields == nil {
			details.Details = &types.Struct{Fields: map[string]*types.Value{}}
		}

		details.Details.Fields[database.RecordIDField] = pb.ToValue(id)
		records = append(records, database.Record{Details: details.Details})
	}

	return
}

func (m *dsObjectStore) GetRelation(relationKey string) (*pbrelation.Relation, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	return m.getRelation(txn, relationKey)
}

// ListRelations retrieves all available relations and sort them in this order:
// 1. extraRelations aggregated from object of specific type (scope objectsOfTheSameType)
// 2. relations aggregated from sets of specific type  (scope setsOfTheSameType)
// 3. user-defined relations aggregated from all objects (scope library)
// 4. the rest of bundled relations (scope library)
func (m *dsObjectStore) ListRelations(objType string) ([]*pbrelation.Relation, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	if objType == "" {
		rels, err := m.listRelations(txn, 0)
		if err != nil {
			return nil, err
		}
		// todo: omit when we will have everything in index
		relsKeys2 := bundle.ListRelationsKeys()
		for _, relKey := range relsKeys2 {
			if pbtypes.HasRelation(rels, relKey.String()) {
				continue
			}

			rel := bundle.MustGetRelation(relKey)
			rel.Scope = pbrelation.Relation_library
			rels = append(rels, rel)
		}
		return rels, nil
	}

	rels, err := m.AggregateRelationsFromObjectsOfType(objType)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate relations from objects: %w", err)
	}

	rels2, err := m.AggregateRelationsFromSetsOfType(objType)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate relations from sets: %w", err)
	}

	for i, rel := range rels2 {
		if pbtypes.HasRelation(rels, rel.Key) {
			continue
		}
		rels = append(rels, rels2[i])
	}

	relsKeys, err := m.listRelationsKeys(txn)
	if err != nil {
		return nil, fmt.Errorf("failed to list relations from store index: %w", err)
	}

	// todo: omit when we will have everything in index
	for _, relKey := range relsKeys {
		if pbtypes.HasRelation(rels, relKey) {
			continue
		}
		rel, err := m.getRelation(txn, relKey)
		if err != nil {
			log.Errorf("relation found in index but failed to retrieve from store")
			continue
		}
		rel.Scope = pbrelation.Relation_library
		rels = append(rels, rel)
	}

	relsKeys2 := bundle.ListRelationsKeys()
	for _, relKey := range relsKeys2 {
		if pbtypes.HasRelation(rels, relKey.String()) {
			continue
		}

		rel := bundle.MustGetRelation(relKey)
		rel.Scope = pbrelation.Relation_library
		rels = append(rels, rel)
	}

	return rels, nil
}

func (m *dsObjectStore) ListRelationsKeys() ([]string, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	return m.listRelationsKeys(txn)
}

func (m *dsObjectStore) AggregateRelationsFromObjectsOfType(objType string) ([]*pbrelation.Relation, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	var rels []*pbrelation.Relation
	objTypeCompact, err := objTypeCompactEncode(objType)
	if err != nil {
		return nil, fmt.Errorf("failed to encode object type '%s': %s", objType, err.Error())
	}
	res, err := localstore.GetKeysByIndexParts(txn, indexObjectTypeRelationObjectId.Prefix, indexObjectTypeRelationObjectId.Name, []string{objTypeCompact}, "/", false, 0)
	if err != nil {
		return nil, err
	}

	relKeys, err := localstore.GetKeyPartFromResults(res, -2, -1, true)
	if err != nil {
		return nil, err
	}

	for _, relKey := range relKeys {
		rel, err := m.getRelation(txn, relKey)
		if err != nil {
			log.Errorf("relation '%s' found in the index but failed to retreive: %s", relKey, err.Error())
			continue
		}

		rel.Scope = pbrelation.Relation_objectsOfTheSameType
		rels = append(rels, rel)
	}

	return rels, nil
}

func (m *dsObjectStore) AggregateRelationsFromSetsOfType(objType string) ([]*pbrelation.Relation, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	var rels []*pbrelation.Relation
	objTypeCompact, err := objTypeCompactEncode(objType)
	if err != nil {
		return nil, err
	}
	res, err := localstore.GetKeysByIndexParts(txn, indexObjectTypeRelationSetId.Prefix, indexObjectTypeRelationSetId.Name, []string{objTypeCompact}, "/", false, 0)
	if err != nil {
		return nil, err
	}

	relKeys, err := localstore.GetKeyPartFromResults(res, -2, -1, true)
	if err != nil {
		return nil, err
	}

	for _, relKey := range relKeys {
		rel, err := m.getRelation(txn, relKey)
		if err != nil {
			log.Errorf("relation '%s' found in the index but failed to retreive: %s", relKey, err.Error())
			continue
		}

		rel.Scope = pbrelation.Relation_setOfTheSameType
		rels = append(rels, rel)
	}

	return rels, nil
}

func (m *dsObjectStore) DeleteObject(id string) error {
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	// todo: remove all indexes with this object
	for _, k := range []ds.Key{
		pagesDetailsBase.ChildString(id),
		pagesSnippetBase.ChildString(id),
		pagesRelationsBase.ChildString(id),
		indexQueueBase.ChildString(id),
	} {
		if err = txn.Delete(k); err != nil {
			return err
		}
	}

	inLinks, err := findInboundLinks(txn, id)
	if err != nil {
		return err
	}

	outLinks, err := findOutboundLinks(txn, id)
	if err != nil {
		return err
	}

	for _, k := range pageLinkKeys(id, inLinks, outLinks) {
		if err := txn.Delete(k); err != nil {
			return err
		}
	}
	if m.fts != nil {
		if err := m.fts.Delete(id); err != nil {
			return err
		}
	}
	return txn.Commit()
}

// RemoveRelationFromCache removes cached relation data
func (m *dsObjectStore) RemoveRelationFromCache(key string) error {
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	for _, k := range []ds.Key{
		relationsBase.ChildString(key),
	} {
		if err = txn.Delete(k); err != nil {
			return err
		}
	}

	return txn.Commit()
}

func (m *dsObjectStore) GetWithLinksInfoByID(id string) (*model.ObjectInfoWithLinks, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	pages, err := getObjectsInfo(txn, []string{id})
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("page not found")
	}
	page := pages[0]

	inboundIds, err := findInboundLinks(txn, id)
	if err != nil {
		return nil, err
	}

	outboundsIds, err := findOutboundLinks(txn, id)
	if err != nil {
		return nil, err
	}

	inbound, err := getObjectsInfo(txn, inboundIds)
	if err != nil {
		return nil, err
	}

	outbound, err := getObjectsInfo(txn, outboundsIds)
	if err != nil {
		return nil, err
	}

	return &model.ObjectInfoWithLinks{
		Id:   id,
		Info: page,
		Links: &model.ObjectLinksInfo{
			Inbound:  inbound,
			Outbound: outbound,
		},
	}, nil
}

func (m *dsObjectStore) GetWithOutboundLinksInfoById(id string) (*model.ObjectInfoWithOutboundLinks, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	pages, err := getObjectsInfo(txn, []string{id})
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("page not found")
	}
	page := pages[0]

	outboundsIds, err := findOutboundLinks(txn, id)
	if err != nil {
		return nil, err
	}

	outbound, err := getObjectsInfo(txn, outboundsIds)
	if err != nil {
		return nil, err
	}

	return &model.ObjectInfoWithOutboundLinks{
		Info:          page,
		OutboundLinks: outbound,
	}, nil
}

func (m *dsObjectStore) GetDetails(id string) (*model.ObjectDetails, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	return getDetails(txn, id)
}

func (m *dsObjectStore) List() ([]*model.ObjectInfo, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	ids, err := findByPrefix(txn, pagesDetailsBase.String()+"/", 0)
	if err != nil {
		return nil, err
	}

	return getObjectsInfo(txn, ids)
}

func (m *dsObjectStore) GetByIDs(ids ...string) ([]*model.ObjectInfo, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	return getObjectsInfo(txn, ids)
}

func (m *dsObjectStore) CreateObject(id string, details *types.Struct, relations *pbrelation.Relations, links []string, snippet string) error {
	m.l.Lock()
	defer m.l.Unlock()
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	// init an empty state to skip nil checks later
	before := model.ObjectInfo{
		Relations: &pbrelation.Relations{},
		Details:   &types.Struct{Fields: map[string]*types.Value{}},
	}

	err = m.updateObjectDetails(txn, id, before, details, relations)
	if err != nil {
		return err
	}

	err = m.updateObjectLinksAndSnippet(txn, id, links, snippet)
	if err != nil {
		return err
	}
	return txn.Commit()
}

func (m *dsObjectStore) UpdateObjectLinksAndSnippet(id string, links []string, snippet string) error {
	m.l.Lock()
	defer m.l.Unlock()
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	err = m.updateObjectLinksAndSnippet(txn, id, links, snippet)
	if err != nil {
		return err
	}
	return txn.Commit()
}

func (m *dsObjectStore) UpdateObjectDetails(id string, details *types.Struct, relations *pbrelation.Relations) error {
	m.l.Lock()
	defer m.l.Unlock()
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()
	var (
		before model.ObjectInfo
	)

	if details != nil || relations != nil {
		exInfo, err := getObjectInfo(txn, id)
		if err != nil {
			log.Debugf("UpdateObject failed to get ex state for object %s: %s", id, err.Error())
		}

		if exInfo != nil {
			before = *exInfo
		} else {
			// init an empty state to skip nil checks later
			before = model.ObjectInfo{
				Relations: &pbrelation.Relations{},
				Details:   &types.Struct{Fields: map[string]*types.Value{}},
			}
		}
	}

	err = m.updateObjectDetails(txn, id, before, details, relations)
	if err != nil {
		return err
	}
	return txn.Commit()
}

func (m *dsObjectStore) updateArchive(txn ds.Txn, id string, links []string) error {
	exLinks, _ := findOutboundLinks(txn, id)
	removedLinks, addedLinks := slice.DifferenceRemovedAdded(exLinks, links)
	getCurrentDetails := func(id string) (*types.Struct, error) {
		det, err := m.GetDetails(id)
		if err != nil {
			return nil, err
		}
		if det == nil || det.Details == nil || det.Details.Fields == nil {
			return &types.Struct{
				Fields: map[string]*types.Value{},
			}, nil
		}

		return det.Details, nil
	}

	setArchived := func(id string, val bool) error {
		det, err := getCurrentDetails(id)
		if err != nil {
			return fmt.Errorf("failed to get current details: %s", err.Error())
		}
		newDet := pbtypes.CopyStruct(det)
		newDet.Fields[bundle.RelationKeyIsArchived.String()] = pbtypes.Bool(val)
		err = m.updateDetails(txn, id, &model.ObjectDetails{det}, &model.ObjectDetails{newDet})
		if err != nil {
			return fmt.Errorf("updateObject failed: %s", err.Error())
		}
		return nil
	}

	var err error
	for _, objId := range removedLinks {
		err = setArchived(objId, false)
		if err != nil {
			log.Errorf("setArchived failed: %s", err.Error())
		}
	}

	for _, objId := range addedLinks {
		err = setArchived(objId, true)
		if err != nil {
			log.Errorf("setArchived failed: %s", err.Error())
		}
	}

	return nil
}

func (m *dsObjectStore) updateObjectLinksAndSnippet(txn ds.Txn, id string, links []string, snippet string) error {
	sbt, err := smartblock.SmartBlockTypeFromID(id)
	if err != nil {
		return fmt.Errorf("failed to extract smartblock type: %w", err)
	}

	if sbt == smartblock.SmartBlockTypeArchive {
		// special case for Archive. We don't need to index the outgoing links for the navigation, instead we should use it to set archived flag on objects
		return m.updateArchive(txn, id, links)
	}

	var addedLinks, removedLinks []string

	exLinks, _ := findOutboundLinks(txn, id)
	removedLinks, addedLinks = slice.DifferenceRemovedAdded(exLinks, links)
	if len(addedLinks) > 0 {
		for _, k := range pageLinkKeys(id, nil, addedLinks) {
			if err := txn.Put(k, nil); err != nil {
				return err
			}
		}
	}

	if len(removedLinks) > 0 {
		for _, k := range pageLinkKeys(id, nil, removedLinks) {
			if err := txn.Delete(k); err != nil {
				return err
			}
		}
	}

	if val, err := txn.Get(pagesSnippetBase.ChildString(id)); err == ds.ErrNotFound || string(val) != snippet {
		if err := m.updateSnippet(txn, id, snippet); err != nil {
			return err
		}
	}

	return nil
}

func (m *dsObjectStore) updateObjectDetails(txn ds.Txn, id string, before model.ObjectInfo, details *types.Struct, relations *pbrelation.Relations) error {
	sbt, err := smartblock.SmartBlockTypeFromID(id)
	if err != nil {
		return fmt.Errorf("failed to extract smartblock type: %w", err)
	}

	if sbt == smartblock.SmartBlockTypeArchive {
		return ErrNotAnObject
	}

	var (
		objTypes []string
	)

	if details != nil {
		objTypes = pbtypes.GetStringList(details, bundle.RelationKeyType.String())
		if err = m.updateDetails(txn, id, &model.ObjectDetails{Details: before.Details}, &model.ObjectDetails{Details: details}); err != nil {
			return err
		}
	}

	if relations != nil && relations.Relations != nil {
		// intentionally do not pass txn, as this tx may be huge
		if err = m.updateRelations(txn, before.ObjectTypeUrls, objTypes, id, before.Relations, relations); err != nil {
			return err
		}
	}

	return nil
}

func (m *dsObjectStore) sendUpdatesToSubscriptions(id string, details *types.Struct) {
	detCopy := pbtypes.CopyStruct(details)
	detCopy.Fields[database.RecordIDField] = pb.ToValue(id)
	for i := range m.subscriptions {
		go func(sub database.Subscription) {
			_ = sub.Publish(id, detCopy)
		}(m.subscriptions[i])
	}
}

func (m *dsObjectStore) AddToIndexQueue(id string) error {
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()
	var buf [8]byte
	size := binary.PutVarint(buf[:], time.Now().Unix())
	if err = txn.Put(indexQueueBase.ChildString(id), buf[:size]); err != nil {
		return err
	}
	return txn.Commit()
}

func (m *dsObjectStore) IndexForEach(f func(id string, tm time.Time) error) error {
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()
	res, err := txn.Query(query.Query{Prefix: indexQueueBase.String()})
	if err != nil {
		return fmt.Errorf("error query txn in datastore: %w", err)
	}
	var idsToRemove []string
	for entry := range res.Next() {
		id := extractIdFromKey(entry.Key)
		ts, _ := binary.Varint(entry.Value)
		if indexErr := f(id, time.Unix(ts, 0)); indexErr != nil {
			log.Warnf("can't index '%s': %v", id, indexErr)
			continue
		}
		idsToRemove = append(idsToRemove, id)
	}

	err = res.Close()
	if err != nil {
		return err
	}

	for _, id := range idsToRemove {
		if err := txn.Delete(indexQueueBase.ChildString(id)); err != nil {
			log.Error("failed to remove id from full text index queue: %s", err.Error())
		}
	}

	return txn.Commit()
}

func (m *dsObjectStore) ListIds() ([]string, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	return findByPrefix(txn, pagesDetailsBase.String()+"/", 0)
}

func (m *dsObjectStore) updateDetails(txn ds.Txn, id string, oldDetails *model.ObjectDetails, newDetails *model.ObjectDetails) error {
	detailsKey := pagesDetailsBase.ChildString(id)
	b, err := proto.Marshal(newDetails)
	if err != nil {
		return err
	}

	err = txn.Put(detailsKey, b)
	if err != nil {
		return err
	}

	err = localstore.UpdateIndexesWithTxn(m, txn, oldDetails, newDetails, id)
	if err != nil {
		return err
	}

	if newDetails != nil && newDetails.Details.Fields != nil {
		m.sendUpdatesToSubscriptions(id, newDetails.Details)
	}

	return nil
}

func (m *dsObjectStore) updateOption(txn ds.Txn, option *pbrelation.RelationOption) error {
	b, err := proto.Marshal(option)
	if err != nil {
		return err
	}
	optionKey := relationsOptionsBase.ChildString(option.Id)

	return txn.Put(optionKey, b)

}

func (m *dsObjectStore) UpdateRelationsInSet(setId, objTypeBefore, objTypeAfter string, relationsBefore, relationsAfter *pbrelation.Relations) error {
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = m.updateRelationsInSet(txn, setId, objTypeBefore, objTypeAfter, relationsBefore, relationsAfter)
	if err != nil {
		return err
	}

	err = m.storeRelations(txn, relationsAfter.Relations)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (m *dsObjectStore) StoreRelations(relations []*pbrelation.Relation) error {
	m.l.Lock()
	defer m.l.Unlock()
	if relations == nil {
		return nil
	}
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	err = m.storeRelations(txn, relations)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (m *dsObjectStore) storeRelations(txn ds.Txn, relations []*pbrelation.Relation) error {
	var relBytes []byte
	var err error
	for _, relation := range relations {
		relCopy := pbtypes.CopyRelation(relation)
		relCopy.SelectDict = nil
		var bundled = true
		if !bundle.HasRelation(relation.Key) {
			bundled = false
			// do not store bundled relations
			relationKey := relationsBase.ChildString(relation.Key)
			relBytes, err = proto.Marshal(relCopy)
			if err != nil {
				return err
			}

			err = txn.Put(relationKey, relBytes)
			if err != nil {
				return err
			}
		}
		_, details := bundle.GetDetailsForRelation(bundled, relCopy)
		id := pbtypes.GetString(details, "id")
		err = m.updateObjectDetails(txn, id, model.ObjectInfo{
			Relations: &pbrelation.Relations{},
			Details:   &types.Struct{Fields: map[string]*types.Value{}},
		}, details, nil)
		if err != nil {
			return err
		}

		err = m.updateObjectLinksAndSnippet(txn, id, nil, relation.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *dsObjectStore) updateRelations(txn ds.Txn, objTypesBefore []string, objTypesAfter []string, id string, relationsBefore *pbrelation.Relations, relationsAfter *pbrelation.Relations) error {
	if relationsAfter == nil {
		return nil
	}

	if relationsBefore != nil && pbtypes.RelationsEqual(relationsBefore.Relations, relationsAfter.Relations) {
		return nil
	}

	relationsKey := pagesRelationsBase.ChildString(id)
	for _, relation := range relationsAfter.Relations {
		if relation.Format == pbrelation.RelationFormat_status || relation.Format == pbrelation.RelationFormat_tag {
			var relBefore *pbrelation.Relation
			if relationsBefore != nil {
				relBefore = pbtypes.GetRelation(relationsBefore.Relations, relation.Key)
			}

			for _, opt := range relation.SelectDict {
				var optBefore *pbrelation.RelationOption
				if relBefore != nil {
					optBefore = pbtypes.GetOption(relBefore.SelectDict, opt.Id)
				}

				if !pbtypes.OptionEqualOmitScope(optBefore, opt) {
					err := m.updateOption(txn, opt)
					if err != nil {
						return err
					}

				}
			}
		}

		err := localstore.AddIndexesWithTxn(m, txn, relation, id)
		if err != nil {
			return err
		}
	}

	err := localstore.UpdateIndexWithTxn(indexObjectTypeRelationObjectId, txn, &relationObjectType{
		relationKeys: pbtypes.GetRelationKeys(relationsBefore.Relations),
		objectTypes:  objTypesBefore,
	}, &relationObjectType{
		relationKeys: pbtypes.GetRelationKeys(relationsAfter.Relations),
		objectTypes:  objTypesAfter,
	}, id)
	if err != nil {
		return err
	}

	err = m.storeRelations(txn, relationsAfter.Relations)
	if err != nil {
		return err
	}

	b, err := proto.Marshal(relationsAfter)
	if err != nil {
		return err
	}
	return txn.Put(relationsKey, b)
}

func (m *dsObjectStore) updateRelationsInSet(txn ds.Txn, setId, objTypesBefore, objTypesAfter string, relationsBefore, relationsAfter *pbrelation.Relations) error {
	return localstore.UpdateIndexWithTxn(indexObjectTypeRelationSetId, txn, &relationObjectType{
		relationKeys: pbtypes.GetRelationKeys(relationsBefore.Relations),
		objectTypes:  []string{objTypesBefore},
	}, &relationObjectType{
		relationKeys: pbtypes.GetRelationKeys(relationsAfter.Relations),
		objectTypes:  []string{objTypesAfter},
	}, setId)
}

func (m *dsObjectStore) updateSnippet(txn ds.Txn, id string, snippet string) error {
	snippetKey := pagesSnippetBase.ChildString(id)
	return txn.Put(snippetKey, []byte(snippet))
}

func (m *dsObjectStore) Prefix() string {
	return pagesPrefix
}

func (m *dsObjectStore) Indexes() []localstore.Index {
	return []localstore.Index{indexObjectTypeRelationObjectId, indexObjectTypeRelationSetId, indexRelationOptionObject, indexRelationObject, indexFormatOptionObject, indexObjectTypeObject}
}

func (m *dsObjectStore) FTSearch() ftsearch.FTSearch {
	return m.fts
}

func (m *dsObjectStore) makeFTSQuery(text string, dsq query.Query) (query.Query, error) {
	if m.fts == nil {
		return dsq, fmt.Errorf("fullText search not configured")
	}
	ids, err := m.fts.Search(text)
	if err != nil {
		return dsq, err
	}
	idsQuery := newIdsFilter(ids)
	dsq.Filters = append([]query.Filter{idsQuery}, dsq.Filters...)
	dsq.Orders = append([]query.Order{idsQuery}, dsq.Orders...)
	return dsq, nil
}

func (m *dsObjectStore) listIdsOfType(txn ds.Txn, ot string) ([]string, error) {
	res, err := localstore.GetKeysByIndexParts(txn, pagesPrefix, indexObjectTypeObject.Name, []string{ot}, "", false, 100)
	if err != nil {
		return nil, err
	}

	return localstore.GetLeavesFromResults(res)
}

func (m *dsObjectStore) listRelationsKeys(txn ds.Txn) ([]string, error) {
	txn, err := m.ds.NewTransaction(true)
	if err != nil {
		return nil, fmt.Errorf("error creating txn in datastore: %w", err)
	}
	defer txn.Discard()

	return findByPrefix(txn, relationsBase.String()+"/", 0)
}

func (m *dsObjectStore) getRelation(txn ds.Txn, key string) (*pbrelation.Relation, error) {
	br, err := bundle.GetRelation(bundle.RelationKey(key))
	if br != nil {
		return br, nil
	}

	res, err := txn.Get(relationsBase.ChildString(key))

	if err != nil {
		return nil, err
	}

	var rel pbrelation.Relation
	if err = proto.Unmarshal(res, &rel); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relation: %s", err.Error())
	}

	return &rel, nil
}

func (m *dsObjectStore) listRelations(txn ds.Txn, limit int) ([]*pbrelation.Relation, error) {
	var rels []*pbrelation.Relation

	res, err := txn.Query(query.Query{
		Prefix:   relationsBase.String(),
		Limit:    limit,
		KeysOnly: false,
	})
	if err != nil {
		return nil, err
	}

	for r := range res.Next() {
		var rel pbrelation.Relation
		if err = proto.Unmarshal(r.Value, &rel); err != nil {
			log.Errorf("listRelations failed to unmarshal relation: %s", err.Error())
			continue
		}
		rels = append(rels, &rel)
	}

	return rels, nil
}

func isObjectBelongToType(txn ds.Txn, id, objType string) (bool, error) {
	objTypeCompact, err := objTypeCompactEncode(objType)
	if err != nil {
		return false, err
	}

	return localstore.HasPrimaryKeyByIndexParts(txn, pagesPrefix, indexObjectTypeObject.Name, []string{objTypeCompact}, "", false, id)
}

func isRelationBelongToType(txn ds.Txn, relKey, objectType string) (bool, error) {
	res, err := localstore.GetKeysByIndexParts(txn, pagesPrefix, indexRelationObject.Name, []string{relKey}, "", false, 0)
	if err != nil {
		return false, err
	}
	i := 0
	for v := range res.Next() {
		i++
		objId, err := localstore.CarveKeyParts(v.Key, -1, 0)
		if err != nil {
			return false, err
		}

		belong, err := isObjectBelongToType(txn, objId, objectType)
		if err != nil {
			return false, err
		}
		if belong {
			return true, nil
		}
	}

	return false, nil
}

/* internal */

func getDetails(txn ds.Txn, id string) (*model.ObjectDetails, error) {
	var details model.ObjectDetails
	if val, err := txn.Get(pagesDetailsBase.ChildString(id)); err != nil && err != ds.ErrNotFound {
		return nil, fmt.Errorf("failed to get details: %w", err)
	} else if err := proto.Unmarshal(val, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

func getRelations(txn ds.Txn, id string) (*pbrelation.Relations, error) {
	var relations pbrelation.Relations
	if val, err := txn.Get(pagesRelationsBase.ChildString(id)); err != nil {
		if err != ds.ErrNotFound {
			return nil, fmt.Errorf("failed to get relations: %w", err)
		}
	} else if err := proto.Unmarshal(val, &relations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relations: %w", err)
	}
	return &relations, nil
}

func getOption(txn ds.Txn, optionId string) (*pbrelation.RelationOption, error) {
	var opt pbrelation.RelationOption
	if val, err := txn.Get(relationsOptionsBase.ChildString(optionId)); err != nil {
		log.Debugf("getOption %s: not found", optionId)
		if err != ds.ErrNotFound {
			return nil, fmt.Errorf("failed to get option from localstore: %w", err)
		}
	} else if err := proto.Unmarshal(val, &opt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal option: %w", err)
	}

	return &opt, nil
}

func getObjectInfo(txn ds.Txn, id string) (*model.ObjectInfo, error) {
	sbt, err := smartblock.SmartBlockTypeFromID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to extract smartblock type: %w", err)
	}
	if sbt == smartblock.SmartBlockTypeArchive {
		return nil, ErrNotAnObject
	}

	var details model.ObjectDetails
	if val, err := txn.Get(pagesDetailsBase.ChildString(id)); err != nil {
		return nil, fmt.Errorf("failed to get details: %w", err)
	} else if err := proto.Unmarshal(val, &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal details: %w", err)
	}

	var relations pbrelation.Relations
	if val, err := txn.Get(pagesRelationsBase.ChildString(id)); err != nil {
		if err != ds.ErrNotFound {
			return nil, fmt.Errorf("failed to get relations: %w", err)
		}

	} else if err := proto.Unmarshal(val, &relations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relations: %w", err)
	}

	if details.Details == nil || details.Details.Fields == nil {
		details.Details = &types.Struct{Fields: map[string]*types.Value{}}
	}
	details.Details.Fields[bundle.RelationKeyId.String()] = pbtypes.String(id)

	var objectTypes []string
	// remove hardcoded type
	otypes := pbtypes.GetStringList(details.Details, bundle.RelationKeyType.String())
	for _, ot := range otypes {
		objectTypes = append(objectTypes, ot)
	}

	var snippet string
	if val, err := txn.Get(pagesSnippetBase.ChildString(id)); err != nil && err != ds.ErrNotFound {
		return nil, fmt.Errorf("failed to get snippet: %w", err)
	} else {
		snippet = string(val)
	}

	// omit decoding page state
	hasInbound, err := hasInboundLinks(txn, id)
	if err != nil {
		return nil, err
	}

	return &model.ObjectInfo{
		Id:              id,
		ObjectType:      sbt.ToProto(),
		Details:         details.Details,
		Relations:       &relations,
		Snippet:         snippet,
		HasInboundLinks: hasInbound,
		ObjectTypeUrls:  objectTypes,
	}, nil
}

func getObjectsInfo(txn ds.Txn, ids []string) ([]*model.ObjectInfo, error) {
	var objects []*model.ObjectInfo
	for _, id := range ids {
		info, err := getObjectInfo(txn, id)
		if err != nil {
			if strings.HasSuffix(err.Error(), "key not found") || err == ErrNotAnObject {
				continue
			}
			return nil, err
		}
		objects = append(objects, info)
	}

	return objects, nil
}

func hasInboundLinks(txn ds.Txn, id string) (bool, error) {
	inboundResults, err := txn.Query(query.Query{
		Prefix:   pagesInboundLinksBase.String() + "/" + id + "/",
		Limit:    1, // we only need to know if there is at least 1 inbound link
		KeysOnly: true,
	})
	if err != nil {
		return false, err
	}

	// max is 1
	inboundLinks, err := localstore.CountAllKeysFromResults(inboundResults)
	return inboundLinks > 0, err
}

// Find to which IDs specified one has outbound links.
func findOutboundLinks(txn ds.Txn, id string) ([]string, error) {
	return findByPrefix(txn, pagesOutboundLinksBase.String()+"/"+id+"/", 0)
}

// Find from which IDs specified one has inbound links.
func findInboundLinks(txn ds.Txn, id string) ([]string, error) {
	return findByPrefix(txn, pagesInboundLinksBase.String()+"/"+id+"/", 0)
}

func findByPrefix(txn ds.Txn, prefix string, limit int) ([]string, error) {
	results, err := txn.Query(query.Query{
		Prefix:   prefix,
		Limit:    limit,
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	return localstore.GetLeavesFromResults(results)
}

func pageLinkKeys(id string, in []string, out []string) []ds.Key {
	var keys = make([]ds.Key, 0, len(in)+len(out))

	// links incoming into specified node id
	for _, from := range in {
		keys = append(keys, inboundLinkKey(from, id), outgoingLinkKey(from, id))
	}

	// links outgoing from specified node id
	for _, to := range out {
		keys = append(keys, outgoingLinkKey(id, to), inboundLinkKey(id, to))
	}

	return keys
}

func outgoingLinkKey(from, to string) ds.Key {
	return pagesOutboundLinksBase.ChildString(from).ChildString(to)
}

func inboundLinkKey(from, to string) ds.Key {
	return pagesInboundLinksBase.ChildString(to).ChildString(from)
}

func newIdsFilter(ids []string) idsFilter {
	f := make(idsFilter)
	for i, id := range ids {
		f[id] = i
	}
	return f
}

type idsFilter map[string]int

func (f idsFilter) Filter(e query.Entry) bool {
	_, ok := f[extractIdFromKey(e.Key)]
	return ok
}

func (f idsFilter) Compare(a, b query.Entry) int {
	aIndex := f[extractIdFromKey(a.Key)]
	bIndex := f[extractIdFromKey(b.Key)]
	if aIndex == bIndex {
		return 0
	} else if aIndex < bIndex {
		return -1
	} else {
		return 1
	}
}

func extractIdFromKey(key string) (id string) {
	i := strings.LastIndexByte(key, '/')
	if i == -1 || len(key)-1 == i {
		return
	}
	return key[i+1:]
}

// temp func until we move to the proper ids
func objTypeCompactEncode(objType string) (string, error) {
	if strings.HasPrefix(objType, addr.BundledObjectTypeURLPrefix) {
		return objType, nil
	}
	if strings.HasPrefix(objType, "ba") {
		return objType, nil
	}
	return "", fmt.Errorf("invalid objType")
}

func GetObjectType(store ObjectStore, url string) (*pbrelation.ObjectType, error) {
	objectType := &pbrelation.ObjectType{}
	if strings.HasPrefix(url, addr.BundledObjectTypeURLPrefix) {
		var err error
		objectType, err = bundle.GetTypeByUrl(url)
		if err != nil {
			if err == bundle.ErrNotFound {
				return nil, fmt.Errorf("unknown object type")
			}
			return nil, err
		}
		return objectType, nil
	} else if !strings.HasPrefix(url, "b") {
		return nil, fmt.Errorf("incorrect object type URL format")
	}

	ois, err := store.GetByIDs(url)
	if err != nil {
		return nil, err
	}
	if len(ois) == 0 {
		return nil, fmt.Errorf("object type not found in the index")
	}

	details := ois[0].Details
	var relations []*pbrelation.Relation
	if ois[0].Relations != nil {
		relations = ois[0].Relations.Relations
	}

	for _, relId := range pbtypes.GetStringList(details, bundle.RelationKeyRecommendedRelations.String()) {
		rk, err := pbtypes.RelationIdToKey(relId)
		if err != nil {
			log.Errorf("GetObjectType failed to get relation key from id: %s (%s)", err.Error(), relId)
			continue
		}

		r := pbtypes.GetRelation(relations, rk)
		if r == nil {
			log.Errorf("invalid state of objectType %s: missing recommended relation %s", url, rk)
			continue
		}
		relCopy := pbtypes.CopyRelation(r)
		relCopy.Scope = pbrelation.Relation_type

		objectType.Relations = append(objectType.Relations, relCopy)
	}

	objectType.Url = url
	if details != nil && details.Fields != nil {
		if v, ok := details.Fields[bundle.RelationKeyName.String()]; ok {
			objectType.Name = v.GetStringValue()
		}
		if v, ok := details.Fields[bundle.RelationKeyRecommendedLayout.String()]; ok {
			objectType.Layout = pbrelation.ObjectTypeLayout(int(v.GetNumberValue()))
		}
		if v, ok := details.Fields[bundle.RelationKeyIconEmoji.String()]; ok {
			objectType.IconEmoji = v.GetStringValue()
		}
	}

	return objectType, err
}