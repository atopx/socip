package socdb

func newDataMap() *dataMap {
	return &dataMap{
		data:      map[dataMapKey]dataMapValue{},
		keyWriter: newKeyWriter(),
	}
}

func (dm *dataMap) get(k dataMapKey) DataType {
	if k == noDataMapKey {
		return nil
	}
	return dm.data[k].data
}

// store stores the value in the dataMap and returns a key for it.
func (dm *dataMap) store(v DataType) (dataMapKey, error) {
	key, err := dm.keyWriter.key(v)
	if err != nil {
		return "", err
	}

	dmv, ok := dm.data[dataMapKey(key)]
	if !ok {
		dmKey := dataMapKey(key)
		dmv = dataMapValue{
			key:  dmKey,
			data: v,
		}
		dm.data[dmKey] = dmv
	}

	return dmv.key, nil
}
