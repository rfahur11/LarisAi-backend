package repositories

import "context"

func getStoreID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value("store_id"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
