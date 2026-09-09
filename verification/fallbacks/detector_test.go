package fallbacks

import (
	"testing"
)

func TestDetectFallbacks(t *testing.T) {
	// 1. Python swallowed error with pass
	pyCode := `
def fetch_token():
    try:
        req = make_request()
        return req.token
    except Exception:
        pass
`
	v := DetectFallbacks("src/auth.py", pyCode, ".py")
	if len(v) == 0 {
		t.Errorf("Expected violation for bare 'pass' in except block, got 0")
	}

	// 2. Python clean error handling with logging
	pyClean := `
def fetch_token():
    try:
        req = make_request()
        return req.token
    except Exception as e:
        logger.error("Failed: %s", e)
        raise
`
	vClean := DetectFallbacks("src/auth.py", pyClean, ".py")
	if len(vClean) != 0 {
		t.Errorf("Expected 0 violations for clean error handling, got %d", len(vClean))
	}

	// 3. TypeScript empty catch block
	tsCode := `
async function loadData() {
    try {
        await api.get();
    } catch (e) {}
}
`
	vTS := DetectFallbacks("src/api.ts", tsCode, ".ts")
	if len(vTS) == 0 {
		t.Errorf("Expected violation for empty catch block in TypeScript, got 0")
	}
}
