package kinesumer

type EmptyCheckpointer struct {
}

func (e *EmptyCheckpointer) DoneC() chan<- *KinesisRecord {
	return nil
}

func (e *EmptyCheckpointer) Begin(chan<- *KinesisRecord) error {
	return nil
}

func (e *EmptyCheckpointer) End() {
}

func (e *EmptyCheckpointer) GetStartSequence(*string) *string {
	return nil
}

func (e *EmptyCheckpointer) Sync() {
}

func (e *EmptyCheckpointer) TryAcquire(shardID *string) error {
	return nil
}

func (e *EmptyCheckpointer) Release(shardID *string) error {
	return nil
}
