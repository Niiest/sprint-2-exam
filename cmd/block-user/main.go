package main

import (
	"example/kafka/user"
	"flag"
	"log"
	"strconv"

	"github.com/lovoo/goka"
)

var (
	blockerUserId = flag.String("blockerUserId", "", "user, which blocks another user")
	blockedUserId = flag.String("blockedUserId", "", "user, which is blocked user")
	// У меня никак не получилось считать Bool для false, использую Int
	isBlocked = flag.Int("isBlocked", 0, "0 to unblock a user, otherwise - block")
	broker    = flag.String("broker", "localhost:9094", "boostrap Kafka broker")
	stream    = flag.String("stream", "users-to-block", "stream name")
)

func main() {
	flag.Parse()

	if *blockerUserId == "" {
		log.Fatal("невозможно определить блокирующего пользователя ''")
	}

	if *blockedUserId == "" {
		log.Fatal("невозможно заблокировать блокируемого пользователя ''")
	}

	blockedUserIdAsInt, err := strconv.Atoi(*blockedUserId)
	if err != nil {
		log.Fatalf("Не удалось распарсить ID блокируемого пользователя: %s", err)
	}

	emitter, err := goka.NewEmitter([]string{*broker}, goka.Stream(*stream), new(user.UserBlockEventCodec))
	if err != nil {
		log.Fatal(err)
	}
	defer emitter.Finish()

	err = emitter.EmitSync(*blockerUserId, &user.UserBlockEvent{BlockedUserId: blockedUserIdAsInt, IsBlocked: *isBlocked > 0})
	if err != nil {
		log.Fatal(err)
	}
}
