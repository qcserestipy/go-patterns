/*
Functional Options Pattern
*/

package funcopts

import "fmt"

type App struct {
	name, version, author string

	users []string
	user  string

	release  string
	releases []string

	httpOnly bool
}

type AppOption func(*App)

func NewApp(name, version, author string, opts ...AppOption) (*App, error) {
	app := &App{
		name:    name,
		version: version,
		author:  author,
	}

	for _, opt := range opts {
		opt(app)
	}

	err := validate(app)
	if err != nil {
		return nil, fmt.Errorf("something went wrong: %v", err)
	}

	return app, nil
}

func WithRelease(release string) AppOption {
	return func(app *App) {
		app.release = release
	}
}

func WithReleases(releases []string) AppOption {
	return func(app *App) {
		app.releases = releases
	}
}

func WithUser(user string) AppOption {
	return func(app *App) {
		app.user = user
	}
}

func WithUsers(user []string) AppOption {
	return func(app *App) {
		app.users = user
	}
}

func WithHttpOnly() AppOption {
	return func(app *App) {
		app.httpOnly = true
	}
}

func validate(a *App) error {
	if len(a.users) > 0 && a.user != "" {
		return fmt.Errorf("cannot set user: %v and users: %v simultaneously", a.user, a.users)
	}
	if len(a.releases) > 0 && a.release != "" {
		return fmt.Errorf("cannot set release: %v and releases: %v simultaneously", a.release, a.releases)
	}
	return nil
}
