package app

import (
	"context"

	"github.com/oklog/run"
)

type App struct {
	services []Service
	runner   *run.Group
}

func NewApp() *App {
	return &App{
		services: make([]Service, 0),
		runner:   &run.Group{},
	}
}

func (a *App) WithService(s Service) *App {
	a.services = append(a.services, s)
	return a
}

func (a *App) Run(ctx context.Context) error {
	for _, service := range a.services {
		a.runner.Add(actor(ctx, service))
	}

	return a.runner.Run()
}
