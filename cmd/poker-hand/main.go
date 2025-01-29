package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/VOVAN1993/poker_hand/internal/hander"
	"github.com/VOVAN1993/poker_hand/internal/server"
)

func main() {
	fmt.Println("poker-hand")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	handManager := hander.NewHandManager()
	if err := handManager.Start(ctx); err != nil {
		fmt.Println("error starting hand manager:", err.Error())
		os.Exit(1)
	}
	server := server.NewServer(handManager)
	server.Start()
}
