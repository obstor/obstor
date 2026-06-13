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
	"bytes"
	"testing"

	iampolicy "github.com/obstor/obstor/pkg/iam/policy"
)

func TestSessionPolicyAllowsDenyOnly(t *testing.T) {
	// Session policy permits only s3:GetObject on bucket1.
	const doc = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::bucket1/*"]
			}
		]
	}`
	subPolicy, err := iampolicy.ParseConfig(bytes.NewReader([]byte(doc)))
	if err != nil {
		t.Fatalf("parse policy: %v", err)
	}

	// A request the session policy does NOT allow: PutObject on bucket2.
	args := iampolicy.Args{
		AccountName: "svc",
		Action:      iampolicy.PutObjectAction,
		BucketName:  "bucket2",
		ObjectName:  "secret",
		DenyOnly:    true, // self-service path; this is the bypass vector
	}

	// The raw evaluation short-circuits to true because of DenyOnly
	if !subPolicy.IsAllowed(args) {
		t.Fatal("precondition: raw DenyOnly evaluation expected to short-circuit true")
	}

	// The fixed evaluation must reject it: the session policy does not allow
	// PutObject on bucket2.
	if sessionPolicyAllows(subPolicy, args) {
		t.Fatal("sessionPolicyAllows must not honor an action outside the session policy even with DenyOnly set")
	}

	// And it must still allow what the session policy genuinely permits.
	okArgs := iampolicy.Args{
		AccountName: "svc",
		Action:      iampolicy.GetObjectAction,
		BucketName:  "bucket1",
		ObjectName:  "file",
		DenyOnly:    true,
	}
	if !sessionPolicyAllows(subPolicy, okArgs) {
		t.Fatal("sessionPolicyAllows must allow an action the session policy permits")
	}

	// IsOwner must likewise not short-circuit the session restriction.
	ownerArgs := args
	ownerArgs.DenyOnly = false
	ownerArgs.IsOwner = true
	if sessionPolicyAllows(subPolicy, ownerArgs) {
		t.Fatal("sessionPolicyAllows must not honor IsOwner to bypass the session policy")
	}
}
