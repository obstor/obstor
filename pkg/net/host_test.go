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

func TestHostIsEmpty(t *testing.T) {
	testCases := []struct {
		host           Host
		expectedResult bool
	}{
		{Host{"", 0, false}, true},
		{Host{"", 0, true}, true},
		{Host{"demo", 9000, false}, false},
		{Host{"demo", 9000, true}, false},
	}

	for i, testCase := range testCases {
		result := testCase.host.IsEmpty()

		if result != testCase.expectedResult {
			t.Fatalf("test %v: result: expected: %v, got: %v", i+1, testCase.expectedResult, result)
		}
	}
}

func TestHostString(t *testing.T) {
	testCases := []struct {
		host        Host
		expectedStr string
	}{
		{Host{"", 0, false}, ""},
		{Host{"", 0, true}, ":0"},
		{Host{"demo", 9000, false}, "demo"},
		{Host{"demo", 9000, true}, "demo:9000"},
	}

	for i, testCase := range testCases {
		str := testCase.host.String()

		if str != testCase.expectedStr {
			t.Fatalf("test %v: string: expected: %v, got: %v", i+1, testCase.expectedStr, str)
		}
	}
}

func TestHostEqual(t *testing.T) {
	testCases := []struct {
		host           Host
		compHost       Host
		expectedResult bool
	}{
		{Host{"", 0, false}, Host{"", 0, true}, false},
		{Host{"demo", 9000, true}, Host{"demo", 9000, false}, false},
		{Host{"", 0, true}, Host{"", 0, true}, true},
		{Host{"demo", 9000, false}, Host{"demo", 9000, false}, true},
		{Host{"demo", 9000, true}, Host{"demo", 9000, true}, true},
	}

	for i, testCase := range testCases {
		result := testCase.host.Equal(testCase.compHost)

		if result != testCase.expectedResult {
			t.Fatalf("test %v: string: expected: %v, got: %v", i+1, testCase.expectedResult, result)
		}
	}
}

func TestHostMarshalJSON(t *testing.T) {
	testCases := []struct {
		host         Host
		expectedData []byte
		expectErr    bool
	}{
		{Host{}, []byte(`""`), false},
		{Host{"demo", 0, false}, []byte(`"demo"`), false},
		{Host{"demo", 0, true}, []byte(`"demo:0"`), false},
		{Host{"demo", 9000, true}, []byte(`"demo:9000"`), false},
		{Host{"demo.obstor.net", 0, false}, []byte(`"demo.obstor.net"`), false},
		{Host{"demo.obstor.net", 9000, true}, []byte(`"demo.obstor.net:9000"`), false},
		{Host{"23.148.200.3", 0, false}, []byte(`"23.148.200.3"`), false},
		{Host{"23.148.200.3", 9000, true}, []byte(`"23.148.200.3:9000"`), false},
		{Host{"demo12", 0, false}, []byte(`"demo12"`), false},
		{Host{"12demo", 0, false}, []byte(`"12demo"`), false},
		{Host{"demo--obstor.net", 0, false}, []byte(`"demo--obstor.net"`), false},
	}

	for i, testCase := range testCases {
		data, err := testCase.host.MarshalJSON()
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

func TestHostUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		data         []byte
		expectedHost *Host
		expectErr    bool
	}{
		{[]byte(`""`), &Host{}, false},
		{[]byte(`"demo"`), &Host{"demo", 0, false}, false},
		{[]byte(`"demo:0"`), &Host{"demo", 0, true}, false},
		{[]byte(`"demo:9000"`), &Host{"demo", 9000, true}, false},
		{[]byte(`"demo.obstor.net"`), &Host{"demo.obstor.net", 0, false}, false},
		{[]byte(`"demo.obstor.net:9000"`), &Host{"demo.obstor.net", 9000, true}, false},
		{[]byte(`"23.148.200.3"`), &Host{"23.148.200.3", 0, false}, false},
		{[]byte(`"23.148.200.3:9000"`), &Host{"23.148.200.3", 9000, true}, false},
		{[]byte(`"demo12"`), &Host{"demo12", 0, false}, false},
		{[]byte(`"12demo"`), &Host{"12demo", 0, false}, false},
		{[]byte(`"demo--obstor.net"`), &Host{"demo--obstor.net", 0, false}, false},
		{[]byte(`":9000"`), &Host{"", 9000, true}, false},
		{[]byte(`"[fe80::8097:76eb:b397:e067%wlp2s0]"`), &Host{"fe80::8097:76eb:b397:e067%wlp2s0", 0, false}, false},
		{[]byte(`"[fe80::8097:76eb:b397:e067]:9000"`), &Host{"fe80::8097:76eb:b397:e067", 9000, true}, false},
		{[]byte(`"fe80::8097:76eb:b397:e067%wlp2s0"`), nil, true},
		{[]byte(`"fe80::8097:76eb:b397:e067%wlp2s0]"`), nil, true},
		{[]byte(`"[fe80::8097:76eb:b397:e067%wlp2s0"`), nil, true},
		{[]byte(`"[[fe80::8097:76eb:b397:e067%wlp2s0]]"`), nil, true},
		{[]byte(`"[[fe80::8097:76eb:b397:e067%wlp2s0"`), nil, true},
		{[]byte(`"demo:"`), nil, true},
		{[]byte(`"demo::"`), nil, true},
		{[]byte(`"demo:90000"`), nil, true},
		{[]byte(`"demo:-10"`), nil, true},
		{[]byte(`"demo-"`), nil, true},
		{[]byte(`":"`), nil, true},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run("", func(t *testing.T) {
			var host Host
			err := host.UnmarshalJSON(testCase.data)
			expectErr := (err != nil)

			if expectErr != testCase.expectErr {
				t.Errorf("error: expected: %v, got: %v", testCase.expectErr, expectErr)
			}

			if !testCase.expectErr {
				if !reflect.DeepEqual(&host, testCase.expectedHost) {
					t.Errorf("host: expected: %#v, got: %#v", testCase.expectedHost, host)
				}
			}
		})
	}
}

func TestParseHost(t *testing.T) {
	testCases := []struct {
		s            string
		expectedHost *Host
		expectErr    bool
	}{
		{"demo", &Host{"demo", 0, false}, false},
		{"demo:0", &Host{"demo", 0, true}, false},
		{"demo:9000", &Host{"demo", 9000, true}, false},
		{"demo.obstor.net", &Host{"demo.obstor.net", 0, false}, false},
		{"demo.obstor.net:9000", &Host{"demo.obstor.net", 9000, true}, false},
		{"23.148.200.3", &Host{"23.148.200.3", 0, false}, false},
		{"23.148.200.3:9000", &Host{"23.148.200.3", 9000, true}, false},
		{"demo12", &Host{"demo12", 0, false}, false},
		{"12demo", &Host{"12demo", 0, false}, false},
		{"demo--obstor.net", &Host{"demo--obstor.net", 0, false}, false},
		{":9000", &Host{"", 9000, true}, false},
		{"demo:", nil, true},
		{"demo::", nil, true},
		{"demo:90000", nil, true},
		{"demo:-10", nil, true},
		{"demo-", nil, true},
		{":", nil, true},
		{"", nil, true},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run("", func(t *testing.T) {
			host, err := ParseHost(testCase.s)
			expectErr := (err != nil)

			if expectErr != testCase.expectErr {
				t.Errorf("error: expected: %v, got: %v", testCase.expectErr, expectErr)
			}

			if !testCase.expectErr {
				if !reflect.DeepEqual(host, testCase.expectedHost) {
					t.Errorf("host: expected: %#v, got: %#v", testCase.expectedHost, host)
				}
			}
		})
	}
}

func TestTrimIPv6(t *testing.T) {
	testCases := []struct {
		IP         string
		expectedIP string
		expectErr  bool
	}{
		{"[fe80::8097:76eb:b397:e067%wlp2s0]", "fe80::8097:76eb:b397:e067%wlp2s0", false},
		{"fe80::8097:76eb:b397:e067%wlp2s0]", "fe80::8097:76eb:b397:e067%wlp2s0", true},
		{"[fe80::8097:76eb:b397:e067%wlp2s0]]", "fe80::8097:76eb:b397:e067%wlp2s0]", false},
		{"[[fe80::8097:76eb:b397:e067%wlp2s0]]", "[fe80::8097:76eb:b397:e067%wlp2s0]", false},
	}

	for i, testCase := range testCases {
		ip, err := trimIPv6(testCase.IP)
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("test %v: error: expected: %v, got: %v", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if ip != testCase.expectedIP {
				t.Fatalf("test %v: IP: expected: %#v, got: %#v", i+1, testCase.expectedIP, ip)
			}
		}
	}
}
