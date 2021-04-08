package crust

import (
	"github.com/dgraph-io/badger"
)

func Unseal(item *badger.Item) ([]byte, error) {
	value, err := item.ValueCopy(nil)
	if err != nil {
		return value, err
	}

	if ok, si := TryGetSealedInfo(value); ok {
		// TODO: Loop for unseal and delete
		return sw.unseal(si.Sbs[0].Path)
	}

	return value, nil
}

func GetSize(item *badger.Item) (int, error) {
	value, err := item.ValueCopy(nil)
	if err != nil {
		return 0, err
	}

	if ok, si := TryGetSealedInfo(value); ok {
		// TODO: Unseal
		// TODO: Loop for unseal and delete
		return si.Sbs[0].Size, nil
	}

	return len(value), nil
}
