package native

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/two-barrels/ari/v6"
)

// File opens a binary recording response without buffering the file in memory.
func (sr *StoredRecording) File(ctx context.Context, key *ari.Key) (*ari.RecordingFile, error) {
	if ctx == nil {
		return nil, errors.New("context not supplied")
	}
	if key == nil || key.ID == "" {
		return nil, errors.New("storedRecording key not supplied")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		sr.client.Options.URL+"/recordings/stored/"+url.PathEscape(key.ID)+"/file", nil)
	if err != nil {
		return nil, err
	}
	if sr.client.Options.Username != "" {
		req.SetBasicAuth(sr.client.Options.Username, sr.client.Options.Password)
	}
	// The regular command timeout covers the entire response body. Streaming
	// needs the caller's context to control that lifetime instead.
	httpClient := *sr.client.httpClient
	httpClient.Timeout = 0
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if err := maybeRequestError(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}
	return &ari.RecordingFile{Body: resp.Body, ContentType: resp.Header.Get("Content-Type"), Size: resp.ContentLength}, nil
}

// StoredRecording provides the ARI StoredRecording accessors for the native client
type StoredRecording struct {
	client *Client
}

// List lists the current stored recordings and returns a list of handles
func (sr *StoredRecording) List(filter *ari.Key) (sx []*ari.Key, err error) {
	var recs []struct {
		Name string `json:"name"`
	}

	if filter == nil {
		filter = sr.client.stamp(ari.NewKey(ari.StoredRecordingKey, ""))
	}

	err = sr.client.get("/recordings/stored", &recs)

	for _, rec := range recs {
		k := sr.client.stamp(ari.NewKey(ari.StoredRecordingKey, rec.Name))
		if filter.Match(k) {
			sx = append(sx, k)
		}
	}

	return
}

// Get gets a lazy handle for the given stored recording name
func (sr *StoredRecording) Get(key *ari.Key) *ari.StoredRecordingHandle {
	return ari.NewStoredRecordingHandle(key, sr, nil)
}

// Data retrieves the state of the stored recording
func (sr *StoredRecording) Data(key *ari.Key) (*ari.StoredRecordingData, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("storedRecording key not supplied")
	}

	data := new(ari.StoredRecordingData)
	if err := sr.client.get("/recordings/stored/"+key.ID, data); err != nil {
		return nil, dataGetError(err, "storedRecording", "%v", key.ID)
	}

	data.Key = sr.client.stamp(key)

	return data, nil
}

// Copy copies a stored recording and returns the new handle
func (sr *StoredRecording) Copy(key *ari.Key, dest string) (*ari.StoredRecordingHandle, error) {
	h, err := sr.StageCopy(key, dest)
	if err != nil {
		// NOTE: return the handle even on failure so that it can be used to
		//   delete the existing stored recording, should the Copy fail.
		//   ARI provides no facility to force-copy a recording.
		return h, err
	}

	return h, h.Exec()
}

// StageCopy creates a `StoredRecordingHandle` with a `Copy` operation staged.
func (sr *StoredRecording) StageCopy(key *ari.Key, dest string) (*ari.StoredRecordingHandle, error) {
	if key == nil || key.ID == "" {
		return nil, errors.New("storedRecording key not supplied")
	}
	var resp struct {
		Name string `json:"name"`
	}

	destKey := sr.client.stamp(ari.NewKey(ari.StoredRecordingKey, dest))

	return ari.NewStoredRecordingHandle(destKey, sr, func(h *ari.StoredRecordingHandle) error {
		path := "/recordings/stored/" + url.PathEscape(key.ID) + "/copy"
		return sr.client.post(path, &resp, &struct {
			DestinationRecordingName string `json:"destinationRecordingName"`
		}{DestinationRecordingName: dest})
	}), nil
}

// Delete deletes the stored recording
func (sr *StoredRecording) Delete(key *ari.Key) error {
	return sr.client.del("/recordings/stored/"+key.ID, nil, "")
}
