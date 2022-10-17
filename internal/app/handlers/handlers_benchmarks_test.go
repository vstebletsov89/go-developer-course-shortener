package handlers

import (
	"bytes"
	"context"
	"github.com/gorilla/mux"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
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
//go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
//В случае успешной оптимизации вы увидите в выводе командной строки результаты
//с отрицательными значениями, означающими уменьшение потребления ресурсов.
//Внимание: к концу текущего спринта покрытие вашего кода автотестами должно быть не менее 40%.

func testBenchmarkRequest(b *testing.B, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, _ := http.NewRequest(method, ts.URL+path, body)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	b.StartTimer() // start timer
	resp, _ := client.Do(req)
	b.StopTimer() // stop all timers

	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	return resp, string(respBody)
}

func AuthHandleMockBenchmark(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserCtx, "4b003ed0-4d8f-46eb-8322-e90174110517")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewRouterBenchmark(config *configs.Config) *mux.Router {
	var storage repository.Repository
	if config.FileStoragePath != "" {
		storage = repository.NewFileRepository(config.FileStoragePath)
	} else {
		storage = repository.NewInMemoryRepository()
	}
	handler := NewHTTPHandler(config, storage)

	log.Printf("Server started on %v", config.ServerAddress)
	type server struct {
		router *mux.Router
	}
	s := &server{
		router: mux.NewRouter(),
	}
	s.router.Use(AuthHandleMockBenchmark)
	s.router.HandleFunc("/", handler.HandlerPOST).Methods(http.MethodPost)
	s.router.HandleFunc("/api/shorten", handler.HandlerJSONPOST).Methods(http.MethodPost)
	s.router.HandleFunc("/api/shorten/batch", handler.HandlerBatchPOST).Methods(http.MethodPost)
	s.router.HandleFunc("/{ID}", handler.HandlerGET).Methods(http.MethodGet)
	s.router.HandleFunc("/api/user/urls", handler.HandlerUserStorageGET).Methods(http.MethodGet)
	s.router.HandleFunc("/ping", handler.HandlerPing).Methods(http.MethodGet)

	return s.router
}

//TODO: refactor to benchmark test (make separate save and get?)
//go test -bench . -benchmem -benchtime 10s -memprofile base.pprof
//go test -bench . -benchmem -benchtime 10s -memprofile result.pprof
//go tool pprof -http=":9090"  bench.test base.pprof
//go tool pprof -http=":9095"  bench.test result.pprof
//go tool pprof base.pprof
//go tool pprof result.pprof
//go tool pprof -http=':8081' -diff_base base.pprof result.pprof
//go tool pprof -top -http=':8081' -diff_base base.pprof result.pprof

func BenchmarkBothHandlersMemoryStorage(b *testing.B) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouterBenchmark(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	b.ResetTimer() // reset timer

	counter := 1 // counter for unique urls
	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop timer
		originalURL := "https://github.com/test_repo" + strconv.Itoa(counter)

		// prepare short url
		resp, body := testBenchmarkRequest(b, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))

		shortURL, _ := url.Parse(body)
		resp.Body.Close()

		// get original url
		resp, _ = testBenchmarkRequest(b, ts, http.MethodGet, shortURL.Path, nil)

		resp.Body.Close()
		counter++
	}
}
