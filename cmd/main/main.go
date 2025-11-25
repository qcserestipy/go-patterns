package main

import (
	"fmt"
	chainopts "go-patterns/pkg/chain-options"
	funcopts "go-patterns/pkg/functional-options"

	"github.com/sirupsen/logrus"
)

func main() {
	app, err := funcopts.NewApp(
		"coolApp", "v1.0", "patrick",
		funcopts.WithUser("user"),
		funcopts.WithHttpOnly(),
	)
	if err != nil {
		logrus.Errorf("error: %v", err)
		return
	}
	fmt.Println(app)

	app2, err := chainopts.NewApp("coolApp", "v1.0", "patrick").
		WithUser("User2").
		WithRelease("stable").
		WithHttpOnly().
		Build()
	if err != nil {
		logrus.Errorf("error: %v", err)
		return
	}

	fmt.Println(app2)
}
