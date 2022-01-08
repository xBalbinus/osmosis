package rate_limit_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	ratelimit "github.com/osmosis-labs/osmosis/x/rate-limit"
)

func TestSizedMap(t *testing.T) {
	valueIndex := 1

	sizedMap := ratelimit.NewSizedMap(5, 2, func() interface{} {
		s := fmt.Sprintf("%d", valueIndex)
		valueIndex++
		return s
	})

	require.Equal(t, 0, sizedMap.Size())

	for i := 1; i <= 10; i++ {
		r := sizedMap.Get(fmt.Sprintf("%d", i))

		require.Equal(t, fmt.Sprintf("%d", i), r)
		require.Equal(t, i, sizedMap.Size())
	}

	for i := 1; i <= 10; i++ {
		r := sizedMap.Get(fmt.Sprintf("%d", i))

		require.Equal(t, fmt.Sprintf("%d", i), r)
		// Size should remain
		require.Equal(t, 10, sizedMap.Size())
	}

	// Add more value
	r := sizedMap.Get("11")
	require.Equal(t, "11", r)
	// In this case, the most old chunk should be cleared, so the size is decreased.
	require.Equal(t, 9, sizedMap.Size())
	// As a result, the oldest 2 values should be cleared too.
	r = sizedMap.Get("1")
	require.Equal(t, "12", r)
	require.Equal(t, 10, sizedMap.Size())
	r = sizedMap.Get("2")
	require.Equal(t, "13", r)
	require.Equal(t, 9, sizedMap.Size())
}

func TestSizedMapBruteRand(t *testing.T) {
	for loop := 0; loop < 10; loop++ {
		valueIndex := 1

		chunkNum := rand.Intn(10) + 2
		limitPerChunk := rand.Intn(100) + 1
		sizedMap := ratelimit.NewSizedMap(chunkNum, limitPerChunk, func() interface{} {
			s := fmt.Sprintf("%d", valueIndex)
			valueIndex++
			return s
		})

		require.Equal(t, 0, sizedMap.Size())

		mapSize := chunkNum * limitPerChunk
		for i := 1; i <= 3000; i++ {
			r := sizedMap.Get(fmt.Sprintf("%d", i))

			require.Equal(t, fmt.Sprintf("%d", i), r)
			if i <= mapSize {
				require.Equal(t, i, sizedMap.Size())
			} else {
				minus := 0
				if i%limitPerChunk != 0 {
					minus = limitPerChunk - i%limitPerChunk
				}

				require.Equal(t, mapSize-minus, sizedMap.Size())
			}

			if i > mapSize && rand.Float64() >= 0.5 {
				fromCache := i - rand.Intn(mapSize-limitPerChunk)
				r = sizedMap.Get(fmt.Sprintf("%d", fromCache))
				require.Equal(t, fmt.Sprintf("%d", fromCache), r)
			}
		}
	}
}
