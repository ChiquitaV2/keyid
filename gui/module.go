package gui

import (
	"github.com/xdave/keyid/interfaces"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(func(client interfaces.Client, shutdowner fx.Shutdowner) {
		Show(client, shutdowner)
	}),
)
