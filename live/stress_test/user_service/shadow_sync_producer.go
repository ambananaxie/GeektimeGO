package main

import (
	"github.com/Shopify/sarama"
)

type ShadowSyncProducer struct {

}

func (s *ShadowSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowSyncProducer) Close() error {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowSyncProducer) IsTransactional() bool {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowSyncProducer) BeginTxn() error {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowSyncProducer) CommitTxn() error {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowSyncProducer) AbortTxn() error {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowSyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowSyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	// TODO implement me
	panic("implement me")
}

