package kinesumer

type Kinesumer interface {
	Begin() (err error)
	End()
	Records() <-chan *KinesisRecord
}
