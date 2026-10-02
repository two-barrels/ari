// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

// Package testfixtures holds version expectations shared by the native client,
// proxy, and specification checker tests. These are derived from the pinned
// Asterisk 23 ARI Swagger metadata; live-server behavior must be tested later.
package testfixtures

type VersionCase struct {
	Method    string
	Path      string
	Version   string
	Available bool
}

var Asterisk20_22_23 = []VersionCase{
	{"GET", "/bridges/{bridgeId}/variable", "20.19.0", false},
	{"GET", "/bridges/{bridgeId}/variable", "20.20.0", true},
	{"GET", "/bridges/{bridgeId}/variable", "22.9.0", false},
	{"GET", "/bridges/{bridgeId}/variable", "22.10.0", true},
	{"GET", "/bridges/{bridgeId}/variable", "23.3.0", false},
	{"GET", "/bridges/{bridgeId}/variable", "23.4.0", true},
	{"POST", "/channels/{channelId}/progress", "20.15.0", false},
	{"POST", "/channels/{channelId}/progress", "20.16.0", true},
	{"POST", "/channels/{channelId}/progress", "22.5.0", false},
	{"POST", "/channels/{channelId}/progress", "22.6.0", true},
	{"POST", "/events/claim", "20.16.0", false},
	{"POST", "/events/claim", "20.17.0", true},
	{"POST", "/events/claim", "22.6.0", false},
	{"POST", "/events/claim", "22.7.0", true},
	{"POST", "/events/claim", "23.0.0", false},
	{"POST", "/events/claim", "23.1.0", true},
	{"POST", "/endpoints/refer", "20.4.0", false},
	{"POST", "/endpoints/refer", "20.5.0", true},
	{"POST", "/endpoints/refer", "22.0.0", true},
	{"POST", "/endpoints/refer", "23.0.0", true},
}
