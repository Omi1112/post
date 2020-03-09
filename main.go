package main

import (
	"github.com/SeijiOmi/posts-service/db"
	"github.com/SeijiOmi/posts-service/server"
)

func main() {
	db.Init()
	server.Init()
	db.Close()
}
