package main

import (
	"io/ioutil"
	"os"

	"github.com/NickRI/wallets-task/transport/restapi"
	"github.com/go-chi/docgen"
	"github.com/go-kit/kit/log"
)

const apiInto = `# Wallet service
This service provide three api endpoints for account and payments actions
`

func main() {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	routes := restapi.MakeRoutes(nil, logger)
	doc := docgen.MarkdownRoutesDoc(routes, docgen.MarkdownOpts{ProjectPath: "github.com/NickRI/wallets-task", Intro: apiInto})

	if err := ioutil.WriteFile("./docs/api.md", []byte(doc), os.ModePerm); err != nil {
		logger.Log("error", err)
	}
}
