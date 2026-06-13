/*
 * MinIO Cloud Storage, (C) 2018 MinIO, Inc.
 * PGG Obstor, (C) 2021-2026 PGG, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package net

import (
	"reflect"
	"testing"
)

func TestURLIsEmpty(t *testing.T) {
	testCases := []struct {
		url            URL
		expectedResult bool
	}{
		{URL{}, true},
		{URL{Scheme: "http", Host: "demo"}, false},
		{URL{Path: "path/to/demo"}, false},
	}

	for i, testCase := range testCases {
		result := testCase.url.IsEmpty()

		if result != testCase.expectedResult {
			t.Fatalf("test %v: result: expected: %v, got: %v", i+1, testCase.expectedResult, result)
		}
	}
}

func TestURLString(t *testing.T) {
	testCases := []struct {
		url         URL
		expectedStr string
	}{
		{URL{}, ""},
		{URL{Scheme: "http", Host: "demo"}, "http://demo"},
		{URL{Scheme: "https", Host: "demo:443"}, "https://demo"},
		{URL{Scheme: "https", Host: "demo.obstor.net:80"}, "https://demo.obstor.net:80"},
		{URL{Scheme: "https", Host: "23.148.200.3:9000", Path: "/"}, "https://23.148.200.3:9000/"},
		{URL{Scheme: "https", Host: "s3.amazonaws.com", Path: "/", RawQuery: "location"}, "https://s3.amazonaws.com/?location"},
		{URL{Scheme: "http", Host: "myobstor:10000", Path: "/mybucket/myobject"}, "http://myobstor:10000/mybucket/myobject"},
		{URL{Scheme: "ftp", Host: "myftp.server:10000", Path: "/myuser"}, "ftp://myftp.server:10000/myuser"},
		{URL{Path: "path/to/demo"}, "path/to/demo"},
	}

	for i, testCase := range testCases {
		str := testCase.url.String()

		if str != testCase.expectedStr {
			t.Fatalf("test %v: string: expected: %v, got: %v", i+1, testCase.expectedStr, str)
		}
	}
}

func TestURLMarshalJSON(t *testing.T) {
	testCases := []struct {
		url          URL
		expectedData []byte
		expectErr    bool
	}{
		{URL{}, []byte(`""`), false},
		{URL{Scheme: "http", Host: "demo"}, []byte(`"http://demo"`), false},
		{URL{Scheme: "https", Host: "demo.obstor.net:0"}, []byte(`"https://demo.obstor.net:0"`), false},
		{URL{Scheme: "https", Host: "23.148.200.3:9000", Path: "/"}, []byte(`"https://23.148.200.3:9000/"`), false},
		{URL{Scheme: "https", Host: "s3.amazonaws.com", Path: "/", RawQuery: "location"}, []byte(`"https://s3.amazonaws.com/?location"`), false},
		{URL{Scheme: "http", Host: "myobstor:10000", Path: "/mybucket/myobject"}, []byte(`"http://myobstor:10000/mybucket/myobject"`), false},
		{URL{Scheme: "ftp", Host: "myftp.server:10000", Path: "/myuser"}, []byte(`"ftp://myftp.server:10000/myuser"`), false},
	}

	for i, testCase := range testCases {
		data, err := testCase.url.MarshalJSON()
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("test %v: error: expected: %v, got: %v", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if !reflect.DeepEqual(data, testCase.expectedData) {
				t.Fatalf("test %v: data: expected: %v, got: %v", i+1, string(testCase.expectedData), string(data))
			}
		}
	}
}

func TestURLUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		data        []byte
		expectedURL *URL
		expectErr   bool
	}{
		{[]byte(`""`), &URL{}, false},
		{[]byte(`"http://demo"`), &URL{Scheme: "http", Host: "demo"}, false},
		{[]byte(`"https://demo.obstor.net:0"`), &URL{Scheme: "https", Host: "demo.obstor.net:0"}, false},
		{[]byte(`"https://23.148.200.3:9000/"`), &URL{Scheme: "https", Host: "23.148.200.3:9000", Path: "/"}, false},
		{[]byte(`"https://s3.amazonaws.com/?location"`), &URL{Scheme: "https", Host: "s3.amazonaws.com", Path: "/", RawQuery: "location"}, false},
		{[]byte(`"http://myobstor:10000/mybucket/myobject//"`), &URL{Scheme: "http", Host: "myobstor:10000", Path: "/mybucket/myobject/"}, false},
		{[]byte(`"ftp://myftp.server:10000/myuser"`), &URL{Scheme: "ftp", Host: "myftp.server:10000", Path: "/myuser"}, false},
		{[]byte(`"http://webhook.server:10000/mywebhook/"`), &URL{Scheme: "http", Host: "webhook.server:10000", Path: "/mywebhook/"}, false},
		{[]byte(`"myserver:1000"`), nil, true},
		{[]byte(`"http://:1000/mybucket"`), nil, true},
		{[]byte(`"https://23.148.200.3:90000/"`), nil, true},
		{[]byte(`"http:/demo"`), nil, true},
	}

	for i, testCase := range testCases {
		var url URL
		err := url.UnmarshalJSON(testCase.data)
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("test %v: error: expected: %v, got: %v", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if !reflect.DeepEqual(&url, testCase.expectedURL) {
				t.Fatalf("test %v: host: expected: %#v, got: %#v", i+1, testCase.expectedURL, url)
			}
		}
	}
}

func TestParseHTTPURL(t *testing.T) {
	testCases := []struct {
		s           string
		expectedURL *URL
		expectErr   bool
	}{
		{"http://demo", &URL{Scheme: "http", Host: "demo"}, false},
		{"https://demo.obstor.net:0", &URL{Scheme: "https", Host: "demo.obstor.net:0"}, false},
		{"https://23.148.200.3:9000/", &URL{Scheme: "https", Host: "23.148.200.3:9000", Path: "/"}, false},
		{"https://s3.amazonaws.com/?location", &URL{Scheme: "https", Host: "s3.amazonaws.com", Path: "/", RawQuery: "location"}, false},
		{"http://myobstor:10000/mybucket//myobject/", &URL{Scheme: "http", Host: "myobstor:10000", Path: "/mybucket/myobject/"}, false},
		{"ftp://myftp.server:10000/myuser", nil, true},
		{"https://my.server:10000000/myuser", nil, true},
		{"myserver:1000", nil, true},
		{"http://:1000/mybucket", nil, true},
		{"https://23.148.200.3:90000/", nil, true},
		{"http:/demo", nil, true},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.s, func(t *testing.T) {
			url, err := ParseHTTPURL(testCase.s)
			expectErr := (err != nil)
			if expectErr != testCase.expectErr {
				t.Fatalf("error: expected: %v, got: %v", testCase.expectErr, expectErr)
			}
			if !testCase.expectErr {
				if !reflect.DeepEqual(url, testCase.expectedURL) {
					t.Fatalf("host: expected: %#v, got: %#v", testCase.expectedURL, url)
				}
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	testCases := []struct {
		s           string
		expectedURL *URL
		expectErr   bool
	}{
		{"http://demo", &URL{Scheme: "http", Host: "demo"}, false},
		{"https://demo.obstor.net:0", &URL{Scheme: "https", Host: "demo.obstor.net:0"}, false},
		{"https://23.148.200.3:9000/", &URL{Scheme: "https", Host: "23.148.200.3:9000", Path: "/"}, false},
		{"https://s3.amazonaws.com/?location", &URL{Scheme: "https", Host: "s3.amazonaws.com", Path: "/", RawQuery: "location"}, false},
		{"http://myobstor:10000/mybucket//myobject/", &URL{Scheme: "http", Host: "myobstor:10000", Path: "/mybucket/myobject/"}, false},
		{"ftp://myftp.server:10000/myuser", &URL{Scheme: "ftp", Host: "myftp.server:10000", Path: "/myuser"}, false},
		{"myserver:1000", nil, true},
		{"http://:1000/mybucket", nil, true},
		{"https://23.148.200.3:90000/", nil, true},
		{"http:/demo", nil, true},
	}

	for i, testCase := range testCases {
		url, err := ParseURL(testCase.s)
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("test %v: error: expected: %v, got: %v", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if !reflect.DeepEqual(url, testCase.expectedURL) {
				t.Fatalf("test %v: host: expected: %#v, got: %#v", i+1, testCase.expectedURL, url)
			}
		}
	}
}
