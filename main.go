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
	/*
		config := goka.DefaultConfig()
		// since the emitter only emits one message, we need to tell the processor
		// to read from the beginning
		// As the processor is slower to start than the emitter, it would not consume the first
		// message otherwise.
		// In production systems however, check whether you really want to read the whole topic on first start, which
		// can be a lot of messages.
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
		goka.ReplaceGlobalConfig(config)
	*/
	go words.RunWordsProcessor(brokers, Words2MaskStream)
	go user.RunUserBlocker(brokers, Users2BlockStream)
	go message.RunMessageFilter(brokers, Emitter2FilterMessagesStream, Filter2FilteredMessagesStream)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigchan
	fmt.Printf("Received signal %v: stopping app\n", sig)
}
