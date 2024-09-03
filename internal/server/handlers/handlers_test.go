package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lambawebdev/metrics/internal/models"
	"github.com/lambawebdev/metrics/internal/server/config"
	"github.com/lambawebdev/metrics/internal/server/middleware"
	"github.com/lambawebdev/metrics/internal/server/storage"
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
		name         string
		readFromFile bool
		want         want
	}{
		{
			name:         "Test ok",
			readFromFile: false,
			want: want{
				code:         200,
				metricValue:  125,
				responseText: `{"Alloc":125}`,
				contentType:  "text/html",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := new(storage.MemStorage)
			storage.Metrics = map[string]interface{}{"Alloc": 125}
			mh := NewMetricHandler(storage)

			w := httptest.NewRecorder()
			mh.GetMetrics(w)

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
			h := NewMetricHandler(storage)

			w := httptest.NewRecorder()
			h.GetMetric(w, request)

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

func TestGetMetricV2(t *testing.T) {
	type want struct {
		code         int
		responseText string
		contentType  string
	}

	tests := []struct {
		name         string
		readFromFile bool
		want         want
		body         *models.Metrics
	}{
		{
			name:         "Test ok",
			readFromFile: false,
			body: &models.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			want: want{
				code:         200,
				responseText: "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":125.44}",
				contentType:  "application/json",
			},
		},
		{
			name:         "Test ok",
			readFromFile: false,
			body: &models.Metrics{
				ID:    "PollCount",
				MType: "counter",
			},
			want: want{
				code:         200,
				responseText: "{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":0}",
				contentType:  "application/json",
			},
		},
		{
			name:         "Test RandomValue",
			readFromFile: false,
			body: &models.Metrics{
				ID:    "RandomValue",
				MType: "gauge",
			},
			want: want{
				code:         200,
				responseText: "{\"id\":\"RandomValue\",\"type\":\"gauge\",\"value\":0}",
				contentType:  "application/json",
			},
		},
		{
			name:         "Test read from file when gauge value not present",
			readFromFile: true,
			body: &models.Metrics{
				ID:    "RandomValue",
				MType: "gauge",
			},
			want: want{
				code:         200,
				responseText: "{\"id\":\"RandomValue\",\"type\":\"gauge\",\"value\":0}",
				contentType:  "application/json",
			},
		},
		{
			name:         "Test read from file when counter value not present",
			readFromFile: true,
			body: &models.Metrics{
				ID:    "RandomValue",
				MType: "counter",
			},
			want: want{
				code:         200,
				responseText: "{\"id\":\"RandomValue\",\"type\":\"counter\",\"delta\":0}",
				contentType:  "application/json",
			},
		},
		{
			name:         "Test read from file when gauge value present",
			readFromFile: true,
			body: &models.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			want: want{
				code:         200,
				responseText: "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":125.44}",
				contentType:  "application/json",
			},
		},
		{
			name:         "Test read from file when not exists metric set",
			readFromFile: true,
			body: &models.Metrics{
				ID:    "GetSetZip",
				MType: "counter",
			},
			want: want{
				code:         200,
				responseText: "{\"id\":\"GetSetZip\",\"type\":\"counter\",\"delta\":0}",
				contentType:  "application/json",
			},
		},
	}

	config.ParseFlags()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.body)
			request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(body))
			request.Header.Set("Content-Type", "application/json")

			s := new(storage.MemStorage)
			s.Metrics = make(map[string]interface{})
			s.AddGauge("Alloc", 125.44)
			mh := NewMetricHandler(s)
			config.SetRestoreMetrics(test.readFromFile)

			w := httptest.NewRecorder()
			mh.GetMetricV2(w, request)

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
			mh := NewMetricHandler(storage)
			mh.UpdateMetric(w, request)

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

func TestUpdateMetricV2(t *testing.T) {
	type want struct {
		code           int
		responseText   string
		contentType    string
		acceptEncoding string
	}

	value1, delta1, delta2 := float64(222.22), int64(5), int64(155)

	tests := []struct {
		name string
		want want
		body *models.Metrics
	}{
		{
			name: "Test update gauge",
			body: &models.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &value1,
			},
			want: want{
				code:           200,
				responseText:   "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":222.22}",
				contentType:    "application/json",
				acceptEncoding: "gzip",
			},
		},
		{
			name: "Test update counter",
			body: &models.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: &delta1,
			},
			want: want{
				code:           200,
				responseText:   "{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":5}",
				contentType:    "application/json",
				acceptEncoding: "gzip",
			},
		},
		{
			name: "Test wrong metric type",
			body: &models.Metrics{
				ID:    "Alloc",
				MType: "wrongType",
				Delta: &delta2,
			},
			want: want{
				code:           400,
				responseText:   "Metric type is not supported!\n",
				contentType:    "text/plain; charset=utf-8",
				acceptEncoding: "gzip",
			},
		},
		{
			name: "Test value is not present",
			body: &models.Metrics{
				ID:    "PollCount",
				MType: "counter",
			},
			want: want{
				code:           400,
				responseText:   "delta have to be present\n",
				contentType:    "text/plain; charset=utf-8",
				acceptEncoding: "gzip",
			},
		},
	}

	storage := new(storage.MemStorage)
	storage.Metrics = make(map[string]interface{})
	mh := NewMetricHandler(storage)

	handler := http.HandlerFunc(middleware.GzipMiddleware(func(w http.ResponseWriter, r *http.Request) {
		mh.UpdateMetricV2(w, r)
	}))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.body)
			request := httptest.NewRequest(http.MethodPost, srv.URL, bytes.NewBuffer(body))
			request.RequestURI = ""
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Accept-Encoding", "gzip")

			res, err := http.DefaultClient.Do(request)
			require.NoError(t, err)

			defer res.Body.Close()

			zr, err := gzip.NewReader(res.Body)
			require.NoError(t, err)

			body, err = io.ReadAll(zr)
			require.NoError(t, err)

			require.NoError(t, err)
			assert.Equal(t, test.want.responseText, string(body))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.acceptEncoding, res.Header.Get("Content-Encoding"))
		})
	}
}
