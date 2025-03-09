package words

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lovoo/goka"
)

var (
	MaskingWordsGroup goka.Group = "masking-words"
)

type WordToMaskEvent struct {
	IsActive bool
}

type WordToMaskEventCodec struct{}

func (uc *WordToMaskEventCodec) Encode(value any) ([]byte, error) {
	if _, isUserBlock := value.(*WordToMaskEvent); !isUserBlock {
		return nil, fmt.Errorf("expected type is *WordToMaskEvent, received %T", value)
	}
	return json.Marshal(value)
}

func (uc *WordToMaskEventCodec) Decode(data []byte) (any, error) {
	var (
		c   WordToMaskEvent
		err error
	)
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("deserialization error: %v", err)
	}
	return &c, nil
}

type WordToMaskValue struct {
	IsActive bool
}

type WordToMaskValueCodec struct{}

func (uc *WordToMaskValueCodec) Encode(value any) ([]byte, error) {
	if _, isUserBlock := value.(*WordToMaskValue); !isUserBlock {
		return nil, fmt.Errorf("expected type is *WordToMaskValue, received %T", value)
	}
	return json.Marshal(value)
}

func (uc *WordToMaskValueCodec) Decode(data []byte) (any, error) {
	var (
		c   WordToMaskValue
		err error
	)
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("deserialization error: %v", err)
	}
	return &c, nil
}

func store(ctx goka.Context, msg interface{}) {
	var maskEvent *WordToMaskEvent
	var ok bool
	var maskValue *WordToMaskValue

	if val := ctx.Value(); val != nil {
		maskValue = val.(*WordToMaskValue)
	} else {
		maskValue = &WordToMaskValue{IsActive: false}
	}

	if maskEvent, ok = msg.(*WordToMaskEvent); !ok || maskEvent == nil {
		return
	}

	maskValue.IsActive = maskEvent.IsActive
	ctx.SetValue(maskValue)
	log.Printf("[proc] key: %s,  msg: %+v, data in group_table %v \n", ctx.Key(), maskEvent, maskValue)
}

func RunWordsProcessor(brokers []string, wordsToMaskStream goka.Stream) {
	g := goka.DefineGroup(MaskingWordsGroup,
		goka.Input(wordsToMaskStream, new(WordToMaskEventCodec), store),
		goka.Persist(new(WordToMaskValueCodec)),
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
