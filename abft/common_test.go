package abft

import (
	"github.com/Ncog-Earth-Chain/forest-base/forest"
	"github.com/Ncog-Earth-Chain/forest-base/inter/idx"
	"github.com/Ncog-Earth-Chain/forest-base/inter/pos"
	"github.com/Ncog-Earth-Chain/forest-base/kvdb"
	"github.com/Ncog-Earth-Chain/forest-base/kvdb/memorydb"
	"github.com/Ncog-Earth-Chain/forest-base/utils/adapters"
	"github.com/Ncog-Earth-Chain/forest-base/vecfc"
)

type applyBlockFn func(block *forest.Block) *pos.Validators

// TestForest extends Forest for tests.
type TestForest struct {
	*IndexedForest

	blocks map[idx.Block]*forest.Block

	applyBlock applyBlockFn
}

// FakeForest creates empty abft with mem store and equal weights of nodes in genesis.
func FakeForest(nodes []idx.ValidatorID, weights []pos.Weight, mods ...memorydb.Mod) (*TestForest, *Store, *EventStore) {
	validators := make(pos.ValidatorsBuilder, len(nodes))
	for i, v := range nodes {
		if weights == nil {
			validators[v] = 1
		} else {
			validators[v] = weights[i]
		}
	}

	openEDB := func(epoch idx.Epoch) kvdb.DropableStore {
		return memorydb.New()
	}
	crit := func(err error) {
		panic(err)
	}
	store := NewStore(memorydb.New(), openEDB, crit, LiteStoreConfig())

	err := store.ApplyGenesis(&Genesis{
		Validators: validators.Build(),
		Epoch:      FirstEpoch,
	})
	if err != nil {
		panic(err)
	}

	input := NewEventStore()

	config := LiteConfig()
	lch := NewIndexedForest(store, input, &adapters.VectorToDagIndexer{vecfc.NewIndex(crit, vecfc.LiteConfig())}, crit, config)

	extended := &TestForest{
		IndexedForest: lch,
		blocks:        map[idx.Block]*forest.Block{},
	}

	blockIdx := idx.Block(0)

	err = extended.Bootstrap(forest.ConsensusCallbacks{
		BeginBlock: func(block *forest.Block) forest.BlockCallbacks {
			blockIdx++
			return forest.BlockCallbacks{
				EndBlock: func() (sealEpoch *pos.Validators) {
					// track blocks
					extended.blocks[blockIdx] = block
					if extended.applyBlock != nil {
						return extended.applyBlock(block)
					}
					return nil
				},
			}
		},
	})
	if err != nil {
		panic(err)
	}

	return extended, store, input
}
