package bot

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCommandParsingCommandForOneDay(t *testing.T) {

	// When:
	shortUrlID, dayDate, err := extractIdAndDate("stats5x20201201")

	// Then:
	assert.Nil(t, err)
	assert.Equal(t, 5, shortUrlID)

	// and
	expectedDate := time.Date(2020, time.December, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedDate, dayDate)
}

func TestCommandExtractIdFromDelete(t *testing.T) {

	// When:
	strID, intID := extractIDFromDeleteCommand("delete8")

	// Then:
	assert.Equal(t, "8", strID)
	assert.Equal(t, 8, intID)
}

func TestCommandExtractIdFromDeleteBigNumber(t *testing.T) {

	// When:
	strID, intID := extractIDFromDeleteCommand("delete678904")

	// Then:
	assert.Equal(t, "678904", strID)
	assert.Equal(t, 678904, intID)
}
