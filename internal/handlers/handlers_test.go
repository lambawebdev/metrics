package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetric(t *testing.T) {
	type want struct {
		code         int
		responseText string
		contentType  string
	}

	type routeParams struct {
		metricType, metricName, metricValue string
	}

	tests := []struct {
		name        string
		want        want
		routeParams routeParams
	}{
		{
			name: "Test ok",
			routeParams: routeParams{
				metricType:  "gauge",
				metricName:  "Alloc",
				metricValue: "155",
			},
			want: want{
				code:         200,
				responseText: "",
				contentType:  "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test wrong metric type",
			routeParams: routeParams{
				metricType:  "wrongType",
				metricName:  "Alloc",
				metricValue: "155",
			},
			want: want{
				code:         400,
				responseText: "Metric type is not supported!\n",
				contentType:  "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test wrong metric value",
			routeParams: routeParams{
				metricType:  "gauge",
				metricName:  "Alloc",
				metricValue: "string",
			},
			want: want{
				code:         400,
				responseText: "",
				contentType:  "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/update", nil)
			request.SetPathValue("type", test.routeParams.metricType)
			request.SetPathValue("name", test.routeParams.metricName)
			request.SetPathValue("value", test.routeParams.metricValue)

			w := httptest.NewRecorder()
			UpdateMetric(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.responseText, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
