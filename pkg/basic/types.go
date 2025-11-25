/*
Basic Constructor Pattern
*/

package basic

type App struct {
	Name    string
	Version string
	Author  string
}

func NewApp(name, version, author string) *App {
	return &App{
		Name:    name,
		Version: version,
		Author:  author,
	}
}
