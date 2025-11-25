/*
 Fluent Builder (Method Chaining) pattern.
*/

package chainopts

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

func NewApp(name, version, author string) *App {
	app := &App{
		name:    name,
		version: version,
		author:  author,
	}
	return app
}

func (a *App) WithRelease(release string) *App {
	a.release = release
	return a
}

func (a *App) WithReleases(releases []string) *App {
	a.releases = releases
	return a
}

func (a *App) WithUser(user string) *App {
	a.user = user
	return a
}

func (a *App) WithUsers(user []string) *App {
	a.users = user
	return a
}

func (a *App) WithHttpOnly() *App {
	a.httpOnly = true
	return a
}

func (a *App) Build() (*App, error) {
	if len(a.users) > 0 && a.user != "" {
		return nil, fmt.Errorf("cannot set user: %v and users: %v simultaneously", a.user, a.users)
	}
	if len(a.releases) > 0 && a.release != "" {
		return nil, fmt.Errorf("cannot set release: %v and releases: %v simultaneously", a.release, a.releases)
	}
	return a, nil
}
