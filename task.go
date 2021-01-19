package gojob

import "context"

type TaskID = int32
type Task func(ctx context.Context, id TaskID) error
