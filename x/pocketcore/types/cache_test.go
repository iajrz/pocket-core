package types

import (
	"encoding/hex"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/pokt-network/pocket-core/types"
	"github.com/stretchr/testify/assert"
)

func InitCacheTest() {
	logger := log.NewNopLogger()
	// init cache in memory
	InitConfig(&HostedBlockchains{
		M: make(map[string]HostedBlockchain),
	}, logger, sdk.DefaultTestingPocketConfig(), false)
}

func TestMain(m *testing.M) {
	InitCacheTest()
	m.Run()
	err := os.RemoveAll("data")
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}

func TestIsUniqueProof(t *testing.T) {
	h := SessionHeader{
		ApplicationPubKey:  "0",
		Chain:              "0001",
		SessionBlockHeight: 0,
	}
	e, _ := GetEvidence(h, RelayEvidence, sdk.NewInt(100000))
	p := RelayProof{
		Entropy: 1,
	}
	p1 := RelayProof{
		Entropy: 2,
	}
	assert.True(t, IsUniqueProof(p, e), "p is unique")
	e.AddProof(p)
	SetEvidence(e)
	e, err := GetEvidence(h, RelayEvidence, sdk.ZeroInt())
	assert.Nil(t, err)
	assert.False(t, IsUniqueProof(p, e), "p is no longer unique")
	assert.True(t, IsUniqueProof(p1, e), "p is unique")

}

func TestAllEvidence_AddGetEvidence(t *testing.T) {
	appPubKey := getRandomPubKey().RawString()
	servicerPubKey := getRandomPubKey().RawString()
	clientPubKey := getRandomPubKey().RawString()
	ethereum := hex.EncodeToString([]byte{0001})
	header := SessionHeader{
		ApplicationPubKey:  appPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}
	proof := RelayProof{
		Entropy:            0,
		RequestHash:        header.HashString(), // fake
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		Blockchain:         ethereum,
		Token: AAT{
			Version:              "0.0.1",
			ApplicationPublicKey: appPubKey,
			ClientPublicKey:      clientPubKey,
			ApplicationSignature: "",
		},
		Signature: "",
	}
	SetProof(header, RelayEvidence, proof, sdk.NewInt(100000))
	assert.True(t, reflect.DeepEqual(GetProof(header, RelayEvidence, 0), proof))
}
func BenchmarkAllEvidence_AddEvidence(b *testing.B) {
	benchmarks := []struct {
		name      string
		customMap bool
	}{
		{name: "stdlib sync map", customMap: false},
		{name: "custom lock", customMap: true},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ClearEvidence()
				globalEvidenceSealedMap = NewCustomSyncMap(bm.customMap)
				appPubKey := getRandomPubKey().RawString()
				servicerPubKey := getRandomPubKey().RawString()
				clientPubKey := getRandomPubKey().RawString()
				ethereum := hex.EncodeToString([]byte{0001})
				header := SessionHeader{
					ApplicationPubKey:  appPubKey,
					Chain:              ethereum,
					SessionBlockHeight: 1,
				}
				proof := RelayProof{
					Entropy:            0,
					RequestHash:        header.HashString(), // fake
					SessionBlockHeight: 1,
					ServicerPubKey:     servicerPubKey,
					Blockchain:         ethereum,
					Token: AAT{
						Version:              "0.0.1",
						ApplicationPublicKey: appPubKey,
						ClientPublicKey:      clientPubKey,
						ApplicationSignature: "",
					},
					Signature: "",
				}
				b.StartTimer()
				SetProof(header, RelayEvidence, proof, sdk.NewInt(100000))
			}
		})
	}
}

func BenchmarkAllEvidence_GetEvidence(b *testing.B) {
	benchmarks := []struct {
		name      string
		customMap bool
	}{
		{name: "stdlib sync map", customMap: false},
		{name: "custom lock", customMap: true},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ClearEvidence()
				globalEvidenceSealedMap = NewCustomSyncMap(bm.customMap)
				appPubKey := getRandomPubKey().RawString()
				servicerPubKey := getRandomPubKey().RawString()
				clientPubKey := getRandomPubKey().RawString()
				ethereum := hex.EncodeToString([]byte{0001})
				header := SessionHeader{
					ApplicationPubKey:  appPubKey,
					Chain:              ethereum,
					SessionBlockHeight: 1,
				}
				proof := RelayProof{
					Entropy:            0,
					RequestHash:        header.HashString(), // fake
					SessionBlockHeight: 1,
					ServicerPubKey:     servicerPubKey,
					Blockchain:         ethereum,
					Token: AAT{
						Version:              "0.0.1",
						ApplicationPublicKey: appPubKey,
						ClientPublicKey:      clientPubKey,
						ApplicationSignature: "",
					},
					Signature: "",
				}
				SetProof(header, RelayEvidence, proof, sdk.NewInt(100000))
				b.StartTimer()
				_ = GetProof(header, RelayEvidence, 0)
			}
		})
	}
}

func TestAllEvidence_DeleteEvidence(t *testing.T) {
	appPubKey := getRandomPubKey().RawString()
	servicerPubKey := getRandomPubKey().RawString()
	clientPubKey := getRandomPubKey().RawString()
	ethereum := hex.EncodeToString([]byte{0001})
	header := SessionHeader{
		ApplicationPubKey:  appPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}
	proof := RelayProof{
		Entropy:            0,
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		RequestHash:        header.HashString(), // fake
		Blockchain:         ethereum,
		Token: AAT{
			Version:              "0.0.1",
			ApplicationPublicKey: appPubKey,
			ClientPublicKey:      clientPubKey,
			ApplicationSignature: "",
		},
		Signature: "",
	}
	SetProof(header, RelayEvidence, proof, sdk.NewInt(100000))
	assert.True(t, reflect.DeepEqual(GetProof(header, RelayEvidence, 0), proof))
	GetProof(header, RelayEvidence, 0)
	_ = DeleteEvidence(header, RelayEvidence)
	assert.Empty(t, GetProof(header, RelayEvidence, 0))

	// concurrent Read & Write to GlobalEvidenceSealedMap
	SetProof(header, RelayEvidence, proof, sdk.NewInt(100000))
	assert.True(t, reflect.DeepEqual(GetProof(header, RelayEvidence, 0), proof))

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 0; i < 100; i++ {
			go GetProof(header, RelayEvidence, 0)
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 100; i++ {
			go DeleteEvidence(header, RelayEvidence)
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestAllEvidence_GetTotalProofs(t *testing.T) {
	appPubKey := getRandomPubKey().RawString()
	servicerPubKey := getRandomPubKey().RawString()
	clientPubKey := getRandomPubKey().RawString()
	ethereum := hex.EncodeToString([]byte{0001})
	header := SessionHeader{
		ApplicationPubKey:  appPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}
	header2 := SessionHeader{
		ApplicationPubKey:  appPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 101,
	}
	proof := RelayProof{
		Entropy:            0,
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		RequestHash:        header.HashString(), // fake
		Blockchain:         ethereum,
		Token: AAT{
			Version:              "0.0.1",
			ApplicationPublicKey: appPubKey,
			ClientPublicKey:      clientPubKey,
			ApplicationSignature: "",
		},
		Signature: "",
	}
	proof2 := RelayProof{
		Entropy:            0,
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		RequestHash:        header.HashString(), // fake
		Blockchain:         ethereum,
		Token: AAT{
			Version:              "0.0.1",
			ApplicationPublicKey: appPubKey,
			ClientPublicKey:      clientPubKey,
			ApplicationSignature: "",
		},
		Signature: "",
	}
	SetProof(header, RelayEvidence, proof, sdk.NewInt(100000))
	SetProof(header, RelayEvidence, proof2, sdk.NewInt(100000))
	SetProof(header2, RelayEvidence, proof2, sdk.NewInt(100000)) // different header so shouldn't be counted
	_, totalRelays := GetTotalProofs(header, RelayEvidence, sdk.NewInt(100000))
	assert.Equal(t, totalRelays, int64(2))
}

func TestSetGetSession(t *testing.T) {

	session := NewTestSession(t, hex.EncodeToString(Hash([]byte("foo"))))
	session2 := NewTestSession(t, hex.EncodeToString(Hash([]byte("bar"))))
	SetSession(session)
	s, found := GetSession(session.SessionHeader)
	assert.True(t, found)
	assert.Equal(t, s, session)
	_, found = GetSession(session2.SessionHeader)
	assert.False(t, found)
	SetSession(session2)
	s, found = GetSession(session2.SessionHeader)
	assert.True(t, found)
	assert.Equal(t, s, session2)
}
func BenchmarkSetSession(b *testing.B) {
	benchmarks := []struct {
		name      string
		customMap bool
	}{
		{name: "stdlib sync map", customMap: false},
		{name: "custom lock", customMap: true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.StopTimer()
			session := NewBenchSession(b, hex.EncodeToString(Hash([]byte("foo"))))
			b.StartTimer()
			SetSession(session)
		})
	}
}

func BenchmarkGetSession(b *testing.B) {
	benchmarks := []struct {
		name      string
		customMap bool
	}{
		{name: "stdlib sync map", customMap: false},
		{name: "custom lock", customMap: true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.StopTimer()
			session := NewBenchSession(b, hex.EncodeToString(Hash([]byte("foo"))))
			SetSession(session)
			b.StartTimer()
			_, _ = GetSession(session.SessionHeader)
		})
	}
}
func TestDeleteSession(t *testing.T) {
	session := NewTestSession(t, hex.EncodeToString(Hash([]byte("foo"))))
	SetSession(session)
	DeleteSession(session.SessionHeader)
	_, found := GetSession(session.SessionHeader)
	assert.False(t, found)
}

func TestClearCache(t *testing.T) {
	session := NewTestSession(t, hex.EncodeToString(Hash([]byte("foo"))))
	SetSession(session)
	ClearSessionCache()
	iter := SessionIterator()
	var count = 0
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		count++
	}
	assert.Zero(t, count)
}

func NewTestSession(t *testing.T, chain string) Session {
	appPubKey := getRandomPubKey()
	var vals []sdk.Address
	for i := 0; i < 5; i++ {
		nodePubKey := getRandomPubKey()
		vals = append(vals, sdk.Address(nodePubKey.Address()))
	}
	return Session{
		SessionHeader: SessionHeader{
			ApplicationPubKey:  appPubKey.RawString(),
			Chain:              chain,
			SessionBlockHeight: 1,
		},
		SessionKey:   appPubKey.RawBytes(), // fake
		SessionNodes: vals,
	}
}
func NewBenchSession(b *testing.B, chain string) Session {
	appPubKey := getRandomPubKey()
	var vals []sdk.Address
	for i := 0; i < 5; i++ {
		nodePubKey := getRandomPubKey()
		vals = append(vals, sdk.Address(nodePubKey.Address()))
	}
	return Session{
		SessionHeader: SessionHeader{
			ApplicationPubKey:  appPubKey.RawString(),
			Chain:              chain,
			SessionBlockHeight: 1,
		},
		SessionKey:   appPubKey.RawBytes(), // fake
		SessionNodes: vals,
	}
}
