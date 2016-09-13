package httperror

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCreation(t *testing.T) {
	eMsg := "These aren't the errors you're looking for."
	e := New(http.StatusNotFound, "err_not_found", eMsg)

	assert.Equal(t, http.StatusNotFound, e.Status(),
		"Error should make its status code accessible.")
	assert.Equal(t, "err_not_found", e.ID(), "Unique error ID should match.")
	assert.Equal(t, eMsg, e.Message(), "Error should have a message.")
	assert.Equal(t, nil, e.Detail(),
		"Newly created error should not have detail when it was not specified.")
}

func TestErrorDetail(t *testing.T) {
	e := New(http.StatusNotFound, "err_missing_sanity", "Missing sanity.")
	detailMsg := "Likely time of loss: when you started developing software."
	e_ := e.WithDetail(detailMsg)

	assert.True(t, e.Equal(e_), "Errors should identify as equal for ease of comparison.")
	assert.Equal(t, detailMsg, e_.Detail(), "Errors with detail should provide access to it.")
	assert.Nil(t, e.Detail(), "Original error should not retain detail.")
}

func TestMarshalling(t *testing.T) {
	e := New(http.StatusInternalServerError, "err_missing_server", "Missing server.")
	repr, err := e.Marshal()
	assert.NoError(t, err, "Marshalling of error should not fail.")
	assert.NotNil(t, repr, "Marshalled representation should not be nil.")
}

func ExampleError_detail() {
	// globally, within your package
	var (
		ErrProcessingFailed = New(http.StatusInternalServerError, "processing_fail",
			"Processing of the specified person failed.")
	)

	// within a request handler function
	if err := DoRiskyProcessing(); err != nil {
		return ErrProcessingFailed.WithDetail(err)
	}
}
