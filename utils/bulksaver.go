package utils

type BulkSaver[T any] interface {
	BulkSave(objs []T) error
}

type BulkSaveHelper[T any] struct {
	objPool   []T
	size      int
	bulkSaver BulkSaver[T]
}

func NewBulkSaveHelper[T any](bulkSaver BulkSaver[T], size int) *BulkSaveHelper[T] {
	return &BulkSaveHelper[T]{objPool: make([]T, 0), size: size, bulkSaver: bulkSaver}
}

func (bsh *BulkSaveHelper[T]) Add(ojb T) {
	bsh.objPool = append(bsh.objPool, ojb)
}

func (bsh *BulkSaveHelper[T]) BulkSave() error {
	err := bsh.bulkSaver.BulkSave(bsh.objPool)
	return err
}

func (bsh *BulkSaveHelper[T]) SaveInBatches() error {
	for len(bsh.objPool) != 0 {
		offset := min(len(bsh.objPool), bsh.size)

		err := bsh.bulkSaver.BulkSave(bsh.objPool[:offset])
		if err != nil {
			return err
		}
		bsh.objPool = bsh.objPool[offset:]
	}
	return nil
}
