package main

import (
	"example/kafka/words"
	"flag"
	"log"
	"strings"

	"github.com/lovoo/goka"
)

var (
	word = flag.String("word", "", "word to mask")
	// У меня никак не получилось считать Bool для false, использую Int
	isActive = flag.Int("isActive", 0, "Is masking active for the specified word")
	broker   = flag.String("broker", "localhost:9094", "boostrap Kafka broker")
	stream   = flag.String("stream", "words-to-mask", "stream name")
)

func main() {
	flag.Parse()

	if *word == "" {
		log.Fatal("не указано слово для маскирования")
	}
	emitter, err := goka.NewEmitter([]string{*broker}, goka.Stream(*stream), new(words.WordToMaskEventCodec))
	if err != nil {
		log.Fatal(err)
	}
	defer emitter.Finish()

	err = emitter.EmitSync(strings.ToLower(strings.Trim(*word, " \n\r")), &words.WordToMaskEvent{IsActive: *isActive > 0})
	if err != nil {
		log.Fatal(err)
	}
}
