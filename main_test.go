package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе в ответе
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: min(100, len(cafeList["moscow"]))},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/cafe?city=moscow&count="+strconv.Itoa(v.count), nil)

		handler.ServeHTTP(response, req)

		// Проверяем, что запрос успешно обработан
		require.Equal(t, http.StatusOK, response.Code)

		// Убираем лишний перевод строки
		body := strings.TrimSpace(response.Body.String())

		// Считаем количество кафе
		var got int
		if body != "" {
			got = len(strings.Split(body, ","))
		}

		assert.Equal(t, v.want, got,
			"count=%d: ожидалось %d кафе, получено %d", v.count, v.want, got)
	}
}
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/cafe?city=moscow&search="+v.search, nil)

		handler.ServeHTTP(response, req)

		// Проверяем, что запрос успешно обработан
		require.Equal(t, http.StatusOK, response.Code)

		// Убираем лишний перевод строки
		body := strings.TrimSpace(response.Body.String())

		var cafes []string
		if body != "" {
			cafes = strings.Split(body, ",")
		}

		// Проверяем количество найденных кафе
		assert.Equal(t, v.wantCount, len(cafes),
			"search=%q: ожидалось %d кафе, получено %d", v.search, v.wantCount, len(cafes))

		// Проверяем, что каждое кафе содержит искомую подстроку (без учёта регистра)
		searchUpper := strings.ToUpper(v.search)
		for _, cafe := range cafes {
			cafe = strings.TrimSpace(cafe)
			assert.True(t, strings.Contains(strings.ToUpper(cafe), searchUpper),
				"кафе %q не содержит подстроку %q", cafe, v.search)
		}
	}
}
