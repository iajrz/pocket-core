package types

import "sync"

type CustomSyncMap struct {
	sync.RWMutex
	useSyncMap        bool
	customEvidenceMap map[string]interface{}
	syncMap           sync.Map
}

func NewCustomSyncMap(useSyncMap bool) *CustomSyncMap {
	return &CustomSyncMap{
		customEvidenceMap: make(map[string]interface{}),
		useSyncMap:        useSyncMap,
	}
}

func (customSyncMap *CustomSyncMap) Load(key string) (value interface{}, ok bool) {
	if customSyncMap.useSyncMap {
		return customSyncMap.syncMap.Load(key)
	} else {
		customSyncMap.RLock()
		result, ok := customSyncMap.customEvidenceMap[key]
		customSyncMap.RUnlock()
		return result, ok
	}
}

func (customSyncMap *CustomSyncMap) Delete(key string) {
	if customSyncMap.useSyncMap {
		customSyncMap.Delete(key)
	} else {
		customSyncMap.Lock()
		delete(customSyncMap.customEvidenceMap, key)
		customSyncMap.Unlock()
	}
}

func (customSyncMap *CustomSyncMap) Store(key string, value interface{}) {
	if customSyncMap.useSyncMap {
		customSyncMap.syncMap.Store(key, value)
	} else {
		customSyncMap.Lock()
		customSyncMap.customEvidenceMap[key] = value
		customSyncMap.Unlock()
	}
}
