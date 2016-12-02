package kinesumer

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	k "github.com/remind101/kinesumer/interface"
)

type ShardWorker struct {
	kinesis                k.Kinesis
	shard                  *kinesis.Shard
	checkpointer           k.Checkpointer
	stream                 string
	pollTime               int
	sequence               string
	stop                   <-chan Unit
	stopped                chan<- Unit
	c                      chan k.Record
	provisioner            k.Provisioner
	errHandler             func(k.Error)
	metricsReporter        k.MetricsReporter
	defaultIteratorType    string
	shardIteratorTimestamp time.Time
	getRecordsThrottle     <-chan time.Time
	GetRecordsLimit        int64
}

func (s *ShardWorker) GetShardIterator(iteratorType string, sequence string, timestamp time.Time) (string, error) {
	var tmp *string
	if len(sequence) > 0 {
		tmp = &sequence
	}
	iter, err := s.kinesis.GetShardIterator(&kinesis.GetShardIteratorInput{
		ShardId:                s.shard.ShardId,
		ShardIteratorType:      &iteratorType,
		StartingSequenceNumber: tmp,
		StreamName:             &s.stream,
		Timestamp:              &timestamp,
	})
	if err != nil {
		return "", err
	}
	return aws.StringValue(iter.ShardIterator), nil
}

func (s *ShardWorker) TryGetShardIterator(iteratorType string, sequence string, timestamp time.Time) string {
	it, err := s.GetShardIterator(iteratorType, sequence, timestamp)
	if err != nil {
		panic(err)
	}
	return it
}

func (s *ShardWorker) GetRecords(it string) ([]*kinesis.Record, string, int64, error) {
	start := time.Now()
	<-s.getRecordsThrottle
	end := time.Now()
	s.metricsReporter.TimeInMilliseconds("kinesumer.get_records.throttle_time", end.Sub(start).Seconds()*1000.0, s.MetricsTags(), 1.0)

	s.metricsReporter.Count("kinesumer.get_records.requests", 1, s.MetricsTags(), 1.0)
	resp, err := s.kinesis.GetRecords(&kinesis.GetRecordsInput{
		Limit:         &s.GetRecordsLimit,
		ShardIterator: &it,
	})
	if err != nil {
		return nil, "", 0, err
	}
	return resp.Records, aws.StringValue(resp.NextShardIterator), aws.Int64Value(resp.MillisBehindLatest), nil
}

func (s *ShardWorker) GetRecordsAndProcess(it, sequence string) (cont bool, nextIt string, nextSeq string) {
	records, nextIt, lag, err := s.GetRecords(it)
	s.metricsReporter.Count("kinesumer.get_records.record_count", int64(len(records)), s.MetricsTags(), 1.0)
	s.metricsReporter.Gauge("kinesumer.get_records.records_received", float64(len(records)), s.MetricsTags(), 1.0)
	if err != nil || len(records) == 0 {
		if err != nil {
			msg := fmt.Sprintf("GetRecords Failed with %d records and lag of %d on shard %s, should wait %d before retrying", len(records), lag, aws.StringValue(s.shard.ShardId), s.pollTime)
			s.errHandler(NewError(EWarn, msg, err))
			s.metricsReporter.Count("kinesumer.get_records.failed", 1, s.MetricsTags(), 1.0)
			nextIt, err = s.GetShardIterator("AFTER_SEQUENCE_NUMBER", sequence, time.Time{})
			if err != nil {
				s.errHandler(NewError(EWarn, "GetShardIterator failed", err))
			}
		}

		if err := s.provisioner.Heartbeat(aws.StringValue(s.shard.ShardId)); err != nil {
			s.errHandler(NewError(EError, "Heartbeat failed", err))
			return true, "", sequence
		}
		// GetRecords is not guaranteed to return records even if there are records to be read.
		// However, if our lag time behind the shard head is <= 3 seconds then there's probably
		// no records.
		if lag <= 3000 /* milliseconds */ {
			s.metricsReporter.Count("kinesumer.get_records.polling", 1, s.MetricsTags(), 1.0)
			select {
			case <-time.NewTimer(time.Duration(s.pollTime) * time.Millisecond).C:
			case <-s.stop:
				return true, "", sequence
			}
		}
	} else {
		for _, rec := range records {
			s.c <- &Record{
				data:               rec.Data,
				partitionKey:       aws.StringValue(rec.PartitionKey),
				sequenceNumber:     aws.StringValue(rec.SequenceNumber),
				shardId:            aws.StringValue(s.shard.ShardId),
				millisBehindLatest: lag,
				checkpointC:        s.checkpointer.DoneC(),
			}

			if err := s.provisioner.Heartbeat(aws.StringValue(s.shard.ShardId)); err != nil {
				s.errHandler(NewError(EError, "Heartbeat failed", err))
				return true, "", sequence
			}
		}
		sequence = aws.StringValue(records[len(records)-1].SequenceNumber)
	}
	return false, nextIt, sequence
}

func (s *ShardWorker) RunWorker() {
	defer func() {
		if val := recover(); val != nil {
			msg := fmt.Sprintf("%v", val)
			s.errHandler(NewError(ECrit, msg, nil))
		}
	}()
	defer func() {
		s.provisioner.Release(aws.StringValue(s.shard.ShardId))
		s.stopped <- Unit{}
	}()

	s.errHandler(NewError(EInfo, "Starting worker", fmt.Errorf("shard: %s", aws.StringValue(s.shard.ShardId))))
	sequence := s.checkpointer.GetStartSequence(aws.StringValue(s.shard.ShardId))
	end := s.shard.SequenceNumberRange.EndingSequenceNumber
	var it string
	if len(sequence) == 0 {
		sequence = aws.StringValue(s.shard.SequenceNumberRange.StartingSequenceNumber)

		s.errHandler(NewError(EWarn, "Using "+s.defaultIteratorType, nil))
		it = s.TryGetShardIterator(s.defaultIteratorType, "", s.shardIteratorTimestamp)
	} else {
		it = s.TryGetShardIterator("AFTER_SEQUENCE_NUMBER", sequence, time.Time{})
	}

	loopStartTime := time.Now()
	watchDog := make(chan Unit)
	go func() {
		lastWatchDogCheckin := time.Now()
		for {
			select {
			case <-watchDog:
				lastWatchDogPet := time.Now()
			case <-time.NewTimer(time.Duraction(1) * time.Second).C:
				elapsedTime := time.Now() - lastWatchDogCheckin
				if elapsedTime > s.provisioner.TTL() { // We missed our Heartbeat
					stackDumpBuffer := make([]byte, 1<<20) // 1 MB stack dump is more than we could possibly need
					stackLen := runtime.Stack(stackDumpBuffer, true)
					err := fmt.Errorf("*** goroutine dump...\n%s\n***", stackDumpBuffer[stackLen])
					s.errHandler(NewError(EError, "Watchdog reached limit", err))
					time.Sleep(s.provisioner.TTL()) // No need to double up a single heartbeat window
					break
				}
			}
		}
	}()
loop:
	for {
		if len(it) == 0 || end != nil && sequence == *end {
			s.errHandler(NewError(EWarn, "Shard has reached its end", nil))
			break loop
		}

		watchDog <- Unit{}
		if err := s.provisioner.Heartbeat(aws.StringValue(s.shard.ShardId)); err != nil {
			s.errHandler(NewError(EError, "Heartbeat failed", err))
			break loop
		}

		select {
		case <-s.stop:
			break loop
		default:
			if brk, nextIt, seq := s.GetRecordsAndProcess(it, sequence); brk {
				break loop
			} else {
				it = nextIt
				sequence = seq
			}
		}
		loopDuration := time.Now().Sub(loopStartTime).Seconds() * 1000.0
		s.metricsReporter.TimeInMilliseconds("kinesumer.shard_worker.loop", loopDuration, s.MetricsTags(), 1.0)
		loopStartTime = time.Now()
	}
}

func (s *ShardWorker) MetricsTags() map[string]string {
	return map[string]string{"shard": aws.StringValue(s.shard.ShardId)}
}
