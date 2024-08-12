package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lambawebdev/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetrics(t *testing.T) {
	type want struct {
		code         int
		metricValue  float64
		responseText string
		contentType  string
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "Test ok",
			want: want{
				code:         200,
				metricValue:  125,
				responseText: `{"Alloc":125}`,
				contentType:  "application/json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", nil)

			storage := new(storage.MemStorage)
			storage.Metrics = map[string]interface{}{"Alloc": 125}

			w := httptest.NewRecorder()
			GetMetrics(w, request, storage)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.JSONEq(t, test.want.responseText, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetMetric(t *testing.T) {
	type want struct {
		code         int
		metricValue  float64
		responseText string
		contentType  string
	}

	type routeParams struct {
		metricType, metricName string
	}

	tests := []struct {
		name        string
		want        want
		routeParams routeParams
	}{
		{
			name: "Test ok",
			routeParams: routeParams{
				metricType: "gauge",
				metricName: "Alloc",
			},
			want: want{
				code:         200,
				metricValue:  125,
				responseText: "125\n",
				contentType:  "application/json",
			},
		},
		{
			name: "Test metric not found",
			routeParams: routeParams{
				metricType: "gauge",
				metricName: "BuckHashSys",
			},
			want: want{
				code:         404,
				metricValue:  125,
				responseText: "Metric not exists!\n",
				contentType:  "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := fmt.Sprintf("/value/%s/%s", test.routeParams.metricType, test.routeParams.metricName)

			request := httptest.NewRequest(http.MethodPost, url, nil)
			request.SetPathValue("type", test.routeParams.metricType)
			request.SetPathValue("name", test.routeParams.metricName)

			storage := new(storage.MemStorage)
			storage.Metrics = map[string]interface{}{"Alloc": 125}

			w := httptest.NewRecorder()
			GetMetric(w, request, storage)

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
			name: "Test update gauge",
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
			name: "Test update counter",
			routeParams: routeParams{
				metricType:  "counter",
				metricName:  "PollCount",
				metricValue: "5",
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
				responseText: "Metric value not supported!\n",
				contentType:  "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := fmt.Sprintf("/update/%s/%s/%s", test.routeParams.metricType, test.routeParams.metricName, test.routeParams.metricValue)

			request := httptest.NewRequest(http.MethodPost, url, nil)
			request.SetPathValue("type", test.routeParams.metricType)
			request.SetPathValue("name", test.routeParams.metricName)
			request.SetPathValue("value", test.routeParams.metricValue)

			storage := new(storage.MemStorage)
			storage.Metrics = make(map[string]interface{})

			w := httptest.NewRecorder()
			UpdateMetric(w, request, storage)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			log.Print(storage)

			require.NoError(t, err)
			assert.Equal(t, test.want.responseText, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
