package message

import (
	"context"
	"encoding/json"
	"example/kafka/user"
	"example/kafka/words"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/lovoo/goka"
)

var (
	MessageFilterGroup goka.Group = "message-filter"
)

// Message — сообщение от пользователя SenderId к пользователю RecepientId
type Message struct {
	SenderId    int
	RecipientId int
	Text        string
	CreatedAt   string
}

type MessageCodec struct{}

// Encode переводит Message в []byte
func (uc *MessageCodec) Encode(value any) ([]byte, error) {
	if _, isMessage := value.(*Message); !isMessage {
		return nil, fmt.Errorf("expected type is *Message, received %T", value)
	}
	return json.Marshal(value)
}

// Decode переводит user из []byte в структуру user.
func (uc *MessageCodec) Decode(data []byte) (any, error) {
	var (
		c   Message
		err error
	)
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("deserialization error: %v", err)
	}
	return &c, nil
}

func shouldDrop(ctx goka.Context, m *Message) bool {
	v := ctx.Lookup(goka.GroupTable(user.BlockedUsersGroup), strconv.Itoa(m.RecipientId))
	shouldDropVal := v != nil && v.(*user.UserBlockValue).Blocks[m.SenderId]

	if shouldDropVal {
		log.Printf("Message from %v to %v will NOT be sent: '%s'", m.SenderId, m.RecipientId, m.Text)
	}

	return shouldDropVal
}

func processMessageText(ctx goka.Context, m *Message) *Message {
	// Ограничимся отсутствием знаков препинания для простоты обработки
	words_array := strings.Split(strings.Trim(m.Text, " \r\n"), " ")
	for i, w := range words_array {
		if entry := ctx.Lookup(goka.GroupTable(words.MaskingWordsGroup), strings.ToLower(w)); entry != nil && entry.(*words.WordToMaskValue).IsActive {
			words_array[i] = strings.Repeat("*", len(w))
		} else {
			words_array[i] = w
		}
	}

	return &Message{
		SenderId:    m.SenderId,
		RecipientId: m.RecipientId,
		Text:        strings.Join(words_array, " "),
		CreatedAt:   m.CreatedAt,
	}
}

func RunMessageFilter(brokers []string, messageTopic goka.Stream, filteredMessagesTopic goka.Stream) {
	g := goka.DefineGroup(MessageFilterGroup,
		goka.Input(messageTopic, new(MessageCodec), func(ctx goka.Context, msg interface{}) {
			if shouldDrop(ctx, msg.(*Message)) {
				return
			}
			m := processMessageText(ctx, msg.(*Message))

			log.Printf("Message from %v to %v is sending: '%s'", m.SenderId, m.RecipientId, m.Text)

			ctx.Emit(filteredMessagesTopic, ctx.Key(), m)
		}),
		goka.Output(filteredMessagesTopic, new(MessageCodec)),
		goka.Lookup(goka.GroupTable(user.BlockedUsersGroup), new(user.UserBlockValueCodec)),
		goka.Lookup(goka.GroupTable(words.MaskingWordsGroup), new(words.WordToMaskValueCodec)),
	)

	p, err := goka.NewProcessor(brokers, g)
	if err != nil {
		log.Fatal(err)
	}
	err = p.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
