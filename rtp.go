// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package ari

// RTPStats is the response from GET /channels/{channelId}/rtp_statistics.
type RTPStats struct {
	TxCount             int64   `json:"txcount"`
	RxCount             int64   `json:"rxcount"`
	TxJitter            float64 `json:"txjitter"`
	RxJitter            float64 `json:"rxjitter"`
	RemoteMaxJitter     float64 `json:"remote_maxjitter"`
	RemoteMinJitter     float64 `json:"remote_minjitter"`
	RemoteNormDevJitter float64 `json:"remote_normdevjitter"`
	RemoteStdDevJitter  float64 `json:"remote_stdevjitter"`
	LocalMaxJitter      float64 `json:"local_maxjitter"`
	LocalMinJitter      float64 `json:"local_minjitter"`
	LocalNormDevJitter  float64 `json:"local_normdevjitter"`
	LocalStdDevJitter   float64 `json:"local_stdevjitter"`
	TxPacketLoss        int64   `json:"txploss"`
	RxPacketLoss        int64   `json:"rxploss"`
	RemoteMaxRxLoss     float64 `json:"remote_maxrxploss"`
	RemoteMinRxLoss     float64 `json:"remote_minrxploss"`
	RemoteNormDevRxLoss float64 `json:"remote_normdevrxploss"`
	RemoteStdDevRxLoss  float64 `json:"remote_stdevrxploss"`
	LocalMaxRxLoss      float64 `json:"local_maxrxploss"`
	LocalMinRxLoss      float64 `json:"local_minrxploss"`
	LocalNormDevRxLoss  float64 `json:"local_normdevrxploss"`
	LocalStdDevRxLoss   float64 `json:"local_stdevrxploss"`
	RTT                 float64 `json:"rtt"`
	MaxRTT              float64 `json:"maxrtt"`
	MinRTT              float64 `json:"minrtt"`
	NormDevRTT          float64 `json:"normdevrtt"`
	StdDevRTT           float64 `json:"stdevrtt"`
	LocalSSRC           int64   `json:"local_ssrc"`
	RemoteSSRC          int64   `json:"remote_ssrc"`
	TxOctetCount        int64   `json:"txoctetcount"`
	RxOctetCount        int64   `json:"rxoctetcount"`
	ChannelUniqueID     string  `json:"channel_uniqueid"`
}
