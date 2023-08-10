package graphql

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/TheRafaBonin/roxy"
	"github.com/gothunder/thunder/pkg/log"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// HandleError handles any error into a gqlerror
func HandleError(ctx context.Context, err error) *gqlerror.Error {
	// Declare some variables
	var gqlErr *gqlerror.Error
	var message string
	var status int

	// Get response
	httpResponse := roxy.GetDefaultHTTPResponse(err)
	message = httpResponse.Message
	status = httpResponse.Status

	// Address OK StatusES (no error)
	is2xx, matchErr := regexp.MatchString("^2[0-9]{2}$", fmt.Sprint(status))
	if matchErr != nil {
		matchErr = roxy.Wrap(matchErr, "error matching 2xx error status")
		log.LogError(ctx, matchErr)
	}

	if is2xx {
		return nil
	}

	// If error is 4xx, we always want to log it as info
	is4xx, matchErr := regexp.MatchString("^4[0-9]{2}$", fmt.Sprint(status))
	if matchErr != nil {
		matchErr = roxy.Wrap(matchErr, "error matching 4xx error status")
		log.LogError(ctx, matchErr)
	}

	if !is4xx {
		err = roxy.SetErrorLogLevel(err, roxy.InfoLevel)
	}

	// If error is 5xx, we always want to omit the error
	is5xx, matchErr := regexp.MatchString("^5[0-9]{2}$", fmt.Sprint(status))
	if matchErr != nil {
		matchErr = roxy.Wrap(matchErr, "error matching 5xx error status")
		log.LogError(ctx, matchErr)
	}

	if !is5xx {
		message = http.StatusText(status)
	}

	// Address the empty message case
	if message == "" {
		message = http.StatusText(status)
	}

	// Logs error and returns the gqlError
	log.LogError(ctx, err)
	gqlErr = &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"status": status,
		},
	}

	return gqlErr
}
