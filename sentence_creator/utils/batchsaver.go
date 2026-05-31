package utils

type BatchSaver[T any] interface {
	BulkSave(objs []T) error
}

type BatchSaveHelper[T any] struct {
	objPool      []T
	size         int
	currentIndex int
	saver        BatchSaver[T]
}

func NewBatchSaveHelper[T any](bulkSaver BatchSaver[T], size int) *BatchSaveHelper[T] {
	return &BatchSaveHelper[T]{objPool: make([]T, size), size: size, currentIndex: 0, saver: bulkSaver}
}

func (bsh *BatchSaveHelper[T]) IsFull() bool {
	return bsh.currentIndex >= bsh.size-1
}

func (bsh *BatchSaveHelper[T]) Add(ojb T) error {
	if bsh.IsFull() {
		err := bsh.BulkSave(true)
		if err != nil {
			return err
		}
	}

	bsh.objPool[bsh.currentIndex] = ojb
	bsh.currentIndex++
	return nil
}

func (bsh *BatchSaveHelper[T]) Empty() {
	clear(bsh.objPool)
	bsh.currentIndex = 0
}

func (bsh *BatchSaveHelper[T]) BulkSave(empty bool) error {
	err := bsh.saver.BulkSave(bsh.objPool)
	if err != nil {
		return err
	}
	if empty {
		bsh.Empty()
	}
	return err
}
