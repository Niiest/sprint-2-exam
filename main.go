package main

import (
	"example/kafka/message"
	"example/kafka/user"
	"example/kafka/words"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lovoo/goka"
)

var brokers = []string{"127.0.0.1:9094"}

var (
	Emitter2FilterMessagesStream  goka.Stream = "messages"
	Filter2FilteredMessagesStream goka.Stream = "filtered-messages"
	Users2BlockStream             goka.Stream = "users-to-block"
	Words2MaskStream              goka.Stream = "words-to-mask"
)

func main() {
	go words.RunWordsProcessor(brokers, Words2MaskStream)
	go user.RunUserBlocker(brokers, Users2BlockStream)
	go message.RunMessageFilter(brokers, Emitter2FilterMessagesStream, Filter2FilteredMessagesStream)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigchan
	fmt.Printf("Received signal %v: stopping app\n", sig)
}
