package cryptorand

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "eb7a03c377c28da97ae97884582e6bd07fa44724af99798b42593355e39f82cb", Hash("Lorem Ipsum dolor sit Amet"))
	assert.Equal(t, "b07f8f8aa86baac568c520b032dae399e558fcaa52f6e4ff26c1f8e72ecbafbd", Hash("Lasnkal jasdflksd"))
	assert.Equal(t, "7c9cc35163a2f58b2548597661ac47f63181d222638dd1f1366343fc85af2df2", Hash("adflsdIMN sdas @424324"))
}
