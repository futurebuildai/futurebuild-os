# Step 75 Deployment Checklist

## Overview

This checklist ensures the Interrogator Agent (Step 75) is production-ready before deployment.

**Feature:** Smart onboarding wizard with conversational extraction and blueprint vision analysis

**Spec:** `/specs/committed/STEP_75_INTERROGATOR_AGENT.md`

**API Endpoint:** `POST /api/v1/agent/onboard`

**Risk Level:** HIGH (Public-facing endpoint handling user uploads and AI operations)

---

## Pre-Deployment

### Code Quality

- [ ] All unit tests pass
  ```bash
  go test -v ./internal/service -run TestInterrogator
  ```

- [ ] All integration tests pass
  ```bash
  go test -v -tags=integration ./test/integration -run TestOnboard
  ```

- [ ] Audit passes (lint + type check)
  ```bash
  make audit
  ```

- [ ] No compiler warnings or errors
  ```bash
  go build ./cmd/api
  ```

- [ ] Code review completed (if applicable)

### Security Review

- [ ] SSRF protection verified
  - [ ] File scheme blocked (`file:///etc/passwd` returns error)
  - [ ] Private IPs blocked (`http://127.0.0.1`, `http://192.168.1.1`)
  - [ ] AWS metadata blocked (`http://169.254.169.254`)
  - [ ] Redirects disabled (no redirect-based SSRF)

- [ ] Input validation tested
  - [ ] Request body size limit enforced (1MB max)
  - [ ] Session ID length limit enforced (100 chars)
  - [ ] Message length limit enforced (10k chars)
  - [ ] Document URL length limit enforced (2k chars)
  - [ ] Current state field limit enforced (50 fields)

- [ ] File upload restrictions tested
  - [ ] Max file size enforced (50MB)
  - [ ] MIME type validation works (only images/PDFs allowed)
  - [ ] Executable files rejected

- [ ] No secrets in logs
  - [ ] API keys not logged
  - [ ] User PII logged only where necessary for compliance

### Performance Testing

- [ ] Load test completed
  ```bash
  # Test with 10 concurrent users
  go test -v -bench=. ./test/integration -run BenchmarkOnboardEndpoint
  ```

- [ ] Latency targets met
  - [ ] p95 < 500ms for text-only extraction
  - [ ] p95 < 5s for blueprint extraction

- [ ] Memory usage acceptable
  - [ ] No memory leaks detected
  - [ ] Max memory usage < 100MB per request

### Documentation

- [ ] API documentation updated
  - [ ] `/docs/API_ONBOARDING.md` reviewed and accurate

- [ ] Frontend integration documented
  - [ ] Onboarding view implementation notes reviewed

- [ ] Monitoring plan reviewed
  - [ ] Know which logs to check for errors
  - [ ] Know which metrics to monitor

---

## Deployment

### Staging Environment

- [ ] Deploy to staging
  ```bash
  # Follow your standard deployment process
  git checkout build
  git pull origin build
  # Deploy to staging
  ```

- [ ] Smoke test: Create project via onboarding wizard
  1. Navigate to onboarding view in frontend
  2. Enter message: "3200 sqft home in Austin"
  3. Verify extraction works (address, GSF)
  4. Provide name: "Test Project"
  5. Verify ready_to_create = true
  6. Click "Create Project"
  7. Verify project created successfully

- [ ] Verify logs show structured output
  ```bash
  # Check staging logs for structured slog output
  tail -f /var/log/futurebuild/api.log | grep onboarding
  ```

- [ ] Test blueprint upload
  1. Upload sample PDF blueprint
  2. Verify vision extraction works
  3. Check confidence scores are present
  4. Verify no errors in logs

- [ ] Verify SSRF protection in staging
  ```bash
  # Attempt SSRF attack (should fail)
  curl -X POST https://staging.futurebuild.com/api/v1/agent/onboard \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "session_id": "test",
      "document_url": "file:///etc/passwd",
      "current_state": {}
    }'
  # Expected: 400 Bad Request with "unsupported URL scheme" message
  ```

- [ ] Check Gemini API quota usage
  - [ ] Verify API calls are being made to Vertex AI
  - [ ] Check quota consumption is reasonable
  - [ ] Ensure no excessive retries

- [ ] Staging soak test (optional)
  - [ ] Run for 1 hour with simulated traffic
  - [ ] Monitor error rates, latency, memory

### Production Deployment

**Deployment Window:** Low-traffic period recommended (late evening/early morning)

- [ ] Deploy to production
  ```bash
  git checkout build
  git pull origin build
  # Follow production deployment process
  ```

- [ ] Verify health check passes
  ```bash
  curl https://api.futurebuild.com/health
  # Expected: 200 OK
  ```

- [ ] Smoke test in production
  1. Create test project via onboarding wizard
  2. Verify extraction works
  3. Delete test project

- [ ] Monitor error rates for 1 hour
  - [ ] Check error rate < 1%
  - [ ] No 500 errors in logs
  - [ ] No panics or crashes

- [ ] Monitor latency
  - [ ] p95 latency acceptable
  - [ ] No timeout errors

- [ ] Check AI quota consumption
  - [ ] Gemini API usage within expected range
  - [ ] No quota exceeded errors

---

## Rollback Plan

If error rate > 5% or critical issues detected:

### Immediate Actions

1. **Revert to previous version**
   ```bash
   git checkout <previous-commit-hash>
   # Redeploy
   ```

2. **Verify rollback successful**
   - [ ] Health check passes
   - [ ] Previous version working correctly
   - [ ] Error rate returns to normal

3. **Disable onboarding feature (graceful degradation)**
   - [ ] Frontend shows manual form instead of wizard
   - [ ] No user-facing errors
   - [ ] Existing functionality unaffected

### Root Cause Analysis

- [ ] Capture logs from failed deployment
- [ ] Identify root cause
- [ ] Document issue
- [ ] Create fix plan
- [ ] Re-test before next deployment

---

## Monitoring

### Alerts to Set Up

- [ ] **Error Rate Alert**
  - Metric: `interrogator_errors_total`
  - Threshold: > 10 errors/minute
  - Action: Page on-call engineer

- [ ] **Latency Alert**
  - Metric: `interrogator_extraction_duration_seconds` (p95)
  - Threshold: > 10 seconds
  - Action: Investigate performance degradation

- [ ] **AI Quota Alert**
  - Metric: Vertex AI quota usage
  - Threshold: > 80% of daily quota
  - Action: Notify team, consider rate limiting

### Dashboards to Monitor

- [ ] **Request Volume**
  - Metric: `interrogator_requests_total`
  - Chart: Time series (last 24h)

- [ ] **Success Rate**
  - Metric: `(total_requests - errors) / total_requests * 100`
  - Chart: Time series (last 24h)
  - Target: > 99%

- [ ] **Extraction Confidence**
  - Metric: `interrogator_confidence_avg` per field
  - Chart: Heatmap by field
  - Monitor: Low confidence fields for prompt tuning

- [ ] **Latency Distribution**
  - Metric: `interrogator_extraction_duration_seconds`
  - Chart: Histogram (p50, p95, p99)
  - Target: p95 < 5s

### Log Queries

**Check onboarding activity:**
```bash
grep "onboarding_message_received" /var/log/futurebuild/api.log | tail -100
```

**Check errors:**
```bash
grep "onboarding_request_failed\|blueprint_download_failed\|ai_extraction_failed" /var/log/futurebuild/api.log
```

**Check extraction metrics:**
```bash
grep "onboarding_message_completed" /var/log/futurebuild/api.log | jq .
```

---

## Post-Deployment

### Day 1

- [ ] Monitor error rates (target: < 1%)
- [ ] Monitor latency (target: p95 < 5s)
- [ ] Check user feedback (if available)
- [ ] Review sample extractions for quality

### Week 1

- [ ] Analyze extraction accuracy
  - [ ] Check confidence scores by field
  - [ ] Identify fields with low confidence
  - [ ] Consider prompt tuning if needed

- [ ] Review AI costs
  - [ ] Gemini API usage vs. budget
  - [ ] Cost per onboarding session

- [ ] Gather user feedback
  - [ ] Are users completing onboarding?
  - [ ] Are extractions accurate?
  - [ ] Any usability issues?

### Month 1

- [ ] Performance review
  - [ ] Latency trends
  - [ ] Error rate trends
  - [ ] User adoption rate

- [ ] Feature iteration
  - [ ] Identify improvement opportunities
  - [ ] Plan next iteration (if needed)

---

## Success Criteria

✅ **Deployment is successful if:**

1. Error rate < 1% for 24 hours
2. p95 latency < 5 seconds
3. No security incidents (SSRF, file upload abuse)
4. Users can complete onboarding end-to-end
5. Extraction confidence scores > 0.8 on average
6. No production incidents or rollbacks required

❌ **Rollback if:**

1. Error rate > 5%
2. Security vulnerability detected
3. Production crashes or panics
4. AI quota exceeded causing failures
5. Data corruption or loss

---

## Contact Information

**On-Call Engineer:** [Insert contact]

**Escalation:** [Insert escalation path]

**Incident Channel:** #incidents (Slack/Teams)

**Monitoring Dashboard:** [Insert URL]

**Logs:** [Insert log access instructions]

---

## Appendix: Testing Commands

### Manual API Test

```bash
# Get auth token
TOKEN=$(curl -X POST https://api.futurebuild.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}' | jq -r .token)

# Test message parsing
curl -X POST https://api.futurebuild.com/api/v1/agent/onboard \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test_123",
    "message": "3200 sqft home in Austin with slab foundation",
    "current_state": {}
  }' | jq .

# Expected response:
# {
#   "session_id": "test_123",
#   "reply": "...",
#   "extracted_values": {
#     "gsf": 3200,
#     "address": "Austin",
#     "foundation_type": "slab"
#   },
#   "confidence_scores": { ... },
#   "ready_to_create": false,
#   "next_priority_field": "name"
# }
```

### Automated Test Suite

```bash
# Unit tests
go test -v ./internal/service -run TestInterrogator

# Integration tests
go test -v -tags=integration ./test/integration -run TestOnboard

# Full suite
make test
make test-integration

# Security tests
go test -v ./internal/service -run TestDownloadImage_Blocks
```

### Load Test

```bash
# Using Apache Bench (ab)
ab -n 100 -c 10 -p onboard_request.json -T application/json \
  -H "Authorization: Bearer $TOKEN" \
  https://api.futurebuild.com/api/v1/agent/onboard

# Using k6 (recommended for production)
k6 run loadtest/onboarding_test.js
```

---

## Version History

**v1.0 (Step 75):** Initial deployment checklist for Interrogator Agent

**Last Updated:** 2026-02-01
