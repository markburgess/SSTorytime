package sst

import "context"

// DeleteChapter removes a chapter via the DeleteChapter PL/pgSQL function.
func DeleteChapter(ctx context.Context, sst PoSST, chapter string) error {
	if sst.Q == nil {
		return ErrNoQuerier
	}
	return sst.Q.CallDeleteChapter(ctx, chapter)
}
