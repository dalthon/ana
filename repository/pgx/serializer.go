package pgx

import (
	"bytes"
	"encoding/gob"
	"errors"

	im "github.com/dalthon/idempotency_manager"
	pgx "github.com/jackc/pgx/v5"
)

func serialize[S any](value *S) []byte {
	if value == nil {
		return []byte{}
	}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	if encoder.Encode(value) != nil {
		panic("Could not encode data.")
	}

	return buffer.Bytes()
}

func deserialize[S any](encoded []byte) *S {
	if len(encoded) == 0 {
		return nil
	}

	decoder := gob.NewDecoder(bytes.NewBuffer(encoded))
	var decoded S

	if decoder.Decode(&decoded) != nil {
		panic("Could not decode data.")
	}

	return &decoded
}

func rowsToTrackedOperation[P any, R any](rows pgx.Rows) *im.TrackedOperation[P, R] {
	var operation im.TrackedOperation[P, R]
	var status string
	var errorMessage string
	var encodedPayload []byte
	var encodedResult []byte

	rows.Next()
	rows.Scan(
		&status,
		&operation.Key,
		&operation.Target,
		&encodedPayload,
		&operation.ReferenceTime,
		&operation.StartedAt,
		&operation.Timeout,
		&operation.Expiration,
		&encodedResult,
		&errorMessage,
	)
	rows.Close()

	switch status {
	case "ready":
		operation.Status = im.Ready
	case "running":
		operation.Status = im.Running
	case "finished":
		operation.Status = im.Finished
	case "failed":
		operation.Status = im.Failed
	}

	if errorMessage != "" {
		operation.Err = errors.New(errorMessage)
	}

	operation.Payload = deserialize[P](encodedPayload)
	operation.Result = deserialize[R](encodedResult)

	return &operation
}
