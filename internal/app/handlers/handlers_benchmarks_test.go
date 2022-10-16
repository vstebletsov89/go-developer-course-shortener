package handlers

import (
	"bytes"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
)

//TODO: REMOVE comments below:
//Theory:
//func BenchmarkXxx(b *testing.B)
//package main
//
//import "testing"
//
//func BenchmarkFibo(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		FiboRecursive(20)
//	}
//}
//Поэтому для точного подсчёта времени у testing.B есть методы, которые:
//останавливают подсчёт — b.StopTimer();
//возобновляют его — b.StartTimer().
//func BenchmarkSortSlice(b *testing.B) {
//	rand.Seed(time.Now().UnixNano())
//
//	for i := 0; i < b.N; i++ {
//		b.StopTimer() // останавливаем таймер
//		slice := make([]int, 10000)
//		for i := 0; i < len(slice); i++ {
//			slice[i] = rand.Intn(1000)
//		}
//		b.StartTimer() // возобновляем таймер
//
//		// сортируем
//		sort.Slice(slice, func(i, j int) bool {
//			return slice[i] < slice[j]
//		})
//	}
//}
//go test -bench . -benchmem -memprofile base.pprof
//# открыть журнал профилирования в браузере
//go tool pprof -http=":9090" bench.test cpu.out
// go tool pprof -http=":9090" bench.test base.pprof
//Инкремент 15
//Добавьте в свой проект бенчмарки, измеряющие скорость выполнения важнейших компонентов вашей системы.
//Проведите анализ использования памяти вашим проектом,
//определите и исправьте неэффективные части кода по следующему алгоритму:
//Используя профилировщик pprof,
//сохраните профиль потребления памяти вашим проектом в директорию profiles с именем base.pprof.
//Изучите полученный профиль, определите и исправьте неэффективные части вашего кода.
//Повторите пункт 1 и сохраните новый профиль потребления памяти в директорию profiles с именем result.pprof.
//Проверьте результат внесённых изменений командой:
//pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
//В случае успешной оптимизации вы увидите в выводе командной строки результаты
//с отрицательными значениями, означающими уменьшение потребления ресурсов.
//Внимание: к концу текущего спринта покрытие вашего кода автотестами должно быть не менее 40%.

func testBenchmarkRequest(b *testing.B, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	assert.NoError(b, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	assert.NoError(b, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(b, err)
	defer resp.Body.Close()

	return resp, string(respBody)
}

func AuthHandleMockBenchmark(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserCtx, "4b003ed0-4d8f-46eb-8322-e90174110517")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewRouterBenchmark(config *configs.Config) chi.Router {
	var storage repository.Repository
	if config.FileStoragePath != "" {
		storage = repository.NewFileRepository(config.FileStoragePath)
	} else {
		storage = repository.NewInMemoryRepository()
	}
	handler := NewHTTPHandler(config, storage)

	log.Printf("Server started on %v", config.ServerAddress)
	r := chi.NewRouter()
	r.Use(AuthHandleMockBenchmark)

	// маршрутизация запросов обработчику
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Get("/{ID}", handler.HandlerGET)

	return r
}

//TODO: refactor to benchmark test (make separate save and get?)
func BenchmarkBothHandlersFileStorage(b *testing.B) {
	homeDir, err := os.UserHomeDir()
	assert.NoError(b, err)

	file, err := os.CreateTemp(homeDir, "test")
	assert.NoError(b, err)
	defer os.RemoveAll(file.Name())

	log.Printf("Temporary file name: %s", file.Name())

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: file.Name(),
	}
	r := NewRouterBenchmark(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	b.ResetTimer() // reset all timers

	counter := 1
	for i := 0; i < b.N; i++ {
		originalURL := "https://github.com/test_repo" + strconv.Itoa(counter)

		// prepare short url
		b.StartTimer() // start timer
		resp, body := testBenchmarkRequest(b, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
		b.StopTimer() // stop all timers

		assert.Equal(b, http.StatusCreated, resp.StatusCode)
		shortURL, err := url.Parse(body)
		assert.NoError(b, err)
		resp.Body.Close()

		// get original url
		b.StartTimer() // start timer
		resp, _ = testBenchmarkRequest(b, ts, http.MethodGet, shortURL.Path, nil)
		b.StopTimer() // stop all timers

		assert.Equal(b, "https://github.com/test_repo"+strconv.Itoa(counter), resp.Header.Get("Location"))
		assert.Equal(b, http.StatusTemporaryRedirect, resp.StatusCode)
		resp.Body.Close()
		counter++
	}
}

//TODO: refactor to benchmark test?
//func TestBothHandlersMemoryStorage(t *testing.T) {
//	config := &configs.Config{
//		ServerAddress:   "localhost:8080",
//		BaseURL:         "http://localhost:8080",
//		FileStoragePath: "",
//	}
//	r := NewRouterBenchmark(config)
//	ts := httptest.NewServer(r)
//	defer ts.Close()
//
//	//сначала подготавливаем сокращенную ссылку через POST
//	originalURL := "https://github.com/test_repo1"
//	resp, body := testBenchmarkRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
//	defer resp.Body.Close()
//	assert.Equal(t, http.StatusCreated, resp.StatusCode)
//	shortURL, err := url.Parse(body)
//	assert.NoError(t, err)
//
//	//получаем оригинальную ссылку через GET запрос
//	resp, _ = testBenchmarkRequest(t, ts, http.MethodGet, shortURL.Path, nil)
//	defer resp.Body.Close()
//	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
//	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
//}
//
////TODO: refactor to benchmark test?
//func TestBothHandlersWithJSON(t *testing.T) {
//	config := &configs.Config{
//		ServerAddress:   "localhost:8080",
//		BaseURL:         "http://localhost:8080",
//		FileStoragePath: "",
//	}
//	r := NewRouterBenchmark(config)
//	ts := httptest.NewServer(r)
//	defer ts.Close()
//
//	//сначала подготавливаем сокращенную ссылку через POST /api/shorten
//	originalURL := "https://github.com/test_repo1"
//	resp, body := testBenchmarkRequest(t, ts, http.MethodPost, "/api/shorten", bytes.NewBufferString(`{"url": "https://github.com/test_repo1"}`))
//	defer resp.Body.Close()
//	assert.Equal(t, http.StatusCreated, resp.StatusCode)
//	var response types.ResponseJSON
//	err := json.Unmarshal([]byte(body), &response)
//	assert.NoError(t, err)
//	shortURL, err := url.Parse(response.Result)
//	assert.NoError(t, err)
//
//	//получаем оригинальную ссылку через GET запрос
//	resp, _ = testBenchmarkRequest(t, ts, http.MethodGet, shortURL.Path, nil)
//	defer resp.Body.Close()
//	assert.Equal(t, originalURL, resp.Header.Get("Location"))
//	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
//}
