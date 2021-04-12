package anytype

import (
	"github.com/anytypeio/go-anytype-middleware/app"
	"github.com/anytypeio/go-anytype-middleware/core/anytype/config"
	"github.com/anytypeio/go-anytype-middleware/core/block"
	"github.com/anytypeio/go-anytype-middleware/core/block/export"
	"github.com/anytypeio/go-anytype-middleware/core/block/meta"
	"github.com/anytypeio/go-anytype-middleware/core/block/process"
	"github.com/anytypeio/go-anytype-middleware/core/event"
	"github.com/anytypeio/go-anytype-middleware/core/history"
	"github.com/anytypeio/go-anytype-middleware/core/indexer"
	"github.com/anytypeio/go-anytype-middleware/core/recordsbatcher"
	"github.com/anytypeio/go-anytype-middleware/core/status"
	"github.com/anytypeio/go-anytype-middleware/core/wallet"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/cafe"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/core"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/datastore/clientds"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/files"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/gateway"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/ipfs/ipfslite"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/localstore/filestore"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/localstore/ftsearch"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/localstore/objectstore"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/pin"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/profilefinder"
	"github.com/anytypeio/go-anytype-middleware/pkg/lib/threads"
	walletUtil "github.com/anytypeio/go-anytype-middleware/pkg/lib/wallet"
	"github.com/anytypeio/go-anytype-middleware/util/builtintemplate"
	"github.com/anytypeio/go-anytype-middleware/util/linkpreview"
)

func StartAccountRecoverApp(eventSender event.Sender, accountPrivKey walletUtil.Keypair) (a *app.App, err error) {
	a = new(app.App)
	device, err := walletUtil.NewRandomKeypair(walletUtil.KeypairTypeDevice)
	if err != nil {
		return nil, err
	}

	a.Register(&config.DefaultConfig).Register(wallet.NewWithRepoPathAndKeys("", accountPrivKey, device)).
		Register(cafe.New()).
		Register(profilefinder.New()).
		Register(eventSender)

	if err = a.Start(); err != nil {
		return
	}
	return a, nil
}

func BootstrapConfigAndWallet(newAccount bool, rootPath, accountId string) ([]app.Component, error) {
	return []app.Component{
		config.New(func(c *config.Config) {
			c.NewAccount = newAccount
		}),
		wallet.NewWithAccountRepo(rootPath, accountId),
	}, nil
}

func StartNewApp(components ...app.Component) (a *app.App, err error) {
	a = new(app.App)
	Bootstrap(a, components...)
	if err = a.Start(); err != nil {
		return
	}
	return
}

func Bootstrap(a *app.App, components ...app.Component) {
	for _, c := range components {
		a.Register(c)
	}
	a.Register(clientds.New()).
		Register(ftsearch.New()).
		Register(objectstore.New()).
		Register(filestore.New()).
		Register(recordsbatcher.New()).
		Register(indexer.New()).
		Register(ipfslite.New()).
		Register(files.New()).
		Register(cafe.New()).
		Register(threads.New()).
		Register(core.New()).
		Register(pin.New()).
		Register(status.New()).
		Register(meta.New()).
		Register(block.New()).
		Register(process.New()).
		Register(history.New()).
		Register(gateway.New()).
		Register(export.New()).
		Register(builtintemplate.New()).
		Register(linkpreview.New())
	return
}