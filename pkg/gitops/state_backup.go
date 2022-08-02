package gitops

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/tidwall/gjson"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type SyncState struct {
	ctx           context.Context
	redisHost     string
	redis         *redis.Client
	saveGMHashes  chan interface{}
	saveK8sHashes chan interface{}

	previousK8sHashes map[string]uint64 // no lock because we only replace the whole map at once
	previousGMHashes  map[string]uint64 // no lock because we only replace the whole map at once
}

// FilterChangedK8s takes Grey Matter config objects, and returns a filtered version of that list, updating the stored
// hashes as a side-effect which don't contain any objects that are the same since the last update. The purpose is to
// return only objects that need to be applied to the environment.
// TODO also return deleted list
func (ss *SyncState) FilterChangedK8s(manifestObjects []client.Object) (filtered []client.Object, deleted []string) {
	newHashes := make(map[string]uint64)
	for _, manifestObject := range manifestObjects {
		// A properly-namespaced key for the object
		key := fmt.Sprintf("%s-%s-%s", manifestObject.GetNamespace(), manifestObject.GetObjectKind(), manifestObject.GetName())
		hash, _ := hashstructure.Hash(manifestObject, hashstructure.FormatV2, nil)
		logger.Info("K8S HASH", "key", key, "hash", hash) // DEBUG
		newHashes[key] = hash                             // store *all* of them in newHashes, to replace previousGMHashes
		// if the hashes don't match, the object has changed, and it should be in the filtered list
		if prevHash, ok := ss.previousK8sHashes[key]; !ok || prevHash != hash {
			filtered = append(filtered, manifestObject)
		}
	}
	// find deleted
	for oldKey := range ss.previousK8sHashes {
		if _, ok := newHashes[oldKey]; !ok {
			deleted = append(deleted, oldKey)
		}
	}

	// save new hash table
	ss.previousK8sHashes = newHashes
	go func() { ss.saveK8sHashes <- struct{}{} }() // asynchronously kick-off asynchronous persistence
	return
}

// FilterChangedGM takes Grey Matter config objects and their kinds, and returned filtered versions of those lists
// which don't contain any objects that are the same since the last update, as well as updating the stored hashes as a
// side-effect. The purpose is to return only objects that need to be applied to the environment.
// TODO also return deleted list
func (ss *SyncState) FilterChangedGM(configObjects []json.RawMessage, kinds []string) (filteredConf []json.RawMessage, filteredKinds []string, deleted []string) {
	newHashes := make(map[string]uint64)
	for i, configObj := range configObjects {
		kind := kinds[i]
		var key string
		keyName := cuemodule.KindToKeyName[kind]
		nameResult := gjson.GetBytes(configObj, keyName)
		zoneResult := gjson.GetBytes(configObj, "zone_key")
		// A properly-namespaced key for the object
		key = fmt.Sprintf("%s-%s-%s", zoneResult.String(), kind, nameResult.String())
		logger.Info("GM HASH", "key", key, "key_name", keyName, "kind", kind) // DEBUG
		hash, _ := hashstructure.Hash(configObj, hashstructure.FormatV2, nil)
		newHashes[key] = hash // store *all* of them in newHashes, to replace previousGMHashes
		if prevHash, ok := ss.previousGMHashes[key]; !ok || prevHash != hash {
			filteredConf = append(filteredConf, configObj)
			filteredKinds = append(filteredKinds, kind)
		}
	}

	// find deleted
	for oldKey := range ss.previousGMHashes {
		if _, ok := newHashes[oldKey]; !ok {
			deleted = append(deleted, oldKey)
		}
	}

	// save new hash table
	ss.previousGMHashes = newHashes
	go func() { ss.saveGMHashes <- struct{}{} }() // asynchronously kick-off asynchronous persistence
	return
}

func newSyncState(defaults cuemodule.Defaults) *SyncState {
	ctx := context.Background() // TODO inject external context

	ss := &SyncState{
		ctx:               ctx,
		redisHost:         defaults.RedisHost,
		redis:             nil, // Filled later by .redisConnect()
		saveGMHashes:      make(chan interface{}),
		saveK8sHashes:     make(chan interface{}),
		previousK8sHashes: make(map[string]uint64),
		previousGMHashes:  make(map[string]uint64),
	}

	// immediately attempt to connect to Redis
	err := ss.redisConnect()
	if err == nil {
		// if we're able to connect immediately, try to load saved GM hashes
		loadedGMHashes := make(map[string]uint64)
		resultGM := ss.redis.Get(ctx, defaults.GitOpsStateKeyGM)
		bsGM, err := resultGM.Bytes()
		if err == nil { // if NO error, unmarshall the map
			err = json.Unmarshal(bsGM, &loadedGMHashes)
			if err == nil { // also no unmarshall error
				ss.previousGMHashes = loadedGMHashes
				logger.Info("Successfully loaded GM object hashes from Redis", "key", defaults.GitOpsStateKeyGM)
			} else {
				logger.Info("Problem unmarshalling GM hashes from Redis", "key", defaults.GitOpsStateKeyGM)
			}
		}
		// if we're able to connect immediately, try to load saved K8s hashes
		loadedK8sHashes := make(map[string]uint64)
		resultK8s := ss.redis.Get(ctx, defaults.GitOpsStateKeyK8s)
		bsK8s, err := resultK8s.Bytes()
		if err == nil { // if NO error, unmarshall the map
			err = json.Unmarshal(bsK8s, &loadedK8sHashes)
			if err == nil { // also no unmarshall error
				ss.previousK8sHashes = loadedK8sHashes
				logger.Info("Successfully loaded K8s object hashes from Redis", "key", defaults.GitOpsStateKeyK8s)
			} else {
				logger.Info("Problem unmarshalling K8s hashes from Redis", "key", defaults.GitOpsStateKeyK8s)
			}
		}
	}

	ss.launchAsyncStateBackupLoop(ctx, defaults)

	return ss
}

func (ss *SyncState) redisConnect() error {
	if ss.redis != nil {
		return nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:       ss.redisHost,
		DB:         0, // TODO don't hard-code this
		MaxRetries: -1,
		// TODO optional configurable credentials
	})
	err := rdb.Ping(ss.ctx).Err()
	if err == nil { // if NO error
		ss.redis = rdb // save client
		logger.Info("Connected to Redis for state backup")
	}
	return err
}

func (ss *SyncState) launchAsyncStateBackupLoop(ctx context.Context, defaults cuemodule.Defaults) {

	go func() {
		// first, wait for a Redis connection
	RetryRedis:
		err := ss.redisConnect()
		if err != nil {
			time.Sleep(30 * time.Second)
			logger.Info(fmt.Sprintf("Waiting another 30 seconds for Redis availability (%v)", err))
			goto RetryRedis
		}

		// then watch the update signal channels and persist the associated key to Redis
		for {
			select {
			case <-ctx.Done():
				return
			case <-ss.saveGMHashes:
				ss.persistStateHashesToRedis(ss.previousGMHashes, defaults.GitOpsStateKeyGM)
			case <-ss.saveK8sHashes:
				ss.persistStateHashesToRedis(ss.previousK8sHashes, defaults.GitOpsStateKeyK8s)
			}
		}

	}()
}

func (ss *SyncState) persistStateHashesToRedis(hashes map[string]uint64, key string) {
	b, _ := json.Marshal(hashes)
	if err := ss.redis.Set(ss.ctx, key, b, 0).Err(); err != nil {
		logger.Error(err, "Failed to save environment state hashes to Redis", "key", key)
	}

}
