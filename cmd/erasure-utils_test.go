/*
 * PGG Obstor, (C) 2021-2026 PGG, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package cmd

import (
	"strings"
	"testing"
)

func TestIsValidBlockHash(t *testing.T) {
	valid := strings.Repeat("a", 64)
	good := []string{
		valid,
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}
	for _, h := range good {
		if !isValidBlockHash(h) {
			t.Fatalf("expected %q to be a valid block hash", h)
		}
	}

	bad := []string{
		"",
		"abc",                       // too short
		valid + "a",                 // too long (65)
		strings.Repeat("A", 64),     // uppercase not allowed
		strings.Repeat("g", 64),     // non-hex
		"../../../../../etc/passwd", // traversal
		"blocks/../../etc/passwd00000000000000000000000000000000000000",
		strings.Repeat("a", 63) + "/", // separator
		strings.Repeat("a", 62) + "..",
	}
	for _, h := range bad {
		if isValidBlockHash(h) {
			t.Fatalf("expected %q to be rejected", h)
		}
	}

	// A real digest from hashBlockData must validate
	if h := hashBlockData([]byte("obstor block")); !isValidBlockHash(h) {
		t.Fatalf("hashBlockData output %q failed validation", h)
	}
}
