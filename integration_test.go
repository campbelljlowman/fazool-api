//go:build integration
// +build integration

package main

import (
	"testing"
	"time"

	"github.com/campbelljlowman/fazool-api/api"
)

func TestIntegrations(t *testing.T){
	router := api.InitializeRoutes()
	go router.Run()
	time.Sleep(3 * time.Second)
	t.Run("test1", testHello1)
	t.Run("test2", testHello2)
}

func testHello2(t *testing.T){
	test := "hello2"
	print(test)
}

func testHello1(t *testing.T){
	test := "hello1"
	print(test)
}

func testNotT(){

}