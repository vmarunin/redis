package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type RedisTestCase struct {
	Name           string
	Method         string
	Body           string
	URL            string
	ResponseStatus int
	Expected       string
}

func TestOptions(t *testing.T) {
	storage := NewAwfulRedisStorage()
	handler := NewAwfulRedisHandler(storage)
	ts := httptest.NewServer(handler)
	client := &http.Client{Timeout: 30 * time.Second}
	urlPrefix := ts.URL + "/redis/v1/"

	for _, suffix := range []string{"key/", "keys"} {
		req, _ := http.NewRequest("OPTIONS", urlPrefix+suffix, nil)
		req.Header.Add("X-Requested-With", "XMLHttpRequest")
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("bad status code, want: 200, have:%v", resp.StatusCode)
		}
	}
}

func TestPOST(t *testing.T) {
	storage := NewAwfulRedisStorage()
	handler := NewAwfulRedisHandler(storage)
	ts := httptest.NewServer(handler)
	client := &http.Client{Timeout: 30 * time.Second}
	urlPrefix := ts.URL + "/redis/v1/"

	for _, suffix := range []string{"key/", "keys"} {
		req, _ := http.NewRequest("POST", urlPrefix+suffix, strings.NewReader(`{"aaa":"bbb"}`))
		req.Header.Add("X-Requested-With", "XMLHttpRequest")
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("bad status code, want: 400, have:%v", resp.StatusCode)
		}
	}
}

func TestRedis(t *testing.T) {
	// return
	storage := NewAwfulRedisStorage()
	handler := NewAwfulRedisHandler(storage)
	ts := httptest.NewServer(handler)
	client := &http.Client{Timeout: 30 * time.Second}
	urlPrefix := ts.URL + "/redis/v1/"

	testCases := []*RedisTestCase{
		{
			Name:           "Set aaa",
			Method:         "PUT",
			Body:           `{"value":"aaa_val", "ttl":60}`,
			URL:            urlPrefix + "key/aaa",
			ResponseStatus: 200,
			Expected:       `{"ok":false,"value":""}`,
		},
		{
			Name:           "Get keys1",
			Method:         "GET",
			Body:           ``,
			URL:            urlPrefix + "keys",
			ResponseStatus: 200,
			Expected:       `["aaa"]`,
		},
		{
			Name:           "Get keys with bad pattern",
			Method:         "GET",
			Body:           ``,
			URL:            urlPrefix + "keys?pattern=[",
			ResponseStatus: 400,
			Expected:       `syntax error in pattern`,
		},
		{
			Name:           "Get aaa",
			Method:         "GET",
			Body:           ``,
			URL:            urlPrefix + "key/aaa",
			ResponseStatus: 200,
			Expected:       `{"ok":true,"value":"aaa_val"}`,
		},
		{
			Name:           "Delete aaa",
			Method:         "DELETE",
			Body:           ``,
			URL:            urlPrefix + "key/aaa",
			ResponseStatus: 200,
			Expected:       `{"ok":true,"value":"aaa_val"}`,
		},
		{
			Name:           "Set aaa with bad json",
			Method:         "PUT",
			Body:           `{"value":"aaa_val"`,
			URL:            urlPrefix + "key/aaa",
			ResponseStatus: 400,
			Expected:       `can't parse inputunexpected EOF`,
		},
		{
			Name:           "Set aaa with bad ttl",
			Method:         "PUT",
			Body:           `{"value":"aaa_val","ttl":1.0}`,
			URL:            urlPrefix + "key/aaa",
			ResponseStatus: 400,
			Expected:       `can't parse ttlstrconv.ParseInt: parsing "1.0": invalid syntax`,
		},
	}

	for _, item := range testCases {
		ok := t.Run(item.Name, func(t *testing.T) {
			req, _ := http.NewRequest(item.Method, item.URL, strings.NewReader(item.Body))
			req.Header.Add("X-Requested-With", "XMLHttpRequest")
			req.Header.Add("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			respStr := string(respBody)

			t.Logf("\nreq body: %s\nresp body: %s", item.Body, respStr)

			if item.ResponseStatus != resp.StatusCode {
				t.Fatalf("bad status code, want: %v, have:%v", item.ResponseStatus, resp.StatusCode)
			}
			if item.ResponseStatus == http.StatusOK {
				expectedData := map[string]interface{}{}
				expectedDecoder := json.NewDecoder(strings.NewReader(item.Expected))
				expectedDecoder.UseNumber()
				expectedDecoder.Decode(&expectedData)
				responseData := map[string]interface{}{}
				responseDecoder := json.NewDecoder(bytes.NewReader(respBody))
				responseDecoder.UseNumber()
				responseDecoder.Decode(&responseData)
				if !reflect.DeepEqual(responseData, expectedData) {
					t.Log("expected", expectedData)
					t.Log("response", responseData)
					t.Fatalf("bad response body\nexpected body: '%s'\nresp body: '%s'", item.Expected, respStr)
				}
			} else {
				if respStr != item.Expected {
					t.Fatalf("bad response body\nexpected body: '%s'\nresp body: '%s'", item.Expected, respStr)
				}
			}
		})
		if !ok {
			break
		}
	}
}
