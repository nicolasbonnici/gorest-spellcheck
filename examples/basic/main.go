package main

import (
	"log"

	"github.com/nicolasbonnici/gorest"
	"github.com/nicolasbonnici/gorest/pluginloader"

	authplugin "github.com/nicolasbonnici/gorest-auth"
	spellcheckplugin "github.com/nicolasbonnici/gorest-spellcheck"
)

func init() {
	pluginloader.RegisterPluginFactory("auth", authplugin.NewPlugin)
	pluginloader.RegisterPluginFactory("spellcheck", spellcheckplugin.NewPlugin)
}

func main() {
	cfg := gorest.Config{
		ConfigPath: ".",
	}

	log.Println("Starting GoREST with Spellcheck Plugin example...")
	log.Println("The spellcheck plugin provides CRUD operations at /api/spellcheck")
	log.Println("Make sure to register and login first using the auth plugin endpoints")

	gorest.Start(cfg)
}
