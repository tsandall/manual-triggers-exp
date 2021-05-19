package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/open-policy-agent/opa/cmd"
	"github.com/open-policy-agent/opa/plugins"
	"github.com/open-policy-agent/opa/plugins/bundle"
	"github.com/open-policy-agent/opa/runtime"
	"github.com/open-policy-agent/opa/util"
)

const Name = "ticker"

type Config struct{}

type Plugin struct {
	manager *plugins.Manager
}

func (t *Plugin) Reconfigure(context.Context, interface{}) {

}

type registration string

func (t *Plugin) Start(context.Context) error {

	go func() {

		// NOTE: The ticker simulates an asynchronous event that kicks off
		// manual triggers.
		ticker := time.NewTicker(time.Second * 2)

		for range ticker.C {

			// TODO: how to compose a custom plugin like this one and discovery?
			//
			// if p := discovery.Lookup(t.manager); p != nil {
			// 	if ch := p.Trigger(); ch != nil {
			// 		update := make(chan *bundle.Status)
			// 		p.RegisterListener(registration(Name), func(s *bundle.Status) {
			// 			update <- s
			// 		})
			// 		ch <- struct{}{}
			// 		st := <-update
			// 		log.Println("last disco status:", st)
			// 	}
			// }

			if p := bundle.Lookup(t.manager); p != nil {

				// Register for bundle status changes.
				update := make(chan map[string]*bundle.Status)
				p.RegisterBulkListener(registration(Name), func(ss map[string]*bundle.Status) {
					update <- ss
				})

				// Trigger each of the loaders.
				var n int
				for _, l := range p.Loaders() {
					if ch := l.Trigger(); ch != nil {
						ch <- struct{}{}
						n++
					}
				}

				// Expect one status update for each loader. The loaders will
				// deliver exactly one status update in response to triggering.
				var st map[string]*bundle.Status
				for i := 0; i < n; i++ {
					st = <-update
				}

				fmt.Println("Last bundle status:", st)
			}

			// TODO: Decision logger triggering can be done here as well -- if a
			// custom decision logger is enabled then the decision logger loop
			// is a no-op.

			// NOTE: Status triggering is not required since status is only sent
			// in response to bundle and plugin state updates (which are only
			// executed in response to manual triggers.)
		}
	}()

	return nil
}

func (p *Plugin) Stop(context.Context) {

}

type Factory struct{}

func (Factory) New(m *plugins.Manager, config interface{}) plugins.Plugin {

	m.UpdatePluginStatus(Name, &plugins.Status{State: plugins.StateOK})

	return &Plugin{manager: m}
}

func (Factory) Validate(_ *plugins.Manager, config []byte) (interface{}, error) {
	parsedConfig := Config{}
	return parsedConfig, util.Unmarshal(config, &parsedConfig)
}

func main() {

	runtime.RegisterPlugin(Name, Factory{})

	if err := cmd.RootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
